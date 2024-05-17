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

package uuid

import (
	"io"

	"hcm/cmd/woa-server/storage/dal/mongo/randutil"
)

// UUID represents a UUID.
type UUID [16]byte

// A source is a UUID generator that reads random values from a io.Reader.
// It should be safe to use from multiple goroutines.
type source struct {
	random io.Reader
}

// new returns a random UUIDv4 with bytes read from the source's random number generator.
func (s *source) new() (UUID, error) {
	var uuid UUID
	_, err := io.ReadFull(s.random, uuid[:])
	if err != nil {
		return UUID{}, err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}

// newSource returns a source that uses a pseudo-random number generator in reandutil package.
// It is intended to be used to initialize the package-global UUID generator.
func newSource() *source {
	return &source{
		random: randutil.NewLockedRand(),
	}
}

// globalSource is a package-global pseudo-random UUID generator.
var globalSource = newSource()

// New returns a random UUIDv4. It uses a global pseudo-random number generator in randutil
// at package initialization.
//
// New should not be used to generate cryptographically-secure random UUIDs.
func New() (UUID, error) {
	return globalSource.new()
}
