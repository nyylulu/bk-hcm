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

// Package dispatcher ...
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"

	"go.mongodb.org/mongo-driver/mongo"
)

// CommittedState the action to be executed in committed state
type CommittedState struct{}

// Name return the name of committed state
func (cs *CommittedState) Name() table.RecycleStatus {
	return table.RecycleStatusCommitted
}

// UpdateState updates state of the task and send it to the next step
func (cs *CommittedState) UpdateState(ctx EventContext, ev *event.Event) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			cs.Name())
		return fmt.Errorf("failed to convert to audit context, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			cs.Name())
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, cs.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, cs.Name())
	}

	// set next state
	if errUpdate := cs.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order state, subOrderId: %s, err: %v", taskCtx.Order.SuborderID,
			errUpdate)
		return errUpdate
	}

	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)
	// 记录日志
	logs.Infof("recycler: status COMMITTED success, first step of recycle, subOrderID: %s", taskCtx.Order.SuborderID)
	return nil
}

// Execute executes action in committed state
func (cs *CommittedState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context")
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", cs.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", cs.Name())
	}
	orderId := taskCtx.Order.SuborderID
	// 记录日志，方便排查问题
	logs.Infof("recycler:logics:cvm:CommittedState:start, orderID: %s", orderId)

	ev := cs.dealCommitTask(taskCtx.Order)

	return cs.UpdateState(taskCtx, ev)
}

func (cs *CommittedState) dealCommitTask(order *table.RecycleOrder) *event.Event {
	txnErr := dal.RunTransaction(kit.New(), func(sc mongo.SessionContext) error {
		if err := cs.initDetectTasks(sc, order); err != nil {
			logs.Errorf("failed to init detection tasks, subOrderId: %s, err: %v", order.SuborderID, err)
			return err
		}

		stage := table.RecycleStageCommit
		status := table.RecycleStatusCommitted

		if err := cs.updateHostInfo(sc, order, stage, status); err != nil {
			logs.Errorf("failed to update recycle hosts, subOrderId: %s, err: %v", order.SuborderID, err)
			return err
		}
		return nil
	})
	if txnErr != nil {
		return &event.Event{Type: event.CommitFailed, Error: txnErr}
	}

	return &event.Event{Type: event.CommitSuccess, Error: nil}
}

func (cs *CommittedState) initDetectTasks(ctx context.Context, order *table.RecycleOrder) error {
	// init tasks and steps
	hosts, err := cs.getRecycleHosts(order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts by key %s, err: %v", order.SuborderID, err)
		return err
	}

	if err := cs.initTasks(ctx, order, hosts); err != nil {
		logs.Warnf("failed to init detection tasks, err: %v", err)
		return err
	}

	return nil
}

func (cs *CommittedState) getRecycleHosts(orderId string) ([]*table.RecycleHost, error) {
	filter := map[string]interface{}{
		"suborder_id": orderId,
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKMaxInstanceLimit,
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle hosts, err: %v", err)
		return nil, err
	}

	return insts, nil
}

// initTaskAndSteps init detection tasks and steps
func (cs *CommittedState) initTasks(ctx context.Context, order *table.RecycleOrder, hosts []*table.RecycleHost) error {
	now := time.Now()

	for index, host := range hosts {
		task := &table.DetectTask{
			OrderID:    order.OrderID,
			SuborderID: order.SuborderID,
			TaskID:     fmt.Sprintf("%s-%d", order.SuborderID, index+1),
			HostID:     host.HostID,
			AssetID:    host.AssetID,
			IP:         host.IP,
			User:       order.User,
			Status:     table.DetectStatusInit,
			Message:    "",
			TotalNum:   0,
			SuccessNum: 0,
			PendingNum: 0,
			FailedNum:  0,
			CreateAt:   now,
			UpdateAt:   now,
		}

		if err := dao.Set().DetectTask().CreateDetectTask(ctx, task); err != nil {
			logs.Warnf("failed to create detection task for ip: %s", host.IP)
		}
	}

	return nil
}

func (cs *CommittedState) updateHostInfo(ctx context.Context, order *table.RecycleOrder, stage table.RecycleStage,
	status table.RecycleStatus) error {

	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(ctx, &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, order id: %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}

func (cs *CommittedState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"success_num": 0,
		"pending_num": order.TotalNum,
		"failed_num":  0,
		"update_at":   time.Now(),
	}

	switch ev.Type {
	case event.CommitSuccess:
		update["stage"] = table.RecycleStageDetect
		update["status"] = table.RecycleStatusDetecting
	case event.CommitFailed:
		update["stage"] = table.RecycleStageTerminate
		update["status"] = table.RecycleStatusTerminate
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
