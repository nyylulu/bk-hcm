/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package operationproduct

// OperationProduct 运营产品
type OperationProduct struct {
	OpProductId          int64  `json:"op_product_id"`
	OpProductName        string `json:"op_product_name"`
	OpProductManagers    string `json:"op_product_managers"`
	OpProductBakManagers string `json:"op_product_bak_managers"`
	PlanProductId        int64  `json:"plan_product_id"`
	PlanProductName      string `json:"plan_product_name"`
	BgId                 int64  `json:"bg_id"`
	BgName               string `json:"bg_name"`
	BgShortName          string `json:"bg_short_name"`
	DeptId               int64  `json:"dept_id"`
	DeptName             string `json:"dept_name"`
}
