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

// Package demandchangelog ...
package demandchangelog

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// DemandChangelogColumns defines all the res_plan_demand_changelog table's columns.
var DemandChangelogColumns = utils.MergeColumns(nil, DemandChangelogColumnDescriptor)

// DemandChangelogColumnDescriptor is DemandChangelogTable's column descriptors.
var DemandChangelogColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "demand_id", NamedC: "demand_id", Type: enumor.String},
	{Column: "ticket_id", NamedC: "ticket_id", Type: enumor.String},
	{Column: "crp_order_id", NamedC: "crp_order_id", Type: enumor.String},
	{Column: "suborder_id", NamedC: "suborder_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "expect_time", NamedC: "expect_time", Type: enumor.String},
	{Column: "obs_project", NamedC: "obs_project", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "zone_name", NamedC: "zone_name", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "os_change", NamedC: "os_change", Type: enumor.Numeric},
	{Column: "cpu_core_change", NamedC: "cpu_core_change", Type: enumor.Numeric},
	{Column: "memory_change", NamedC: "memory_change", Type: enumor.Numeric},
	{Column: "disk_size_change", NamedC: "disk_size_change", Type: enumor.Numeric},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DemandChangelogTable is used to save DemandChangelogTable's data.
type DemandChangelogTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// DemandID 预测需求表ID
	DemandID string `db:"demand_id" json:"demand_id" validate:"lte=64"`
	// TicketID 预测订单ID
	TicketID string `db:"ticket_id" json:"ticket_id" validate:"lte=64"`
	// CrpOrderID crp订单ID
	CrpOrderID string `db:"crp_order_id" json:"crp_order_id" validate:"lte=64"`
	// SuborderID 主机申领子订单ID
	SuborderID string `db:"suborder_id" json:"suborder_id" validate:"lte=64"`
	// Type 变更类型
	Type enumor.DemandChangelogType `db:"type" json:"type"`
	// ExpectTime 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01
	ExpectTime string `db:"expect_time" json:"expect_time" validate:"lte=16"`
	// ObsProject 项目类型
	ObsProject enumor.ObsProject `db:"obs_project" json:"obs_project" validate:"lte=64"`
	// RegionName 地区/城市名称
	RegionName string `db:"region_name" json:"region_name" validate:"lte=64"`
	// ZoneName 可用区名称
	ZoneName string `db:"zone_name" json:"zone_name" validate:"lte=64"`
	// DeviceType 机型规格
	DeviceType string `db:"device_type" json:"device_type" validate:"lte=64"`
	// OsChange 实例变更数
	OSChange *types.Decimal `db:"os_change" json:"os_change"`
	// CpuCoreChange CPU变更数
	CpuCoreChange *int64 `db:"cpu_core_change" json:"cpu_core_change"`
	// MemoryChange 内存变更数
	MemoryChange *int64 `db:"memory_change" json:"memory_change"`
	// DiskSizeChange 磁盘变更数
	DiskSizeChange *int64 `db:"disk_size_change" json:"disk_size_change"`
	// Remark 备注
	Remark string `db:"remark" json:"remark" validate:"lte=1024"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r DemandChangelogTable) TableName() table.Name {
	return table.ResPlanDemandChangelogTable
}

// InsertValidate validate resource plan demand on insertion.
func (r DemandChangelogTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.DemandID) == 0 {
		return errors.New("demand id can not be empty")
	}

	if len(r.TicketID) == 0 && len(r.SuborderID) == 0 {
		return errors.New("ticket id or suborder id must be provided")
	}

	if err := r.Type.Validate(); err != nil {
		return err
	}

	if err := r.resourceInsertValidate(); err != nil {
		return err
	}

	return nil
}

func (r DemandChangelogTable) resourceInsertValidate() error {
	if r.OSChange == nil {
		return errors.New("os change can not be nil")
	}

	if r.CpuCoreChange == nil {
		return errors.New("cpu core change can not be nil")
	}

	if r.MemoryChange == nil {
		return errors.New("memory change can not be nil")
	}

	if r.DiskSizeChange == nil {
		return errors.New("disk size change can not be nil")
	}

	return nil
}

// UpdateValidate validate resource plan demand on update.
func (r DemandChangelogTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Type) > 0 {
		if err := r.Type.Validate(); err != nil {
			return err
		}
	}

	return nil
}
