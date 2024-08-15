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

import (
	"errors"
	"time"
)

// GradeCfg defines a resource grade config's detail information
type GradeCfg struct {
	ID           uint64       `json:"id" bson:"id"`
	ResourceType ResourceType `json:"resource_type" bson:"resource_type"`
	DeviceType   string       `json:"device_type" bson:"device_type"`
	Grade        int          `json:"grade" bson:"grade"`
	GradeTag     string       `json:"grade_tag" bson:"grade_tag"`
	CreateAt     time.Time    `json:"create_at" bson:"create_at"`
	UpdateAt     time.Time    `json:"update_at" bson:"update_at"`
}

// Validate whether GradeCfg is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GradeCfg) Validate() error {
	if len(param.ResourceType) == 0 {
		return errors.New("resource_type cannot be empty")
	}

	if len(param.DeviceType) == 0 {
		return errors.New("device_type cannot be empty")
	}

	if len(param.GradeTag) == 0 {
		return errors.New("grade_tag cannot be empty")
	}

	return nil
}
