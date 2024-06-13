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
func (r *DateRange) Validate() error {
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
