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
)

// RollingQuotaOffsetCreateReq create request
type RollingQuotaOffsetCreateReq struct {
	QuotaOffsets []RollingQuotaOffsetCreate `json:"quota_offsets" validate:"required,max=100"`
}

// Validate validate
func (r *RollingQuotaOffsetCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.QuotaOffsets {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// RollingQuotaOffsetCreate create request
type RollingQuotaOffsetCreate struct {
	BkBizID     int64  `json:"bk_biz_id" validate:"required"`
	BkBizName   string `json:"bk_biz_name" validate:"required"`
	Year        int64  `json:"year" validate:"required"`
	Month       int64  `json:"month" validate:"required"`
	QuotaOffset int64  `json:"quota_offset" validate:"required"`
}

// Validate validate
func (r *RollingQuotaOffsetCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// RollingQuotaOffsetListResult list rolling quota offset result.
type RollingQuotaOffsetListResult struct {
	Count   uint64                            `json:"count"`
	Details []tablers.RollingQuotaOffsetTable `json:"details"`
}

// RollingQuotaOffsetListReq list request
type RollingQuotaOffsetListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"field" validate:"omitempty"`
}

// Validate validate
func (r *RollingQuotaOffsetListReq) Validate() error {
	return validator.Validate.Struct(r)
}

// RollingQuotaOffsetBatchUpdateReq batch update request
type RollingQuotaOffsetBatchUpdateReq struct {
	QuotaOffsets []RollingQuotaOffsetUpdateReq `json:"quota_offsets" validate:"required,max=100"`
}

// Validate validate
func (r *RollingQuotaOffsetBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.QuotaOffsets {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// RollingQuotaOffsetUpdateReq batch update request
type RollingQuotaOffsetUpdateReq struct {
	ID          string `json:"id" validate:"required"`
	BkBizID     int64  `json:"bk_biz_id"`
	BkBizName   string `json:"bk_biz_name"`
	Year        int64  `json:"year"`
	Month       int64  `json:"month"`
	QuotaOffset *int64 `json:"quota_offset"`
}

// Validate validate
func (r *RollingQuotaOffsetUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}
