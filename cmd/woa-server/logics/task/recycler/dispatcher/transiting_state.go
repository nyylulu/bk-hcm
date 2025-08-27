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

// Package dispatcher defines the logic of recycling task dispatching
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// TransitingState the action to be executed in transiting state
type TransitingState struct{}

// UpdateState update next state
func (ts *TransitingState) UpdateState(ctx EventContext, ev *event.Event) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			ts.Name())
		return fmt.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			ts.Name())
	}

	// set next state
	if errUpdate := ts.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("recycler:logics:cvm:TransitingState:failed, failed to update recycle order state, subOrderId: %s, "+
			"err: %v", taskCtx.Order.SuborderID, errUpdate)
		return errUpdate
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("recycler:logics:cvm:TransitingState:failed, failed to add order to dispatch, "+
			"for dispatcher is nil, subOrderId: %s, state: %s", taskCtx.Order.SuborderID, ts.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, ts.Name())
	}

	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)

	// 记录日志
	logs.Infof("recycler: finish transfer state, subOrderId: %s, ev: %+v, recycleType: %s", taskCtx.Order.SuborderID,
		cvt.PtrToVal(ev), taskCtx.Order.RecycleType)
	return nil
}

// Name return the name of transiting state
func (ts *TransitingState) Name() table.RecycleStatus {
	return table.RecycleStatusTransiting
}

// Execute executes action in transiting state
func (ts *TransitingState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context")
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", ts.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", ts.Name())
	}
	orderId := taskCtx.Order.SuborderID
	// run transit tasks
	ev := taskCtx.Dispatcher.transit.DealRecycleOrder(taskCtx.Order)

	// 记录日志，方便排查问题
	logs.Infof("recycler:logics:cvm:TransitingState:end, orderID: %s, recycleType: %s, ev: %+v", orderId,
		taskCtx.Order.RecycleType, cvt.PtrToVal(ev))

	return ts.UpdateState(ctx, ev)
}

func (ts *TransitingState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	switch ev.Type {
	case event.TransitSuccess:
		update["stage"] = table.RecycleStageReturn
		update["status"] = table.RecycleStatusReturning
		if order.ResourceType == table.ResourceTypeOthers || (order.ResourceType == table.ResourceTypePm &&
			order.RecycleType == table.RecycleTypeRegular) {
			update["stage"] = table.RecycleStageDone
			update["status"] = table.RecycleStatusDone
			update["handler"] = "AUTO"
			update["success_num"] = order.TotalNum
			update["pending_num"] = 0
			update["failed_num"] = 0
		}
	case event.TransitFailed:
		update["status"] = table.RecycleStatusTransitFailed
		update["handler"] = recovertask.Handler
		if ev.Error != nil {
			update["message"] = ev.Error.Error()
		}
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
