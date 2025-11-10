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

package config

import (
	"errors"
	"strings"
	"time"
)

// CreateApplyOrderStatisticsConfigParam 创建申请单统计配置请求参数
type CreateApplyOrderStatisticsConfigParam struct {
	YearMonth   string   `json:"year_month" validate:"required"`
	SubOrderIDs []string `json:"sub_order_ids"`
	StartAt     string   `json:"start_at"`
	EndAt       string   `json:"end_at"`
	Memo        string   `json:"memo" validate:"required,max=255"`
}

// Validate 验证创建申请单统计配置请求参数
func (c *CreateApplyOrderStatisticsConfigParam) Validate() error {
	if len(c.YearMonth) == 0 {
		return errors.New("year_month is required")
	}

	// 验证年月格式 YYYY-MM
	parts := strings.Split(c.YearMonth, "-")
	if len(parts) != 2 {
		return errors.New("year_month format must be YYYY-MM")
	}

	if len(c.Memo) == 0 {
		return errors.New("memo is required")
	}

	if len(c.Memo) > 255 {
		return errors.New("memo length must not exceed 255 characters")
	}

	// 子单号和时间段必须二选一
	hasSubOrderIDs := len(c.SubOrderIDs) > 0
	hasTimeRange := len(c.StartAt) > 0 && len(c.EndAt) > 0

	if !hasSubOrderIDs && !hasTimeRange {
		return errors.New("sub_order_ids and time range (start_at, end_at) must choose one")
	}

	if hasSubOrderIDs && hasTimeRange {
		return errors.New("sub_order_ids and time range (start_at, end_at) can not both be set")
	}

	// 验证子单号数量
	if hasSubOrderIDs && len(c.SubOrderIDs) > 100 {
		return errors.New("sub_order_ids length must not exceed 100")
	}

	// 验证时间段
	if hasTimeRange {
		if len(c.StartAt) == 0 {
			return errors.New("start_at is required when using time range")
		}
		if len(c.EndAt) == 0 {
			return errors.New("end_at is required when using time range")
		}
		// 验证日期格式 YYYY-MM-DD
		startParts := strings.Split(c.StartAt, "-")
		if len(startParts) != 3 {
			return errors.New("start_at format must be YYYY-MM-DD")
		}
		endParts := strings.Split(c.EndAt, "-")
		if len(endParts) != 3 {
			return errors.New("end_at format must be YYYY-MM-DD")
		}
	}

	return nil
}

// CreateApplyOrderStatisticsConfigResult 创建申请单统计配置响应结果
type CreateApplyOrderStatisticsConfigResult struct {
	ID string `json:"id"`
}

// CvmApplyOrderStatisticsConfig 申请单统计配置实体
type CvmApplyOrderStatisticsConfig struct {
	ID         string    `json:"id" bson:"id"`
	YearMonth  string    `json:"year_month" bson:"year_month"`
	BkBizID    int64     `json:"bk_biz_id" bson:"bk_biz_id"`
	SubOrderID string    `json:"sub_order_id" bson:"sub_order_id"`
	StartAt    string    `json:"start_at" bson:"start_at"`
	EndAt      string    `json:"end_at" bson:"end_at"`
	Memo       string    `json:"memo" bson:"memo"`
	Extension  string    `json:"extension" bson:"extension"`
	Creator    string    `json:"creator" bson:"creator"`
	Reviser    string    `json:"reviser" bson:"reviser"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}
