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
	"math/rand"
	"sync"
	"sync/atomic"
)

// MultiSecret 轮询多秘钥支持
type MultiSecret struct {
	// first as main secret
	secrets      []BaseSecret
	currentIndex atomic.Uint64
}

// GetSecretId 默认返回主密钥
func (m *MultiSecret) GetSecretId() string {
	return m.secrets[0].CloudSecretID
}

// GetToken 不支持token，默认返回空
func (m *MultiSecret) GetToken() string {
	return ""
}

// GetSecretKey 默认返回主密钥
func (m *MultiSecret) GetSecretKey() string {
	return m.secrets[0].CloudSecretKey
}

// GetCredential 按序获取备用秘钥
func (m *MultiSecret) GetCredential() (ak string, sk string, token string) {
	idx := m.currentIndex.Add(1)
	idx = (idx - 1) % uint64(len(m.secrets))
	return m.secrets[idx].CloudSecretID, m.secrets[idx].CloudSecretKey, ""
}

// GetSecrets 获取所有秘钥
func (m *MultiSecret) GetSecrets() []BaseSecret {
	return m.secrets
}

// setRandomIndex 设置随机起始点
func (m *MultiSecret) setRandomIndex() {
	m.currentIndex.Add(uint64(rand.Intn(len(m.secrets))))
}

// NewMultiSecret new multi secret.
func NewMultiSecret(mainSecret BaseSecret, backupSecrets ...BaseSecret) *MultiSecret {
	secrets := make([]BaseSecret, 0, len(backupSecrets)+1)
	secrets = append(secrets, mainSecret)
	secrets = append(secrets, backupSecrets...)
	ms := &MultiSecret{secrets: secrets, currentIndex: atomic.Uint64{}}
	return ms
}

// NewMultiSecretWithRandomIndex new multi secret with random index
func NewMultiSecretWithRandomIndex(mainSecret BaseSecret, backupSecrets ...BaseSecret) *MultiSecret {
	ms := NewMultiSecret(mainSecret, backupSecrets...)
	ms.setRandomIndex()
	return ms
}

// MultiSecretMutex 轮询多秘钥支持 - 互斥锁版本
type MultiSecretMutex struct {
	// first as main secret
	secrets []BaseSecret
	mu      sync.Mutex
	muIdx   uint64
}

// GetSecretId 默认返回主密钥
func (m *MultiSecretMutex) GetSecretId() string {
	return m.secrets[0].CloudSecretID
}

// GetToken 不支持token，默认返回空
func (m *MultiSecretMutex) GetToken() string {
	return ""
}

// GetSecretKey 默认返回主密钥
func (m *MultiSecretMutex) GetSecretKey() string {
	return m.secrets[0].CloudSecretKey
}

// GetCredential 按序获取备用秘钥
func (m *MultiSecretMutex) GetCredential() (ak string, sk string, token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muIdx++
	idx := m.muIdx % uint64(len(m.secrets))
	return m.secrets[idx].CloudSecretID, m.secrets[idx].CloudSecretKey, ""
}

// NewMultiSecretMutex new mutex version multi secret.
func NewMultiSecretMutex(mainSecret BaseSecret, backupSecrets ...BaseSecret) *MultiSecretMutex {
	secrets := make([]BaseSecret, 0, len(backupSecrets)+1)
	secrets = append(secrets, mainSecret)
	secrets = append(secrets, backupSecrets...)
	return &MultiSecretMutex{secrets: secrets}
}
