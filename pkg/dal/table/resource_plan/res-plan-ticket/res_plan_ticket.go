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

package resplanticket

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanTicketColumns defines all the resource plan ticket table's columns.
var ResPlanTicketColumns = utils.MergeColumns(nil, ResPlanTicketColumnDescriptor)

// ResPlanTicketColumnDescriptor is ResPlanTicketTable's column descriptors.
var ResPlanTicketColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "applicant", NamedC: "applicant", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "bk_product_id", NamedC: "bk_product_id", Type: enumor.Numeric},
	{Column: "bk_product_name", NamedC: "bk_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "demand_class", NamedC: "demand_class", Type: enumor.String},
	{Column: "os", NamedC: "os", Type: enumor.Numeric},
	{Column: "cpu_core", NamedC: "cpu_core", Type: enumor.Numeric},
	{Column: "memory", NamedC: "memory", Type: enumor.Numeric},
	{Column: "disk_size", NamedC: "disk_size", Type: enumor.Numeric},
	{Column: "remark", NamedC: "remark", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "submitted_at", NamedC: "submitted_at", Type: enumor.Time},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanTicketTable is used to save resource's resource plan ticket information.
type ResPlanTicketTable struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// Applicant 申请人
	Applicant string `db:"applicant" json:"applicant" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BkBizName 业务名称
	BkBizName string `db:"bk_biz_name" json:"bk_biz_name" validate:"lte=64"`
	// BkProductID 运营产品ID
	BkProductID int64 `db:"bk_product_id" json:"bk_product_id"`
	// BkProductName 运营产品名称
	BkProductName string `db:"bk_product_name" json:"bk_product_name" validate:"lte=64"`
	// PlanProductID 规划产品ID
	PlanProductID int64 `db:"plan_product_id" json:"plan_product_id"`
	// PlanProductName 规划产品名称
	PlanProductName string `db:"plan_product_name" json:"plan_product_name" validate:"lte=64"`
	// VirtualDeptID 虚拟部门ID
	VirtualDeptID int64 `db:"virtual_dept_id" json:"virtual_dept_id"`
	// VirtualDeptName 虚拟部门名称
	VirtualDeptName string `db:"virtual_dept_name" json:"virtual_dept_name" validate:"lte=64"`
	// DemandClass 预测的需求类型
	DemandClass string `db:"demand_class" json:"demand_class" validate:"lte=16"`
	// OS OS数，单位：台
	OS int64 `db:"os" json:"os"`
	// CpuCore CPU核心数，单位：台
	CpuCore int64 `db:"cpu_core" json:"cpu_core"`
	// Memory 内存大小，单位：GB
	Memory int64 `db:"memory" json:"memory"`
	// DiskSize 云盘大小，单位：GB
	DiskSize int64 `db:"disk_size" json:"disk_size"`
	// Remark 预测说明，最短20，最长1024
	Remark string `db:"remark" json:"remark" validate:"lte=1024"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// SubmittedAt 提单或改单的时间
	SubmittedAt types.Time `db:"submitted_at" json:"submitted_at"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r ResPlanTicketTable) TableName() table.Name {
	return table.ResPlanTicketTable
}

// InsertValidate validate resource plan ticket on insertion.
func (r ResPlanTicketTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.Applicant) == 0 {
		return errors.New("applicant can not be empty")
	}

	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(r.BkBizName) == 0 {
		return errors.New("bk biz name can not be empty")
	}

	if r.BkProductID <= 0 {
		return errors.New("bk product id should be > 0")
	}

	if len(r.BkProductName) == 0 {
		return errors.New("bk product name can not be empty")
	}

	if r.PlanProductID <= 0 {
		return errors.New("plan product id should be > 0")
	}

	if len(r.PlanProductName) == 0 {
		return errors.New("plan product name can not be empty")
	}

	if r.VirtualDeptID <= 0 {
		return errors.New("virtual dept id should be > 0")
	}

	if len(r.VirtualDeptName) == 0 {
		return errors.New("virtual dept name can not be empty")
	}

	if len(r.DemandClass) == 0 {
		return errors.New("demand class can not be empty")
	}

	if r.OS < 0 {
		return errors.New("os should be >= 0")
	}

	if r.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if r.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	if r.DiskSize < 0 {
		return errors.New("disk size should be >= 0")
	}

	if len(r.Remark) < 20 {
		return errors.New("len remark should be >= 20")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// UpdateValidate validate resource plan ticket on update.
func (r ResPlanTicketTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.BkBizID < 0 {
		return errors.New("bk biz id should be >= 0")
	}

	if r.BkProductID < 0 {
		return errors.New("bk product id should be >= 0")
	}

	if r.PlanProductID < 0 {
		return errors.New("plan product id should be >= 0")
	}

	if r.VirtualDeptID < 0 {
		return errors.New("virtual dept id should be >= 0")
	}

	if r.OS < 0 {
		return errors.New("os should be >= 0")
	}

	if r.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if r.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	if r.DiskSize < 0 {
		return errors.New("disk size should be >= 0")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
