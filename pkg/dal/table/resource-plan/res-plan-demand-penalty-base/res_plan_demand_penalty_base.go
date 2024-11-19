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

// Package demandpenaltybase ...
package demandpenaltybase

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// DemandPenaltyBaseColumns defines all the res_plan_demand_penalty_base table's columns.
var DemandPenaltyBaseColumns = utils.MergeColumns(nil, DemandPenaltyBaseColumnDescriptor)

// DemandPenaltyBaseColumnDescriptor is DemandPenaltyBaseTable's column descriptors.
var DemandPenaltyBaseColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "week", NamedC: "week", Type: enumor.Numeric},
	{Column: "year_week", NamedC: "year_week", Type: enumor.Numeric},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "area_name", NamedC: "area_name", Type: enumor.String},
	{Column: "device_family", NamedC: "device_family", Type: enumor.String},
	{Column: "cpu_core", NamedC: "cpu_core", Type: enumor.Numeric},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DemandPenaltyBaseTable is used to save DemandPenaltyBaseTable's data.
type DemandPenaltyBaseTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Year 需求年
	Year int `db:"year" json:"year"`
	// Month 需求月
	Month int `db:"month" json:"month"`
	// Week 需求周
	Week int `db:"week" json:"week"`
	// YearWeek 全年需求周
	YearWeek int `db:"year_week" json:"year_week"`
	// Source 数据来源
	Source enumor.DemandPenaltyBaseSource `db:"source" json:"source"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BkBizName 业务名称
	BkBizName string `db:"bk_biz_name" json:"bk_biz_name" validate:"lte=64"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// OpProductName 运营产品名称
	OpProductName string `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	// PlanProductID 规划产品ID
	PlanProductID int64 `db:"plan_product_id" json:"plan_product_id"`
	// PlanProductName 规划产品名称
	PlanProductName string `db:"plan_product_name" json:"plan_product_name" validate:"lte=64"`
	// VirtualDeptID 虚拟部门ID
	VirtualDeptID int64 `db:"virtual_dept_id" json:"virtual_dept_id"`
	// VirtualDeptName 虚拟部门名称
	VirtualDeptName string `db:"virtual_dept_name" json:"virtual_dept_name" validate:"lte=64"`
	// AreaName 地域名称
	AreaName string `db:"area_name" json:"area_name" validate:"lte=64"`
	// DeviceFamily 机型族
	DeviceFamily string `db:"device_family" json:"device_family" validate:"lte=64"`
	// CpuCore 预测CPU核数
	CpuCore *int64 `db:"cpu_core" json:"cpu_core"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r DemandPenaltyBaseTable) TableName() table.Name {
	return table.ResPlanDemandPenaltyBaseTable
}

// InsertValidate validate resource plan demand on insertion.
func (r DemandPenaltyBaseTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.Year <= 0 {
		return errors.New("year should be > 0")
	}

	if r.Month <= 0 || r.Month > 12 {
		return errors.New("month should be > 0 and <= 12")
	}

	if r.Week <= 0 {
		return errors.New("week should be > 0")
	}

	if r.YearWeek <= 0 {
		return errors.New("year week should be > 0")
	}

	if err := r.Source.Validate(); err != nil {
		return err
	}

	if err := r.bizInsertValidate(); err != nil {
		return err
	}

	if len(r.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	if len(r.DeviceFamily) == 0 {
		return errors.New("device family can not be empty")
	}

	if r.CpuCore == nil {
		return errors.New("cpu core can not be nil")
	}

	return nil
}

func (r DemandPenaltyBaseTable) bizInsertValidate() error {
	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(r.BkBizName) == 0 {
		return errors.New("bk biz name can not be empty")
	}

	if r.OpProductID <= 0 {
		return errors.New("op product id should be > 0")
	}

	if len(r.OpProductName) == 0 {
		return errors.New("op product name can not be empty")
	}

	if r.PlanProductID <= 0 {
		return errors.New("plan product id should be > 0")
	}

	if len(r.PlanProductName) == 0 {
		return errors.New("plan product name can not be empty")
	}

	if r.VirtualDeptID <= 0 {
		return errors.New("virtual dept id should be > 0")
	}

	if len(r.VirtualDeptName) == 0 {
		return errors.New("virtual dept name can not be empty")
	}

	return nil
}

// UpdateValidate validate resource plan demand on update.
func (r DemandPenaltyBaseTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.Year < 0 {
		return errors.New("year should be >= 0")
	}

	if r.Month < 0 || r.Month > 12 {
		return errors.New("month should be >= 0 and <= 12")
	}

	if r.Week < 0 {
		return errors.New("week should be >= 0")
	}

	if r.YearWeek < 0 {
		return errors.New("year week should be >= 0")
	}

	if len(r.Source) > 0 {
		if err := r.Source.Validate(); err != nil {
			return err
		}
	}

	if err := r.bizUpdateValidate(); err != nil {
		return err
	}

	return nil
}

func (r DemandPenaltyBaseTable) bizUpdateValidate() error {
	if r.BkBizID < 0 {
		return errors.New("bk biz id should be >= 0")
	}

	if r.OpProductID < 0 {
		return errors.New("op product id should be >= 0")
	}

	if r.PlanProductID < 0 {
		return errors.New("plan product id should be >= 0")
	}

	if r.VirtualDeptID < 0 {
		return errors.New("virtual dept id should be >= 0")
	}

	return nil
}
