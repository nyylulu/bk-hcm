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

// Package table pool task
package table

// PoolTaskStatus resource pool task status
type PoolTaskStatus struct {
	Phase      OpTaskPhase `json:"phase" bson:"phase"`
	Message    string      `json:"message" bson:"message"`
	TotalNum   uint        `json:"total_num" bson:"total_num"`
	SuccessNum uint        `json:"success_num" bson:"success_num"`
	PendingNum uint        `json:"pending_num" bson:"pending_num"`
	FailedNum  uint        `json:"failed_num" bson:"failed_num"`
}
