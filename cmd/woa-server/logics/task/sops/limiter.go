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

package sops

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// 定义一个全局的速率限制器映射
var sopsApiLimiters = make(map[sopsApi]*rate.Limiter)
var mu sync.Mutex

// sopsLimit 标准运维限流大小
type sopsLimit float64

const (
	// writeLimit 标准运维写接口限制10/s
	writeLimit sopsLimit = 10
	// readLimit  标准运维读接口限制15/s，理论上20/s, 实际上15的时候就会出现较多失败
	readLimit sopsLimit = 15
)

// sopsApi 标准运维接口
type sopsApi string

const (
	// createTask 标准运维创建任务
	createTask sopsApi = "create_task"
	// startTask 标准运维启动任务
	startTask sopsApi = "start_task"
	// getTaskStatus 标准运维获取任务状态
	getTaskStatus sopsApi = "get_task_status"
)

// getSopsLimiter 获取或创建一个Sops的限频器
func getSopsLimiter(api sopsApi, limit sopsLimit) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if _, exist := sopsApiLimiters[api]; !exist {
		// 标准运维并发能力有限，限制桶大小
		burst := int(limit/2) + 1
		sopsApiLimiters[api] = rate.NewLimiter(rate.Limit(limit), burst)
	}

	return sopsApiLimiters[api]
}

// WaitSopsCreateTaskLimiter 标准运维-创建任务限频
func WaitSopsCreateTaskLimiter(ctx context.Context, timeout time.Duration) error {
	timedCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	limiter := getSopsLimiter(createTask, writeLimit)
	return limiter.Wait(timedCtx)
}

// WaitSopsStartTaskLimiter 标准运维-启动任务限频
func WaitSopsStartTaskLimiter(ctx context.Context, timeout time.Duration) error {
	timedCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	limiter := getSopsLimiter(startTask, writeLimit)
	return limiter.Wait(timedCtx)
}

// WaitSopsGetTaskStatusLimiter 标准运维-查询任务限频
func WaitSopsGetTaskStatusLimiter(ctx context.Context, timeout time.Duration) error {
	timedCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	limiter := getSopsLimiter(getTaskStatus, readLimit)
	return limiter.Wait(timedCtx)
}
