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

// Package dispatcher defines the interface of dispatcher
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/logs"
)

// AuditingState the action to be executed in auditing state
type AuditingState struct{}

// Name return the name of auditing state
func (as *AuditingState) Name() table.RecycleStatus {
	return table.RecycleStatusAudit
}

// Execute executes action in auditing state
func (as *AuditingState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*AuditContext)
	if !ok {
		logs.Errorf("failed to convert to audit context")
		return errors.New("failed to convert to audit context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", as.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", as.Name())
	}
	orderId := taskCtx.Order.SuborderID

	ev := as.dealAuditTask(taskCtx.Order, taskCtx.Approval, taskCtx.Remark)

	// set next state
	if errUpdate := as.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order %s state, err: %v", orderId, errUpdate)
		return errUpdate
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("failed to add order to dispatch, for dispatcher is nil, order id: %s, state: %s", orderId,
			as.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, order id: %s, state: %s",
			orderId, as.Name())
	}

	taskCtx.Dispatcher.Add(orderId)

	return nil
}

func (as *AuditingState) dealAuditTask(order *table.RecycleOrder, approval bool, remark string) *event.Event {
	ev := &event.Event{
		Type:  event.AuditApproved,
		Error: nil,
	}

	stage := table.RecycleStageTransit
	status := table.RecycleStatusTransiting

	if !approval {
		ev.Type = event.AuditRejected
		ev.Error = fmt.Errorf("audit remark: %s", remark)
		stage = table.RecycleStageTerminate
		status = table.RecycleStatusRejected
	}

	if err := as.updateHostInfo(order, stage, status); err != nil {
		logs.Errorf("failed to update recycle hosts, order id: %s, err: %v")
		ev.Type = event.AuditRejected
	}

	return ev
}

func (as *AuditingState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	switch ev.Type {
	case event.AuditApproved:
		update["stage"] = table.RecycleStageTransit
		update["status"] = table.RecycleStatusTransiting
	case event.AuditRejected:
		update["stage"] = table.RecycleStageTerminate
		update["status"] = table.RecycleStatusRejected
		if ev.Error != nil {
			update["message"] = ev.Error.Error()
		}
	default:
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}

func (as *AuditingState) updateHostInfo(order *table.RecycleOrder, stage table.RecycleStage,
	status table.RecycleStatus) error {

	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, order id: %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
