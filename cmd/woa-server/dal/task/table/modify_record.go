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

// Package table defines the resource apply order modify record table structure
package table

import (
	"time"

	"hcm/pkg/criteria/enumor"
)

// ModifyRecord defines a resource apply order modify record's detail information
type ModifyRecord struct {
	ID         uint64        `json:"id" bson:"id"`
	SuborderID string        `json:"suborder_id" bson:"suborder_id"`
	User       string        `json:"bk_username" bson:"bk_username"`
	Details    *ModifyDetail `json:"details" bson:"details"`
	CreateAt   time.Time     `json:"create_at" bson:"create_at"`
}

// ModifyDetail apply order modify details with previous and current data
type ModifyDetail struct {
	PreData *ModifyData `json:"pre_data" bson:"pre_data"`
	CurData *ModifyData `json:"cur_data" bson:"cur_data"`
}

// ModifyData apply order modified data
type ModifyData struct {
	TotalNum    uint              `json:"total_num" bson:"total_num"`
	Replicas    uint              `json:"replicas" bson:"replicas"`
	Region      string            `json:"region" bson:"region"`
	Zone        string            `json:"zone" bson:"zone"`
	DeviceType  string            `json:"device_type" bson:"device_type"`
	ImageId     string            `json:"image_id" bson:"image_id"`
	DiskSize    int64             `json:"disk_size" bson:"disk_size"`
	DiskType    enumor.DiskType   `json:"disk_type" bson:"disk_type"`
	NetworkType string            `json:"network_type" bson:"network_type"`
	Vpc         string            `json:"vpc" bson:"vpc"`
	Subnet      string            `json:"subnet" bson:"subnet"`
	SystemDisk  enumor.DiskSpec   `json:"system_disk"`
	DataDisk    []enumor.DiskSpec `json:"data_disk"`
}
