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

package plan

import (
	"time"

	"hcm/pkg/criteria/enumor"
)

// getResPlanMonthStartAndEnd 获取当前月的第一天和最后一天
// 当第一天所在周落在本月更多，则将当周的第一天当作本月的第一天，否则将下一周的第一天当作本月的第一天
// 当最后一天所在周落在本月更多，则将当周的最后一天当作本月的最后一天，否则将上一周的最后一天当作本月的最后一天
func getResPlanMonthStartAndEnd() (time.Time, time.Time) {
	now := time.Now()

	// 获取本月的第一天和最后一天
	currentYear, currentMonth, _ := now.Date()
	location := now.Location()
	firstDay := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
	lastDay := firstDay.AddDate(0, 1, -1)

	firstWeekday := firstDay.Weekday()
	if firstWeekday >= 1 && firstWeekday <= 4 {
		// 如果本月第一天的所在周落在本月更多，则将当周的第一天当作本月的第一天
		firstDay = firstDay.AddDate(0, 0, -int(firstWeekday-1))
	} else {
		// 否则将下周的第一天当作本月的第一天
		firstDay = firstDay.AddDate(0, 0, (8-int(firstWeekday))%7)
	}

	lastWeekday := lastDay.Weekday()
	if lastWeekday >= 4 || lastWeekday == 0 {
		// 如果本月最后一天的所在周落在本月更多，则将当周的最后一天当作本月的最后一天
		lastDay = lastDay.AddDate(0, 0, (7-int(lastWeekday))%7)
	} else {
		// 否则将上周的最后一天当作本月的最后一天
		lastDay = lastDay.AddDate(0, 0, -int(lastWeekday))
	}

	return firstDay, lastDay
}

// isAboutToExpire check whether the cvm and cbs plan is about to expire
func isAboutToExpire(expectedTime *time.Time) bool {
	monthStart, monthEnd := getResPlanMonthStartAndEnd()

	// 如果期望时间早于本月（即已过期），是否还属于“即将”过期？
	if expectedTime.Before(monthStart) || expectedTime.After(monthEnd) {
		return false
	}
	return true
}

// getDemandStatus 获取需求状态
func getDemandStatus(expectedStart, expectedEnd *time.Time) enumor.DemandStatus {
	monthStart, monthEnd := getResPlanMonthStartAndEnd()

	// 未到申领时间
	if expectedStart.After(monthEnd) {
		return enumor.DemandStatusNotReady
	}

	// 已过期
	if expectedEnd.Before(monthStart) {
		return enumor.DemandStatusExpired
	}
	return enumor.DemandStatusCanApply
}
