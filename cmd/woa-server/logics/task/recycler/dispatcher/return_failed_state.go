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

// Package dispatcher implements state machine dispatcher
package dispatcher

import (
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
)

// ReturnFailedState the action to be executed in return failed state
type ReturnFailedState struct{}

// Name return the name of return failed state
func (rs *ReturnFailedState) Name() table.RecycleStatus {
	return table.RecycleStatusReturnFailed
}

// Execute executes action in return failed state
func (rs *ReturnFailedState) Execute(ctx EventContext) error {
	return nil
}

// UpdateState update next state
func (rs *ReturnFailedState) UpdateState(ctx EventContext, ev *event.Event) error {
	return nil
}
