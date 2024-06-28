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

// Package task scheduler
package task

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateApplyTicket create or update apply ticket
func (s *service) UpdateApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to save apply ticket, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().UpdateApplyTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyTicket get apply ticket
func (s *service) GetApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyTicketReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyAudit get apply ticket audit info
func (s *service) GetApplyAudit(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyAuditReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyAudit(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// AuditApplyTicket audit apply ticket
func (s *service) AuditApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyAuditReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().AuditTicket(cts.Kit, input); err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AutoAuditApplyTicket system automatic audit apply ticket
func (s *service) AutoAuditApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyAutoAuditReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().AutoAuditTicket(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to auto audit apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// ApproveApplyTicket approve or reject apply ticket
func (s *service) ApproveApplyTicket(cts *rest.Contexts) (any, error) {
	input := new(types.ApproveApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().ApproveTicket(cts.Kit, input); err != nil {
		logs.Errorf("failed to approve apply ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateApplyOrder creates apply order
func (s *service) CreateApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.ApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		logs.Errorf("no permission to create apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizId, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().CreateApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyOrder gets apply order info
func (s *service) GetApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bkBizID := range input.BkBizID {
		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bkBizID,
		})
	}
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get apply order, inputBizIDs: %v, err: %v, rid: %s",
			input.BkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetBizApplyOrder gets given business's apply order info
func (s *service) GetBizApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetBizApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get biz apply order, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// check permission
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizID,
	})
	if err != nil {
		logs.Errorf("no permission to get biz apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
			input.BkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	param := &types.GetApplyParam{
		BkBizID: []int64{input.BkBizID},
		Start:   input.Start,
		End:     input.End,
		Page:    input.Page,
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(cts.Kit, param)
	if err != nil {
		logs.Errorf("failed to get biz apply order, param: %+v, err: %v, rid: %s", param, err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyStatus gets apply order status
func (s *service) GetApplyStatus(cts *rest.Contexts) (any, error) {
	orderId, err := strconv.Atoi(cts.Request.PathParameter("order_id"))
	if err != nil {
		logs.Errorf("failed to get apply order status, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if orderId <= 0 {
		logs.Errorf("failed to get apply order status, for invalid order id %d <= 0, rid: %s", orderId, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "order_id")
	}

	input := &types.GetApplyParam{
		OrderID: []uint64{uint64(orderId)},
	}

	rst, err := s.logics.Scheduler().GetApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order status, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDetail gets apply order detail info
func (s *service) GetApplyDetail(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyDetailReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyGenerate gets apply order generate records
func (s *service) GetApplyGenerate(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyGenerateReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyGenerate(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply generate record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyInit gets apply order init records
func (s *service) GetApplyInit(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyInitReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply init record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply init record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyInit(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply init record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDiskCheck gets apply order disk check records
func (s *service) GetApplyDiskCheck(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyInitReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyDiskCheck(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply disk check record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDeliver gets apply order deliver records
func (s *service) GetApplyDeliver(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyDeliverReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyDeliver(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply deliver record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDevice get apply order delivered devices
func (s *service) GetApplyDevice(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// 解析参数里的业务ID，用于鉴权，是必传参数
	bkBizIDs, err := s.parseInputForBkBizID(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to parse input for bizID, err: %+v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	// 主机申领-业务粒度
	authAttrs := make([]meta.ResourceAttribute, 0)
	for _, bkBizID := range bkBizIDs {
		authAttrs = append(authAttrs, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Find}, BizID: bkBizID,
		})
	}
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, authAttrs...)
	if err != nil {
		logs.Errorf("no permission to get apply device, bizIDs: %v, err: %v, rid: %s", bkBizIDs, err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Scheduler().GetApplyDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

func (s *service) parseInputForBkBizID(kt *kit.Kit, input *types.GetApplyDeviceReq) ([]int64, error) {
	filterMap, err := input.GetFilter()
	if err != nil {
		logs.Errorf("failed to parse input filter, err: %v, input: %+v, rid: %s", err, input, kt.Rid)
		return nil, err
	}

	var bkBizIDs []int64
	paramMap, ok := filterMap["$and"].([]map[string]interface{})
	if !ok {
		return nil, errf.Newf(errf.InvalidParameter, "filter is illegal")
	}

	for _, paramItem := range paramMap {
		condMap, ok := paramItem["bk_biz_id"]
		if !ok {
			continue
		}
		// 如果找到了业务ID，但解析失败则break
		fieldMap, ok := condMap.(map[string]interface{})
		if !ok {
			break
		}
		numbers, ok := fieldMap["$in"].([]interface{})
		if !ok {
			logs.Errorf("bk_biz_id value is not []interface, fieldMap: %+v, rid: %s", fieldMap, kt.Rid)
			return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id is illegal")
		}

		for _, val := range numbers {
			number, ok := val.(json.Number)
			if !ok {
				logs.Errorf("bk_biz_id value is not json.Number, val: %+v, valType: %+v, rid: %s",
					val, reflect.TypeOf(val), kt.Rid)
				return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id value is not json.Number")
			}
			bkBizID, err := number.Int64()
			if err != nil {
				logs.Errorf("bk_biz_id value is not int64, number: %+v, valType: %+v, err: %v, rid: %s",
					number, reflect.TypeOf(number), err, kt.Rid)
				return nil, err
			}
			bkBizIDs = append(bkBizIDs, bkBizID)
		}
		break
	}

	if len(bkBizIDs) <= 0 {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id is required")
	}

	return bkBizIDs, nil
}

// GetDeliverDeviceByOrder get delivered devices by order id
func (s *service) GetDeliverDeviceByOrder(cts *rest.Contexts) (any, error) {
	input := new(types.GetDeliverDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply delivered device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules: []querybuilder.Rule{
			querybuilder.AtomRule{
				Field:    "order_id",
				Operator: querybuilder.OperatorEqual,
				Value:    input.OrderId,
			}},
	}
	if len(input.SuborderId) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "suborder_id",
			Operator: querybuilder.OperatorEqual,
			Value:    input.SuborderId,
		})
	}
	param := &types.GetApplyDeviceReq{
		Filter: &querybuilder.QueryFilter{
			Rule: rule,
		},
	}

	rst, err := s.logics.Scheduler().GetApplyDevice(cts.Kit, param)
	if err != nil {
		logs.Errorf("failed to get apply device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	type deviceBriefInfo struct {
		Ip      string `json:"ip" bson:"ip"`
		AssetId string `json:"asset_id" bson:"asset_id"`
	}
	type getDeviceBriefRst struct {
		Count int64              `json:"count"`
		Info  []*deviceBriefInfo `json:"info"`
	}

	briefRst := &getDeviceBriefRst{
		Count: rst.Count,
		Info:  make([]*deviceBriefInfo, 0),
	}
	for _, device := range rst.Info {
		briefRst.Info = append(briefRst.Info, &deviceBriefInfo{
			Ip:      device.Ip,
			AssetId: device.AssetId,
		})
	}

	return briefRst, nil
}

// ExportDeliverDevice export delivered devices
func (s *service) ExportDeliverDevice(cts *rest.Contexts) (any, error) {
	input := new(types.ExportDeliverDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// 主机申领-业务粒度
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: input.BkBizId,
	})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.Scheduler().ExportDeliverDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetMatchDevice get apply order match devices
func (s *service) GetMatchDevice(cts *rest.Contexts) (any, error) {
	input := new(types.GetMatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetMatchDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// MatchDevice execute apply order match devices
func (s *service) MatchDevice(cts *rest.Contexts) (any, error) {
	input := new(types.MatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to match devices, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().MatchDevice(cts.Kit, input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// MatchPoolDevice execute apply order match devices from resource pool
func (s *service) MatchPoolDevice(cts *rest.Contexts) (any, error) {
	input := new(types.MatchPoolDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to match pool devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to match pool devices, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	if err := s.logics.Scheduler().MatchPoolDevice(cts.Kit, input); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// PauseApplyOrder pauses apply order
func (s *service) PauseApplyOrder(_ *rest.Contexts) (any, error) {
	// TODO
	return nil, nil
}

// ResumeApplyOrder resumes apply order
func (s *service) ResumeApplyOrder(_ *rest.Contexts) (any, error) {
	// TODO
	return nil, nil
}

// StartApplyOrder start apply order
func (s *service) StartApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.StartApplyOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to start apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to start apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to start apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to start apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to start apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
				bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().StartApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to start recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TerminateApplyOrder terminate apply order
func (s *service) TerminateApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.TerminateApplyOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to terminate apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to terminate apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, input.SuborderID)
	if err != nil {
		logs.Errorf("failed to terminate apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to terminate apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to terminate apply order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().TerminateApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to terminate recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ModifyApplyOrder modify apply order
func (s *service) ModifyApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.ModifyApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to modify apply order, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	suborderIDs := []string{input.SuborderID}
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, suborderIDs)
	if err != nil {
		logs.Errorf("failed to modify apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to modify apply order, failed to check permission, bizID: %d, err: %v, rid: %s",
				bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	if err = s.logics.Scheduler().ModifyApplyOrder(cts.Kit, input); err != nil {
		logs.Errorf("failed to modify recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// RecommendApplyOrder get apply order modification recommendation
func (s *service) RecommendApplyOrder(cts *rest.Contexts) (any, error) {
	input := new(types.RecommendApplyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to recommend apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to recommend apply order, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	// get orders' biz id list
	suborderIDs := []string{input.SuborderID}
	bizIds, err := s.getApplyOrderBizIds(cts.Kit, suborderIDs)
	if err != nil {
		logs.Errorf("failed to recommend apply order, for get order biz id err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.Newf(common.CCErrCommParamsIsInvalid, "get order biz id err: %v", err)
	}

	if len(bizIds) == 0 {
		err = errors.New("biz id list is empty")
		logs.Errorf("failed to recommend apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// check permission
	for _, bizId := range bizIds {
		err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: meta.ZiYanResource, Action: meta.Create}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("no permission to terminate apply order, failed to check permission, bizID: %d, "+
				"err: %v, rid: %s", bizId, err, cts.Kit.Rid)
			return nil, err
		}
	}

	rst, err := s.logics.Scheduler().RecommendApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to recommend recycle order, input: %+v, err: %v, rid: %s", input, err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyModify get apply order modification records
func (s *service) GetApplyModify(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyModifyReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetApplyModify(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

func (s *service) getApplyOrderBizIds(kit *kit.Kit, suborderIds []string) ([]int64, error) {
	filter := map[string]interface{}{}

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
	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return bizIds, err
	}

	for _, inst := range insts {
		bizIds = append(bizIds, inst.BkBizId)
	}

	bizIds = util.IntArrayUnique(bizIds)

	return bizIds, nil
}
