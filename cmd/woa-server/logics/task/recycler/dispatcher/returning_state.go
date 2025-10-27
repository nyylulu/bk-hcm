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

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	srlogics "hcm/cmd/woa-server/logics/short-rental"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// ReturningState the action to be executed in returning state
type ReturningState struct {
	ShortRentalLogic srlogics.Logics
}

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
			// 等待2分钟后重新查询
			time.Sleep(time.Minute * 2)
			taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)
		}()
		return nil
	}
	if taskCtx.Dispatcher == nil {
		logs.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, rs.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, rs.Name())
	}
	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)

	// 记录日志
	logs.Infof("recycler: finish return state, subOrderId: %s, ev: %+v", taskCtx.Order.SuborderID, cvt.PtrToVal(ev))
	return nil
}

// Execute executes action in returning state
func (rs *ReturningState) Execute(ctx EventContext) error {
	kt := core.NewBackendKit()
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context, rid: %s", kt.Rid)
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil, rid: %s", rs.Name(), kt.Rid)
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", rs.Name())
	}
	orderId := taskCtx.Order.SuborderID
	// run return tasks
	ev := taskCtx.Dispatcher.returner.DealRecycleOrder(kt, taskCtx.Order)
	// 记录日志，方便排查问题
	logs.Infof("recycler:logics:cvm:ReturningState:end, subOrderID: %s, ev: %+v, rid: %s",
		orderId, cvt.PtrToVal(ev), kt.Rid)
	return rs.UpdateState(ctx, ev)
}

func (rs *ReturningState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	isFinished := false
	var shortRentalReturnedStatus enumor.ShortRentalStatus
	switch ev.Type {
	case event.ReturnSuccess:
		update["stage"] = table.RecycleStageReturnPlan
		update["status"] = table.RecycleStatusReturningPlan
		update["handler"] = "AUTO"
		isFinished = true
		shortRentalReturnedStatus = enumor.ShortRentalStatusDone
	case event.ReturnFailed:
		update["status"] = table.RecycleStatusReturnFailed
		if ev.Error != nil {
			update["message"] = ev.Error.Error()
		}
	case event.ReturnHandling:
		logs.Infof("recycle return order is handling, subOrderId: %s, type: %s", order.SuborderID, ev.Type)
		// CRP侧未更新，不修改DB中Order状态
		return nil
	default:
		logs.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
		return fmt.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}
	if isFinished {
		tmpKit := core.NewBackendKit()
		// 根据回收子订单ID更新短租回收的状态
		if err := rs.ShortRentalLogic.UpdateReturnedStatusBySubOrderID(tmpKit, order.SuborderID,
			shortRentalReturnedStatus); err != nil {
			logs.Errorf("failed to update short rental returned record status, subOrderID: %s, err: %v, rid: %s",
				order.SuborderID, err, tmpKit.Rid)
			return fmt.Errorf("failed to terminate order %s, err:%v", order.SuborderID, err)
		}
	}

	return nil
}
