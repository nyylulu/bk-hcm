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

// Package rollingserver ...
package rollingserver

import (
	"errors"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
)

// QuotaMonth quota month YYYY-DD
type QuotaMonth string

// Validate whether QuotaMonth is valid.
func (q QuotaMonth) Validate() error {
	_, err := time.Parse(constant.YearMonthLayout, string(q))
	if err != nil {
		return err
	}

	return nil
}

// GetYearMonth get year month.
func (q QuotaMonth) GetYearMonth() (int64, int64, error) {
	var year, month int64
	quotaTime, err := time.Parse(constant.YearMonthLayout, string(q))
	if err != nil {
		return year, month, err
	}

	year = int64(quotaTime.Year())
	month = int64(quotaTime.Month())
	return year, month, nil
}

// GetTime get time.Time
func (q QuotaMonth) GetTime() (time.Time, error) {
	return time.Parse(constant.YearMonthLayout, string(q))
}

// AdjustQuotaOffsetsResp is adjust rolling quota offset configs response.
type AdjustQuotaOffsetsResp struct {
	IDs []string `json:"ids"`
}

// AdjustQuotaOffsetsReq is adjust rolling quota offset configs request.
type AdjustQuotaOffsetsReq struct {
	BkBizIDs    []int64                      `json:"bk_biz_ids" validate:"required,max=100"`
	AdjustType  enumor.QuotaOffsetAdjustType `json:"adjust_type" validate:"required"`
	QuotaOffset int64                        `json:"quota_offset" validate:"required"`
	AdjustMonth AdjustMonthRange             `json:"adjust_month" validate:"required"`
}

// Validate whether AdjustQuotaOffsetsReq is valid.
func (a AdjustQuotaOffsetsReq) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	for _, id := range a.BkBizIDs {
		if id <= 0 {
			return errors.New("bk_biz_id should be > 0")
		}
	}

	if err := a.AdjustType.Validate(); err != nil {
		return err
	}

	if a.QuotaOffset < 0 {
		return errors.New("quota_offset should be >= 0")
	}

	if err := a.AdjustMonth.Validate(); err != nil {
		return err
	}

	return nil
}

// AdjustMonthRange is adjust month range.
type AdjustMonthRange struct {
	Start QuotaMonth `json:"start" validate:"required"`
	End   QuotaMonth `json:"end" validate:"required"`
}

// Validate whether AdjustMonthRange is valid.
func (a AdjustMonthRange) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	startTime, err := a.Start.GetTime()
	if err != nil {
		return err
	}
	endTime, err := a.End.GetTime()
	if err != nil {
		return err
	}

	if startTime.After(endTime) {
		return errors.New("start time cannot be later than end time")
	}

	return nil
}

// CreateBizQuotaConfigsReq is create biz quota configs request.
type CreateBizQuotaConfigsReq struct {
	BkBizIDs   []int64    `json:"bk_biz_ids" validate:"omitempty,max=100"`
	QuotaMonth QuotaMonth `json:"quota_month" validate:"required"`
	Quota      int64      `json:"quota" validate:"required"`
}

// Validate whether CreateBizQuotaConfigsReq is valid.
func (r CreateBizQuotaConfigsReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, id := range r.BkBizIDs {
		if id <= 0 {
			return errors.New("bk biz id should be > 0")
		}
	}

	if err := r.QuotaMonth.Validate(); err != nil {
		return err
	}

	return nil
}

// CreateBizQuotaConfigsResp is create biz quota configs response.
type CreateBizQuotaConfigsResp struct {
	IDs []string `json:"ids"`
}

// GetGlobalQuotaConfigResp get global quota config response.
type GetGlobalQuotaConfigResp struct {
	ID          string        `json:"id"`
	GlobalQuota int64         `json:"global_quota"`
	BizQuota    int64         `json:"biz_quota"`
	UnitPrice   types.Decimal `json:"unit_price"`
	Creator     string        `json:"creator"`
	Reviser     string        `json:"reviser"`
	CreatedAt   types.Time    `json:"created_at"`
	UpdatedAt   types.Time    `json:"updated_at"`
}

// ListBizsWithExistQuotaReq is list biz with exist quota request.
type ListBizsWithExistQuotaReq struct {
	QuotaMonth QuotaMonth `json:"quota_month" validate:"required"`
}

