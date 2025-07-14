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
	"hcm/pkg/cc"
)

// WrapZiyanMultiSecret wrap ziyan multi secret. 每次调用返回独立的secret id 列表
func WrapZiyanMultiSecret(mainSecret *BaseSecret) *MultiSecret {
	if mainSecret == nil {
		return nil
	}
	secrets := cc.HCService().ZiyanSecrets
	var backupSecrets = make([]BaseSecret, 0)
	for _, secret := range secrets {
		if mainSecret.CloudAccountID == "" || mainSecret.CloudAccountID != secret.SubAccountID {
			continue
		}
		backupSecrets = append(backupSecrets, BaseSecret{
			CloudSecretID:  secret.ID,
			CloudSecretKey: secret.Key,
			CloudAccountID: secret.SubAccountID,
		})
	}

	return NewMultiSecretWithRandomIndex(*mainSecret, backupSecrets...)
}
