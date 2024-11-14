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

// Package greenchannel ...
package greenchannel

import (
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"net/http"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		client:     c.Client,
		authorizer: c.Authorizer,
		gcLogics:   c.GcLogic,
	}
	h := rest.NewHandler()
	h.Path("/green_channels")

	s.initService(h)
	h.Load(c.WebService)

	// 业务下的接口
	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}/green_channels")
	s.bizService(bizH)
	bizH.Load(c.WebService)
}

type service struct {
	authorizer auth.Authorizer
	client     *client.ClientSet
	gcLogics   gclogics.Logics
}

// initService 资源下的接口
func (s *service) initService(h *rest.Handler) {
	h.Add("GetGreenChannelCpuCoreSummary", http.MethodPost, "/cpu_core/summary", s.GetCpuCoreSummary)
	h.Add("ListGreenChannelStatisticalRecord", http.MethodPost, "/statistical_record/list", s.ListStatisticalRecord)
}

// bizService 业务下的接口
func (s *service) bizService(h *rest.Handler) {
	h.Add("GetGreenChannelBizCpuCoreSummary", http.MethodPost, "/cpu_core/summary", s.GetBizCpuCoreSummary)
}
