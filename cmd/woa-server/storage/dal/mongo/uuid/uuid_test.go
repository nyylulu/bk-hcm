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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// GODRIVER-2349
// Test that initializing many package-global UUID sources concurrently never leads to any duplicate
// UUIDs being generated.
func TestGlobalSource(t *testing.T) {
	t.Run("exp rand 1 UUID x 1,000,000 goroutines using a global source", func(t *testing.T) {
		// Read a UUID from each of 1,000,000 goroutines and assert that there is never a duplicate value.
		const iterations = 1e6
		uuids := new(sync.Map)
		var wg sync.WaitGroup
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(i int) {
				defer wg.Done()
				uuid, err := New()
				require.NoError(t, err, "new() error")
				_, ok := uuids.Load(uuid)
				require.Falsef(t, ok, "New returned a duplicate UUID on iteration %d: %v", i, uuid)
				uuids.Store(uuid, true)
			}(i)
		}
		wg.Wait()
	})
	t.Run("exp rand 1 UUID x 1,000,000 goroutines each initializing a new source", func(t *testing.T) {
		// Read a UUID from each of 1,000,000 goroutines and assert that there is never a duplicate value.
		// The goal is to emulate many separate Go driver processes starting at the same time and
		// initializing the uuid package at the same time.
		const iterations = 1e6
		uuids := new(sync.Map)
		var wg sync.WaitGroup
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(i int) {
				defer wg.Done()
				s := newSource()
				uuid, err := s.new()
				require.NoError(t, err, "new() error")
				_, ok := uuids.Load(uuid)
				require.Falsef(t, ok, "New returned a duplicate UUID on iteration %d: %v", i, uuid)
				uuids.Store(uuid, true)
			}(i)
		}
		wg.Wait()
	})
	t.Run("exp rand 1,000 UUIDs x 1,000 goroutines each initializing a new source", func(t *testing.T) {
		// Read 1,000 UUIDs from each goroutine and assert that there is never a duplicate value, either
		// from the same goroutine or from separate goroutines.
		const iterations = 1000
		uuids := new(sync.Map)
		var wg sync.WaitGroup
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(i int) {
				defer wg.Done()
				s := newSource()
				for j := 0; j < iterations; j++ {
					uuid, err := s.new()
					require.NoError(t, err, "new() error")
					_, ok := uuids.Load(uuid)
					require.Falsef(t, ok, "goroutine %d returned a duplicate UUID on iteration %d: %v", i, j, uuid)
					uuids.Store(uuid, true)
				}
			}(i)
		}
		wg.Wait()
	})
}

func BenchmarkUuidGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := New()
		if err != nil {
			panic(err)
		}
	}
}
