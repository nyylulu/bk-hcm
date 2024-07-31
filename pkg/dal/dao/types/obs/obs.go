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

package obs

import (
	"hcm/pkg/api/core"
	tableobs "hcm/pkg/dal/table/obs"
)

// ListOBSBillItemHuaweiDetails list account bill summary daily details.
type ListOBSBillItemHuaweiDetails struct {
	Count   *uint64                      `json:"count,omitempty"`
	Details []tableobs.OBSBillItemHuawei `json:"details,omitempty"`
}

// ListOBSBillItemAwsDetails Obs aws bill item
type ListOBSBillItemAwsDetails = core.ListResultT[tableobs.OBSBillItemAws]

// ListOBSBillItemGcpDetails Obs gcp bill item
type ListOBSBillItemGcpDetails = core.ListResultT[tableobs.OBSBillItemGcp]
