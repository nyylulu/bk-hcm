/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package dispatcher implements the dispatcher of recycle task
package dispatcher

import (
	"hcm/cmd/woa-server/dal/task/table"
)

// DetectFailedState the action to be executed in detect failed state
type DetectFailedState struct{}

// Name return the name of detect failed state
func (ds *DetectFailedState) Name() table.RecycleStatus {
	return table.RecycleStatusDetectFailed
}

// Execute executes action in detect failed state
func (ds *DetectFailedState) Execute(ctx EventContext) error {
	return nil
}
