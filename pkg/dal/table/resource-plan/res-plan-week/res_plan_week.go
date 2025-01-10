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

// Package resplanweek ...
package resplanweek

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanWeekColumns defines all the res_plan_week table's columns.
var ResPlanWeekColumns = utils.MergeColumns(nil, ResPlanWeekColumnDescriptor)

// ResPlanWeekColumnDescriptor is ResPlanWeekTable's column descriptors.
var ResPlanWeekColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "year_week", NamedC: "year_week", Type: enumor.Numeric},
	{Column: "start", NamedC: "start", Type: enumor.Numeric},
	{Column: "end", NamedC: "end", Type: enumor.Numeric},
	{Column: "is_holiday", NamedC: "is_holiday", Type: enumor.Numeric},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanWeekTable is used to save ResPlanWeekTable's data.
type ResPlanWeekTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Year 需求所属年
	Year int `db:"year" json:"year"`
	// Month 需求所属月
	Month int `db:"month" json:"month"`
	// YearWeek 需求所属周（全年范围内）
	YearWeek int `db:"year_week" json:"year_week"`
	// Start 期望到货时间范围，YYYYMMDD
	Start int `db:"start" json:"start"`
	// End 期望到货时间范围，YYYYMMDD
	End int `db:"end" json:"end"`
	// IsHoliday 是否节假日
	IsHoliday *enumor.ResPlanWeekHolidayStatus `db:"is_holiday" json:"is_holiday"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the ResPlanWeekTable's database table name.
func (r ResPlanWeekTable) TableName() table.Name {
	return table.ResPlanWeekTable
}

// InsertValidate validate resource plan week on insertion.
func (r ResPlanWeekTable) InsertValidate() error {
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

	if r.YearWeek <= 0 {
		return errors.New("year_week should be > 0")
	}

	if r.Start <= 0 || r.End <= 0 {
		return errors.New("start and end can not be empty")
	}

	if r.Start >= r.End {
		return errors.New("start should be < end")
	}

	if err := r.IsHoliday.Validate(); err != nil {
		return err
	}

	return nil
}

// UpdateValidate validate resource plan week on update.
func (r ResPlanWeekTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.Year < 0 {
		return errors.New("year should be >= 0")
	}

	if r.Month < 0 || r.Month > 12 {
		return errors.New("month should be >= 0 and <= 12")
	}

	if r.YearWeek < 0 {
		return errors.New("year_week should be >= 0")
	}

	if r.Start < 0 || r.End < 0 {
		return errors.New("start and end should be >= 0")
	}

	if r.Start > 0 && r.End > 0 {
		if r.Start >= r.End {
			return errors.New("start should be < end")
		}
	}

	if r.IsHoliday != nil {
		if err := r.IsHoliday.Validate(); err != nil {
			return err
		}
	}

	return nil
}
