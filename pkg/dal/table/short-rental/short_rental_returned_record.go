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

// Package shortrental ...
package shortrental

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ShortRentalReturnedRecordColumns defines all the short_rental_returned_record table's columns.
var ShortRentalReturnedRecordColumns = utils.MergeColumns(nil, ShortRentalReturnedRecordColumnDescriptor)

// ShortRentalReturnedRecordColumnDescriptor defines all the short_rental_returned_record table's column descriptors.
var ShortRentalReturnedRecordColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "order_id", NamedC: "order_id", Type: enumor.Numeric},
	{Column: "suborder_id", NamedC: "suborder_id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "returned_date", NamedC: "returned_date", Type: enumor.Numeric},
	{Column: "physical_device_family", NamedC: "physical_device_family", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "returned_core", NamedC: "returned_core", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ShortRentalReturnedRecordTable defines the short_rental_returned_record table.
type ShortRentalReturnedRecordTable struct {
	ID                   string                   `db:"id" json:"id" validate:"lte=64"`
	BkBizID              int64                    `db:"bk_biz_id" json:"bk_biz_id"`
	BkBizName            string                   `db:"bk_biz_name" json:"bk_biz_name" validate:"lte=64"`
	OpProductID          int64                    `db:"op_product_id" json:"op_product_id"`
	OpProductName        string                   `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	PlanProductID        int64                    `db:"plan_product_id" json:"plan_product_id"`
	PlanProductName      string                   `db:"plan_product_name" json:"plan_product_name" validate:"lte=64"`
	VirtualDeptID        int64                    `db:"virtual_dept_id" json:"virtual_dept_id"`
	VirtualDeptName      string                   `db:"virtual_dept_name" json:"virtual_dept_name" validate:"lte=64"`
	OrderID              int64                    `db:"order_id" json:"order_id"`
	SuborderID           string                   `db:"suborder_id" json:"suborder_id" validate:"lte=64"`
	Year                 int64                    `db:"year" json:"year"`
	Month                int64                    `db:"month" json:"month"`
	ReturnedDate         int64                    `db:"returned_date" json:"returned_date"`
	PhysicalDeviceFamily string                   `db:"physical_device_family" json:"physical_device_family" validate:"lte=64"`
	RegionID             string                   `db:"region_id" json:"region_id" validate:"lte=64"`
	RegionName           string                   `db:"region_name" json:"region_name" validate:"lte=64"`
	Status               enumor.ShortRentalStatus `db:"status" json:"status" validate:"lte=64"`
	ReturnedCore         *uint64                  `db:"returned_core" json:"returned_core"`
	Creator              string                   `db:"creator" json:"creator" validate:"lte=64"`
	Reviser              string                   `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt            types.Time               `db:"created_at" json:"created_at"`
	UpdatedAt            types.Time               `db:"updated_at" json:"updated_at"`
}

// TableName is the ShortRentalReturnedRecordTable's database table name.
func (s ShortRentalReturnedRecordTable) TableName() table.Name {
	return table.ShortRentalReturnedRecordTable
}

// InsertValidate validates the ShortRentalReturnedRecordTable on insertion.
func (s ShortRentalReturnedRecordTable) InsertValidate() error {
	if len(s.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if s.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(s.BkBizName) == 0 {
		return errors.New("bk biz name can not be empty")
	}

	if len(s.SuborderID) == 0 {
		return errors.New("suborder id can not be empty")
	}

	if len(s.PhysicalDeviceFamily) == 0 {
		return errors.New("physical device family can not be empty")
	}

	if len(s.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(s.Status) == 0 {
		return errors.New("status can not be empty")
	}

	if len(s.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}
	if len(s.Creator) == 0 {
		return errors.New("creator is required")
	}

	return validator.Validate.Struct(s)
}

// UpdateValidate validates the ShortRentalReturnedRecordTable on update.
func (s ShortRentalReturnedRecordTable) UpdateValidate() error {
	if len(s.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(s.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if len(s.Creator) > 0 {
		return errors.New("creator can not be updated")
	}
	return validator.Validate.Struct(s)
}
