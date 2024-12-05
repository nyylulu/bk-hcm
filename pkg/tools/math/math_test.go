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

package math

import (
	"testing"
)

func TestRoundToDecimalPlaces1(t *testing.T) {
	type args struct {
		f      float64
		places int
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"0-0", args{1, 0}, 1, false},
		{".2-0", args{1.23, 0}, 1, false},
		{".2-1", args{1.23, 1}, 1.2, false},
		{".2-2", args{1.23, 2}, 1.23, false},
		{".2-3", args{1.23, 3}, 1.23, false},
		{"nan", args{123.45678, 309}, 0, true},
		{"error", args{123.45, -1}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RoundToDecimalPlaces(tt.args.f, tt.args.places)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoundToDecimalPlaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoundToDecimalPlaces() got = %v, want %v", got, tt.want)
			}
		})
	}
}
