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

// Package table defines the table structure
package table

import "time"

// RecallOrder defines a resource recall order's detail information
type RecallOrder struct {
	ID            uint64          `json:"id" bson:"id"`
	User          string          `json:"bk_username" bson:"bk_username"`
	Spec          *RecallTaskSpec `json:"spec" bson:"spec"`
	RecyclePolicy *RecyclePolicy  `json:"recycle_policy" bson:"recycle_policy"`
	Status        *PoolTaskStatus `json:"status" bson:"status"`
	CreateAt      time.Time       `json:"create_at" bson:"create_at"`
	UpdateAt      time.Time       `json:"update_at" bson:"update_at"`
}

// RecyclePolicy resource recycle policy
type RecyclePolicy struct {
	ImageID string `json:"image_id" bson:"image_id"`
	OsType  string `json:"os_type" bson:"os_type"`
}
