/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package tableapplystat package table apply statistics config.
package tableapplystat

import (
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// CvmApplyOrderStatisticsConfigTableColumns defines all the cvm apply order statistics config table's columns.
var CvmApplyOrderStatisticsConfigTableColumns = utils.MergeColumns(
	nil, CvmApplyOrderStatisticsConfigTableColumnDescriptors)

// CvmApplyOrderStatisticsConfigTableColumnDescriptors is cvm apply order statistics config table column descriptors.
var CvmApplyOrderStatisticsConfigTableColumnDescriptors = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "year_month", NamedC: "year_month", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "sub_order_ids", NamedC: "sub_order_ids", Type: enumor.String},
	{Column: "start_at", NamedC: "start_at", Type: enumor.String},
	{Column: "end_at", NamedC: "end_at", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// CvmApplyOrderStatisticsConfigTable define cvm apply order statistics config table.
type CvmApplyOrderStatisticsConfigTable struct {
	// ID 配置ID
	ID string `db:"id" json:"id"`
	// YearMonth 年月，格式：YYYY-MM
	YearMonth string `db:"year_month" json:"year_month"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// SubOrderIDs 子单号列表，逗号分隔
	SubOrderIDs string `db:"sub_order_ids" json:"sub_order_ids"`
	// StartAt 开始时间，格式：YYYY-MM-DD
	StartAt string `db:"start_at" json:"start_at"`
	// EndAt 结束时间，格式：YYYY-MM-DD
	EndAt string `db:"end_at" json:"end_at"`
	// Memo 备注
	Memo string `db:"memo" json:"memo"`
	// Extension 扩展字段，JSON格式
	Extension types.JsonField `db:"extension" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName return cvm apply order statistics config table name.
func (t CvmApplyOrderStatisticsConfigTable) TableName() table.Name {
	return table.CvmApplyOrderStatisticsConfigTable
}

// Columns return cvm apply order statistics config table columns.
func (t CvmApplyOrderStatisticsConfigTable) Columns() *utils.Columns {
	return CvmApplyOrderStatisticsConfigTableColumns
}

// ColumnDescriptors define cvm apply order statistics config table column descriptor.
func (t CvmApplyOrderStatisticsConfigTable) ColumnDescriptors() utils.ColumnDescriptors {
	return CvmApplyOrderStatisticsConfigTableColumnDescriptors
}

// InsertValidate cvm apply order statistics config table when insert.
func (t CvmApplyOrderStatisticsConfigTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.YearMonth) == 0 {
		return errors.New("year_month is required")
	}

	if t.BkBizID <= 0 {
		return errors.New("bk_biz_id is required and must be greater than 0")
	}

	if len(t.Memo) == 0 {
		return errors.New("memo is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	hasSubIDs := strings.TrimSpace(t.SubOrderIDs) != ""
	hasStart := strings.TrimSpace(t.StartAt) != ""
	hasEnd := strings.TrimSpace(t.EndAt) != ""

	if hasSubIDs && (hasStart || hasEnd) {
		return errors.New("sub_order_ids and time range can not be set together")
	}

	if !hasSubIDs {
		if hasStart != hasEnd {
			return errors.New("start_at and end_at must both be set when using time range")
		}
		if !hasStart {
			return errors.New("either sub_order_ids or time range (start_at & end_at) must be set")
		}
	}

	return nil
}

// UpdateValidate cvm apply order statistics config table when update.
func (t CvmApplyOrderStatisticsConfigTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be updated")
	}

	if len(t.YearMonth) != 0 {
		return errors.New("year_month can not be updated")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if t.BkBizID != 0 && t.BkBizID <= 0 {
		return errors.New("bk_biz_id must be greater than 0")
	}

	if len(t.SubOrderIDs) > 64 {
		return fmt.Errorf("sub_order_ids length can not exceed 64")
	}

	if len(t.StartAt) > 64 {
		return fmt.Errorf("start_at length can not exceed 64")
	}

	if len(t.EndAt) > 64 {
		return fmt.Errorf("end_at length can not exceed 64")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}
