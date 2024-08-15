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

// LaunchTask defines a resource launch task's detail information
type LaunchTask struct {
	ID       uint64          `json:"id" bson:"id"`
	User     string          `json:"bk_username" bson:"bk_username"`
	Status   *PoolTaskStatus `json:"status" bson:"status"`
	CreateAt time.Time       `json:"create_at" bson:"create_at"`
	UpdateAt time.Time       `json:"update_at" bson:"update_at"`
}
