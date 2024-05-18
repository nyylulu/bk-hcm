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

package table

import "time"

// OpRecord defines a resource operation record's detail information
type OpRecord struct {
	ID       uint64            `json:"id" bson:"id"`
	HostID   int64             `json:"bk_host_id" bson:"bk_host_id"`
	Labels   map[string]string `json:"labels" bson:"labels"`
	OpType   OpType            `json:"op_type" bson:"op_type"`
	TaskID   uint64            `json:"task_id" bson:"task_id"`
	Phase    OpTaskPhase       `json:"phase" bson:"phase"`
	Message  string            `json:"message" bson:"message"`
	Operator string            `json:"operator" bson:"operator"`
	CreateAt time.Time         `json:"create_at" bson:"create_at"`
	UpdateAt time.Time         `json:"update_at" bson:"update_at"`
}
