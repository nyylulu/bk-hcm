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

package task

import (
	"net/http"

	taskLogics "hcm/cmd/woa-server/logics/task"
	"hcm/cmd/woa-server/service/capability"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/thirdparty/iamapi"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	logics := taskLogics.New(c.SchedulerIf, c.RecyclerIf, c.InformerIf, c.OperationIf)
	s := &service{
		logics: logics,
		Cc:     c.EsbClient.Cmdb(),
	}
	h := rest.NewHandler()
	h.Path("/task")

	s.initOperationService(h)
	s.initRecyclerService(h)
	s.initSchedulerService(h)

	h.Load(c.WebService)
}

type service struct {
	logics taskLogics.Logics
	Cc     cmdb.Client
	Iam    iamapi.IAMClientInterface
}

func (s *service) initOperationService(h *rest.Handler) {
	h.Add("GetApplyStatistics", http.MethodPost, "/find/operation/apply/statistics", s.GetApplyStatistics)
}

func (s *service) initRecyclerService(h *rest.Handler) {
	h.Add("GetRecyclability", http.MethodPost, "/findmany/recycle/recyclability", s.GetRecyclability)
	h.Add("PreviewRecycleOrder", http.MethodPost, "/preview/recycle/order", s.PreviewRecycleOrder)
	h.Add("AuditRecycleOrder", http.MethodPost, "/audit/recycle/order", s.AuditRecycleOrder)
	h.Add("CreateRecycleOrder", http.MethodPost, "/create/recycle/order", s.CreateRecycleOrder)
	h.Add("GetRecycleOrder", http.MethodPost, "/findmany/recycle/order", s.GetRecycleOrder)
	h.Add("GetBizRecycleOrder", http.MethodPost, "/findmany/biz/recycle/order", s.GetBizRecycleOrder)
	h.Add("GetRecycleDetect", http.MethodPost, "/findmany/recycle/detect", s.GetRecycleDetect)
	h.Add("ListDetectHost", http.MethodPost, "/list/recycle/detect/host", s.ListDetectHost)
	h.Add("GetRecycleDetectStep", http.MethodPost, "/findmany/recycle/detect/step", s.GetRecycleDetectStep)
	h.Add("StartRecycleOrder", http.MethodPost, "/start/recycle/order", s.StartRecycleOrder)
	h.Add("StartRecycleDetect", http.MethodPost, "/start/recycle/detect", s.StartRecycleDetect)
	h.Add("ReviseRecycleOrder", http.MethodPost, "/revise/recycle/order", s.ReviseRecycleOrder)
	h.Add("PauseRecycleOrder", http.MethodPost, "/pause/recycle", s.PauseRecycleOrder)
	h.Add("ResumeRecycleOrder", http.MethodPost, "/resume/recycle/order", s.ResumeRecycleOrder)
	h.Add("TerminateRecycleOrder", http.MethodPost, "/terminate/recycle/order", s.TerminateRecycleOrder)
	h.Add("GetRecycleOrderHost", http.MethodPost, "/findmany/recycle/host", s.GetRecycleOrderHost)
	h.Add("GetRecycleRecordDeviceType", http.MethodGet, "/find/recycle/record/devicetype", s.GetRecycleRecordDeviceType)
	h.Add("GetRecycleRecordRegion", http.MethodGet, "/find/recycle/record/region", s.GetRecycleRecordRegion)
	h.Add("GetRecycleRecordZone", http.MethodGet, "/find/recycle/record/zone", s.GetRecycleRecordZone)
	h.Add("GetBizHostToRecycle", http.MethodPost, "/find/recycle/biz/host", s.GetBizHostToRecycle)

	// configs related api
	h.Add("GetRecycleStageCfg", http.MethodGet, "/find/config/recycle/stage", s.GetRecycleStageCfg)
	h.Add("GetRecycleStatusCfg", http.MethodGet, "/find/config/recycle/status", s.GetRecycleStatusCfg)
	h.Add("GetDetectStatusCfg", http.MethodGet, "/find/config/recycle/detect/status", s.GetDetectStatusCfg)
	h.Add("GetDetectStepCfg", http.MethodGet, "/find/config/recycle/detect/step", s.GetDetectStepCfg)
}

func (s *service) initSchedulerService(h *rest.Handler) {
	h.Add("UpdateApplyTicket", http.MethodPost, "/update/apply/ticket", s.UpdateApplyTicket)
	h.Add("GetApplyTicket", http.MethodPost, "/get/apply/ticket", s.GetApplyTicket)
	h.Add("GetApplyAudit", http.MethodPost, "/get/apply/ticket/audit", s.GetApplyAudit)
	h.Add("AuditApplyTicket", http.MethodPost, "/audit/apply/ticket", s.AuditApplyTicket)
	h.Add("UpdateApplyTicket", http.MethodPost, "/autoaudit/apply/ticket", s.AutoAuditApplyTicket)
	h.Add("AutoAuditApplyTicket", http.MethodPost, "/approve/apply/ticket", s.ApproveApplyTicket)
	h.Add("CreateApplyOrder", http.MethodPost, "/create/apply", s.CreateApplyOrder)
	h.Add("GetApplyOrder", http.MethodPost, "/findmany/apply", s.GetApplyOrder)
	h.Add("GetBizApplyOrder", http.MethodPost, "/findmany/biz/apply", s.GetBizApplyOrder)
	h.Add("GetApplyStatus", http.MethodGet, "/find/apply/status/{order_id}", s.GetApplyStatus)
	h.Add("GetApplyDetail", http.MethodPost, "/find/apply/detail", s.GetApplyDetail)
	h.Add("GetApplyGenerate", http.MethodPost, "/find/apply/record/generate", s.GetApplyGenerate)
	h.Add("GetApplyInit", http.MethodPost, "/find/apply/record/init", s.GetApplyInit)
	h.Add("GetApplyDiskCheck", http.MethodPost, "/find/apply/record/disk_check", s.GetApplyDiskCheck)
	h.Add("GetApplyDeliver", http.MethodPost, "/find/apply/record/deliver", s.GetApplyDeliver)
	h.Add("GetApplyDevice", http.MethodPost, "/findmany/apply/device", s.GetApplyDevice)
	h.Add("GetDeliverDeviceByOrder", http.MethodPost, "/findmany/apply/deliver/device", s.GetDeliverDeviceByOrder)
	h.Add("ExportDeliverDevice", http.MethodPost, "/export/apply/deliver/device", s.ExportDeliverDevice)
	h.Add("MatchDevice", http.MethodPost, "/findmany/apply/match/device", s.GetMatchDevice)
	h.Add("MatchDevice", http.MethodPost, "/commit/apply/match", s.MatchDevice)
	h.Add("MatchPoolDevice", http.MethodPost, "/commit/apply/pool/match", s.MatchPoolDevice)
	h.Add("PauseApplyOrder", http.MethodPost, "/pause/apply", s.PauseApplyOrder)
	h.Add("ResumeApplyOrder", http.MethodPost, "/resume/apply", s.ResumeApplyOrder)
	h.Add("StartApplyOrder", http.MethodPost, "/start/apply", s.StartApplyOrder)
	h.Add("TerminateApplyOrder", http.MethodPost, "/terminate/apply", s.TerminateApplyOrder)
	h.Add("ModifyApplyOrder", http.MethodPost, "/modify/apply", s.ModifyApplyOrder)
	h.Add("RecommendApplyOrder", http.MethodPost, "/recommend/apply", s.RecommendApplyOrder)
	h.Add("GetApplyModify", http.MethodPost, "/find/apply/record/modify", s.GetApplyModify)
}
