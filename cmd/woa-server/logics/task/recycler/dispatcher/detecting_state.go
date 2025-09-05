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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// DetectingState the action to be executed in detecting state
type DetectingState struct{}

// Name return the name of detecting state
func (ds *DetectingState) Name() table.RecycleStatus {
	return table.RecycleStatusDetecting
}

// UpdateState update next state
func (ds *DetectingState) UpdateState(ctx EventContext, ev *event.Event) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			ds.Name())
		return fmt.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			ds.Name())
	}

	if errUpdate := ds.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order state, subOrderId: %s, err: %v", taskCtx.Order.SuborderID,
			errUpdate)
		return errUpdate
	}

	// need not dispatch if next state is audit
	if ev.Type == event.DetectSuccess && taskCtx.Order.ResourceType == table.ResourceTypePm &&
		taskCtx.Order.RecycleType == table.RecycleTypeRegular && taskCtx.Order.TotalNum > 10 {
		return nil
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID,
			ds.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, ds.Name())
	}

	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)
	// 记录日志
	logs.Infof("recycler: success detect state, subOrderId: %s", taskCtx.Order.SuborderID)
	return nil
}

// Execute executes action in detecting state
func (ds *DetectingState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context")
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
	}
	orderId := taskCtx.Order.SuborderID
	kt := core.NewBackendKit()
	kt.Ctx = taskCtx.Dispatcher.ctx
	logs.Infof("DetectingState: start detect order: %s, rid: %s", orderId, kt.Rid)
	ev := ds.dealDetectTask(kt, taskCtx)
	logs.Infof("DetectingState: end detect order: %s, ev: %+v, rid: %s", orderId, cvt.PtrToVal(ev), kt.Rid)

	return ds.UpdateState(taskCtx, ev)
}

func (ds *DetectingState) dealDetectTask(kt *kit.Kit, ctx *CommonContext) *event.Event {
	orderId := ctx.Order.SuborderID

	// init recycle host status
	stage := table.RecycleStageDetect
	status := table.RecycleStatusDetecting
	if err := ds.updateHostInfo(orderId, stage, status); err != nil {
		logs.Errorf("failed to update recycle hosts, order id: %s, err: %v, rid: %s", orderId, err, kt.Rid)
		return &event.Event{Type: event.DetectFailed, Error: err}
	}
	// run detection tasks
	if err := ctx.Dispatcher.detector.Detect(kt, ctx.Order); err != nil {
		logs.Errorf("failed to run detection tasks, err: %v, rid: %s", err, kt.Rid)
		return &event.Event{Type: event.DetectFailed, Error: err}
	}

	if err := ctx.Dispatcher.detector.CheckDetectStatus(kt, orderId); err != nil {
		logs.Errorf("recycle detection failed, order id: %s, err: %v, rid: %s", orderId, err, kt.Rid)
		return &event.Event{Type: event.DetectFailed, Error: err}
	}

	return &event.Event{Type: event.DetectSuccess, Error: nil}
}

func (ds *DetectingState) updateHostInfo(orderId string, stage table.RecycleStage,
	status table.RecycleStatus) error {

	filter := mapstr.MapStr{
		"suborder_id": orderId,
	}

	update := mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, order id: %s, err: %v", orderId, err)
		return err
	}

	return nil
}

func (ds *DetectingState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	switch ev.Type {
	case event.DetectSuccess:
		if order.ResourceType == table.ResourceTypePm && order.RecycleType == table.RecycleTypeRegular &&
			order.TotalNum > 10 {
			update["stage"] = table.RecycleStageAudit
			update["status"] = table.RecycleStatusAudit
			update["handler"] = recovertask.Handler
		} else {
			update["stage"] = table.RecycleStageTransit
			update["status"] = table.RecycleStatusTransiting
		}
		// 清空之前产生的message，比如检测失败，避免影响后续流程展示
		update["message"] = ""
	case event.DetectFailed:
		update["stage"] = table.RecycleStageDetect
		update["status"] = table.RecycleStatusDetectFailed
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
