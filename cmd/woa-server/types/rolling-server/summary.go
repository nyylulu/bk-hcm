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

// Package rollingserver ...
package rollingserver

import (
	"fmt"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/times"
)

// RollingServerDateRange is rollinfg server date range.
type RollingServerDateRange struct {
	Start RollingServerDateTimeItem `json:"start" validate:"required"`
	End   RollingServerDateTimeItem `json:"end" validate:"required"`
}

// Validate validate.
func (r *RollingServerDateRange) Validate() error {
	startDate, err := r.Start.Validate()
	if err != nil {
		return err
	}

	endDate, err := r.End.Validate()
	if err != nil {
		return err
	}

	if startDate.After(endDate) {
		return fmt.Errorf("start date should be no later than end date")
	}
	return validator.Validate.Struct(r)
}

// RollingServerDateTimeItem defines resource rolling server datetime item.
type RollingServerDateTimeItem struct {
	Year  int `json:"year" validate:"required,min=2000,max=9999"`
	Month int `json:"month" validate:"required,min=1,max=12"`
	Day   int `json:"day" validate:"required,min=1,max=31"`
}

// Validate validate.
func (r *RollingServerDateTimeItem) Validate() (time.Time, error) {
	if err := validator.Validate.Struct(r); err != nil {
		return time.Time{}, err
	}

	return times.ParseDay(fmt.Sprintf("%d-%02d-%02d", r.Year, r.Month, r.Day))
}

// CpuCoreSummaryReq is cpu core summary request.
type CpuCoreSummaryReq struct {
	RollingServerDateRange `json:",inline"`
	BkBizIDs               []int64            `json:"bk_biz_ids" validate:"omitempty,max=100"`
	OrderIDs               []int              `json:"order_ids" validate:"omitempty,max=100"`
	SubOrderIDs            []string           `json:"suborder_ids" validate:"omitempty,max=100"`
	ReturnedWay            enumor.ReturnedWay `json:"returned_way" validate:"omitempty"`
	AppliedType            enumor.AppliedType `json:"applied_type" validate:"omitempty"`
	InstanceGroup          string             `json:"instance_group" validate:"omitempty"`
	CoreType               *enumor.CoreType   `json:"core_type" validate:"omitempty"`
}

// Validate validate.
func (r *CpuCoreSummaryReq) Validate() error {
	if err := r.RollingServerDateRange.Validate(); err != nil {
		return err
	}
	if len(r.BkBizIDs) > 100 {
		return fmt.Errorf("bk_biz_ids should <= 100")
	}
	if len(r.OrderIDs) > 100 {
		return fmt.Errorf("order_ids should <= 100")
	}
	if len(r.SubOrderIDs) > 100 {
		return fmt.Errorf("suborder_ids should <= 100")
	}
	if err := r.ReturnedWay.Validate(); len(r.ReturnedWay) > 0 && err != nil {
		return err
	}
	if err := r.AppliedType.Validate(); len(r.AppliedType) > 0 && err != nil {
		return err
	}
	return validator.Validate.Struct(r)
}

// CpuCoreSummaryResp is cpu core summary response.
type CpuCoreSummaryResp struct {
	SumDeliveredCore       uint64 `json:"sum_delivered_core"`
	SumReturnedAppliedCore uint64 `json:"sum_returned_applied_core"`
}

// CreateAppliedRecordData create applied record Data
type CreateAppliedRecordData struct {
	BizID       int64              `json:"bk_biz_id"`
	OrderID     uint64             `json:"order_id"`
	SubOrderID  string             `json:"suborder_id"`
	DeviceType  string             `json:"device_type"`
	Count       int                `json:"count"`
	AppliedType enumor.AppliedType `json:"applied_type"`
}

// OldVersionCoreType 之前版本是没有core_type的字段的，这里定义一个空字符串，用于给其他地方兼容之前没有该字段的滚服申请滚服记录
const OldVersionCoreType = enumor.CoreType("")
