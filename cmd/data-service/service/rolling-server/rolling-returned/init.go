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

// Package rollingreturned ...
package rollingreturned

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initialize the rolling returned record service
func InitService(cap *capability.Capability) {
	svc := &service{
		dao: cap.Dao,
	}
	h := rest.NewHandler()
	h.Add("BatchCreateRollingReturnedRecord", http.MethodPost, "/rolling_servers/returned_records/batch/create",
		svc.BatchCreateRollingReturnedRecord)
	h.Add("DeleteRollingReturnedRecord", http.MethodDelete, "/rolling_servers/returned_records/batch",
		svc.DeleteRollingReturnedRecord)
	h.Add("BatchUpdateRollingReturnedRecord", http.MethodPatch, "/rolling_servers/returned_records/batch",
		svc.BatchUpdateRollingReturnedRecord)
	h.Add("ListRollingReturnedRecord", http.MethodPost, "/rolling_servers/returned_records/list",
		svc.ListRollingReturnedRecord)
	h.Add("GetRollingReturnedCoreSum", http.MethodPost, "/rolling_servers/returned_records/returned_core/sum",
		svc.GetRollingReturnedCoreSum)

	h.Load(cap.WebService)
}

type service struct {
	dao dao.Set
}
