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

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/logs"
)

// CommittedState the action to be executed in committed state
type CommittedState struct{}

// Name return the name of committed state
func (cs *CommittedState) Name() table.RecycleStatus {
	return table.RecycleStatusCommitted
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

	// set next state
	if errUpdate := cs.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order %s state, err: %v", orderId, errUpdate)
		return errUpdate
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("failed to add order to dispatch, for dispatcher is nil, order id: %s, state: %s", orderId,
			cs.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, order id: %s, state: %s", orderId,
			cs.Name())
	}

	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)

	// 记录日志
	logs.Infof("recycler:logics:cvm:CommittedState:end, orderID: %s", orderId)

	return nil
}

func (cs *CommittedState) dealCommitTask(order *table.RecycleOrder) *event.Event {
	if err := cs.initDetectTasks(order); err != nil {
		logs.Warnf("failed to init detection tasks, order %s, err: %v", order.SuborderID, err)
		return &event.Event{Type: event.CommitFailed, Error: err}
	}

	stage := table.RecycleStageCommit
	status := table.RecycleStatusCommitted

	if err := cs.updateHostInfo(order, stage, status); err != nil {
		logs.Errorf("failed to update recycle hosts, order id: %s, err: %v")
		return &event.Event{Type: event.CommitFailed, Error: err}
	}

	return &event.Event{Type: event.CommitSuccess, Error: nil}
}

func (cs *CommittedState) initDetectTasks(order *table.RecycleOrder) error {
	// init tasks and steps
	hosts, err := cs.getRecycleHosts(order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts by key %s, err: %v", order.SuborderID, err)
		return err
	}

	if err := cs.initTasks(order, hosts); err != nil {
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
		Limit: common.BKMaxInstanceLimit,
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle hosts, err: %v", err)
		return nil, err
	}

	return insts, nil
}

// initTaskAndSteps init detection tasks and steps
func (cs *CommittedState) initTasks(order *table.RecycleOrder, hosts []*table.RecycleHost) error {
	now := time.Now()
	for index, host := range hosts {
		task := &table.DetectTask{
			OrderID:    order.OrderID,
			SuborderID: order.SuborderID,
			TaskID:     fmt.Sprintf("%s-%d", order.SuborderID, index+1),
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

		if err := dao.Set().DetectTask().CreateDetectTask(context.Background(), task); err != nil {
			logs.Warnf("failed to create detection task for ip: %s", host.IP)
		}
	}

	return nil
}

func (cs *CommittedState) updateHostInfo(order *table.RecycleOrder, stage table.RecycleStage,
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
	default:
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
