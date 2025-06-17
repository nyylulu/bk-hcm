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

// Package table defines the recycle order's detail information
package table

import (
	"time"
)

// RecycleOrder defines a recycle order's detail information
type RecycleOrder struct {
	OrderID       uint64        `json:"order_id" bson:"order_id"`
	SuborderID    string        `json:"suborder_id" bson:"suborder_id"`
	BizID         int64         `json:"bk_biz_id" bson:"bk_biz_id"`
	BizName       string        `json:"bk_biz_name" bson:"bk_biz_name"`
	User          string        `json:"bk_username" bson:"bk_username"`
	ResourceType  ResourceType  `json:"resource_type" bson:"resource_type"`
	RecycleType   RecycleType   `json:"recycle_type" bson:"recycle_type"`
	ReturnPlan    RetPlanType   `json:"return_plan" bson:"return_plan"`
	SkipConfirm   bool          `json:"skip_confirm" bson:"skip_confirm"`
	Pool          PoolType      `json:"pool_type" bson:"pool_type"`
	CostConcerned bool          `json:"cost_concerned" bson:"cost_concerned"`
	Stage         RecycleStage  `json:"stage" bson:"stage"`
	Status        RecycleStatus `json:"status" bson:"status"`
	Message       string        `json:"message" bson:"message"`
	Handler       string        `json:"handler" bson:"handler"`
	TotalNum      uint          `json:"total_num" bson:"total_num"`
	SuccessNum    uint          `json:"success_num" bson:"success_num"`
	PendingNum    uint          `json:"pending_num" bson:"pending_num"`
	FailedNum     uint          `json:"failed_num" bson:"failed_num"`
	Remark        string        `json:"remark" bson:"remark"`
	CreateAt      time.Time     `json:"create_at" bson:"create_at"`
	UpdateAt      time.Time     `json:"update_at" bson:"update_at"`
	// 提交时间，用于计算耗时
	CommittedAt time.Time `json:"committed_at" bson:"committed_at"`
}
