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

package table

import (
	"math"
	"testing"
)

// TestRecycleType_getRecycleTypePriority 测试回收类型优先级规则
func TestRecycleType_getRecycleTypePriority(t *testing.T) {
	type args struct {
		recycleTypeSeq []RecycleType
	}
	tests := []struct {
		name string
		rt   RecycleType
		args args
		want int
	}{
		{
			name: "裁撤优先级最高", rt: RecycleTypeDissolve, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental, RecycleTypeRollServer, RecycleTypeDissolve},
			}, want: math.MinInt,
		},
		{
			name: "滚服-优先短租", rt: RecycleTypeRollServer, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental, RecycleTypeRollServer},
			}, want: 1,
		},
		{
			name: "短租-优先短租", rt: RecycleTypeShortRental, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental, RecycleTypeRollServer},
			}, want: 0,
		},
		{
			name: "滚服-默认优先级", rt: RecycleTypeRollServer, args: args{
				recycleTypeSeq: []RecycleType{},
			}, want: 0,
		},
		{
			name: "短租-默认优先级", rt: RecycleTypeShortRental, args: args{
				recycleTypeSeq: []RecycleType{},
			}, want: 1,
		},
		{
			name: "滚服-优先级仅指定短租", rt: RecycleTypeRollServer, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental},
			}, want: 1,
		},
		{
			name: "短租-优先级仅指定短租", rt: RecycleTypeShortRental, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental},
			}, want: 0,
		},
		{
			name: "常规-优先常规", rt: RecycleTypeRegular, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeRegular, RecycleTypeRollServer},
			}, want: 0,
		},
		{
			name: "滚服-优先常规和滚服", rt: RecycleTypeRollServer, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeRegular, RecycleTypeRollServer},
			}, want: 1,
		},
		{
			name: "短租-优先常规和滚服", rt: RecycleTypeShortRental, args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeRegular, RecycleTypeRollServer},
			}, want: 3,
		},
		{
			name: "未知类型", rt: RecycleType("未知类型"), args: args{
				recycleTypeSeq: []RecycleType{RecycleTypeShortRental, RecycleTypeRollServer},
			}, want: math.MaxInt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rt.getRecycleTypePriority(tt.args.recycleTypeSeq); got != tt.want {
				t.Errorf("getRecycleTypePriority() = %v, want %v", got, tt.want)
			}
		})
	}
}
