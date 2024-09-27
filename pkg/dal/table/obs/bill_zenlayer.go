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

// Package obs ...
package obs

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// OBSBillItemZenlayerColumns defines account_bill_summary's columns.
var OBSBillItemZenlayerColumns = utils.MergeColumns(nil, OBSBillItemZenlayerColumnDescriptor)

// OBSBillItemZenlayerColumnDescriptor is AwsBill's column descriptors.
var OBSBillItemZenlayerColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "main_account_id", NamedC: "main_account_id", Type: enumor.String},
	{Column: "bill_year", NamedC: "bill_year", Type: enumor.Numeric},
	{Column: "bill_month", NamedC: "bill_month", Type: enumor.Numeric},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "set_index", NamedC: "set_index", Type: enumor.String},

	{Column: "yearMonth", NamedC: "yearMonth", Type: enumor.Numeric},
	{Column: "billing_main_account_id", NamedC: "billing_main_account_id", Type: enumor.String},
	{Column: "billing_sub_account_id", NamedC: "billing_sub_account_id", Type: enumor.String},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "currency", NamedC: "currency", Type: enumor.String},
	{Column: "rate", NamedC: "rate", Type: enumor.Numeric},
	{Column: "productid", NamedC: "productid", Type: enumor.Numeric},
	{Column: "real_cost", NamedC: "real_cost", Type: enumor.Numeric},
	{Column: "city", NamedC: "city", Type: enumor.String},
	{Column: "contract_period", NamedC: "contract_period", Type: enumor.String},
	{Column: "description", NamedC: "description", Type: enumor.String},
	{Column: "group_uid", NamedC: "group_uid", Type: enumor.String},
	{Column: "pay_amount", NamedC: "pay_amount", Type: enumor.Numeric},
	{Column: "pay_content", NamedC: "pay_content", Type: enumor.String},
	{Column: "price", NamedC: "price", Type: enumor.Numeric},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "uid", NamedC: "uid", Type: enumor.String},
	{Column: "zen_order_no", NamedC: "zen_order_no", Type: enumor.String},
	{Column: "accept_amount", NamedC: "accept_amount", Type: enumor.Numeric},
	{Column: "bill_monthly", NamedC: "bill_monthly", Type: enumor.String},
	{Column: "cpu", NamedC: "cpu", Type: enumor.String},
	{Column: "disk", NamedC: "disk", Type: enumor.String},
	{Column: "memory", NamedC: "memory", Type: enumor.String},
}

// OBSBillItemZenlayer Zenlayer bill item
type OBSBillItemZenlayer struct {
	ID            string `json:"id" db:"id"`
	Vendor        string `json:"vendor" db:"vendor"`
	MainAccountID string `json:"main_account_id" db:"main_account_id"`
	BillYear      int64  `json:"bill_year" db:"bill_year"`
	BillMonth     int64  `json:"bill_month" db:"bill_month"`
	SetIndex      string `json:"set_index" db:"set_index"`

	YearMonth            int32          `json:"year_month" db:"yearMonth"`
	BillingMainAccountId string         `json:"billing_main_account_id" db:"billing_main_account_id"`
	BillingSubAccountId  string         `json:"billing_sub_account_id" db:"billing_sub_account_id"`
	Cost                 *types.Decimal `json:"cost" db:"cost"`
	Currency             string         `json:"currency" db:"currency"`
	Rate                 float64        `json:"rate" db:"rate"`
	ProductID            int32          `json:"productid" db:"productid"`
	RealCost             *types.Decimal `json:"real_cost" db:"real_cost"`
	City                 string         `json:"city" db:"city"`
	ContractPeriod       string         `json:"contract_period" db:"contract_period"`
	Description          string         `json:"description" db:"description"`
	GroupUid             string         `json:"group_uid" db:"group_uid"`
	PayAmount            *types.Decimal `json:"pay_amount" db:"pay_amount"`
	PayContent           string         `json:"pay_content" db:"pay_content"`
	Price                *types.Decimal `json:"price" db:"price"`
	Type                 string         `json:"type" db:"type"`
	UID                  string         `json:"uid" db:"uid"`
	ZenOrderNo           string         `json:"zen_order_no" db:"zen_order_no"`
	AcceptAmount         *types.Decimal `json:"accept_amount" db:"accept_amount"`
	BillMonthly          string         `json:"bill_monthly" db:"bill_monthly"`
	Cpu                  string         `json:"cpu" db:"cpu"`
	Disk                 string         `json:"disk" db:"disk"`
	Memory               string         `json:"memory" db:"memory"`
}

// TableName 返回月度汇总账单表名
func (bih *OBSBillItemZenlayer) TableName() table.Name {
	return table.OBSBillZenlayerItemTable
}

// InsertValidate validate account bill summary on insert
func (bih *OBSBillItemZenlayer) InsertValidate() error {
	if len(bih.ID) == 0 {
		return errors.New("id is required")
	}
	if err := validator.Validate.Struct(bih); err != nil {
		return err
	}
	return nil
}
