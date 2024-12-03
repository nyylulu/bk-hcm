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

package times

import (
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
)

// ConvStdTimeFormat 转为HCM标准时间格式
func ConvStdTimeFormat(t time.Time) string {
	return t.In(time.Local).Format(constant.TimeStdFormat)
}

// ConvStdTimeNow 转为HCM标准时间格式的当前时间
func ConvStdTimeNow() time.Time {
	return time.Now().In(time.Local)
}

// ParseToStdTime parse layout format time to std time.
func ParseToStdTime(layout, t string) (string, error) {
	tm, err := time.Parse(layout, t)
	if err != nil {
		return "", err
	}

	return tm.In(time.Local).Format(constant.TimeStdFormat), nil
}

// Day 24 hours
const Day = time.Hour * 24

// DateRange define a date range that includes the start and end parameters.
// The date can be year, month or day.
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Validate validates the date range.
func (r DateRange) Validate() error {
	start, err := ParseDay(r.Start)
	if err != nil {
		return err
	}

	end, err := ParseDay(r.End)
	if err != nil {
		return err
	}

	if start.After(end) {
		return errf.New(errf.InvalidParameter, "start should be no later than end")
	}

	return nil
}

// GetTimeDate get start and end time.
func (r DateRange) GetTimeDate() (time.Time, time.Time, error) {
	start, err := ParseDay(r.Start)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end, err := ParseDay(r.End)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, end, nil
}

// ParseDay parse day from string.
func ParseDay(formattedDay string) (time.Time, error) {
	if len(formattedDay) == 0 {
		return time.Time{}, errf.New(errf.InvalidParameter, "empty date time")
	}

	d, err := time.Parse(constant.DateLayout, formattedDay)
	if err != nil {
		return time.Time{}, errf.Newf(errf.InvalidParameter, "invalid date time format, should be like %s, err: %v",
			constant.DateLayout, err)
	}

	return d, nil
}

// DaysInMonth 返回给定年份和月份的天数
func DaysInMonth(year int, month time.Month) int {
	// 获取下个月的第一天
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	// 获取本月的最后一天
	lastOfThisMonth := firstOfNextMonth.AddDate(0, 0, -1)

	return lastOfThisMonth.Day()
}

// GetMonthDays 获取指定年月的天数列表
func GetMonthDays(year int, month time.Month) []int {
	lastDay := DaysInMonth(year, month)
	// 创建日期列表
	days := make([]int, lastDay)
	for day := 1; day <= int(lastDay); day++ {
		days[day-1] = day
	}
	return days
}

// GetDataIntDate 如：year:2021, month：1, day: 2 => 20210102
func GetDataIntDate(year, month, day int) int {
	return year*10000 + month*100 + day
}

// DateTimeItem defines green channel datetime item.
type DateTimeItem struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// Validate validate.
func (r *DateTimeItem) Validate() error {
	if r.Year < 0 {
		return errf.New(errf.InvalidParameter, "year should be >= 0")
	}

	if r.Month < 1 || r.Month > 12 {
		return errf.New(errf.InvalidParameter, "month should be in [1, 12]")
	}

	if r.Day < 1 || r.Day > 31 {
		return errf.New(errf.InvalidParameter, "day should be in [1, 31]")
	}

	return nil
}

// GetTime get time.
func (r *DateTimeItem) GetTime() time.Time {
	return time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)
}

// GetMondayOfWeek 获取本周的周一日期
func GetMondayOfWeek(now time.Time) time.Time {
	weekday := now.Weekday()

	// 计算距离本周一的天数差
	daysToMonday := int(time.Monday - weekday)
	if weekday == time.Sunday {
		daysToMonday = -6
	}

	// 计算本周一的日期
	monday := now.AddDate(0, 0, daysToMonday)
	return monday
}
