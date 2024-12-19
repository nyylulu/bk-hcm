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

package caiche

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// GetTokenReq get token request
type GetTokenReq struct {
	ID       string      `json:"id" validate:"omitempty"`
	JsonRPC  string      `json:"jsonrpc" validate:"omitempty"`
	Params   GrantParams `json:"params" validate:"required"`
	Reason   string      `json:"reason" validate:"omitempty"`
	XTraceID string      `json:"x_trace_id" validate:"omitempty"`
}

// GrantParams grant params
type GrantParams struct {
	AppKey    string    `json:"app_key" validate:"required"`
	AppSecret string    `json:"app_secret" validate:"required"`
	GrantType GrantType `json:"grant_type" validate:"required"`
}

// ListDeviceReq list device request
type ListDeviceReq struct {
	PlanProduct           []string              `json:"plan_product,omitempty" validate:"omitempty"`
	AbolishPrincipal      []string              `json:"abolish_principal,omitempty" validate:"omitempty"`
	SvrOwnerAssetID       []string              `json:"svr_owner_asset_id,omitempty" validate:"omitempty"`
	ExpectAbolishDate     string                `json:"expect_abolish_date,omitempty" validate:"omitempty"`
	SvrAssetID            []string              `json:"svr_asset_id,omitempty" validate:"omitempty"`
	AbolishPhase          []enumor.AbolishPhase `json:"abolish_phase,omitempty" validate:"omitempty"`
	ServerLanIP           []string              `json:"server_lan_ip,omitempty" validate:"omitempty"`
	ProjectName           []string              `json:"project_name,omitempty" validate:"omitempty"`
	DeptName              []string              `json:"dept_name,omitempty" validate:"omitempty"`
	Module                []string              `json:"module,omitempty" validate:"omitempty"`
	VirtualDepartmentName []string              `json:"virtual_department_name,omitempty" validate:"omitempty"`
	ObsBG                 []string              `json:"obs_bg,omitempty" validate:"omitempty"`
	PageIndex             uint                  `json:"page_index" validate:"required,min=1"`
	PageSize              uint                  `json:"page_size" validate:"required,min=1,max=500"`
}

// Validate ...
func (d *ListDeviceReq) Validate() error {
	return validator.Validate.Struct(d)
}
