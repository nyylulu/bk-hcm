/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package task task
package task

import (
	"errors"
	"fmt"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetBizRecyclability get biz recyclability
func (s *service) GetBizRecyclability(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.getRecyclability(cts, bkBizIDMap, meta.Biz, meta.Recycle)
}

// GetRecyclability check whether hosts can be recycled or not
func (s *service) GetRecyclability(cts *rest.Contexts) (any, error) {
	return s.getRecyclability(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Recycle)
}

// getRecyclability check whether hosts can be recycled or not
func (s *service) getRecyclability(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.RecycleCheckReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to check resource recyclability, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to check resource recyclability, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().RecycleCheck(cts.Kit, input, bkBizIDMap, resType, action)
	if err != nil {
		logs.Errorf("failed to check resource recyclability, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// PreviewBizRecycleOrder preview biz recycle order
func (s *service) PreviewBizRecycleOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check preview biz recycle order permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.previewRecycleOrder(cts, bkBizIDMap)
}

// PreviewRecycleOrder preview recycle order
func (s *service) PreviewRecycleOrder(cts *rest.Contexts) (any, error) {
	return s.previewRecycleOrder(cts, make(map[int64]struct{}))
}

// previewRecycleOrder get preview recycle orders before commit
func (s *service) previewRecycleOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}) (any, error) {
	input := new(types.PreviewRecycleReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to preview recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to preview recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().PreviewRecycleOrder(cts.Kit, input, bkBizIDMap)
	if err != nil {
		logs.Errorf("failed to preview recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// AuditRecycleOrder audit recycle orders by resource administrator
func (s *service) AuditRecycleOrder(cts *rest.Contexts) (any, error) {
	input := new(types.AuditRecycleReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to audit recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if cts.Kit.User == "" {
		logs.Errorf("failed to audit recycle order, for invalid user is empty, rid: %s", cts.Kit.Rid)
		return nil, errf.New(common.CCErrCommParamsIsInvalid, "failed to recycle check, for invalid user is empty")
	}
	input.Operator = cts.Kit.User

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to preview recycle order, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Recycler().AuditRecycleOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to preview recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateBizRecycleOrder create biz recycle order
func (s *service) CreateBizRecycleOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.createRecycleOrder(cts, bkBizIDMap, meta.Biz, meta.Recycle)
}

// CreateRecycleOrder create recycle order
func (s *service) CreateRecycleOrder(cts *rest.Contexts) (any, error) {
	return s.createRecycleOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Recycle)
}

// createRecycleOrder create and start recycle orders
func (s *service) createRecycleOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.CreateRecycleReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to create recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().CreateRecycleOrder(cts.Kit, input, bkBizIDMap, resType, action)
	if err != nil {
		logs.Errorf("failed to create recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecycleBizOrder get recycle biz order
func (s *service) GetRecycleBizOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get recycle biz order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BizID = []int64{bkBizID}

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to get recycle biz order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	return s.getRecycleOrder(cts.Kit, input)
}

// GetRecycleOrder get recycle order
func (s *service) GetRecycleOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	return s.getRecycleOrder(cts.Kit, input)
}

// getRecycleOrder gets recycle order info
func (s *service) getRecycleOrder(kt *kit.Kit, input *types.GetRecycleOrderReq) (any, error) {
	// 主机回收-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bizID := range input.BizID {
		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bizID,
		})
	}
	err := s.authorizer.AuthorizeWithPerm(kt, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get recycle order, bizIDs: %v, err: %v, rid: %s", input.BizID, err, kt.Rid)
		return nil, err
	}

	rst, err := s.logics.Recycler().GetRecycleOrder(kt, input)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizRecycleOrder gets given business's recycle order info
func (s *service) GetBizRecycleOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetBizRecycleReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode get biz recycle order request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to validate get biz recycle order request, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// check permission
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Recycle}, BizID: input.BkBizID,
	})
	if err != nil {
		logs.Errorf("no permission to get biz recycle order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	param := &types.GetRecycleOrderReq{
		BizID: []int64{input.BkBizID},
		Start: input.Start,
		End:   input.End,
		Page:  input.Page,
	}

	rst, err := s.logics.Recycler().GetRecycleOrder(cts.Kit, param)
	if err != nil {
		logs.Errorf("failed to get biz recycle order, param: %+v, err: %v, rid: %s", param, err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizRecycleDetect get biz recycle detect
func (s *service) GetBizRecycleDetect(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz recycle detect permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetRecycleDetect(cts)
}

// GetRecycleDetect gets recycle detection task info
func (s *service) GetRecycleDetect(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleDetectReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get recycle detection task info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get recycle detection task info, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().GetRecycleDetect(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get recycle detection task info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return rst, nil
}

// ListBizDetectHost list biz detect host
func (s *service) ListBizDetectHost(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check list biz detect host permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.ListDetectHost(cts)
}

// ListDetectHost gets recycle detection host list
func (s *service) ListDetectHost(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleDetectReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to list recycle detection host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to list recycle detection host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().ListDetectHost(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to list recycle detection host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizRecycleDetectStep get biz recycle detect step
func (s *service) GetBizRecycleDetectStep(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check list biz detect host permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	return s.GetRecycleDetectStep(cts)
}

// GetRecycleDetectStep gets recycle detection step info
func (s *service) GetRecycleDetectStep(cts *rest.Contexts) (any, error) {
	input := new(types.GetDetectStepReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get recycle detection step info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get recycle detection step info, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Recycler().GetRecycleDetectStep(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get recycle detection step info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizRecycleOrderHost get biz recycle order host
func (s *service) GetBizRecycleOrderHost(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleHostReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get biz recycle host info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	input.BizID = []int64{bkBizID}

	err = input.Validate()
	if err != nil {
		logs.Errorf("failed to get biz recycle host info, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	return s.getRecycleOrderHost(cts.Kit, input)
}

// GetRecycleOrderHost get recycle order host
func (s *service) GetRecycleOrderHost(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleHostReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get recycle host info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get recycle host info, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	return s.getRecycleOrderHost(cts.Kit, input)
}

// getRecycleOrderHost gets recycle host info in certain order
func (s *service) getRecycleOrderHost(kt *kit.Kit, input *types.GetRecycleHostReq) (any, error) {
	// 主机回收-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bizID := range input.BizID {
		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bizID,
		})
	}
	err := s.authorizer.AuthorizeWithPerm(kt, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get recycle order host, bizIDs: %v, err: %v, rid: %s",
			input.BizID, err, kt.Rid)
		return nil, err
	}

	rst, err := s.logics.Recycler().GetRecycleHost(kt, input)
	if err != nil {
		logs.Errorf("failed to get recycle host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecycleRecordDeviceType gets recycle record device type list
func (s *service) GetRecycleRecordDeviceType(cts *rest.Contexts) (any, error) {
	rst, err := s.logics.Recycler().GetRecycleRecordDeviceType(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get recycle record device type list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecycleRecordRegion gets recycle record region list
func (s *service) GetRecycleRecordRegion(cts *rest.Contexts) (any, error) {
	rst, err := s.logics.Recycler().GetRecycleRecordRegion(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get recycle record region list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecycleRecordZone gets recycle record zone list
func (s *service) GetRecycleRecordZone(cts *rest.Contexts) (any, error) {
	rst, err := s.logics.Recycler().GetRecycleRecordZone(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get recycle record zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizHostToRecycle gets business hosts in recycle module
func (s *service) GetBizHostToRecycle(cts *rest.Contexts) (any, error) {
	input := new(types.GetRecycleBizHostReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get biz host to recycle, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get biz host to recycle, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// 主机回收-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Recycle}, BizID: input.BizID,
	})
	if err != nil {
		logs.Errorf("no permission to get biz host to recycle, bizID: %d, err: %v, rid: %s",
			input.BizID, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Recycler().GetRecycleBizHost(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get recycle host info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// StartBizRecycleOrder start biz recycle order
func (s *service) StartBizRecycleOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.startRecycleOrder(cts, bkBizIDMap, meta.Biz, meta.Recycle)
}

// StartRecycleOrder start recycle order
func (s *service) StartRecycleOrder(cts *rest.Contexts) (any, error) {
	return s.startRecycleOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Recycle)
}

// startRecycleOrder start recycle order
func (s *service) startRecycleOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.StartRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to start recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to start recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getOrderBizIds(cts.Kit, input.OrderID, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to start recycle order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to start recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to start recycle order, failed to check permission, bizID: %d, err: %v, rid: %s",
				bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Recycler().StartRecycleOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to start recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (s *service) getOrderBizIds(kit *kit.Kit, orderIds []uint64, suborderIds []string) ([]int64, error) {
	filter := map[string]interface{}{}
	if len(orderIds) > 0 {
		filter["order_id"] = mapstr.MapStr{
			common.BKDBIN: orderIds,
		}
	}

	if len(suborderIds) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			common.BKDBIN: suborderIds,
		}
	}

	bizIds := make([]int64, 0)
	page := metadata.BasePage{
		Start: 0,
		Limit: 500,
	}
	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return bizIds, err
	}

	for _, inst := range insts {
		bizIds = append(bizIds, inst.BizID)
	}

	bizIds = util.IntArrayUnique(bizIds)

	return bizIds, nil
}

// StartRecycleDetect starts recycle detection task
func (s *service) StartRecycleDetect(cts *rest.Contexts) (any, error) {
	input := new(types.StartDetectTaskReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to start recycle detection task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to start recycle detection task, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getOrderBizIds(cts.Kit, []uint64{}, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to start recycle detection task, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to start recycle detection task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Recycle}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to start recycle detection task, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Recycler().StartDetectTask(cts.Kit, input); err != nil {
		logs.Errorf("failed to start detection task, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ReviseBizRecycleOrder revise biz recycle order
func (s *service) ReviseBizRecycleOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.reviseRecycleOrder(cts, bkBizIDMap, meta.Biz, meta.Recycle)
}

// ReviseRecycleOrder revise recycle order
func (s *service) ReviseRecycleOrder(cts *rest.Contexts) (any, error) {
	return s.reviseRecycleOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Recycle)
}

// reviseRecycleOrder revise recycle orders to remove detection failed hosts
func (s *service) reviseRecycleOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.ReviseRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to revise recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to revise recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getOrderBizIds(cts.Kit, []uint64{}, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to revise recycle order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to revise recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to revise recycle order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Recycler().ReviseRecycleOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to revise recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// PauseRecycleOrder pauses recycle order
func (s *service) PauseRecycleOrder(_ *rest.Contexts) (any, error) {
	// TODO
	return nil, nil
}

// ResumeRecycleOrder resumes recycle order
func (s *service) ResumeRecycleOrder(cts *rest.Contexts) (any, error) {
	input := new(types.ResumeRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to resumes recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to resumes recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getOrderBizIds(cts.Kit, []uint64{}, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to resumes recycle order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, fmt.Errorf("get order biz id err: %v", err))
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to resumes recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Recycle}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to terminate recycle order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Recycler().ResumeRecycleOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to resume recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TerminateBizRecycleOrder terminate biz recycle order
func (s *service) TerminateBizRecycleOrder(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.terminateRecycleOrder(cts, bkBizIDMap, meta.Biz, meta.Recycle)
}

// TerminateRecycleOrder terminate recycle order
func (s *service) TerminateRecycleOrder(cts *rest.Contexts) (any, error) {
	return s.terminateRecycleOrder(cts, make(map[int64]struct{}), meta.ZiYanResource, meta.Recycle)
}

// terminateRecycleOrder terminates recycle order
func (s *service) terminateRecycleOrder(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.TerminateRecycleOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to terminate recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to terminate recycle order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getOrderBizIds(cts.Kit, []uint64{}, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to terminate recycle order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to terminate recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to terminate recycle order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Recycler().TerminateRecycleOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to terminate recycle order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetRecycleStageCfg get recycle stage config
func (s *service) GetRecycleStageCfg(_ *rest.Contexts) (any, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"stage":       table.RecycleStageCommit,
				"description": table.RecycleStageDescCommit,
			},
			{
				"stage":       table.RecycleStageDetect,
				"description": table.RecycleStageDescDetect,
			},
			{
				"stage":       table.RecycleStageAudit,
				"description": table.RecycleStageDescAudit,
			},
			{
				"stage":       table.RecycleStageTransit,
				"description": table.RecycleStageDescTransit,
			},
			{
				"stage":       table.RecycleStageReturn,
				"description": table.RecycleStageDescReturn,
			},
			{
				"stage":       table.RecycleStageDone,
				"description": table.RecycleStageDescDone,
			},
			{
				"stage":       table.RecycleStageTerminate,
				"description": table.RecycleStageDescTerminate,
			},
		},
	}

	return rst, nil
}

// GetRecycleStatusCfg get recycle status config
func (s *service) GetRecycleStatusCfg(_ *rest.Contexts) (any, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"status":      table.RecycleStatusUncommit,
				"description": table.RecycleStatusDescUncommit,
			},
			{
				"status":      table.RecycleStatusCommitted,
				"description": table.RecycleStatusDescCommitted,
			},
			{
				"status":      table.RecycleStatusDetecting,
				"description": table.RecycleStatusDescDetecting,
			},
			{
				"status":      table.RecycleStatusDetectFailed,
				"description": table.RecycleStatusDescDetectFailed,
			},
			{
				"status":      table.RecycleStatusAudit,
				"description": table.RecycleStatusDescAudit,
			},
			{
				"status":      table.RecycleStatusRejected,
				"description": table.RecycleStatusDescRejected,
			},
			{
				"status":      table.RecycleStatusTransiting,
				"description": table.RecycleStatusDescTransiting,
			},
			{
				"status":      table.RecycleStatusTransitFailed,
				"description": table.RecycleStatusDescTransitFailed,
			},
			{
				"status":      table.RecycleStatusReturning,
				"description": table.RecycleStatusDescReturning,
			},
			{
				"status":      table.RecycleStatusReturnFailed,
				"description": table.RecycleStatusDescReturnFailed,
			},
			{
				"status":      table.RecycleStatusDone,
				"description": table.RecycleStatusDescDone,
			},
			{
				"status":      table.RecycleStatusTerminate,
				"description": table.RecycleStatusDescTerminate,
			},
		},
	}

	return rst, nil
}

// GetDetectStatusCfg get recycle detection status config
func (s *service) GetDetectStatusCfg(_ *rest.Contexts) (any, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"status":      table.DetectStatusInit,
				"description": "未执行",
			},
			{
				"status":      table.DetectStatusRunning,
				"description": "执行中",
			},
			{
				"status":      table.DetectStatusSuccess,
				"description": "成功",
			},
			{
				"status":      table.DetectStatusFailed,
				"description": "失败",
			},
		},
	}

	return rst, nil
}

// GetDetectStepCfg gets recycle detection task step config info
func (s *service) GetDetectStepCfg(cts *rest.Contexts) (any, error) {
	rst, err := s.logics.Recycler().GetDetectStepCfg(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get recycle detection step config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
