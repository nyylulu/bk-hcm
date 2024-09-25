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

package meta

import (
	"net/http"

	"hcm/cmd/woa-server/logics/meta"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the meta service.
func InitService(c *capability.Capability) {
	s := &service{
		dao:        c.Dao,
		authorizer: c.Authorizer,
		logics:     meta.New(c.EsbClient, c.Authorizer),
	}
	h := rest.NewHandler()

	s.initMetaService(h)

	h.Load(c.WebService)
}

type service struct {
	dao        dao.Set
	authorizer auth.Authorizer
	logics     meta.Logics
}

func (s *service) initMetaService(h *rest.Handler) {
	h.Add("ListDiskType", http.MethodGet, "/meta/disk_type/list", s.ListDiskType)
	h.Add("ListObsProject", http.MethodGet, "/meta/obs_project/list", s.ListObsProject)
	h.Add("ListRegion", http.MethodGet, "/meta/region/list", s.ListRegion)
	h.Add("ListZone", http.MethodPost, "/meta/zone/list", s.ListZone)
	h.Add("ListDeviceClass", http.MethodGet, "/meta/device_class/list", s.ListDeviceClass)
	h.Add("ListDeviceType", http.MethodPost, "/meta/device_type/list", s.ListDeviceType)
	h.Add("ListBizsByOpProduct", http.MethodPost, "/metas/bizs/by/op_product/list", s.ListBizsByOpProduct)
	h.Add("ListOpProducts", http.MethodPost, "/metas/op_products/list", s.ListOpProducts)
	h.Add("ListPlanProducts", http.MethodPost, "/metas/plan_products/list", s.ListPlanProducts)
}
