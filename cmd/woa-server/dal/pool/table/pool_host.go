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

// PoolHost defines a resource pool host's detail information
type PoolHost struct {
	HostID   int64             `json:"bk_host_id" bson:"bk_host_id"`
	Labels   map[string]string `json:"labels" bson:"labels"`
	Status   *PoolHostStatus   `json:"status" bson:"status"`
	CreateAt time.Time         `json:"create_at" bson:"create_at"`
	UpdateAt time.Time         `json:"update_at" bson:"update_at"`
}

// PoolHostStatus resource pool host's status
type PoolHostStatus struct {
	Phase      PoolHostPhase `json:"phase" bson:"phase"`
	LaunchID   uint64        `json:"launch_id" bson:"launch_id"`
	LaunchTime time.Time     `json:"launch_time" bson:"launch_time"`
	RecallID   uint64        `json:"recall_id" bson:"recall_id"`
	RecallTime time.Time     `json:"recall_time" bson:"recall_time"`
	DrawTime   time.Time     `json:"draw_time" bson:"draw_time"`
	ReturnTime time.Time     `json:"return_time" bson:"return_time"`
}
