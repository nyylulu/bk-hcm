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
	"time"
)

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
