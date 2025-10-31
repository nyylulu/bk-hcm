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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-transfer-applied-record"
	cvt "hcm/pkg/tools/converter"
)

// TransferAppliedRecordBatchCreateReq create request
type TransferAppliedRecordBatchCreateReq struct {
	Records []TransferAppliedRecordCreateReq `json:"records" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *TransferAppliedRecordBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, record := range r.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// TransferAppliedRecordCreateReq create request
type TransferAppliedRecordCreateReq struct {
	AppliedType    enumor.AppliedType `json:"applied_type" validate:"required"`
	BkBizID        int64              `json:"bk_biz_id" validate:"required"`
	SubTicketID    string             `json:"sub_ticket_id" validate:"required"`
	Year           int                `json:"year" validate:"required"`
	TechnicalClass string             `json:"technical_class" validate:"required"`
	ObsProject     enumor.ObsProject  `json:"obs_project" validate:"required"`
	ExpectedCore   *int64             `json:"expected_core" validate:"required"`
	AppliedCore    *int64             `json:"applied_core" validate:"required"`
}

// Validate validate
func (r *TransferAppliedRecordCreateReq) Validate() error {
	if cvt.PtrToVal(r.ExpectedCore) < 0 || cvt.PtrToVal(r.AppliedCore) < 0 {
		return errf.New(errf.InvalidParameter, "expected_core and applied_core must be non-negative")
	}
	return validator.Validate.Struct(r)
}

// TransferAppliedRecordBatchUpdateReq batch update request
type TransferAppliedRecordBatchUpdateReq struct {
	Records []TransferAppliedRecordUpdateReq `json:"records" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *TransferAppliedRecordBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, record := range r.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// TransferAppliedRecordUpdateReq update request
type TransferAppliedRecordUpdateReq struct {
	ID             string            `json:"id" validate:"required"`
	TechnicalClass string            `json:"technical_class"`
	ObsProject     enumor.ObsProject `json:"obs_project"`
	ExpectedCore   *int64            `json:"expected_core"`
	AppliedCore    *int64            `json:"applied_core"`
	Reviser        string            `json:"reviser"`
}

// Validate validate
func (r *TransferAppliedRecordUpdateReq) Validate() error {
	if cvt.PtrToVal(r.ExpectedCore) < 0 || cvt.PtrToVal(r.AppliedCore) < 0 {
		return errf.New(errf.InvalidParameter, "expected_core and applied_core must be non-negative")
	}
	return validator.Validate.Struct(r)
}

// ResPlanTransferAppliedRecordListResult list result
type ResPlanTransferAppliedRecordListResult types.ListResult[tablers.ResPlanTransferAppliedRecordTable]

// TransferAppliedRecordListReq list request
type TransferAppliedRecordListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *TransferAppliedRecordListReq) Validate() error {
	return r.ListReq.Validate()
}

// TransferAppliedRecordDeleteReq delete request
type TransferAppliedRecordDeleteReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *TransferAppliedRecordDeleteReq) Validate() error {
	return validator.Validate.Struct(r)
}
