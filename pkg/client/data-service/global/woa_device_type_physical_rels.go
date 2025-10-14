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

// ListWoaDeviceTypePhysicalRel list woa device type physical rel
func (b *ResourcePlanClient) ListWoaDeviceTypePhysicalRel(kt *kit.Kit, req *core.ListReq) (
	*rpproto.WoaDeviceTypePhysicalRelListResult, error) {

	return common.Request[core.ListReq, rpproto.WoaDeviceTypePhysicalRelListResult](
		b.client, rest.POST, kt, req, "/res_plans/woa_device_type_physical_rels/list")
}

// BatchCreateWoaDeviceTypePhysicalRel batch create woa device type physical rel
func (b *ResourcePlanClient) BatchCreateWoaDeviceTypePhysicalRel(kt *kit.Kit,
	req *rpproto.WoaDeviceTypePhysicalRelBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.WoaDeviceTypePhysicalRelBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/woa_device_type_physical_rels/batch/create")
}

// BatchUpdateWoaDeviceTypePhysicalRel update woa device type physical rel
func (b *ResourcePlanClient) BatchUpdateWoaDeviceTypePhysicalRel(kt *kit.Kit,
	req *rpproto.WoaDeviceTypePhysicalRelBatchUpdateReq) error {

	return common.RequestNoResp[rpproto.WoaDeviceTypePhysicalRelBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/woa_device_type_physical_rels/batch")
}

// DeleteWoaDeviceTypePhysicalRel delete woa device type physical rel
func (b *ResourcePlanClient) DeleteWoaDeviceTypePhysicalRel(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/woa_device_type_physical_rels/batch")
}
