/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package table defines the return task table structure
package table

import (
	"time"
)

// ReturnTask defines a device return task's detail information
type ReturnTask struct {
	OrderID            uint64       `json:"order_id" bson:"order_id"`
	SuborderID         string       `json:"suborder_id" bson:"suborder_id"`
	ResourceType       ResourceType `json:"resource_type"`
	RecycleType        RecycleType  `json:"recycle_type" bson:"recycle_type"`
	ReturnPlan         RetPlanType  `json:"return_plan" bson:"return_plan"`
	SkipConfirm        bool         `json:"skip_confirm" bson:"skip_confirm"`
	Status             ReturnStatus `json:"status" bson:"status"`
	Message            string       `json:"message" bson:"message"`
	TaskID             string       `json:"task_id" bson:"task_id"`
	TaskLink           string       `json:"task_link" bson:"task_link"`
	ReturnForecast     bool         `json:"return_forecast"`
	ReturnForecastTime string       `json:"return_forecast_time"`
	User               string       `json:"user" bson:"user"`
	BkBizID            int64        `json:"bk_biz_id" bson:"bk_biz_id"`
	CreateAt           time.Time    `json:"create_at" bson:"create_at"`
	UpdateAt           time.Time    `json:"update_at" bson:"update_at"`
}
