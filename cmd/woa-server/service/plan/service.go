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
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/pkg/dal/dao"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the plan service.
func InitService(c *capability.Capability) {
	s := &service{
		dao:            c.Dao,
		planController: c.PlanController,
		esbClient:      c.EsbClient,
		authorizer:     c.Authorizer,
		logics:         biz.New(c.EsbClient, c.Authorizer),
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
	logics         biz.Logics
}

func (s *service) initPlanService(h *rest.Handler) {
	// biz
	h.Add("GetBizOrgRel", http.MethodGet, "/bizs/{bk_biz_id}/org/relation", s.GetBizOrgRel)

	// meta
	h.Add("ListDemandClass", http.MethodGet, "/plan/demand_class/list", s.ListDemandClass)
	h.Add("ListResMode", http.MethodGet, "/plan/res_mode/list", s.ListResMode)
	h.Add("ListDemandSource", http.MethodGet, "/plan/demand_source/list", s.ListDemandSource)
	h.Add("ListResPlanTicketStatus", http.MethodGet, "/plan/res_plan_ticket_status/list", s.ListRPTicketStatus)
	h.Add("GetDemandAvailableTime", http.MethodPost, "/plans/demands/available_times/get", s.GetDemandAvailableTime)

	// ticket
	h.Add("ListResPlanTicket", http.MethodPost, "/plans/resources/tickets/list", s.ListResPlanTicket)
	h.Add("ListBizResPlanTicket", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/tickets/list",
		s.ListBizResPlanTicket)
	h.Add("CreateBizResPlanTicket", http.MethodPost, "/plan/resource/ticket/create", s.CreateBizResPlanTicket)
	h.Add("GetResPlanTicket", http.MethodGet, "/plans/resources/tickets/{id}", s.GetResPlanTicket)
	h.Add("GetBizResPlanTicket", http.MethodGet, "/bizs/{bk_biz_id}/plans/resources/tickets/{id}",
		s.GetBizResPlanTicket)

	// demand
	h.Add("ListResPlanDemand", http.MethodPost, "/plans/resources/demands/list", s.ListResPlanDemand)
	h.Add("ListBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/list",
		s.ListBizResPlanDemand)
	h.Add("GetPlanDemandDetail", http.MethodGet, "/plans/demands/{id}", s.GetPlanDemandDetail)
	h.Add("GetBizPlanDemandDetail", http.MethodGet, "/bizs/{bk_biz_id}/plans/demands/{id}", s.GetBizPlanDemandDetail)
	h.Add("ListPlanDemandChangelog", http.MethodPost, "/plans/demands/change_logs/list", s.ListPlanDemandChangeLog)
	h.Add("AdjustBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/adjust",
		s.AdjustBizResPlanDemand)
	h.Add("CancelBizResPlanDemand", http.MethodPost, "/bizs/{bk_biz_id}/plans/resources/demands/cancel",
		s.CancelBizResPlanDemand)

	// verify
	h.Add("VerifyResPlanDemand", http.MethodPost, "/plans/resources/demands/verify", s.VerifyResPlanDemand)
	h.Add("GetCvmChargeTypeDeviceType", http.MethodPost, "/config/findmany/config/cvm/charge_type/device_type",
		s.GetCvmChargeTypeDeviceType)
}
