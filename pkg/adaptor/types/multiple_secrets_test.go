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

package types

import (
	"math"
	"testing"
)

func TestNewMultiSecret(t *testing.T) {
	main, backups := prepareGetCred()
	t.Run("test_over_flow", func(t *testing.T) {
		ms := NewMultiSecret(main, backups...)
		ak, _, _ := ms.GetCredential()
		if ak != main.CloudSecretID {
			t.Errorf("first secret not match, want: %s, got: %s", main.CloudSecretID, ak)
		}
		// when uint64 overflowed, it will be reset to zero
		ms.currentIndex.Add(math.MaxUint64)
		for i := 0; i < len(backups)+1; i++ {
			ak, sk, idx := ms.GetCredential()
			if ak != ms.secrets[i].CloudSecretID {
				t.Errorf("secret not match after overflow: %d, want: %s, got: %s", i, ms.secrets[i].CloudSecretID, ak)
			}
			t.Log(ak, sk, idx)
		}
		for i := 0; i < len(backups)+1; i++ {
			ak, sk, idx := ms.GetCredential()
			if ak != ms.secrets[i].CloudSecretID {
				t.Errorf("secret not match after second iterate: %d, want: %s, got: %s",
					i, ms.secrets[i].CloudSecretID, ak)
			}
			t.Log(ak, sk, idx)
		}
	})

}

// slightly slower than atomic version
func BenchmarkGetCredentialMutex(b *testing.B) {
	main, backups := prepareGetCred()
	ms := NewMultiSecretMutex(main, backups...)
	for i := 0; i < b.N; i++ {
		ak, sk, idx := ms.GetCredential()
		if len(ak+sk+idx) == 0 {
			b.Errorf("len equal zero: %d, %s, %s, %s", i, ak, sk, idx)
		}
	}
}

func BenchmarkGetCredential(b *testing.B) {
	main, backups := prepareGetCred()
	ms := NewMultiSecret(main, backups...)
	for i := 0; i < b.N; i++ {
		ak, sk, idx := ms.GetCredential()
		if len(ak+sk+idx) == 0 {
			b.Errorf("len equal zero: %d, %s, %s, %s", i, ak, sk, idx)
		}
	}
}

func prepareGetCred() (BaseSecret, []BaseSecret) {

	backupSecrets := []BaseSecret{
		{CloudSecretID: "b", CloudSecretKey: "b", CloudAccountID: "2"},
		{CloudSecretID: "c", CloudSecretKey: "c", CloudAccountID: "3"},
		{CloudSecretID: "d", CloudSecretKey: "d", CloudAccountID: "4"},
		{CloudSecretID: "e", CloudSecretKey: "e", CloudAccountID: "5"},
	}
	main := BaseSecret{CloudSecretID: "a", CloudSecretKey: "a", CloudAccountID: "1"}
	return main, backupSecrets
}
