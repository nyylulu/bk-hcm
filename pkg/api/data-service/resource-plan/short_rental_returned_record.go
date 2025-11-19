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

package resourceplan

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	shortrental "hcm/pkg/dal/dao/types/short-rental"
)

// ShortRentalReturnedRecord defines the short_rental_returned_record.
type ShortRentalReturnedRecord struct {
	ID                   string `json:"id"`
	BkBizID              int64  `json:"bk_biz_id"`
	BkBizName            string `json:"bk_biz_name"`
	OpProductID          int64  `json:"op_product_id"`
	OpProductName        string `json:"op_product_name"`
	PlanProductID        int64  `json:"plan_product_id"`
	PlanProductName      string `json:"plan_product_name"`
	VirtualDeptID        int64  `json:"virtual_dept_id"`
	VirtualDeptName      string `json:"virtual_dept_name"`
	OrderID              int64  `json:"order_id"`
	SuborderID           string `json:"suborder_id"`
	Year                 int64  `json:"year"`
	Month                int    `json:"month"`
	ReturnedDate         int    `json:"returned_date"`
	PhysicalDeviceFamily string `json:"physical_device_family"`
	RegionID             string `json:"region_id"`
	RegionName           string `json:"region_name"`
	Status               string `json:"status"`
	ReturnedCore         int64  `json:"returned_core"`
	Creator              string `json:"creator"`
	Reviser              string `json:"reviser"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// ShortRentalReturnedRecordCreateReq defines the request for creating a short rental returned record.
type ShortRentalReturnedRecordCreateReq struct {
	BkBizID              int64                    `json:"bk_biz_id" validate:"required"`
	BkBizName            string                   `json:"bk_biz_name" validate:"required"`
	OpProductID          int64                    `json:"op_product_id" validate:"required"`
	OpProductName        string                   `json:"op_product_name" validate:"required"`
	PlanProductID        int64                    `json:"plan_product_id" validate:"required"`
	PlanProductName      string                   `json:"plan_product_name" validate:"required"`
	VirtualDeptID        int64                    `json:"virtual_dept_id" validate:"required"`
	VirtualDeptName      string                   `json:"virtual_dept_name" validate:"required"`
	OrderID              int64                    `json:"order_id" validate:"required"`
	SuborderID           string                   `json:"suborder_id" validate:"required"`
	Year                 int64                    `json:"year" validate:"required,gt=0"`
	Month                int64                    `json:"month" validate:"required,gte=1,lte=12"`
	ReturnedDate         int64                    `json:"returned_date" validate:"required"`
	PhysicalDeviceFamily string                   `json:"physical_device_family" validate:"required"`
	RegionID             string                   `json:"region_id" validate:"required"`
	RegionName           string                   `json:"region_name" validate:"required"`
	Status               enumor.ShortRentalStatus `json:"status" validate:"required"`
	ReturnedCore         *uint64                  `json:"returned_core" validate:"required"`
}

// Validate validates the ShortRentalReturnedRecordCreateReq.
func (req *ShortRentalReturnedRecordCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ShortRentalReturnedRecordBatchCreateReq defines the request for batch creating short rental returned records.
type ShortRentalReturnedRecordBatchCreateReq struct {
	Records []ShortRentalReturnedRecordCreateReq `json:"records"`
}

// Validate validates the ShortRentalReturnedRecordBatchCreateReq.
func (req *ShortRentalReturnedRecordBatchCreateReq) Validate() error {
	if len(req.Records) == 0 {
		return errf.New(errf.InvalidParameter, "records is required")
	}
	for _, record := range req.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ShortRentalReturnedRecordUpdateReq defines the request for updating a short rental returned record.
type ShortRentalReturnedRecordUpdateReq struct {
	ID           string                   `json:"id" validate:"required"`
	Status       enumor.ShortRentalStatus `json:"status,omitempty"`
	ReturnedCore *uint64                  `json:"returned_core,omitempty"`
	Year         *int64                   `json:"year,omitempty" validate:"omitempty,gt=0"`
	Month        *int64                   `json:"month,omitempty" validate:"omitempty,gte=1,lte=12"`
	ReturnedDate *int64                   `json:"returned_date,omitempty" validate:"omitempty,gte=1"`
}

// Validate validates the ShortRentalReturnedRecordUpdateReq.
func (req *ShortRentalReturnedRecordUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ShortRentalReturnedRecordBatchUpdateReq defines the request for batch updating short rental returned records.
type ShortRentalReturnedRecordBatchUpdateReq struct {
	Records []ShortRentalReturnedRecordUpdateReq `json:"records"`
}

// Validate validates the ShortRentalReturnedRecordBatchUpdateReq.
func (req *ShortRentalReturnedRecordBatchUpdateReq) Validate() error {
	if len(req.Records) == 0 {
		return errf.New(errf.InvalidParameter, "records is required")
	}
	for _, record := range req.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ShortRentalReturnedRecordListResult list result
type ShortRentalReturnedRecordListResult = types.ListResult[ShortRentalReturnedRecord]

// ShortRentalReturnedRecordSumReq sum request
type ShortRentalReturnedRecordSumReq struct {
	PhysicalDeviceFamilies []string `json:"physical_device_families"`
	RegionNames            []string `json:"region_names"`
	OpProductIDs           []int64  `json:"op_product_ids"`
	Year                   int64    `json:"year" validate:"required,gt=0"`
	Month                  int64    `json:"month" validate:"required,gte=1,lte=12"`
}

// Validate validate
func (r *ShortRentalReturnedRecordSumReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ShortRentalReturnedRecordSumResult sum result
type ShortRentalReturnedRecordSumResult struct {
	Records []*shortrental.SumShortRentalReturnedRecord `json:"records"`
}
