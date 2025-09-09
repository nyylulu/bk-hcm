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
	plantypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// GetTransferQuotaConfigs 获取预测转移额度配置
func (s *service) GetTransferQuotaConfigs(cts *rest.Contexts) (interface{}, error) {
	result, err := s.planController.GetPlanTransferQuotaConfigs(cts.Kit)
	if err != nil {
		logs.Errorf("get plan transfer quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateTransferQuotaConfigs 更新预测转移额度配置
func (s *service) UpdateTransferQuotaConfigs(cts *rest.Contexts) (interface{}, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Update}})
	if err != nil {
		logs.Errorf("update transfer quota configs auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(plantypes.UpdatePlanTransferQuotaConfigsReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("update transfer quota configs decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("update transfer quota configs validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err = s.planController.UpdatePlanTransferQuotaConfigs(cts.Kit, req); err != nil {
		logs.Errorf("update transfer quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// --- res plan transfer applied record ---

// ListResPlanTransferAppliedRecord list resource plan transfer applied record.
func (s *service) ListResPlanTransferAppliedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(plantypes.ListResPlanTransferAppliedRecordReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan transfer applied record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource transfer applied record parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanTransferAppliedRecord(cts.Kit, req)
}

// ListBizResPlanTransferAppliedRecord list biz res plan transfer applied record.
func (s *service) ListBizResPlanTransferAppliedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(plantypes.ListResPlanTransferAppliedRecordReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan transfer applied record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource transfer applied record parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	bkBizIDs, err := s.bizLogics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// if bizID not in authorized bkBizIDs, return empty response.
	if !slice.IsItemInSlice(bkBizIDs, bizID) {
		return nil, errf.New(errf.PermissionDenied, "no permission")
	}

	return s.listResPlanTransferAppliedRecord(cts.Kit, req)
}

// listResPlanTransferAppliedRecord general logic for list transfer applied record
func (s *service) listResPlanTransferAppliedRecord(kt *kit.Kit, req *plantypes.ListResPlanTransferAppliedRecordReq) (
	interface{}, error) {

	tarReq := &rpproto.TransferAppliedRecordListReq{ListReq: req.ListReq}
	resp, err := s.client.DataService().Global.ResourcePlan.ListResPlanTransferAppliedRecord(kt, tarReq)
	if err != nil {
		logs.Errorf("failed to list res plan transfer applied record, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return resp, nil
}

// ListResPlanTransferQuotaSummary 查询资源下转移额度使用概览信息.
func (s *service) ListResPlanTransferQuotaSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(plantypes.ListResPlanTransferQuotaSummaryReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan transfer applied quota, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource transfer applied quota parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanTransferQuotaSummary(cts.Kit, req)
}

// ListBizResPlanTransferQuotaSummary 查询业务下转移额度使用概览信息.
func (s *service) ListBizResPlanTransferQuotaSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(plantypes.ListResPlanTransferQuotaSummaryReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan transfer applied quota, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource transfer applied quota parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	bkBizIDs, err := s.bizLogics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	req.BkBizIDs = []int64{bizID}

	// if bizID not in authorized bkBizIDs, return empty response.
	if !slice.IsItemInSlice(bkBizIDs, bizID) {
		return nil, errf.New(errf.PermissionDenied, "no permission")
	}

	return s.listResPlanTransferQuotaSummary(cts.Kit, req)
}

// listResPlanTransferQuotaSummary general logic for list transfer quota summary
func (s *service) listResPlanTransferQuotaSummary(kt *kit.Kit, req *plantypes.ListResPlanTransferQuotaSummaryReq) (
	*plantypes.ResPlanTransferQuotaSummaryResp, error) {

	// 不带业务ID的查询已使用额度
	appliedRules := []*filter.AtomRule{tools.RuleEqual("year", req.Year)}
	if len(req.AppliedType) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("applied_type", req.AppliedType))
	}
	if len(req.SubTicketID) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("sub_ticket_id", req.SubTicketID))
	}
	if len(req.TechnicalClass) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("technical_class", req.TechnicalClass))
	}
	if len(req.ObsProject) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("obs_project", req.ObsProject))
	}

	// 带业务ID的查询已使用额度
	bizRules := make([]*filter.AtomRule, len(appliedRules))
	copy(bizRules, appliedRules)
	if len(req.BkBizIDs) > 0 {
		bizRules = append(bizRules, tools.RuleIn("bk_biz_id", req.BkBizIDs))
	}

	// 查询当前业务-已使用的额度
	bizTarReq := &rpproto.TransferAppliedRecordListReq{ListReq: core.ListReq{
		Filter: tools.ExpressionAnd(bizRules...),
		Page:   core.NewDefaultBasePage(),
	}}
	bizAppliedQuota, err := s.client.DataService().Global.ResourcePlan.SumResPlanTransferAppliedRecord(kt, bizTarReq)
	if err != nil {
		logs.Errorf("failed to list res plan transfer applied record quota, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// 查询CRP侧的预测额度
	demands, err := s.planController.QueryCrpDemandsQuota(kt, req.ObsProject, req.TechnicalClass)
	if err != nil {
		logs.Errorf("failed to query ieg demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var remainQuota int64
	for _, demand := range demands {
		remainQuota += demand.CoreAmount
	}

	return &plantypes.ResPlanTransferQuotaSummaryResp{
		UsedQuota:   bizAppliedQuota.SumAppliedCore,
		RemainQuota: remainQuota,
	}, nil
}
