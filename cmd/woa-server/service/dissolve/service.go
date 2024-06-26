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

package dissolve

import (
	"net/http"

	"hcm/cmd/woa-server/logics/dissolve"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		logics: dissolve.New(c.Dao, c.EsbClient, c.EsCli, c.Conf.ResDissolve.OriginDate),
	}
	h := rest.NewHandler()

	s.initDissolveService(h)

	h.Load(c.WebService)
}

type service struct {
	logics dissolve.Logics
}

func (s *service) initDissolveService(h *rest.Handler) {
	// recycle module
	h.Add("CreateRecycledModule", http.MethodPost, "/dissolve/recycled_module/create", s.CreateRecycledModule)
	h.Add("UpdateRecycledModule", http.MethodPut, "/dissolve/recycled_module/update", s.UpdateRecycledModule)
	h.Add("ListRecycledModule", http.MethodPost, "/dissolve/recycled_module/list", s.ListRecycledModule)
	h.Add("DeleteRecycledModule", http.MethodDelete, "/dissolve/recycled_module/delete", s.DeleteRecycledModule)

	// recycle host
	h.Add("CreateRecycledHost", http.MethodPost, "/dissolve/recycled_host/create", s.CreateRecycledHost)
	h.Add("UpdateRecycledHost", http.MethodPut, "/dissolve/recycled_host/update", s.UpdateRecycledHost)
	h.Add("ListRecycledHost", http.MethodPost, "/dissolve/recycled_host/list", s.ListRecycledHost)
	h.Add("DeleteRecycledHost", http.MethodDelete, "/dissolve/recycled_host/delete", s.DeleteRecycledHost)

	// resource dissolve
	h.Add("ListOriginHost", http.MethodPost, "/dissolve/host/origin/list", s.ListOriginHost)
	h.Add("ListCurrentHost", http.MethodPost, "/dissolve/host/current/list", s.ListCurHost)
	h.Add("ListResDissolveTable", http.MethodPost, "/dissolve/table/list", s.ListResDissolveTable)
}
