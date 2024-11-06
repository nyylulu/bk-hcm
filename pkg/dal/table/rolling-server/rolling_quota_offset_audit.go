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

package rollingserver

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// RollingQuotaOffsetAuditColumns defines all the rolling quota offset audit table's columns.
var RollingQuotaOffsetAuditColumns = utils.MergeColumns(nil, RollingQuotaOffsetAuditColumnDescriptor)

// RollingQuotaOffsetAuditColumnDescriptor is RollingQuotaOffsetAuditTable's column descriptors.
var RollingQuotaOffsetAuditColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "offset_config_id", NamedC: "offset_config_id", Type: enumor.String},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "quota_offset", NamedC: "quota_offset", Type: enumor.Numeric},
	{Column: "rid", NamedC: "rid", Type: enumor.String},
	{Column: "app_code", NamedC: "app_code", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// RollingQuotaOffsetAuditTable is used to save rolling quota offset audit table.
type RollingQuotaOffsetAuditTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// OffsetConfigID 配额偏移配置ID
	OffsetConfigID string `db:"offset_config_id" json:"offset_config_id" validate:"lte=64"`
	// Operator 操作人
	Operator string `db:"operator" json:"operator" validate:"lte=64"`
	// QuotaOffset CPU核心配额偏移量
	QuotaOffset *int64 `db:"quota_offset" json:"quota_offset"`
	// Rid ...
	Rid string `db:"rid" json:"rid" validate:"lte=64"`
	// AppCode ...
	AppCode string `db:"app_code" json:"app_code" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the rolling quota offset audit table's name.
func (r RollingQuotaOffsetAuditTable) TableName() table.Name {
	return table.RollingQuotaOffsetAuditTable
}

// InsertValidate validate rolling quota offset audit on insert.
func (r RollingQuotaOffsetAuditTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.OffsetConfigID) == 0 {
		return errors.New("offset_config_id can not be empty")
	}

	if len(r.Operator) == 0 {
		return errors.New("operator can not be empty")
	}

	if len(r.Rid) == 0 {
		return errors.New("rid can not be empty")
	}

	return nil
}

// UpdateValidate validate rolling quota offset audit on update.
func (r RollingQuotaOffsetAuditTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	return nil
}
