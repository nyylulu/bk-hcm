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

// Package woadevicetype ...
package woadevicetype

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// WoaDeviceTypePhysicalRelColumns defines all the woa_device_type_physical_rel table's columns.
var WoaDeviceTypePhysicalRelColumns = utils.MergeColumns(nil, WoaDeviceTypePhysicalRelColumnDescriptor)

// WoaDeviceTypePhysicalRelColumnDescriptor defines woa_device_type_physical_rel table's columns.
var WoaDeviceTypePhysicalRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "device_type", NamedC: "device_type", Type: enumor.String},
	{Column: "physical_device_family", NamedC: "physical_device_family", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// WoaDeviceTypePhysicalRelTable woa device type physical rel table.
type WoaDeviceTypePhysicalRelTable struct {
	ID                   string     `db:"id" json:"id" validate:"lte=64"`
	DeviceType           string     `db:"device_type" json:"device_type" validate:"lte=64"`
	PhysicalDeviceFamily string     `db:"physical_device_family" json:"physical_device_family" validate:"lte=64"`
	CreatedAt            types.Time `db:"created_at" json:"created_at"`
	UpdatedAt            types.Time `db:"updated_at" json:"updated_at"`
}

// TableName is the WoaDeviceTypePhysicalRelTable's database table name.
func (w WoaDeviceTypePhysicalRelTable) TableName() table.Name {
	return table.WoaDeviceTypePhysicalRelTable
}

// InsertValidate woa device type physical rel table when insert.
func (w WoaDeviceTypePhysicalRelTable) InsertValidate() error {
	if err := validator.Validate.Struct(w); err != nil {
		return err
	}

	if len(w.ID) == 0 {
		return errors.New("id is required")
	}

	if len(w.DeviceType) == 0 {
		return errors.New("device_type is required")
	}

	if len(w.PhysicalDeviceFamily) == 0 {
		return errors.New("physical_device_family is required")
	}

	return nil
}

// UpdateValidate woa device type physical rel table when update.
func (w WoaDeviceTypePhysicalRelTable) UpdateValidate() error {
	return validator.Validate.Struct(w)
}
