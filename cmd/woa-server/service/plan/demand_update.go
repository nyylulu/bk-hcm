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

package plan

import (
	"errors"
	"slices"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AdjustBizResPlanDemand adjust biz resource plan demand.
func (s *service) AdjustBizResPlanDemand(cts *rest.Contexts) (rst interface{}, err error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.AdjustRPDemandReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode adjust biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate adjust biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Update}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// 验证预测提报参数
	if err = s.validateAdjustResPlan(req, bkBizID); err != nil {
		return nil, err
	}

	// get biz org relation.
	bizOrgRel, err := s.bizLogics.GetBizOrgRel(cts.Kit, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ticketID, err := s.planController.AdjustBizResPlanDemand(cts.Kit, req, bkBizID, bizOrgRel)
	if err != nil {
		logs.Errorf("failed to adjust biz resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil
}

func (s *service) validateAdjustResPlan(req *ptypes.AdjustRPDemandReq, bkBizID int64) error {
	for _, item := range req.Adjusts {
		// 只允许931业务提报滚服项目
		if item.OriginalInfo != nil && item.OriginalInfo.ObsProject == enumor.ObsProjectRollServer &&
			bkBizID != enumor.ResourcePlanRollServerBiz {
			return errf.Newf(errf.InvalidParameter, "this business origin does not support rolling server project")
		}

		if item.UpdatedInfo != nil && item.UpdatedInfo.ObsProject == enumor.ObsProjectRollServer &&
			bkBizID != enumor.ResourcePlanRollServerBiz {
			return errf.Newf(errf.InvalidParameter, "this business updated does not support rolling server project")
		}
	}
	return nil
}

// areAllCrpDemandBelongToBiz return whether all input crp demand ids belong to input biz.
func (s *service) areAllCrpDemandBelongToBiz(kt *kit.Kit, crpDemandIDs []int64, bkBizID int64) (bool, error) {
	listOpt := &types.ListOption{
		Fields: []string{"bk_biz_id"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	notAllBelong := slices.ContainsFunc(rst.Details, func(ele rpcd.ResPlanCrpDemandTable) bool {
		return ele.BkBizID != bkBizID
	})

	return !notAllBelong, nil
}

// examineDemandClass examine whether all demands are the same demand class, and return the demand class.
func (s *service) examineDemandClass(kt *kit.Kit, crpDemandIDs []int64) (enumor.DemandClass, error) {
	if len(crpDemandIDs) == 0 {
		return "", errors.New("crp demand ids is empty")
	}

	listOpt := &types.ListOption{
		Fields: []string{"demand_class"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := s.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(rst.Details) == 0 {
		logs.Errorf("list resource plan demand, but len detail is 0, rid: %s", kt.Rid)
		return "", errors.New("list resource plan demand, but len detail is 0")
	}

	demandClass := rst.Details[0].DemandClass
	for _, detail := range rst.Details {
		if detail.DemandClass != demandClass {
			logs.Errorf("not all demand classes are the same, rid: %s", kt.Rid)
			return "", errors.New("not all demand classes are the same")
		}
	}

	return demandClass, nil
}

// CancelBizResPlanDemand cancel biz resource plan demand.
func (s *service) CancelBizResPlanDemand(cts *rest.Contexts) (rst interface{}, err error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.CancelRPDemandReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode cancel biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate cancel biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Delete}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// get biz org relation.
	bizOrgRel, err := s.bizLogics.GetBizOrgRel(cts.Kit, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ticketID, err := s.planController.CancelBizResPlanDemand(cts.Kit, req, bkBizID, bizOrgRel)
	if err != nil {
		logs.Errorf("failed to adjust biz resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil

}
