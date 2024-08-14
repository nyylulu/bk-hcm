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

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/logs"
)

// TransitingState the action to be executed in transiting state
type TransitingState struct{}

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

	// 记录日志，方便排查问题
	logs.Infof("recycler:logics:cvm:TransitingState:start, orderID: %s", orderId)

	// run transit tasks
	ev := taskCtx.Dispatcher.transit.DealRecycleOrder(taskCtx.Order)

	// set next state
	if errUpdate := ts.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("recycler:logics:cvm:TransitingState:failed, failed to update recycle order %s state, err: %v",
			orderId, errUpdate)
		return errUpdate
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("recycler:logics:cvm:TransitingState:failed, failed to add order to dispatch, "+
			"for dispatcher is nil, order id: %s, state: %s", orderId, ts.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, order id: %s, state: %s",
			orderId, ts.Name())
	}

	taskCtx.Dispatcher.Add(orderId)

	// 记录日志
	logs.Infof("recycler:logics:cvm:TransitingState:end, orderID: %s", orderId)

	return nil
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
		update["handler"] = "dommyzhang;forestchen"
	default:
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
