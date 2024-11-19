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
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-changelog"

	"github.com/shopspring/decimal"
)

// DemandChangelogCreateReq create request
type DemandChangelogCreateReq struct {
	Changelogs []DemandChangelogCreate `json:"changelogs" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandChangelogCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Changelogs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandChangelogCreate create request
type DemandChangelogCreate struct {
	DemandID       string                     `json:"demand_id" validate:"required"`
	TicketID       string                     `json:"ticket_id" validate:"omitempty"`
	CrpOrderID     string                     `json:"crp_order_id" validate:"omitempty"`
	SuborderID     string                     `json:"suborder_id" validate:"omitempty"`
	Type           enumor.DemandChangelogType `json:"type" validate:"required"`
	ExpectTime     string                     `json:"expect_time" validate:"required"`
	ObsProject     enumor.ObsProject          `json:"obs_project" validate:"required"`
	RegionName     string                     `json:"region_name" validate:"required"`
	ZoneName       string                     `json:"zone_name" validate:"omitempty"`
	DeviceType     string                     `json:"device_type" validate:"required"`
	OSChange       *decimal.Decimal           `json:"os_change" validate:"required"`
	CpuCoreChange  *int64                     `json:"cpu_core_change" validate:"required"`
	MemoryChange   *int64                     `json:"memory_change" validate:"required"`
	DiskSizeChange *int64                     `json:"disk_size_change" validate:"required"`
	Remark         string                     `json:"remark" validate:"omitempty"`
}

// Validate validate
func (r *DemandChangelogCreate) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, r.ExpectTime)
	if err != nil {
		return err
	}

	return nil
}

// DemandChangelogBatchUpdateReq batch update request
type DemandChangelogBatchUpdateReq struct {
	Changelogs []DemandChangelogUpdateReq `json:"changelogs" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandChangelogBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Changelogs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandChangelogUpdateReq batch update request
type DemandChangelogUpdateReq struct {
	ID             string                     `json:"id" validate:"required"`
	DemandID       string                     `json:"demand_id"`
	TicketID       string                     `json:"ticket_id"`
	CrpOrderID     string                     `json:"crp_order_id"`
	SuborderID     string                     `json:"suborder_id"`
	Type           enumor.DemandChangelogType `json:"type"`
	ExpectTime     string                     `json:"expect_time"`
	ObsProject     enumor.ObsProject          `json:"obs_project"`
	RegionName     string                     `json:"region_name"`
	ZoneName       string                     `json:"zone_name"`
	DeviceType     string                     `json:"device_type"`
	OSChange       *decimal.Decimal           `json:"os_change"`
	CpuCoreChange  *int64                     `json:"cpu_core_change"`
	MemoryChange   *int64                     `json:"memory_change"`
	DiskSizeChange *int64                     `json:"disk_size_change"`
	Remark         string                     `json:"remark"`
}

// Validate validate
func (r *DemandChangelogUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, r.ExpectTime)
	if err != nil {
		return err
	}

	return nil
}

// DemandChangelogListResult list result
type DemandChangelogListResult types.ListResult[tablers.DemandChangelogTable]

// DemandChangelogListReq list request
type DemandChangelogListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *DemandChangelogListReq) Validate() error {
	return r.ListReq.Validate()
}
