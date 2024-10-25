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
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	cvt "hcm/pkg/tools/converter"
)

// RollingGlobalConfigColumns defines all the rolling_global_config table's columns.
var RollingGlobalConfigColumns = utils.MergeColumns(nil, RollingGlobalConfigColumnDescriptor)

// RollingGlobalConfigColumnDescriptor is RollingGlobalConfigTable's column descriptors.
var RollingGlobalConfigColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "global_quota", NamedC: "global_quota", Type: enumor.Numeric},
	{Column: "biz_quota", NamedC: "biz_quota", Type: enumor.Numeric},
	{Column: "unit_price", NamedC: "unit_price", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// RollingGlobalConfigTable is used to save rolling_global_config table.
type RollingGlobalConfigTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// GlobalQuota 全局总配额
	GlobalQuota *int64 `db:"global_quota" json:"global_quota"`
	// BizQuota 单业务基础配额
	BizQuota *int64 `db:"biz_quota" json:"biz_quota"`
	// UnitPrice 单价
	UnitPrice *types.Decimal `db:"unit_price" json:"unit_price"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the rolling_global_config table's name.
func (r RollingGlobalConfigTable) TableName() table.Name {
	return table.RollingGlobalConfigTable
}

// InsertValidate validate rolling_global_config on insert.
func (r RollingGlobalConfigTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if cvt.PtrToVal(r.GlobalQuota) <= 0 {
		return errors.New("global quota is required")
	}

	if cvt.PtrToVal(r.BizQuota) <= 0 {
		return errors.New("biz quota is required")
	}

	if r.UnitPrice == nil {
		return errors.New("unit price is required")
	}
	if cvt.PtrToVal(r.UnitPrice).IsNegative() {
		return errors.New("unit price should be positive")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate rolling_global_config on update.
func (r RollingGlobalConfigTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if cvt.PtrToVal(r.GlobalQuota) < 0 {
		return errors.New("global quota id should be >= 0")
	}

	if cvt.PtrToVal(r.BizQuota) < 0 {
		return errors.New("biz quota should be >= 0")
	}

	if cvt.PtrToVal(r.UnitPrice).IsNegative() {
		return errors.New("unit price should be positive")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
