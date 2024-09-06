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

package module

import (
	"errors"

	"hcm/cmd/woa-server/common/util"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// RecycleModuleColumns defines all the recycle module table's columns.
var RecycleModuleColumns = utils.MergeColumns(nil, RecycleModuleColumnDescriptor)

// RecycleModuleColumnDescriptor is RecycleModuleInfo's column descriptors.
var RecycleModuleColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "start_time", NamedC: "start_time", Type: enumor.String},
	{Column: "end_time", NamedC: "end_time", Type: enumor.String},
	{Column: "which_stages", NamedC: "which_stages", Type: enumor.Numeric},
	{Column: "recycle_type", NamedC: "recycle_type", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// RecycleModuleTable is used to save recycle module information.
type RecycleModuleTable struct {
	// ID 自增ID
	ID string `db:"id" json:"id"`
	// Name 模块名称
	Name *string `db:"name" validate:"lte=255" json:"name"`
	// StartTime 开始日期，格式为yyyy-mm-dd
	StartTime *string `db:"start_time" json:"start_time"`
	// EndTime 结束日期，格式为yyyy-mm-dd
	EndTime *string `db:"end_time" json:"end_time"`
	// WhichStages 裁撤模块信息分类，用于前端展示
	WhichStages *int `db:"which_stages" json:"which_stages"`
	// RecycleType 裁撤类型，recycle_type有两个值，0表示全裁，1表示部分裁
	RecycleType *RecycleType `db:"recycle_type" json:"recycle_type"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// RecycleType 裁撤类型
type RecycleType int

const (
	// All 全裁
	All RecycleType = 0
	// Part 部分裁
	Part RecycleType = 1
)

func isRecycleType(val RecycleType) bool {
	if val != All && val != Part {
		return false
	}

	return true
}

// TableName is the recycle module database table name.
func (r RecycleModuleTable) TableName() table.Name {
	return table.RecycleModuleInfo
}

// InsertValidate validate recycle module on insertion.
func (r RecycleModuleTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.StartTime == nil {
		return errors.New("start_time can not be empty")
	}

	if !util.IsDate(*r.StartTime) {
		return errors.New("start_time is not date type")
	}

	if r.EndTime == nil {
		return errors.New("end_time can not be empty")
	}

	if !util.IsDate(*r.EndTime) {
		return errors.New("end_time is not date type")
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if r.RecycleType == nil {
		return errors.New("recycle_type can not be empty")
	}

	if !isRecycleType(*r.RecycleType) {
		return errors.New("recycle_type is invalid")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate recycle module on update.
func (r RecycleModuleTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.StartTime != nil && !util.IsDate(*r.StartTime) {
		return errors.New("start_time is not date type")
	}

	if r.EndTime != nil && !util.IsDate(*r.EndTime) {
		return errors.New("end_time is not date type")
	}

	if r.RecycleType != nil && !isRecycleType(*r.RecycleType) {
		return errors.New("recycle_type is invalid")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
