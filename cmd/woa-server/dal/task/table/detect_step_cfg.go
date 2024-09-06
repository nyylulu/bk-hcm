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

// Package table defines all table names
package table

import (
	"time"
)

// DetectStepCfg defines a recycle detection step config's detail information
type DetectStepCfg struct {
	ID          int64          `json:"id" bson:"id"`
	Sequence    int            `json:"sequence" bson:"sequence"`
	Name        DetectStepName `json:"name" bson:"name"`
	Description string         `json:"description" bson:"description"`
	Enable      bool           `json:"enable" bson:"enable"`
	Retry       int            `json:"retry" bson:"retry"`
	CreateAt    time.Time      `json:"create_at" bson:"create_at"`
	UpdateAt    time.Time      `json:"update_at" bson:"update_at"`
}
