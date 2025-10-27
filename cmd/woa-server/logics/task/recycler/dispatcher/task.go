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

// Package dispatcher defines the recycle order processing task
package dispatcher

import (
	"hcm/cmd/woa-server/dal/task/table"
	srlogics "hcm/cmd/woa-server/logics/short-rental"
	"hcm/cmd/woa-server/logics/task/recycler/event"
)

// Task recycle order processing task
type Task struct {
	State       ActionState
	orderStatus table.RecycleStatus
	srLogic     srlogics.Logics
}

// NewTask creates a task instance
func NewTask(status table.RecycleStatus, shortRentalLogic srlogics.Logics) *Task {
	task := &Task{
		orderStatus: status,
		srLogic:     shortRentalLogic,
	}
	task.initState()
	return task
}

func (t *Task) initState() {
	switch t.orderStatus {
	case table.RecycleStatusUncommit:
		t.State = &UncommitState{}
	case table.RecycleStatusCommitted:
		t.State = &CommittedState{
			ShortRentalLogic: t.srLogic,
		}
	case table.RecycleStatusDetecting:
		t.State = &DetectingState{}
	case table.RecycleStatusDetectFailed:
		t.State = &DetectFailedState{}
	case table.RecycleStatusAudit:
		t.State = &AuditingState{
			ShortRentalLogic: t.srLogic,
		}
	case table.RecycleStatusRejected:
		t.State = &AuditRejectedState{}
	case table.RecycleStatusTransiting:
		t.State = &TransitingState{
			ShortRentalLogic: t.srLogic,
		}
	case table.RecycleStatusTransitFailed:
		t.State = &TransitFailedState{}
	case table.RecycleStatusReturning:
		t.State = &ReturningState{
			ShortRentalLogic: t.srLogic,
		}
	case table.RecycleStatusReturnFailed:
		t.State = &ReturnFailedState{}
	case table.RecycleStatusReturningPlan:
		t.State = &ReturningPlanState{}
	case table.RecycleStatusReturnPlanFailed:
		t.State = &ReturnPlanFailedState{}
	case table.RecycleStatusDone:
		t.State = &DoneState{}
	case table.RecycleStatusTerminate:
		t.State = &TerminateState{}
	default:
		t.State = &DftState{}
	}
}

// ActionState represents the action to be executed in a given state
type ActionState interface {
	Name() table.RecycleStatus
	Execute(ctx EventContext) error
	UpdateState(ctx EventContext, ev *event.Event) error
}

// DftState the action to be executed in default state
type DftState struct{}

// Name return the name of default state
func (rs *DftState) Name() table.RecycleStatus {
	return table.RecycleStatusDefault
}

// Execute executes action in default state
func (rs *DftState) Execute(ctx EventContext) error {
	return nil
}

// UpdateState update next state
func (rs *DftState) UpdateState(ctx EventContext, ev *event.Event) error {
	return nil
}
