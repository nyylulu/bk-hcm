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

package bill

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
)

// AwsSPSavedCostReq saving plans req
type AwsSPSavedCostReq struct {
	RootAccountID       string         `json:"root_account_id" validate:"omitempty"`
	MainAccountIDs      []string       `json:"main_account_ids" validate:"omitempty,max=10"`
	MainAccountCloudIDs []string       `json:"main_account_cloud_ids" validate:"omitempty,max=10"`
	ProductIDs          []int64        `json:"product_ids" validate:"omitempty,max=10"`
	Year                uint           `json:"year" validate:"required"`
	Month               uint           `json:"month" validate:"required,min=1,max=12"`
	StartDay            uint           `json:"start_day" validate:"required,min=1,max=31"`
	EndDay              uint           `json:"end_day" validate:"required,min=1,max=31"`
	Page                *core.BasePage `json:"page" validate:"required"`
}

// Validate validate saving plans req
func (s AwsSPSavedCostReq) Validate() error {
	return validator.Validate.Struct(s)
}

// AwsAccountSPCost ...
type AwsAccountSPCost struct {
	MainAccountID          string   `json:"main_account_id"`
	MainAccountCloudID     string   `json:"main_account_cloud_id"`
	MainAccountManagers    []string `json:"main_account_managers"`
	MainAccountBakManagers []string `json:"main_account_bak_managers"`
	ProductId              int64    `json:"product_id"`

	SpArn           string   `json:"sp_arn"`
	SpManagers      []string `json:"sp_managers"`
	SpBakManagers   []string `json:"sp_bak_managers"`
	UnblendedCost   string   `json:"unblended_cost"`
	SPEffectiveCost string   `json:"sp_effective_cost"`
	SPNetCost       string   `json:"sp_net_cost"`
	SPSavedCost     string   `json:"sp_saved_cost"`
}

// AwsSPCostResult ...
type AwsSPCostResult struct {
	Count      uint64             `json:"count"`
	Details    []AwsAccountSPCost `json:"details"`
	BatchTotal string             `json:"batch_total"`
}
