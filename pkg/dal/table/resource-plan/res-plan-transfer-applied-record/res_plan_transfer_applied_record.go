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
 * to the current version of the project delivered to anyone in the future.
 */

// Package transferappliedrecord ...
package transferappliedrecord

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	cvt "hcm/pkg/tools/converter"
)

// ResPlanTransferAppliedRecordColumns defines all the resource plan transfer applied record table's columns.
var ResPlanTransferAppliedRecordColumns = utils.MergeColumns(nil, ResPlanTransferAppliedRecordColumnDescriptor)

// ResPlanTransferAppliedRecordColumnDescriptor is ResPlanTransferAppliedRecordTable's column descriptors.
var ResPlanTransferAppliedRecordColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "applied_type", NamedC: "applied_type", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "sub_ticket_id", NamedC: "sub_ticket_id", Type: enumor.String},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "technical_class", NamedC: "technical_class", Type: enumor.String},
	{Column: "obs_project", NamedC: "obs_project", Type: enumor.String},
	{Column: "expected_core", NamedC: "expected_core", Type: enumor.Numeric},
	{Column: "applied_core", NamedC: "applied_core", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanTransferAppliedRecordTable is used to save resource's resource plan transfer applied record information.
type ResPlanTransferAppliedRecordTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// AppliedType 转移类型
	AppliedType enumor.AppliedType `db:"applied_type" json:"applied_type"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// SubTicketID 预测调整子单号
	SubTicketID string `db:"sub_ticket_id" json:"sub_ticket_id" validate:"lte=64"`
	// Year 预测调整时间-年
	Year int `db:"year" json:"year"`
	// TechnicalClass 技术分类
	TechnicalClass string `db:"technical_class" json:"technical_class" validate:"lte=64"`
	// ObsProject 项目类型
	ObsProject enumor.ObsProject `db:"obs_project" json:"obs_project" validate:"lte=64"`
	// ExpectedCore 预期转移的核心数
	ExpectedCore *int64 `db:"expected_core" json:"expected_core" validate:"min=0"`
	// AppliedCore 成功转移的核心数
	AppliedCore *int64 `db:"applied_core" json:"applied_core" validate:"min=0"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at" validate:"isdefault"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at" validate:"isdefault"`
}

// TableName is the recycleRecord's database table name.
func (r ResPlanTransferAppliedRecordTable) TableName() table.Name {
	return table.ResPlanTransferAppliedRecordTable
}

// InsertValidate validate resource plan transfer applied record on insertion.
func (r ResPlanTransferAppliedRecordTable) InsertValidate() error {
	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if err := r.AppliedType.Validate(); err != nil {
		return err
	}

	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(r.SubTicketID) == 0 {
		return errors.New("sub ticket id can not be empty")
	}

	if r.Year <= 0 {
		return errors.New("year should be > 0")
	}

	if len(r.TechnicalClass) == 0 {
		return errors.New("technical class can not be empty")
	}

	if len(r.ObsProject) == 0 {
		return errors.New("obs project can not be empty")
	}

	if err := r.ObsProject.ValidateResPlan(); err != nil {
		return err
	}

	if cvt.PtrToVal(r.ExpectedCore) < 0 {
		return errors.New("expected core should be >= 0")
	}

	if cvt.PtrToVal(r.AppliedCore) < 0 {
		return errors.New("applied core should be >= 0")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate resource plan transfer applied record on update.
func (r ResPlanTransferAppliedRecordTable) UpdateValidate() error {
	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.AppliedType) > 0 {
		if err := r.AppliedType.Validate(); err != nil {
			return err
		}
	}

	if r.BkBizID < 0 {
		return errors.New("bk biz id should be > 0")
	}

	if r.Year < 0 {
		return errors.New("year should be > 0")
	}

	if len(r.ObsProject) > 0 {
		if err := r.ObsProject.ValidateResPlan(); err != nil {
			return err
		}
	}

	if cvt.PtrToVal(r.ExpectedCore) < 0 {
		return errors.New("expected core should be >= 0")
	}

	if cvt.PtrToVal(r.AppliedCore) < 0 {
		return errors.New("applied core should be >= 0")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
