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

package demandtime

import (
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/times"
)

// DemandYearMonthWeek is the year, month and week of the month from a demand perspective.
type DemandYearMonthWeek struct {
	Year  int        `json:"year"`
	Month time.Month `json:"month"`
	// Week 需求年月周
	Week int `json:"week"`
	// YearWeek 需求全年周
	YearWeek int `json:"year_week"`
}

// GetDemandYearMonthWeek returns the year, month and week of the month based on the input time from a demand
// perspective.
// 需求年月周：当一周内出现跨月（自然月）的情况，则根据该周周一所属月划为该月的需求年月周。
// 举例1：2024-08-26 ～ 2024-09-01分别是周一至周日，则该周划为2024-08需求年月的最后一周
// 举例2：2024-09-30 ～ 2024-10-06分别是周一至周日，则该周划为2024-09需求年月的最后一周
func GetDemandYearMonthWeek(t time.Time) DemandYearMonthWeek {
	year, month := GetDemandYearMonth(t)

	// 从输入时间t逐周往前回溯：
	// 若前一周的需求年月与当前不同，则需求年月周++
	// 若前一周的需求年与当前不同，则需求全年周++
	// 举例：2024年9月8日的需求年月是“2024年9月”，前一周（2024年9月1日）的需求年月是“2024年8月”，两者的需求年月不同，
	//   因此，2024年9月8日的需求年月周为“2024年9月第1周”
	week := 1
	yearWeek := 1
	for tPrevWeek := t.AddDate(0, 0, -7); ; tPrevWeek = tPrevWeek.AddDate(0, 0, -7) {
		yearOfPrevWeek, monthOfPrevWeek := GetDemandYearMonth(tPrevWeek)
		if yearOfPrevWeek != year {
			break
		}

		if monthOfPrevWeek == month {
			week += 1
		}

		yearWeek += 1
	}

	return DemandYearMonthWeek{
		Year:     year,
		Month:    month,
		Week:     week,
		YearWeek: yearWeek,
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
	startDate, endDate := getDemandMonthStartEnd(t)

	return times.DateRange{
		Start: startDate.Format(constant.DateLayout),
		End:   endDate.Format(constant.DateLayout),
	}
}

// GetDemandYearMonth returns the year, month based on the input time from a demand perspective.
func GetDemandYearMonth(t time.Time) (year int, month time.Month) {
	// 无论如何，需求所属年月均以所在周的周一为准
	weekdays := Weekdays(t)
	return weekdays[0].Year(), weekdays[0].Month()
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

// getDemandMonthStartEnd 获取给定时间的需求年月的第一天和最后一天
// 当输入时间在所在周跨月，则统一将该周周一所在月作为需求月
// 当需求月第一天所在周跨月时，将第二周的第一天当作本月的第一天
// 无论如何，最后一周的最后一天都是本月的最后一天，即使最后一周出现跨月
func getDemandMonthStartEnd(t time.Time) (time.Time, time.Time) {
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

	return startDate, endDate
}

// IsDayCrossMonth 给定日期，判断日期是否在周纬度跨月且该日期属于下个月
func IsDayCrossMonth(t time.Time) bool {
	weekdays := Weekdays(t)

	if weekdays[0].Month() != t.Month() {
		return true
	}
	return false
}

// IsAboutToExpire check whether the cvm and cbs plan is about to expire
func IsAboutToExpire(expectedTime *time.Time) bool {
	monthStart, monthEnd := getDemandMonthStartEnd(time.Now())

	// 如果期望时间早于本月（即已过期），是否还属于“即将”过期？
	if expectedTime.Before(monthStart) || expectedTime.After(monthEnd) {
		return false
	}
	return true
}

// GetDemandStatus 获取需求状态
func GetDemandStatus(expectedStart, expectedEnd *time.Time) enumor.DemandStatus {
	monthStart, monthEnd := getDemandMonthStartEnd(time.Now())

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

// GetDemandStatusByExpectTime 根据期望交付时间获取需求状态
func GetDemandStatusByExpectTime(expectTime string) (enumor.DemandStatus, error) {
	t, err := time.Parse(constant.DateLayout, expectTime)
	if err != nil {
		return "", err
	}

	monthStart, monthEnd := getDemandMonthStartEnd(time.Now())
	// 未到申领时间
	if t.After(monthEnd) {
		return enumor.DemandStatusNotReady, nil
	}

	// 已过期
	if t.Before(monthStart) {
		return enumor.DemandStatusExpired, nil
	}

	return enumor.DemandStatusCanApply, nil
}
