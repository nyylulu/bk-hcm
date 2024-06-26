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

package plan

import (
	"container/list"
	"sync"
)

// UniQueue is a unique FIFO queue, which only pushes values that are not already in the queue.
type UniQueue struct {
	mu    sync.Mutex
	queue *list.List
	cache map[string]interface{}
}

// NewUniQueue creates a unique queue instance.
func NewUniQueue() *UniQueue {
	return &UniQueue{
		queue: list.New(),
		cache: make(map[string]interface{}),
	}
}

// Enqueue add the value to the end if the value is not already in the queue.
func (c *UniQueue) Enqueue(value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[value]; ok {
		return
	}

	c.queue.PushBack(value)
	c.cache[value] = value
}

// Pop gets and removes the front item in the queue.
func (c *UniQueue) Pop() (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.queue.Len() == 0 {
		return "", false
	}

	front := c.queue.Front()
	id := front.Value.(string)
	c.queue.Remove(front)
	delete(c.cache, id)

	return id, true
}

// Clear removes all the item in the queue.
func (c *UniQueue) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for c.queue.Len() > 0 {
		front := c.queue.Front()
		c.queue.Remove(front)
		delete(c.cache, front.Value.(string))
	}
}
