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

// --- resource pool business ---

// ListResPoolBiz list resource pool business
func (b *RollingServerClient) ListResPoolBiz(kt *kit.Kit, req *rsproto.ResourcePoolBusinessListReq) (
	*rsproto.ResourcePoolBusinessListResult, error) {

	return common.Request[rsproto.ResourcePoolBusinessListReq, rsproto.ResourcePoolBusinessListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/respool_bizs/list")
}

// BatchCreateResPoolBiz batch create resource pool business
func (b *RollingServerClient) BatchCreateResPoolBiz(kt *kit.Kit, req *rsproto.ResourcePoolBusinessCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rsproto.ResourcePoolBusinessCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/respool_bizs/batch/create")
}

// BatchUpdateResPoolBiz update resource pool business
func (b *RollingServerClient) BatchUpdateResPoolBiz(kt *kit.Kit,
	req *rsproto.ResourcePoolBusinessBatchUpdateReq) error {

	return common.RequestNoResp[rsproto.ResourcePoolBusinessBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/respool_bizs/batch")
}

// DeleteResPoolBiz delete resource pool business
func (b *RollingServerClient) DeleteResPoolBiz(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/respool_bizs/batch")
}

// --- rolling global config ---

// ListGlobalConfig list global config
func (b *RollingServerClient) ListGlobalConfig(kt *kit.Kit, req *rsproto.RollingGlobalConfigListReq) (
	*rsproto.RollingGlobalConfigListResult, error) {

	return common.Request[rsproto.RollingGlobalConfigListReq, rsproto.RollingGlobalConfigListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/global_configs/list")
}

// BatchCreateGlobalConfig batch create global config
func (b *RollingServerClient) BatchCreateGlobalConfig(kt *kit.Kit, req *rsproto.RollingGlobalConfigCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rsproto.RollingGlobalConfigCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/global_configs/batch/create")
}

// BatchUpdateGlobalConfig update global config
func (b *RollingServerClient) BatchUpdateGlobalConfig(kt *kit.Kit,
	req *rsproto.RollingGlobalConfigBatchUpdateReq) error {

	return common.RequestNoResp[rsproto.RollingGlobalConfigBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/global_configs/batch")
}

// DeleteGlobalConfig delete global config
func (b *RollingServerClient) DeleteGlobalConfig(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/global_configs/batch")
}

// --- rolling quota config ---

// ListQuotaConfig list quota config
func (b *RollingServerClient) ListQuotaConfig(kt *kit.Kit, req *rsproto.RollingQuotaConfigListReq) (
	*rsproto.RollingQuotaConfigListResult, error) {

	return common.Request[rsproto.RollingQuotaConfigListReq, rsproto.RollingQuotaConfigListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/quota_configs/list")
}

// BatchCreateQuotaConfig batch create quota config
func (b *RollingServerClient) BatchCreateQuotaConfig(kt *kit.Kit, req *rsproto.RollingQuotaConfigCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rsproto.RollingQuotaConfigCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/quota_configs/batch/create")
}

// BatchUpdateQuotaConfig update quota config
func (b *RollingServerClient) BatchUpdateQuotaConfig(kt *kit.Kit,
	req *rsproto.RollingQuotaConfigBatchUpdateReq) error {

	return common.RequestNoResp[rsproto.RollingQuotaConfigBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/quota_configs/batch")
}

// DeleteQuotaConfig delete quota config
func (b *RollingServerClient) DeleteQuotaConfig(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/quota_configs/batch")
}

// --- rolling quota offset ---

// ListQuotaOffset list quota offset
func (b *RollingServerClient) ListQuotaOffset(kt *kit.Kit, req *rsproto.RollingQuotaOffsetListReq) (
	*rsproto.RollingQuotaOffsetListResult, error) {

	return common.Request[rsproto.RollingQuotaOffsetListReq, rsproto.RollingQuotaOffsetListResult](
		b.client, rest.POST, kt, req, "/rolling_servers/quota_offsets/list")
}

// BatchCreateQuotaOffset batch create quota offset
func (b *RollingServerClient) BatchCreateQuotaOffset(kt *kit.Kit, req *rsproto.RollingQuotaOffsetCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rsproto.RollingQuotaOffsetCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/rolling_servers/quota_offsets/batch/create")
}

// BatchUpdateQuotaOffset update quota offset
func (b *RollingServerClient) BatchUpdateQuotaOffset(kt *kit.Kit,
	req *rsproto.RollingQuotaOffsetBatchUpdateReq) error {

	return common.RequestNoResp[rsproto.RollingQuotaOffsetBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/rolling_servers/quota_offsets/batch")
}

// DeleteQuotaOffset delete quota offset
func (b *RollingServerClient) DeleteQuotaOffset(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/rolling_servers/quota_offsets/batch")
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
