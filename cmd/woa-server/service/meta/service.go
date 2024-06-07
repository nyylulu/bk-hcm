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

	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initial the meta service.
func InitService(c *capability.Capability) {
	s := &service{
		dao: c.Dao,
	}
	h := rest.NewHandler()

	s.initMetaService(h)

	h.Load(c.WebService)
}

type service struct {
	dao dao.Set
}

func (s *service) initMetaService(h *rest.Handler) {
	h.Add("ListDiskType", http.MethodGet, "/meta/disk_type/list", s.ListDiskType)
	h.Add("ListObsProject", http.MethodGet, "/meta/obs_project/list", s.ListObsProject)
}
