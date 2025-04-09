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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDaysInMonth(t *testing.T) {
	testCases := []struct {
		year  int
		month time.Month
		days  int
	}{
		{
			year:  2024,
			month: time.Month(5),
			days:  31,
		},
		{
			year:  2024,
			month: time.Month(4),
			days:  30,
		},
	}
	for _, testCase := range testCases {
		days := DaysInMonth(testCase.year, testCase.month)
		assert.Equal(t, days, testCase.days)
	}
}

func TestGetMondayOfWeek(t *testing.T) {
	testCases := []struct {
		now    time.Time
		monday time.Time
	}{{
		now:    time.Date(2024, 11, 18, 0, 0, 0, 0, time.Local),
		monday: time.Date(2024, 11, 18, 0, 0, 0, 0, time.Local),
	},
		{
			now:    time.Date(2024, 11, 21, 0, 0, 0, 0, time.Local),
			monday: time.Date(2024, 11, 18, 0, 0, 0, 0, time.Local),
		},
		{
			now:    time.Date(2024, 11, 24, 0, 0, 0, 0, time.Local),
			monday: time.Date(2024, 11, 18, 0, 0, 0, 0, time.Local),
		},
	}
	for _, testCase := range testCases {
		monday := GetMondayOfWeek(testCase.now)
		assert.Equal(t, monday, testCase.monday)
	}
}

func TestGetNextMondayOfWeek(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "monday",
			args: args{t: time.Date(2024, 11, 18, 0, 0, 0, 0, time.Local)},
			want: time.Date(2024, 11, 25, 0, 0, 0, 0, time.Local),
		},
		{
			name: "tuesday",
			args: args{t: time.Date(2024, 11, 19, 0, 0, 0, 0, time.Local)},
			want: time.Date(2024, 11, 25, 0, 0, 0, 0, time.Local),
		},
		{
			name: "sunday",
			args: args{t: time.Date(2024, 11, 24, 0, 0, 0, 0, time.Local)},
			want: time.Date(2024, 11, 25, 0, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetNextMondayOfWeek(tt.args.t), "GetNextMondayOfWeek(%v)", tt.args.t)
		})
	}
}

func TestIsLastNDaysOfMonth(t *testing.T) {
	type args struct {
		t     time.Time
		lastN int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{
			t:     time.Date(2024, 2, 28, 13, 28, 59, 0, time.Local),
			lastN: 1,
		}, false},
		{"test2", args{
			t:     time.Date(2025, 2, 28, 13, 28, 59, 0, time.Local),
			lastN: 1,
		}, true},
		{"test3", args{
			t:     time.Date(2024, 11, 28, 0, 0, 0, 0, time.Local),
			lastN: 3,
		}, true},
		{"test4", args{
			t:     time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local),
			lastN: 7,
		}, true},
		{"test5", args{
			t:     time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local),
			lastN: 6,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsLastNDaysOfMonth(tt.args.t, tt.args.lastN), "IsLastNDaysOfMonth(%v, %v)",
				tt.args.t, tt.args.lastN)
		})
	}
}

func TestDaysUntilEndOfTheMonth(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test1", args{
			t: time.Date(2024, 2, 28, 13, 28, 59, 0, time.Local),
		}, 2},
		{"test2", args{
			t: time.Date(2025, 2, 28, 13, 28, 59, 0, time.Local),
		}, 1},
		{"test3", args{
			t: time.Date(2024, 11, 28, 0, 0, 0, 0, time.Local),
		}, 3},
		{"test4", args{
			t: time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local),
		}, 7},
		{"test5", args{
			t: time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local),
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DaysUntilEndOfTheMonth(tt.args.t), "DaysUntilEndOfTheMonth(%v)", tt.args.t)
		})
	}
}
