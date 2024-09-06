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

// Package randutil provides common random number utilities.
package randutil

import (
	crand "crypto/rand"
	"fmt"
	"io"

	xrand "hcm/cmd/woa-server/storage/dal/mongo/randutil/rand"
)

// NewLockedRand returns a new "x/exp/rand" pseudo-random number generator seeded with a
// cryptographically-secure random number.
// It is safe to use from multiple goroutines.
func NewLockedRand() *xrand.Rand {
	var randSrc = new(xrand.LockedSource)
	randSrc.Seed(cryptoSeed())
	return xrand.New(randSrc)
}

// cryptoSeed returns a random uint64 read from the "crypto/rand" random number generator. It is
// intended to be used to seed pseudorandom number generators at package initialization. It panics
// if it encounters any errors.
func cryptoSeed() uint64 {
	var b [8]byte
	_, err := io.ReadFull(crand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("failed to read 8 bytes from a \"crypto/rand\".Reader: %v", err))
	}

	return (uint64(b[0]) << 0) | (uint64(b[1]) << 8) | (uint64(b[2]) << 16) | (uint64(b[3]) << 24) |
		(uint64(b[4]) << 32) | (uint64(b[5]) << 40) | (uint64(b[6]) << 48) | (uint64(b[7]) << 56)
}
