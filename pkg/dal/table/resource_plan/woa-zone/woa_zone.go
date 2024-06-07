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

package woazone

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// WoaZoneColumns defines all the woa zone status table's columns.
var WoaZoneColumns = utils.MergeColumns(nil, WoaZoneColumnDescriptor)

// WoaZoneColumnDescriptor is WoaZoneTable's column descriptors.
var WoaZoneColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "zone_id", NamedC: "zone_id", Type: enumor.String},
	{Column: "zone_name", NamedC: "zone_name", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "area_id", NamedC: "area_id", Type: enumor.String},
	{Column: "area_name", NamedC: "area_name", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// WoaZoneTable is used to save resource's woa zone status information.
type WoaZoneTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// ZoneID 可用区ID
	ZoneID string `db:"zone_id" json:"zone_id" validate:"lte=64"`
	// ZoneName 可用区名称
	ZoneName string `db:"zone_name" json:"zone_name" validate:"lte=64"`
	// RegionID 地区/城市ID
	RegionID string `db:"region_id" json:"region_id" validate:"lte=64"`
	// RegionName 地区/城市名称
	RegionName string `db:"region_name" json:"region_name" validate:"lte=64"`
	// AreaID 地域ID
	AreaID string `db:"area_id" json:"area_id" validate:"lte=64"`
	// AreaName 地域名称
	AreaName string `db:"area_name" json:"area_name" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (t WoaZoneTable) TableName() table.Name {
	return table.WoaZoneTable
}

// InsertValidate validate woa zone status on insertion.
func (t WoaZoneTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(t.ZoneID) == 0 {
		return errors.New("zone id can not be empty")
	}

	if len(t.ZoneName) == 0 {
		return errors.New("zone name can not be empty")
	}

	if len(t.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(t.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(t.AreaID) == 0 {
		return errors.New("area id can not be empty")
	}

	if len(t.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	return nil
}

// UpdateValidate validate woa zone status on update.
func (t WoaZoneTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	return nil
}
