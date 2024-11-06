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

// ResourcePoolBusinessCreateReq create request
type ResourcePoolBusinessCreateReq struct {
	Businesses []ResourcePoolBusinessCreate `json:"businesses" validate:"required,max=100"`
}

// Validate validate
func (r *ResourcePoolBusinessCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Businesses {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResourcePoolBusinessCreate create request
type ResourcePoolBusinessCreate struct {
	BkBizID   int64  `json:"bk_biz_id" validate:"required"`
	BkBizName string `json:"bk_biz_name" validate:"required"`
}

// Validate validate
func (r *ResourcePoolBusinessCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// ResourcePoolBusinessListResult list result.
type ResourcePoolBusinessListResult struct {
	Count   uint64                              `json:"count"`
	Details []tablers.ResourcePoolBusinessTable `json:"details"`
}

// ResourcePoolBusinessListReq list request
type ResourcePoolBusinessListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"field" validate:"omitempty"`
}

// Validate validate
func (r *ResourcePoolBusinessListReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResourcePoolBusinessBatchUpdateReq batch update request
type ResourcePoolBusinessBatchUpdateReq struct {
	Businesses []ResourcePoolBusinessUpdateReq `json:"businesses" validate:"required,max=100"`
}

// Validate validate
func (r *ResourcePoolBusinessBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Businesses {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResourcePoolBusinessUpdateReq batch update request
type ResourcePoolBusinessUpdateReq struct {
	ID        string `json:"id" validate:"required"`
	BkBizID   int64  `json:"bk_biz_id"`
	BkBizName string `json:"bk_biz_name"`
}

// Validate validate
func (r *ResourcePoolBusinessUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}
