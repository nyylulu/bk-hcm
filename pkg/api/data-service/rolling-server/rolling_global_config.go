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

// Package rollingserver ...
package rollingserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// RollingGlobalConfigCreateReq create request
type RollingGlobalConfigCreateReq struct {
	GlobalConfigs []RollingGlobalConfigCreate `json:"global_configs" validate:"required,max=1"`
}

// Validate validate
func (r *RollingGlobalConfigCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.GlobalConfigs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// RollingGlobalConfigCreate create request
type RollingGlobalConfigCreate struct {
	GlobalQuota *int64          `json:"global_quota" validate:"required"`
	BizQuota    *int64          `json:"biz_quota" validate:"required"`
	UnitPrice   decimal.Decimal `json:"unit_price" validate:"required"`
}

// Validate validate
func (r *RollingGlobalConfigCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// RollingGlobalConfigListResult list rolling global config result.
type RollingGlobalConfigListResult struct {
	Count   uint64                             `json:"count"`
	Details []tablers.RollingGlobalConfigTable `json:"details"`
}

// RollingGlobalConfigListReq list request
type RollingGlobalConfigListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"field" validate:"omitempty"`
}

// Validate validate
func (r *RollingGlobalConfigListReq) Validate() error {
	return validator.Validate.Struct(r)
}

// RollingGlobalConfigBatchUpdateReq batch update request
type RollingGlobalConfigBatchUpdateReq struct {
	GlobalConfigs []RollingGlobalConfigUpdateReq `json:"global_configs" validate:"required,max=1"`
}

// Validate validate
func (r *RollingGlobalConfigBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.GlobalConfigs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// RollingGlobalConfigUpdateReq batch update request
type RollingGlobalConfigUpdateReq struct {
	ID          string           `json:"id" validate:"required"`
	GlobalQuota *int64           `json:"global_quota"`
	BizQuota    *int64           `json:"biz_quota"`
	UnitPrice   *decimal.Decimal `json:"unit_price"`
}

// Validate validate
func (r *RollingGlobalConfigUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}
