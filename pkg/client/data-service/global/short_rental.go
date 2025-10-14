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
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// --- short rental returned record ---

// ListShortRentalReturnedRecord list short rental returned record
func (b *ResourcePlanClient) ListShortRentalReturnedRecord(kt *kit.Kit, req *core.ListReq) (
	*rpproto.ShortRentalReturnedRecordListResult, error) {

	return common.Request[core.ListReq, rpproto.ShortRentalReturnedRecordListResult](
		b.client, rest.POST, kt, req, "/short_rental/returned_records/list")
}

// BatchCreateShortRentalReturnedRecord batch create short rental returned record
func (b *ResourcePlanClient) BatchCreateShortRentalReturnedRecord(kt *kit.Kit,
	req *rpproto.ShortRentalReturnedRecordBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.ShortRentalReturnedRecordBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/short_rental/returned_records/batch/create")
}

// BatchUpdateShortRentalReturnedRecord update short rental returned record
func (b *ResourcePlanClient) BatchUpdateShortRentalReturnedRecord(kt *kit.Kit,
	req *rpproto.ShortRentalReturnedRecordBatchUpdateReq) error {

	return common.RequestNoResp[rpproto.ShortRentalReturnedRecordBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/short_rental/returned_records/batch")
}

// DeleteShortRentalReturnedRecord delete short rental returned record
func (b *ResourcePlanClient) DeleteShortRentalReturnedRecord(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/short_rental/returned_records/batch")
}

// SumReturnedCore sum short rental returned record returned core
func (b *ResourcePlanClient) SumReturnedCore(kt *kit.Kit, req *rpproto.ShortRentalReturnedRecordSumReq) (
	*rpproto.ShortRentalReturnedRecordSumResult, error) {

	return common.Request[rpproto.ShortRentalReturnedRecordSumReq, rpproto.ShortRentalReturnedRecordSumResult](
		b.client, rest.POST, kt, req, "/short_rental/returned_records/sum")
}
