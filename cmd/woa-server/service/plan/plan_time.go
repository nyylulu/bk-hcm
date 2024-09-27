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

	"hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/times"
)

// GetDemandYearMonthWeek returns the year, month and week of the month based on the input time from a demand
// perspective.
// 需求年月周：当一周内出现跨月（自然月）的情况，则根据哪个自然月的日期更多，划为该“月”的需求年月周。
// 举例：2024-08-26 ～ 2024-09-01分别是周一至周日，而8月的日期比9月的日期多，则该周划为2024-08需求年月的最后一周
func GetDemandYearMonthWeek(t time.Time) plan.DemandYearMonthWeek {
	year, month := GetDemandYearMonth(t)

	// 从输入时间t逐周往前回溯，若前一周的需求年月与当前不同，则需求年月周++
	// 举例：2024年9月8日的需求年月是“2024年9月”，前一周（2024年9月1日）的需求年月是“2024年8月”，两者的需求年月不同，
	//   因此，2024年9月8日的需求年月周为“2024年9月第1周”
	week := 1
	for tPrevWeek := t.AddDate(0, 0, -7); ; tPrevWeek = tPrevWeek.AddDate(0, 0, -7) {
		_, monthOfPrevWeek := GetDemandYearMonth(tPrevWeek)
		if monthOfPrevWeek != month {
			break
		}
		week += 1
	}

	return plan.DemandYearMonthWeek{
		Year:  year,
		Month: month,
		Week:  week,
	}
}

// GetDemandDateRangeInWeek get the date range of a week based on the input time from a demand perspective.
func GetDemandDateRangeInWeek(t time.Time) times.DateRange {
	weekdays := Weekdays(t)

	return times.DateRange{
		Start: weekdays[0].Format(constant.DateLayout),
		End:   weekdays[6].Format(constant.DateLayout),
	}
}

// GetDemandDateRangeInMonth get the date range of a month based on the input time from a demand perspective.
func GetDemandDateRangeInMonth(t time.Time) times.DateRange {
	year, month := GetDemandYearMonth(t)

	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	weekdaysOfFirstDay := Weekdays(firstDay)
	weekdaysOfLastDay := Weekdays(lastDay)

	startDate := weekdaysOfFirstDay[0]
	// 若自然月的第一天的需求年月与输入时间的需求年月不同，则输入时间的需求年月起始时间往后推一周
	if _, monthOfFirstDay := GetDemandYearMonth(firstDay); monthOfFirstDay != month {
		startDate = startDate.AddDate(0, 0, 7)
	}

	endDate := weekdaysOfLastDay[6]
	// 若自然月的最后一天的需求年月与输入时间的需求年月不同，则输入时间的需求年月结束时间往前推一周
	if _, monthOfLastDay := GetDemandYearMonth(lastDay); monthOfLastDay != month {
		endDate = endDate.AddDate(0, 0, -7)
	}

	return times.DateRange{
		Start: startDate.Format(constant.DateLayout),
		End:   endDate.Format(constant.DateLayout),
	}
}

// GetDemandYearMonth returns the year, month based on the input time from a demand perspective.
func GetDemandYearMonth(t time.Time) (year int, month time.Month) {
	weekdays := Weekdays(t)

	// 当一周内出现跨月（自然月）的情况，则根据哪个自然月的日期更多，划为该“月”的年月周
	year = weekdays[3].Year()
	month = weekdays[3].Month()

	return
}

// Weekdays returns the weekdays from Monday to Sunday around the input time.
func Weekdays(t time.Time) (week [7]time.Time) {
	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset = -6
	}

	monday := t.AddDate(0, 0, offset)

	for i := 0; i < 7; i++ {
		week[i] = monday.AddDate(0, 0, i)
	}
	return
}

// getResPlanMonthStartAndEnd 获取当前月的第一天和最后一天
// 当第一天所在周落在本月更多，则将当周的第一天当作本月的第一天，否则将下一周的第一天当作本月的第一天
// 当最后一天所在周落在本月更多，则将当周的最后一天当作本月的最后一天，否则将上一周的最后一天当作本月的最后一天
func getResPlanMonthStartAndEnd(expectTime time.Time) (time.Time, time.Time) {
	// 获取本月的第一天和最后一天
	currentYear, currentMonth, _ := expectTime.Date()
	location := expectTime.Location()
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
	monthStart, monthEnd := getResPlanMonthStartAndEnd(time.Now())

	// 如果期望时间早于本月（即已过期），是否还属于“即将”过期？
	if expectedTime.Before(monthStart) || expectedTime.After(monthEnd) {
		return false
	}
	return true
}

// getDemandStatus 获取需求状态
func getDemandStatus(expectedStart, expectedEnd *time.Time) enumor.DemandStatus {
	monthStart, monthEnd := getResPlanMonthStartAndEnd(time.Now())

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
