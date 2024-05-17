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

// ZoneLeftIP cvm zone with left ip info
type ZoneLeftIP struct {
	ID              uint64    `json:"id" bson:"id"`
	Region          string    `json:"region" bson:"region"`
	Zone            string    `json:"zone" bson:"zone"`
	ZoneCn          string    `json:"zone_cn" bson:"zone_cn"`
	CmdbZoneName    string    `json:"cmdb_zone_name" bson:"cmdb_zone_name"`
	LeftIPThreshold uint      `json:"left_ip_threshold" bson:"left_ip_threshold"`
	LeftIPNum       uint      `json:"left_ip_num" bson:"left_ip_num"`
	EnableAlarm     bool      `json:"enable_alarm" bson:"enable_alarm"`
	UpdateAt        time.Time `json:"update_at" bson:"update_at"`
}
