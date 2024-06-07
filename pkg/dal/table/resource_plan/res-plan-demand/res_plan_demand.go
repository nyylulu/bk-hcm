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

package resplandemand

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanDemandColumns defines all the resource plan demand table's columns.
var ResPlanDemandColumns = utils.MergeColumns(nil, ResPlanDemandColumnDescriptor)

// ResPlanDemandColumnDescriptor is ResPlanDemandTable's column descriptors.
var ResPlanDemandColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "ticket_id", NamedC: "ticket_id", Type: enumor.String},
	{Column: "obs_project", NamedC: "obs_project", Type: enumor.String},
	{Column: "expect_time", NamedC: "expect_time", Type: enumor.String},
	{Column: "zone_id", NamedC: "zone_id", Type: enumor.String},
	{Column: "zone_name", NamedC: "zone_name", Type: enumor.String},
	{Column: "region_id", NamedC: "region_id", Type: enumor.String},
	{Column: "region_name", NamedC: "region_name", Type: enumor.String},
	{Column: "area_id", NamedC: "area_id", Type: enumor.String},
	{Column: "area_name", NamedC: "area_name", Type: enumor.String},
	{Column: "demand_source", NamedC: "demand_source", Type: enumor.String},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "cvm", NamedC: "cvm", Type: enumor.Json},
	{Column: "cbs", NamedC: "cbs", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanDemandTable is used to save resource's resource plan demand information.
type ResPlanDemandTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// TicketID 单据表唯一ID
	TicketID string `db:"ticket_id" json:"ticket_id" validate:"lte=64"`
	// ObsProject 申请人
	ObsProject enumor.ObsProject `db:"obs_project" json:"obs_project" validate:"lte=64"`
	// ExpectTime 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01
	ExpectTime string `db:"expect_time" json:"expect_time" validate:"lte=64"`
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
	// DemandSource 需求分类/变更原因
	DemandSource string `db:"demand_source" json:"demand_source" validate:"lte=64"`
	// Remark 需求备注
	Remark string `db:"remark" json:"remark" validate:"lte=255"`
	// Cvm cvm信息
	Cvm types.JsonField `db:"cvm" json:"cvm"`
	// Cbs cbs信息
	Cbs types.JsonField `db:"cbs" json:"cbs"`
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
func (r ResPlanDemandTable) TableName() table.Name {
	return table.ResPlanDemandTable
}

// InsertValidate validate resource plan demand on insertion.
func (r ResPlanDemandTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.TicketID) == 0 {
		return errors.New("ticket id can not be empty")
	}

	if err := r.ObsProject.Validate(); err != nil {
		return err
	}

	if len(r.ExpectTime) == 0 {
		return errors.New("expect time can not be empty")
	}

	if len(r.ZoneID) == 0 {
		return errors.New("zone id can not be empty")
	}

	if len(r.ZoneName) == 0 {
		return errors.New("zone name can not be empty")
	}

	if len(r.RegionID) == 0 {
		return errors.New("region id can not be empty")
	}

	if len(r.RegionName) == 0 {
		return errors.New("region name can not be empty")
	}

	if len(r.AreaID) == 0 {
		return errors.New("area id can not be empty")
	}

	if len(r.AreaName) == 0 {
		return errors.New("area name can not be empty")
	}

	if len(r.DemandSource) == 0 {
		return errors.New("demand source can not be empty")
	}

	if len(r.Remark) == 0 {
		return errors.New("remark can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate resource plan demand on update.
func (r ResPlanDemandTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ObsProject) > 0 {
		if err := r.ObsProject.Validate(); err != nil {
			return err
		}
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
