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
	"net/http"

	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/plan"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
)

// InitService initial the plan service.
func InitService(c *capability.Capability) {
	s := &service{
		dao:            c.Dao,
		planController: c.PlanController,
		esbClient:      c.EsbClient,
		authorizer:     c.Authorizer,
		bizLogics:      c.BizLogic,
		client:         c.Client,
	}
	h := rest.NewHandler()

	s.initPlanService(h)

	h.Load(c.WebService)
}

type service struct {
	dao            dao.Set
	esbClient      esb.Client
	planController *plan.Controller
	authorizer     auth.Authorizer
	bizLogics      biz.Logics
	client         *client.ClientSet
}

func (s *service) initPlanService(h *rest.Handler) {
	// biz
	h.Add("GetBizOrgRel", http.MethodGet, "/bizs/{bk_biz_id}/org/relation", s.GetBizOrgRel)

	// meta
	// TODO: 这里的url跟meta包里的url边界划分不清晰
	h.Add("ListDemandClass", http.MethodGet, "/plan/demand_class/list", s.ListDemandClass)
	h.Add("ListResMode", http.MethodGet, "/plan/res_mode/list", s.ListResMode)
	h.Add("ListDemandSource", http.MethodGet, "/plan/demand_source/list", s.ListDemandSource)
	h.Add("ListResPlanTicketStatus", http.MethodGet, "/plan/res_plan_ticket_status/list", s.ListRPTicketStatus)
	h.Add("GetDemandAvailableTime", http.MethodPost, "/plans/demands/available_times/get", s.GetDemandAvailableTime)

	// ticket
	h.Add("ListResPlanTicket", http.MethodPost, "/plans/resources/tickets/list", s.ListResPlanTicket)
	h.Add("ListBizResPlanTicket", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/tickets/list",
		s.ListBizResPlanTicket)
	h.Add("CreateBizResPlanTicket", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/tickets/create",
		s.CreateBizResPlanTicket)
	h.Add("GetResPlanTicket", http.MethodGet, "/plans/resources/tickets/{id}", s.GetResPlanTicket)
	h.Add("GetBizResPlanTicket", http.MethodGet,
		"/bizs/{bk_biz_id}/plans/resources/tickets/{id}", s.GetBizResPlanTicket)
	h.Add("GetResPlanTicketAudit", http.MethodGet,
		"/plans/resources/tickets/{ticket_id}/audit", s.GetResPlanTicketAudit)
	h.Add("GetBizResPlanTicketAudit", http.MethodGet,
		"/bizs/{bk_biz_id}/plans/resources/tickets/{ticket_id}/audit", s.GetBizResPlanTicketAudit)
	h.Add("ApproveResPlanTicketITSMNode", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/approve_itsm_node", s.ApproveResPlanTicketITSMNode)
	h.Add("ApproveBizResPlanTicketITSMNode", http.MethodPost,
		"/bizs/{bk_biz_id}/plans/resources/tickets/{ticket_id}/approve_itsm_node", s.ApproveBizResPlanTicketITSMNode)

	// demand
	h.Add("ListResPlanDemand", http.MethodPost, "/plans/resources/demands/list", s.ListResPlanDemand)
	h.Add("ListBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/list",
		s.ListBizResPlanDemand)
	h.Add("GetPlanDemandDetail", http.MethodGet, "/plans/demands/{id}", s.GetPlanDemandDetail)
	h.Add("GetBizPlanDemandDetail", http.MethodGet, "/bizs/{bk_biz_id}/plans/demands/{id}", s.GetBizPlanDemandDetail)
	h.Add("ListBizPlanDemandChangeLog", http.MethodPost, "/bizs/{bk_biz_id}/plans/demands/change_logs/list",
		s.ListBizPlanDemandChangeLog)
	h.Add("ListPlanDemandChangelog", http.MethodPost, "/plans/demands/change_logs/list", s.ListPlanDemandChangeLog)
	h.Add("AdjustBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/adjust",
		s.AdjustBizResPlanDemand)
	h.Add("CancelBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/cancel",
		s.CancelBizResPlanDemand)

	// verify
	h.Add("VerifyResPlanDemandV2", http.MethodPost, "/plans/resources/demands/verify", s.VerifyResPlanDemandV2)
	h.Add("GetCvmChargeTypeDeviceTypeV2", http.MethodPost, "/config/findmany/config/cvm/charge_type/device_type",
		s.GetCvmChargeTypeDeviceTypeV2)

	// repair history data
	h.Add("RepairResPlanDemand", http.MethodPost, "/plans/resources/demands/repair", s.RepairResPlanDemand)
	// penalty
	h.Add("CalcPenaltyBase", http.MethodPost, "/plans/penalty/base/calc", s.CalcPenaltyBase)
	h.Add("CalcAndPushPenaltyRatio", http.MethodPost, "/plans/penalty/ratio/push", s.CalcAndPushPenaltyRatio)
	h.Add("PushExpireNotification", http.MethodPost, "/plans/demands/expire_notifications/push",
		s.PushExpireNotification)

	// demand week
	h.Add("ImportDemandWeek", http.MethodPost, "/plans/demand_week/import", s.ImportDemandWeek)

	// woa device type
	h.Add("ListDeviceType", http.MethodPost, "/plans/device_types/list", s.ListDeviceType)
	h.Add("CreateDeviceType", http.MethodPost, "/plans/device_types/batch/create", s.CreateDeviceType)
	h.Add("UpdateDeviceType", http.MethodPatch, "/plans/device_types/batch", s.UpdateDeviceType)
	h.Add("DeleteDeviceType", http.MethodDelete, "/plans/device_types/batch", s.DeleteDeviceType)
}
