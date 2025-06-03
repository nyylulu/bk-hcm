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

package pool

import (
	"context"
	"net/http"

	"hcm/cmd/woa-server/logics/pool"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		authorizer: c.Authorizer,
		logics:     pool.New(context.Background(), c.Conf.ClientConfig, c.ThirdCli, c.CmdbCli),
	}
	h := rest.NewHandler()

	s.initPoolService(h)

	h.Load(c.WebService)

	// 业务下的接口
	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}")
	bizService(bizH, s)

	bizH.Load(c.WebService)
}

type service struct {
	logics     pool.Logics
	authorizer auth.Authorizer
}

func (s *service) initPoolService(h *rest.Handler) {
	h.Add("CreateLaunchTask", http.MethodPost, "/pool/create/launch/task", s.CreateLaunchTask)
	h.Add("CreateRecallTask", http.MethodPost, "/pool/create/recall/task", s.CreateRecallTask)
	h.Add("GetLaunchTask", http.MethodPost, "/pool/findmany/launch/task", s.GetLaunchTask)
	h.Add("GetRecallTask", http.MethodPost, "/pool/findmany/recall/task", s.GetRecallTask)
	h.Add("GetLaunchHost", http.MethodPost, "/pool/findmany/launch/host", s.GetLaunchHost)
	h.Add("GetRecallHost", http.MethodPost, "/pool/findmany/recall/host", s.GetRecallHost)
	h.Add("GetIdleHost", http.MethodPost, "/pool/findmany/idle/host", s.GetIdleHost)
	h.Add("DrawHost", http.MethodPost, "/pool/draw/host", s.DrawHost)
	h.Add("ReturnHost", http.MethodPost, "/pool/return/host", s.ReturnHost)
	h.Add("CreateRecallOrder", http.MethodPost, "/pool/create/recall/order", s.CreateRecallOrder)
	h.Add("GetRecallOrder", http.MethodPost, "/pool/find/recall/order", s.GetRecallOrder)
	h.Add("GetRecalledInstance", http.MethodPost, "/pool/find/recall/order/instance", s.GetRecalledInstance)

	h.Add("GetRecallDetail", http.MethodPost, "/pool/findmany/recall/detail", s.GetRecallDetail)

	h.Add("GetLaunchMatchDevice", http.MethodPost, "/pool/findmany/launch/match/device", s.GetLaunchMatchDevice)
	h.Add("GetRecallMatchDevice", http.MethodPost, "/pool/findmany/recall/match/device", s.GetRecallMatchDevice)

	h.Add("ResumeRecycleTask", http.MethodPost, "/pool/resume/recycle/task", s.ResumeRecycleTask)

	// configs related api
	h.Add("CreateGradeCfg", http.MethodPost, "/pool/create/config/grade", s.CreateGradeCfg)
	h.Add("GetGradeCfg", http.MethodGet, "/pool/find/config/grade", s.GetGradeCfg)
	h.Add("GetRecallStatusCfg", http.MethodGet, "/pool/find/config/recall/status", s.GetRecallStatusCfg)
	h.Add("GetTaskStatusCfg", http.MethodGet, "/pool/find/config/task/status", s.GetTaskStatusCfg)
	h.Add("GetDeviceType", http.MethodGet, "/pool/find/config/devicetype", s.GetDeviceType)
}

// bizService 业务下的接口
func bizService(h *rest.Handler, s *service) {
	h.Add("GetBizRecallMatchDevice", http.MethodPost, "/pool/findmany/recall/match/device", s.GetBizRecallMatchDevice)
}
