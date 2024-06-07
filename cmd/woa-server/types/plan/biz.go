/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package plan ...
package plan

// BizOrgRel is GetBizOrgRel result.
type BizOrgRel struct {
	BizID           int64  `json:"bk_biz_id"`
	BizName         string `json:"bk_biz_name"`
	BkProductID     int64  `json:"bk_product_id"`
	BkProductName   string `json:"bk_product_name"`
	PlanProductID   int64  `json:"plan_product_id"`
	PlanProductName string `json:"plan_product_name"`
	VirtualDeptID   int64  `json:"virtual_dept_id"`
	VirtualDeptName string `json:"virtual_dept_name"`
}
