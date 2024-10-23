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

// ReturningState the action to be executed in returning state
type ReturningState struct{}

// Name return the name of returning state
func (rs *ReturningState) Name() table.RecycleStatus {
	return table.RecycleStatusReturning
}

// UpdateState update next state
func (rs *ReturningState) UpdateState(ctx EventContext, ev *event.Event) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to return status, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			rs.Name())
		return fmt.Errorf("failed to convert to return status, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			rs.Name())
	}

	// set next state
	if errUpdate := rs.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order state, subOrderId: %s, err: %v", taskCtx.Order.SuborderID,
			errUpdate)
		return errUpdate
	}

	if ev.Type == event.ReturnHandling {
		if taskCtx.Dispatcher == nil {
			logs.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
				taskCtx.Order.SuborderID, rs.Name())
			return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
				taskCtx.Order.SuborderID, rs.Name())
		}

		go func() {
			// query every 5 minutes
			time.Sleep(time.Minute * 5)
			taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)
		}()
	}

	// 记录日志
	logs.Infof("recycler: finish return state, subOrderId: %s", taskCtx.Order.SuborderID)
	return nil
}

// Execute executes action in returning state
func (rs *ReturningState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context")
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", rs.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", rs.Name())
	}
	orderId := taskCtx.Order.SuborderID
	// 记录日志，方便排查问题
	logs.Infof("recycler:logics:cvm:ReturningState:start, orderID: %s", orderId)
	// run return tasks
	ev := taskCtx.Dispatcher.returner.DealRecycleOrder(taskCtx.Order)
	return rs.UpdateState(ctx, ev)
}

func (rs *ReturningState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	switch ev.Type {
	case event.ReturnSuccess:
		update["stage"] = table.RecycleStageDone
		update["status"] = table.RecycleStatusDone
		update["handler"] = "AUTO"
	case event.ReturnFailed:
		update["status"] = table.RecycleStatusReturnFailed
	case event.ReturnHandling:
		logs.Infof("recycle return order is handling, subOrderId: %s, type: %s", order.SuborderID, ev.Type)
	default:
		logs.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
		return fmt.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
