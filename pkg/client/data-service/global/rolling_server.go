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

package global

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// RollingServerClient is data service rolling server api client.
type RollingServerClient struct {
	client rest.ClientInterface
}

// NewRollingServerClient create a new rolling server api client.
func NewRollingServerClient(client rest.ClientInterface) *RollingServerClient {
	return &RollingServerClient{
		client: client,
	}
}

// --- rolling applied record ---

// ListAppliedRecord list applied record
func (b *RollingServerClient) ListAppliedRecord(kt *kit.Kit, req *rsproto.RollingAppliedRecordListReq) (
	*rsproto.RollingAppliedRecordListResult, error) {
	return common.Request[rsproto.RollingAppliedRecordListReq, rsproto.RollingAppliedRecordListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/applied_records/list")
}

// CreateAppliedRecord create applied record
func (b *RollingServerClient) CreateAppliedRecord(kt *kit.Kit, req *rsproto.RollingAppliedRecordCreateReq) (
	*core.BatchCreateResult, error) {
	return common.Request[rsproto.RollingAppliedRecordCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/applied_records/batch/create")
}

// UpdateAppliedRecord update applied record
func (b *RollingServerClient) UpdateAppliedRecord(kt *kit.Kit, req *rsproto.RollingAppliedRecordUpdateReq) error {
	return common.RequestNoResp[rsproto.RollingAppliedRecordUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/applied_records/batch")
}

// BatchDeleteAppliedRecord delete applied record
func (b *RollingServerClient) BatchDeleteAppliedRecord(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/applied_records/batch")
}

// --- rolling returned record ---

// ListReturnedRecord list returned record
func (b *RollingServerClient) ListReturnedRecord(kt *kit.Kit, req *rsproto.RollingReturnedRecordListReq) (
	*rsproto.RollingReturnedRecordListResult, error) {
	return common.Request[rsproto.RollingReturnedRecordListReq, rsproto.RollingReturnedRecordListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/returned_records/list")
}

// CreateReturnedRecord create returned record
func (b *RollingServerClient) CreateReturnedRecord(kt *kit.Kit, req *rsproto.RollingReturnedRecordCreateReq) (
	*core.BatchCreateResult, error) {
	return common.Request[rsproto.RollingReturnedRecordCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/returned_records/batch/create")
}

// UpdateReturnedRecord update returned record
func (b *RollingServerClient) UpdateReturnedRecord(kt *kit.Kit, req *rsproto.RollingReturnedRecordUpdateReq) error {
	return common.RequestNoResp[rsproto.RollingReturnedRecordUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/returned_records/batch")
}

// BatchDeleteReturnedRecord delete returned record
func (b *RollingServerClient) BatchDeleteReturnedRecord(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/returned_records/batch")
}
