/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package concurrence

import (
	"testing"
	"time"
)

func TestBatchReadResults_Normal(t *testing.T) {
	ch := make(chan int, 5)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
	}()

	results, ok := BatchReadChannel(ch, 5, 10*time.Millisecond)
	if !ok {
		t.Error("expected ok=true, got false")
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestBatchReadResults_Timeout(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
		time.Sleep(20 * time.Millisecond)
		ch <- 2
		close(ch)
	}()

	results, ok := BatchReadChannel(ch, 5, 10*time.Millisecond)
	if !ok {
		t.Error("expected ok=true, got false")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestBatchReadResults_MaxCount(t *testing.T) {
	ch := make(chan int, 5)
	for i := 0; i < 5; i++ {
		ch <- i
	}
	close(ch)

	results, ok := BatchReadChannel(ch, 3, 10*time.Millisecond)
	if !ok {
		t.Error("expected ok=true, got false")
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestBatchReadResults_NilChannel(t *testing.T) {
	results, ok := BatchReadChannel[int](nil, 5, 10*time.Millisecond)
	if ok {
		t.Error("expected ok=false, got true")
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}

func TestBatchReadResults_ZeroN(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	close(ch)

	results, ok := BatchReadChannel(ch, 0, 10*time.Millisecond)
	if !ok {
		t.Error("expected ok=true, got false")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestBatchReadResults_ChannelClosed(t *testing.T) {
	ch := make(chan int)
	close(ch)

	results, ok := BatchReadChannel(ch, 5, 10*time.Millisecond)
	if ok {
		t.Error("expected ok=false, got true")
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}
