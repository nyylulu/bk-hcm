/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package greenchannel ...
package greenchannel

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/times"
)

// DateRange is green channel date range.
type DateRange struct {
	Start times.DateTimeItem `json:"start" validate:"required"`
	End   times.DateTimeItem `json:"end" validate:"required"`
}

// Validate validate.
func (r *DateRange) Validate() error {
	if err := r.Start.Validate(); err != nil {
		return err
	}
	startDate := r.Start.GetTime()

	if err := r.End.Validate(); err != nil {
		return err
	}
	endDate := r.End.GetTime()

	if startDate.After(endDate) {
		return fmt.Errorf("start date should be no later than end date")
	}

	return validator.Validate.Struct(r)
}

// CpuCoreSummaryReq is cpu core summary request.
type CpuCoreSummaryReq struct {
	DateRange `json:",inline"`
	BkBizIDs  []int64 `json:"bk_biz_ids" validate:"omitempty,max=100"`
}

// Validate validate.
func (r *CpuCoreSummaryReq) Validate() error {
	if err := r.DateRange.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(r)
}

// CpuCoreSummaryResp is cpu core summary response.
type CpuCoreSummaryResp struct {
	SumDeliveredCore uint64 `json:"sum_delivered_core"`
}

// AggregateCount is aggregate count.
type AggregateCount struct {
	Count uint64 `bson:"count"`
}

// StatisticalRecordReq is statistical record request.
type StatisticalRecordReq struct {
	DateRange `json:",inline"`
	BkBizIDs  []int64        `json:"bk_biz_ids" validate:"omitempty,max=100,dive,gt=0"`
	Page      *core.BasePage `json:"page" validate:"required"`
}

// Validate validate.
func (s *StatisticalRecordReq) Validate() error {
	if err := s.DateRange.Validate(); err != nil {
		return err
	}

	if err := s.Page.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(s)
}

// StatisticalRecordResp is statistical record response.
type StatisticalRecordResp = core.ListResultT[StatisticalRecordItem]

// StatisticalRecordItem is statistical record item.
type StatisticalRecordItem struct {
	BizID            int64  `json:"bk_biz_id" bson:"bk_biz_id"`
	OrderCount       uint64 `json:"order_count" bson:"order_count"`
	SumDeliveredCore uint64 `json:"sum_delivered_core" bson:"sum_delivered_core"`
	SumAppliedCore   uint64 `json:"sum_applied_core" bson:"sum_applied_core"`
}

// GetConfigsResp is get config response.
type GetConfigsResp struct {
	Config `json:",inline"`
}

// UpdateConfigsReq is update config request.
type UpdateConfigsReq struct {
	BizQuota       *int64 `json:"biz_quota" validate:"omitempty"`
	IEGQuota       *int64 `json:"ieg_quota" validate:"omitempty"`
	AuditThreshold *int64 `json:"audit_threshold" validate:"omitempty"`
}

// Validate UpdateConfigsReq
func (u *UpdateConfigsReq) Validate() error {
	return validator.Validate.Struct(u)
}

// Config is config.
type Config struct {
	BizQuota       int64          `json:"biz_quota"`
	IEGQuota       int64          `json:"ieg_quota"`
	AuditThreshold int64          `json:"audit_threshold"`
	CvmApplyConfig CvmApplyConfig `json:"cvm_apply_config"`
}

// CvmApplyConfig is cvm apply config.
type CvmApplyConfig struct {
	Enabled      bool     `json:"enabled"`
	DeviceGroups []string `json:"device_groups"` // 机型族
	CpuMaxLimit  int64    `json:"cpu_max_limit"` // 核心数
}
