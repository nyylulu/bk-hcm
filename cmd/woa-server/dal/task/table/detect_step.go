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

// Package table defines all the table schema
package table

import (
	"time"
)

// DetectStep defines a recycle detection step's detail information
type DetectStep struct {
	ID         string         `json:"id" bson:"id"`
	OrderID    uint64         `json:"order_id" bson:"order_id"`
	SuborderID string         `json:"suborder_id" bson:"suborder_id"`
	TaskID     string         `json:"task_id" bson:"task_id"`
	StepID     int            `json:"step_id" bson:"step_id"`
	StepName   DetectStepName `json:"step_name" bson:"step_name"`
	StepDesc   string         `json:"step_desc" bson:"step_desc"`
	IP         string         `json:"ip" bson:"ip"`
	User       string         `json:"bk_username" bson:"bk_username"`
	RetryTime  uint32         `json:"retry_time" bson:"retry_time"`
	Status     DetectStatus   `json:"status" bson:"status"`
	Message    string         `json:"message" bson:"message"`
	Skip       int            `json:"skip" bson:"skip"`
	Log        string         `json:"log" bson:"log"`
	StartAt    time.Time      `json:"start_at" bson:"start_at"`
	EndAt      time.Time      `json:"end_at" bson:"end_at"`
	CreateAt   time.Time      `json:"create_at" bson:"create_at"`
	UpdateAt   time.Time      `json:"update_at" bson:"update_at"`
}
