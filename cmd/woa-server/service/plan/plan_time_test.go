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
	"testing"
	"time"
)

func TestGetDemandYearMonthWeek(t *testing.T) {
	date := time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local)
	ymw := GetDemandYearMonthWeek(date)
	if !(ymw.Year == 2024 && ymw.Month == time.August && ymw.Week == 5) {
		t.Errorf("test get demand year month week failed, got: %+v", ymw)
		return
	}
}

func TestGetDemandDateRangeInWeek(t *testing.T) {
	date := time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local)
	dr := GetDemandDateRangeInWeek(date)
	if !(dr.Start == "2024-08-26" && dr.End == "2024-09-01") {
		t.Errorf("test get demand date range in week failed, got: %+v", dr)
		return
	}
}

func TestGetDemandDateRangeInMonth(t *testing.T) {
	date := time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local)
	dr := GetDemandDateRangeInMonth(date)
	if !(dr.Start == "2024-07-29" && dr.End == "2024-09-01") {
		t.Errorf("test get demand date range in week failed, got: %+v", dr)
		return
	}
}
