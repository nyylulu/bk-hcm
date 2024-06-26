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

package cvm

import (
	"net/http"

	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/cvm"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		authorizer: c.Authorizer,
		logics:     cvm.New(c.ThirdCli, c.Conf.ClientConfig, config.New(c.ThirdCli)),
	}
	h := rest.NewHandler()

	s.initCvmService(h)

	h.Load(c.WebService)
}

type service struct {
	logics     cvm.Logics
	authorizer auth.Authorizer
}

func (s *service) initCvmService(h *rest.Handler) {
	h.Add("CreateApplyOrder", http.MethodPost, "/cvm/create/apply/order", s.CreateApplyOrder)
	h.Add("GetApplyOrderById", http.MethodPost, "/cvm/find/apply/order", s.GetApplyOrderById)
	h.Add("GetApplyOrder", http.MethodPost, "/cvm/findmany/apply/order", s.GetApplyOrder)
	h.Add("GetApplyDevice", http.MethodPost, "/cvm/findmany/apply/device", s.GetApplyDevice)
	h.Add("GetCapacity", http.MethodPost, "/cvm/find/capacity", s.GetCapacity)

	h.Add("GetApplyStatusCfg", http.MethodGet, "/cvm/find/config/apply/status", s.GetApplyStatusCfg)
}
