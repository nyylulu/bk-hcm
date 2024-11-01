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

package rollingserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/runtime/filter"
)

// QuotaOffsetAuditCreateReq create request
type QuotaOffsetAuditCreateReq struct {
	QuotaOffsetsAudit []QuotaOffsetAuditCreate `json:"quota_offsets_audit" validate:"required,max=100"`
}

// Validate validate
func (r *QuotaOffsetAuditCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.QuotaOffsetsAudit {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// QuotaOffsetAuditCreate create request
type QuotaOffsetAuditCreate struct {
	OffsetConfigID string `json:"offset_config_id" validate:"required"`
	Operator       string `json:"operator" validate:"required"`
	QuotaOffset    *int64 `json:"quota_offset" validate:"required"`
	Rid            string `json:"rid" validate:"required"`
	AppCode        string `json:"app_code" validate:"omitempty"`
}

// Validate validate
func (r *QuotaOffsetAuditCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// QuotaOffsetAuditListResult list rolling quota offset result.
type QuotaOffsetAuditListResult struct {
	Count   uint64                                 `json:"count"`
	Details []tablers.RollingQuotaOffsetAuditTable `json:"details"`
}

// QuotaOffsetAuditListReq list request
type QuotaOffsetAuditListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"field" validate:"omitempty"`
}

// Validate validate
func (r *QuotaOffsetAuditListReq) Validate() error {
	return validator.Validate.Struct(r)
}

// QuotaOffsetAuditBatchUpdateReq batch update request
type QuotaOffsetAuditBatchUpdateReq struct {
	QuotaOffsetsAudit []QuotaOffsetAuditUpdateReq `json:"quota_offsets_audit" validate:"required,max=100"`
}

// Validate validate
func (r *QuotaOffsetAuditBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.QuotaOffsetsAudit {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// QuotaOffsetAuditUpdateReq batch update request
type QuotaOffsetAuditUpdateReq struct {
	ID             string `json:"id" validate:"required"`
	OffsetConfigID string `json:"offset_config_id"`
	Operator       string `json:"operator"`
	QuotaOffset    *int64 `json:"quota_offset"`
	Rid            string `json:"rid"`
	AppCode        string `json:"app_code"`
}

// Validate validate
func (r *QuotaOffsetAuditUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}
