/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package task scheduler
package task

import (
	"errors"

	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// ConfirmBizApplyModify confirm biz apply modify
func (s *service) ConfirmBizApplyModify(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	bkBizIDMap := make(map[int64]struct{})
	bkBizIDMap[bkBizID] = struct{}{}
	return s.confirmApplyModify(cts, bkBizIDMap, meta.Biz, meta.Create)
}

// confirmApplyModify 用户确认单据变更的回调接口
func (s *service) confirmApplyModify(cts *rest.Contexts, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
	action meta.Action) (any, error) {

	input := new(types.ConfirmApplyModifyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to confirm apply modify, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to confirm apply modify, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, []string{input.SuborderID})
	if err != nil {
		logs.Errorf("failed to confirm apply modify, for get order biz id err: %v, input: %+v, rid: %s",
			err, cvt.PtrToVal(input), cts.Kit.Rid)
		return nil, errf.Newf(pkg.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to confirm apply modify, err: %v, input: %+v, rid: %s",
			err, cvt.PtrToVal(input), cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错
		if _, ok := bkBizIDMap[bizId]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID is located is not in "+
				"the bizIDMap:%+v passed in", bizId, bkBizIDMap)
		}

		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to confirm apply modify, failed to check permission, bizID: %d, "+
				"err: %v, user: %s, rid: %s", bizId, err, cts.Kit.User, cts.Kit.Rid)
			return nil, err
		}
	}

	confirmResp, err := s.logics.Scheduler().ConfirmApplyModify(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to confirm apply modify, err: %v, input: %+v, rid: %s",
			err, cvt.PtrToVal(input), cts.Kit.Rid)
		return nil, err
	}

	// 记录回调成功的日志，方便排查问题
	logs.Infof("confirm apply modify success, subOrderID: %s, modifyID: %d, input: %+v, confirmResp: %+v, "+
		"user: %s, rid: %s", input.SuborderID, input.ModifyID, cvt.PtrToVal(input), cvt.PtrToVal(confirmResp),
		cts.Kit.User, cts.Kit.Rid)

	return confirmResp, nil
}
