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

package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-week"
)

// ResPlanWeekBatchCreateReq create request
type ResPlanWeekBatchCreateReq struct {
	Weeks []ResPlanWeekCreateReq `json:"weeks" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanWeekBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Weeks {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanWeekCreateReq create request
type ResPlanWeekCreateReq struct {
	Year      int                              `json:"year" validate:"required"`
	Month     int                              `json:"month" validate:"required"`
	YearWeek  int                              `json:"year_week" validate:"required"`
	Start     int                              `json:"start" validate:"required"`
	End       int                              `json:"end" validate:"required"`
	IsHoliday *enumor.ResPlanWeekHolidayStatus `json:"is_holiday" validate:"required"`
}

// Validate validate
func (r *ResPlanWeekCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResPlanWeekBatchUpdateReq batch update request
type ResPlanWeekBatchUpdateReq struct {
	Weeks []ResPlanWeekUpdateReq `json:"weeks" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanWeekBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Weeks {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanWeekUpdateReq batch update request
type ResPlanWeekUpdateReq struct {
	ID        string                           `json:"id" validate:"required"`
	Start     int                              `json:"start"`
	End       int                              `json:"end"`
	IsHoliday *enumor.ResPlanWeekHolidayStatus `json:"is_holiday"`
}

// Validate validate
func (r *ResPlanWeekUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResPlanWeekListResult list result
type ResPlanWeekListResult types.ListResult[tablers.ResPlanWeekTable]

// ResPlanWeekListReq list request
type ResPlanWeekListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *ResPlanWeekListReq) Validate() error {
	return r.ListReq.Validate()
}
