/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package metric

import (
	"regexp"
	"testing"
)

func TestGetTCloudSecretID(t *testing.T) {
	type args struct {
		authHeader string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "xxx header",
			args: args{
				authHeader: "TC3-HMAC-SHA256 Credential=xxxxxxxx/2024-12-04/clb/tc3_request, SignedHeaders=content-type;host, Signature=957bd833ee997ac688162f033ae911d9d99b2c4504cfb0be37f20871fe8b4834",
			},
			want: "xxxxxxxx",
		},
		{
			name: "normal_ak",
			args: args{
				authHeader: "TC3-HMAC-SHA256 Credential=AKIDz8krbsJ5yKBZQpn74WFkmLPx3EXAMPLE/2020-03-19/2020-03-19/ap-guangzhou/tc3_request",
			},
			want: "AKIDz8krbsJ5yKBZQpn74WFkmLPx3EXAMPLE",
		},
		{
			name: "empty_string",
			args: args{
				authHeader: "   ",
			},
			want: "",
		},
		{
			name: "empty_cred",
			args: args{
				authHeader: "TC3-HMAC-SHA256 Credential=",
			},
			want: "",
		},
		{
			name: "empty_cred2",
			args: args{
				authHeader: "TC3-HMAC-SHA256 Credential=/",
			},
			want: "",
		},
		{
			name: "slashes",
			args: args{
				authHeader: "///",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTCloudSecretID(tt.args.authHeader); got != tt.want {
				t.Errorf("GetTCloudSecretID() = %v, want %v", got, tt.want)
			}

		})
	}
}

func BenchmarkGetTCloudSecredID(b *testing.B) {
	const payload = "TC3-HMAC-SHA256 Credential=AKIDz8krbsJ5yKBZQpn74WFkmLPx3EXAMPLE/2024-12-04/clb/tc3_request, SignedHeaders=content-type;host, Signature=957bd833ee997ac688162f033ae911d9d99b2c4504cfb0be37f20871fe8b4834"

	b.Run("bench-GetTCloudSecretID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ak := GetTCloudSecretID(payload)
			if ak != "AKIDz8krbsJ5yKBZQpn74WFkmLPx3EXAMPLE" {
				b.Errorf("ak not equal")
			}
		}
	})

	// 大概需要50倍时间
	re := regexp.MustCompile(`TC3-HMAC-SHA256 Credential=\w+/`)
	b.Run("bench-GetTCloudSecretIDRegexp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			found := re.FindString(payload)
			if len(found) < 28 {
				b.Errorf("ak not found")
			}
			if found[27:len(found)-1] != "AKIDz8krbsJ5yKBZQpn74WFkmLPx3EXAMPLE" {
				b.Errorf("ak not equal")
			}
		}
	})

}
