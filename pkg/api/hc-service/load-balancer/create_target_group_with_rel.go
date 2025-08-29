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

package hclb

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// CreateTargetGroupWithRelReq 创建目标组并绑定监听器请求
type CreateTargetGroupWithRelReq struct {
	Vendor              enumor.Vendor     `json:"vendor" validate:"required"`
	LoadBalancerID      string            `json:"lb_id" validate:"required"`
	ListenerID          string            `json:"listener_id" validate:"required"`
	ListenerRuleID      string            `json:"listener_rule_id"`
	RuleType            enumor.RuleType   `json:"rule_type" validate:"required"`
	Targets             []*RegisterTarget `json:"targets" validate:"required"`
	ManagementDetailIDs []string          `json:"management_detail_ids"`
}

// Validate 验证请求参数
func (req *CreateTargetGroupWithRelReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateTargetGroupWithRelResult 创建目标组并绑定监听器结果
type CreateTargetGroupWithRelResult struct {
	TargetGroupID string `json:"target_group_id"`
}
