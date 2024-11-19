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

// Package resourceplan ...
package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-penalty-base"
)

// DemandPenaltyBaseCreateReq create request
type DemandPenaltyBaseCreateReq struct {
	PenaltyBases []DemandPenaltyBaseCreate `json:"penalty_bases" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandPenaltyBaseCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.PenaltyBases {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandPenaltyBaseCreate create request
type DemandPenaltyBaseCreate struct {
	Year            int                            `json:"year" validate:"required"`
	Month           int                            `json:"month" validate:"required"`
	Week            int                            `json:"week" validate:"required"`
	YearWeek        int                            `json:"year_week" validate:"required"`
	Source          enumor.DemandPenaltyBaseSource `json:"source" validate:"required"`
	BkBizID         int64                          `json:"bk_biz_id" validate:"required"`
	BkBizName       string                         `json:"bk_biz_name" validate:"required"`
	OpProductID     int64                          `json:"op_product_id" validate:"required"`
	OpProductName   string                         `json:"op_product_name" validate:"required"`
	PlanProductID   int64                          `json:"plan_product_id" validate:"required"`
	PlanProductName string                         `json:"plan_product_name" validate:"required"`
	VirtualDeptID   int64                          `json:"virtual_dept_id" validate:"required"`
	VirtualDeptName string                         `json:"virtual_dept_name" validate:"required"`
	AreaName        string                         `json:"area_name" validate:"required"`
	DeviceFamily    string                         `json:"device_family" validate:"required"`
	CpuCore         *int64                         `json:"cpu_core" validate:"required"`
}

// Validate validate
func (r *DemandPenaltyBaseCreate) Validate() error {
	return validator.Validate.Struct(r)
}

// DemandPenaltyBaseBatchUpdateReq batch update request
type DemandPenaltyBaseBatchUpdateReq struct {
	PenaltyBases []DemandPenaltyBaseUpdateReq `json:"penalty_bases" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandPenaltyBaseBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.PenaltyBases {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandPenaltyBaseUpdateReq batch update request
type DemandPenaltyBaseUpdateReq struct {
	ID              string `json:"id" validate:"required"`
	OpProductID     int64  `json:"op_product_id"`
	OpProductName   string `json:"op_product_name"`
	PlanProductID   int64  `json:"plan_product_id"`
	PlanProductName string `json:"plan_product_name"`
	VirtualDeptID   int64  `json:"virtual_dept_id"`
	VirtualDeptName string `json:"virtual_dept_name"`
	CpuCore         *int64 `json:"cpu_core"`
}

// Validate validate
func (r *DemandPenaltyBaseUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// DemandPenaltyBaseListResult list result
type DemandPenaltyBaseListResult types.ListResult[tablers.DemandPenaltyBaseTable]

// DemandPenaltyBaseListReq list request
type DemandPenaltyBaseListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *DemandPenaltyBaseListReq) Validate() error {
	return r.ListReq.Validate()
}