// Validate whether ListBizsWithExistQuota is valid.
func (r ListBizsWithExistQuotaReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.QuotaMonth.Validate(); err != nil {
		return err
	}

	return nil
}

// ListBizsWithExistQuotaResp is list biz with exist quota response.
type ListBizsWithExistQuotaResp struct {
	Details []*ListBizsWithExistQuotaItem `json:"details"`
}

// ListBizsWithExistQuotaItem is list biz with exist quota item.
type ListBizsWithExistQuotaItem struct {
	ID        string `json:"id"`
	BkBizID   int64  `json:"bk_biz_id"`
	BkBizName string `json:"bk_biz_name"`
	Quota     int64  `json:"quota"`
}

// ListBizQuotaConfigsReq is list biz quota configs request.
type ListBizQuotaConfigsReq struct {
	BkBizIDs   []int64                        `json:"bk_biz_ids" validate:"omitempty,max=100"`
	AdjustType []enumor.QuotaOffsetAdjustType `json:"adjust_type" validate:"omitempty"`
	Revisers   []string                       `json:"revisers" validate:"omitempty,max=100"`
	QuotaMonth QuotaMonth                     `json:"quota_month" validate:"required"`
	Page       *core.BasePage                 `json:"page" validate:"required"`
}

// Validate whether ListBizQuotaConfigsReq is valid.
func (r ListBizQuotaConfigsReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, id := range r.BkBizIDs {
		if id <= 0 {
			return errors.New("bk biz id should be > 0")
		}
	}

	for _, t := range r.AdjustType {
		if err := t.Validate(); err != nil {
			return err
		}
	}

	if err := r.QuotaMonth.Validate(); err != nil {
		return err
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// ListBizQuotaConfigsResp is list biz quota configs response.
type ListBizQuotaConfigsResp struct {
	Count   uint64                     `json:"count"`
	Details []*ListBizQuotaConfigsItem `json:"details"`
}

// ListBizQuotaConfigsItem is list biz quota configs item.
type ListBizQuotaConfigsItem struct {
	ID             string                        `json:"id"`
	OffsetConfigID *string                       `json:"offset_config_id"`
	Year           int64                         `json:"year"`
	Month          int64                         `json:"month"`
	BkBizID        int64                         `json:"bk_biz_id"`
	BkBizName      string                        `json:"bk_biz_name"`
	Quota          *int64                        `json:"quota"`
	AdjustType     *enumor.QuotaOffsetAdjustType `json:"adjust_type"`
	QuotaOffset    *uint64                       `json:"quota_offset"`
	Creator        *string                       `json:"creator"`
	Reviser        *string                       `json:"reviser"`
	CreatedAt      *types.Time                   `json:"created_at"`
	UpdatedAt      *types.Time                   `json:"updated_at"`
}

// ListQuotaOffsetsAdjustRecordsReq is list quota offsets adjust records request.
type ListQuotaOffsetsAdjustRecordsReq struct {
	OffsetConfigIds []string       `json:"offset_config_ids" validate:"required,max=100"`
	Page            *core.BasePage `json:"page" validate:"required"`
}

// Validate whether ListQuotaOffsetsAdjustRecordsReq is valid.
func (r ListQuotaOffsetsAdjustRecordsReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// ListQuotaOffsetsAdjustRecordsResp is list quota offsets adjust records response.
type ListQuotaOffsetsAdjustRecordsResp struct {
	Count   uint64                               `json:"count"`
	Details []*ListQuotaOffsetsAdjustRecordsItem `json:"details"`
}

// ListQuotaOffsetsAdjustRecordsItem is list quota offsets adjust records item.
type ListQuotaOffsetsAdjustRecordsItem struct {
	ID             string                       `json:"id"`
	OffsetConfigID string                       `json:"offset_config_id"`
	Operator       string                       `json:"operator"`
	AdjustType     enumor.QuotaOffsetAdjustType `json:"adjust_type"`
	QuotaOffset    uint64                       `json:"quota_offset"`
	CreatedAt      types.Time                   `json:"created_at"`
}

// GetBizBizQuotaConfigsReq is get biz quota configs request.
type GetBizBizQuotaConfigsReq struct {
	QuotaMonth QuotaMonth `json:"quota_month" validate:"required"`
}

// Validate whether GetBizBizQuotaConfigsReq is valid.
func (r GetBizBizQuotaConfigsReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.QuotaMonth.Validate(); err != nil {
		return err
	}

	return nil
}
