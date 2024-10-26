/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package recycle

import (
	"fmt"
	"sync"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recoverDetectedOrder 恢复状态为RecycleStatusDetecting的回收订单
func (r *recycleRecoverer) recoverDetectedOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	logs.Infof("recycle recover: start recover detect recycle order, suborderId: %s, rid: %s", order.SuborderID, kt.Rid)

	ev := &event.Event{Type: event.DetectSuccess}
	hostCount, err := r.getRecycleHostsCount(kt, order.SuborderID, table.RecycleStatusDetecting)
	if err != nil {
		logs.Errorf("failed to get host count by suborderId and status, err: %v, subOrderId: %s, status: %s, rid: %s",
			err, order.SuborderID, table.RecycleStatusDetecting, kt.Rid)
		ev = &event.Event{Type: event.DetectFailed, Error: err}
	}
	// 正在执行检测任务的任务数
	taskCount, err := r.getDetectTaskCount(kt, order.SuborderID, table.DetectStatusRunning)
	if err != nil {
		logs.Errorf("failed to get detect task count by suborderId and status, err: %v, subOrderId: %s, status: %s, "+
			"rid: %s", err, order.SuborderID, table.RecycleStatusDetecting, kt.Rid)
		ev = &event.Event{Type: event.DetectFailed, Error: err}
	}
	// 未开始执行检测任务，将订单加入队列重新执行
	if hostCount == 0 || taskCount == 0 {
		logs.Infof("recycle detect: no host is being detected and the detection process can be reentered, "+
			"subOrderId: %s, rid: %s", order.SuborderID, kt.Rid)
		r.recyclerIf.GetDispatcher().Add(order.SuborderID)
		return nil
	}

	if err = r.dealDetectingTask(kt, order); err != nil {
		logs.Errorf("failed to recover detecting task, subOrderId: %s, err: %v, rid: %s", order.SuborderID, err, kt.Rid)
		ev = &event.Event{Type: event.DetectFailed, Error: err}
	}

	task, taskCtx := r.newTask(order)
	if ev.Type == event.DetectSuccess {
		if err := r.recyclerIf.CheckDetectStatus(order.SuborderID); err != nil {
			logs.Errorf("failed to check detect task status, subOrderId: %s, err: %v, rid: %s", order.SuborderID, err,
				kt.Rid)
			ev = &event.Event{Type: event.DetectFailed, Error: err}
		}
	}

	// 设置下一步状态
	if err = task.State.UpdateState(taskCtx, ev); err != nil {
		logs.Errorf("failed to update state and set next status task by suborderId: %s, err: %v, rid: %s",
			order.SuborderID, err, kt.Rid)
		return err
	}
	logs.Infof("finish recover detect recycle order, subOrderId: %s, event: %+v, rid: %s", order.SuborderID, *ev,
		kt.Rid)
	return nil
}

// dealDetectingTask 恢复detecting状态回收订单
func (r *recycleRecoverer) dealDetectingTask(kt *kit.Kit, order *table.RecycleOrder) error {
	// 更新状态以重入检测步骤
	if err := r.updateTasksInit(kt, order.SuborderID); err != nil {
		logs.Errorf("failed to update detectTask status to init, subOrderId: %s, err: %v, rid: %s", order.SuborderID,
			err, kt.Rid)
		return err
	}

	detectingTasks, err := r.getDetectTasks(kt, order.SuborderID, table.DetectStatusInit)
	if err != nil {
		logs.Errorf("failed to get detectTasks, subOrderId: %s, status: %s, err: %v, rid: %s", order.SuborderID,
			table.DetectStatusInit, err, kt.Rid)
		return err
	}

	// run recycle tasks
	wg := sync.WaitGroup{}
	for _, detectTask := range detectingTasks {
		wg.Add(1)
		go func(task *table.DetectTask) {
			defer wg.Done()
			logs.Infof("start to recover detectTask, taskID: %s, subOrderId: %s, rid: %s", task.TaskID, task.SuborderID,
				kt.Rid)
			if err := r.recoverDetectTask(kt, task); err != nil {
				logs.Errorf("failed to recover detect detectTask, subOrderId: %s, taskID: %s, err: %v, rid: %s",
					order.SuborderID, task.TaskID, err, kt.Rid)
				return
			}
			logs.Infof("success to recover detectTask, taskID: %s, subOrderId: %s, rid: %s", task.TaskID,
				task.SuborderID, kt.Rid)
		}(detectTask)
	}
	wg.Wait()

	return nil
}

