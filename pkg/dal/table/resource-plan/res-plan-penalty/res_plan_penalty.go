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

package resplanpenalty

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanPenaltyColumns defines all the resource plan penalty table's columns.
var ResPlanPenaltyColumns = utils.MergeColumns(nil, ResPlanPenaltyColumnDescriptor)

// ResPlanPenaltyColumnDescriptor is ResPlanPenaltyTable's column descriptors.
var ResPlanPenaltyColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "year_month", NamedC: "year_month", Type: enumor.String},
	{Column: "penalty_cpu_core", NamedC: "penalty_cpu_core", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanPenaltyTable is used to save resource's resource plan penalty information.
type ResPlanPenaltyTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// 罚金所属年月，格式为YYYY-MM
	YearMonth string `db:"year_month" json:"year_month" validate:"max=64"`
	// 惩罚核心数
	PenaltyCpuCore float64 `db:"penalty_cpu_core" json:"penalty_cpu_core"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r ResPlanPenaltyTable) TableName() table.Name {
	return table.ResPlanPenaltyTable
}

// InsertValidate validate resource plan penalty on insertion.
func (r ResPlanPenaltyTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.OpProductID <= 0 {
		return errors.New("op product id should be > 0")
	}

	if len(r.YearMonth) == 0 {
		return errors.New("year month can not be empty")
	}

	if r.PenaltyCpuCore < 0 {
		return errors.New("penalty cpu core should be >= 0")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate resource plan penalty on update.
func (r ResPlanPenaltyTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.OpProductID < 0 {
		return errors.New("op product id should be >= 0")
	}

	if r.PenaltyCpuCore < 0 {
		return errors.New("penalty cpu core should be >= 0")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
