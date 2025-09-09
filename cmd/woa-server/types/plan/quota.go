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

// Package plan ...
package plan

import (
	"errors"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// ListResPlanTransferAppliedRecordReq is list resource plan transfer applied record request.
type ListResPlanTransferAppliedRecordReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate.
func (r ListResPlanTransferAppliedRecordReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.Page != nil {
		if err := r.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ListResPlanTransferQuotaSummaryReq is list resource plan transfer quota summary request.
type ListResPlanTransferQuotaSummaryReq struct {
	Year           int64               `json:"year" validate:"required"`
	BkBizIDs       []int64             `json:"bk_biz_id" validate:"omitempty,max=100"`
	AppliedType    []string            `json:"applied_type" validate:"omitempty,max=100"`
	SubTicketID    []string            `json:"sub_ticket_id" validate:"omitempty,max=100"`
	TechnicalClass []string            `json:"technical_class" validate:"omitempty,max=100"`
	ObsProject     []enumor.ObsProject `json:"obs_project" validate:"omitempty,max=100"`
}

// Validate validate.
func (r ListResPlanTransferQuotaSummaryReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, bkBizID := range r.BkBizIDs {
		if bkBizID <= 0 {
			return errors.New("bk biz id should be > 0")
		}
	}

	for _, obsProject := range r.ObsProject {
		if err := obsProject.ValidateResPlan(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanTransferQuotaSummaryResp is resource plan transfer quota summary response.
type ResPlanTransferQuotaSummaryResp struct {
	UsedQuota   int64 `json:"used_quota"`   // 已使用额度
	RemainQuota int64 `json:"remain_quota"` // 剩余额度
}

// UpdatePlanTransferQuotaConfigsReq is update plan transfer quota config request.
type UpdatePlanTransferQuotaConfigsReq struct {
	Quota      *int64 `json:"quota" validate:"omitempty"`
	AuditQuota *int64 `json:"audit_quota" validate:"omitempty"`
}

// Validate UpdateConfigsReq
func (u *UpdatePlanTransferQuotaConfigsReq) Validate() error {
	return validator.Validate.Struct(u)
}

// TransferQuotaConfig is plan transfer quota config.
type TransferQuotaConfig struct {
	Quota      int64 `json:"quota"`       // 预测转移额度
	AuditQuota int64 `json:"audit_quota"` // 预测转移审批额度
}