// recoverDetectTask 恢复检测步骤
func (r *recycleRecoverer) recoverDetectTask(kt *kit.Kit, task *table.DetectTask) error {
	// 更新检测步骤状态以便再次重入检测步骤
	if err := r.updateStepInit(kt, task); err != nil {
		logs.Errorf("failed to update detect task status to init, taskID: %s, err: %v, rid: %s", task.TaskID, err,
			kt.Rid)
		return fmt.Errorf("failed to update detect task status to init, taskID: %s, err: %v, rid: %s", task.TaskID, err,
			kt.Rid)
	}

	dealedStepNum := task.SuccessNum + task.FailedNum
	// 未开始执行检测步骤
	if dealedStepNum == 0 {
		r.recyclerIf.RunRecycleTask(task, 0)
		return nil
	}
	// 未执行完成检测步骤
	if dealedStepNum <= task.TotalNum {
		r.recyclerIf.RunRecycleTask(task, dealedStepNum)
		return nil
	}

	logs.Errorf("too many detect task steps, subOrderId: %s, dealedStepNum: %d, total: %d, rid: %s", task.SuborderID,
		dealedStepNum, task.TotalNum, kt.Rid)
	return fmt.Errorf("too many detect task steps, subOrderId: %s, rid: %s", task.SuborderID, kt.Rid)
}

// updateStepInit update step status to init to start detect step
func (r *recycleRecoverer) updateStepInit(kt *kit.Kit, task *table.DetectTask) error {
	stepId := task.SuccessNum + task.FailedNum + 1
	filter := &mapstr.MapStr{
		"step_id":     stepId,
		"suborder_id": task.SuborderID,
		"status":      table.DetectStatusRunning,
	}
	doc := &mapstr.MapStr{
		"status":    table.DetectStatusInit,
		"message":   "init",
		"update_at": time.Now(),
	}

	if err := dao.Set().DetectStep().UpdateDetectStep(kt.Ctx, filter, doc); err != nil {
		logs.Errorf("failed to update recycle step, err: %v, step id: %s, rid: %s", err, stepId, kt.Rid)
		return err
	}
	return nil
}

// getDetectTasks 获取检测任务
func (r *recycleRecoverer) getDetectTasks(kt *kit.Kit, subOrderId string,
	status table.DetectStatus) ([]*table.DetectTask, error) {

	filter := map[string]interface{}{
		"suborder_id": subOrderId,
		"status":      status,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}

	tasks, err := dao.Set().DetectTask().FindManyDetectTask(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get detectTasks by subOrderId and status, subOrderId: %s, err: %v, rid: %s", subOrderId,
			err, kt.Rid)
		return nil, err
	}
	return tasks, nil
}

// getDetectTaskCount get count of detect task by orderId and status
func (r *recycleRecoverer) getDetectTaskCount(kt *kit.Kit, subOrderId string, status table.DetectStatus) (uint64,
	error) {
	filter := map[string]interface{}{
		"suborder_id": subOrderId,
		"status":      status,
	}

	cnt, err := dao.Set().DetectTask().CountDetectTask(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("get detect task count failed, err: %v, subOrderId: %s, status: %s, rid: %s", err, subOrderId,
			status, kt.Rid)
		return 0, err
	}
	return cnt, nil
}

func (r *recycleRecoverer) updateTasksInit(kt *kit.Kit, subOrderId string) error {
	filter := &mapstr.MapStr{
		"suborder_id": subOrderId,
		"status":      table.DetectStatusRunning,
	}
	doc := &mapstr.MapStr{
		"status":    table.DetectStatusInit,
		"message":   "init",
		"update_at": time.Now(),
	}

	if err := dao.Set().DetectTask().UpdateDetectTasks(kt, filter, doc); err != nil {
		logs.Errorf("failed to update recycle detect task, subOrderId: %s, err: %v, rid: %s", subOrderId, err, kt.Rid)
		return err
	}
	return nil
}
