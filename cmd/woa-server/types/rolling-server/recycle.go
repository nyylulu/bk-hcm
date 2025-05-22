/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package rollingserver ...
package rollingserver

import (
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/criteria/enumor"
)

const (
	// CalculateMatchSixtyDay 滚服机器回收匹配时间段-60天
	CalculateMatchSixtyDay = 60
	// CalculateMatchNinetyDay 滚服机器回收匹配时间段-90天
	CalculateMatchNinetyDay = 90
)

// RecycleMatchDateRange recycle match date range
type RecycleMatchDateRange struct {
	// Year 记录账单的年份
	Start int `json:"start"`
	// Month 记录账单的月份
	End int `json:"end"`
}

// RecycleHostMatchInfo recycle host match info
type RecycleHostMatchInfo struct {
	*table.RecycleHost    `json:",inline"`
	IsMatched             bool             `json:"is_matched"`
	MatchAppliedIDCoreMap map[string]int64 `json:"match_applied_id_core_map"`
}

// ReturnedRecordInfo returned record info
type ReturnedRecordInfo struct {
	AppliedRecordID string             `json:"applied_record_id"`
	DeviceGroup     string             `json:"device_group"`
	CoreType        enumor.CoreType    `json:"core_type"`
	ReturnedWay     enumor.ReturnedWay `json:"returned_way"`
}

// AppliedRecordKey applied record key
type AppliedRecordKey struct {
	DeviceGroup string          `json:"device_group"`
	CoreType    enumor.CoreType `json:"core_type"`
}

// AppliedRecordInfo applied record info
type AppliedRecordInfo struct {
	SumCpuCore       int64            `json:"sum_cpu_core"`
	AppliedIDCoreMap map[string]int64 `json:"applied_id_core_map"`
	AppliedIDs       []string         `json:"applied_ids"`
}
