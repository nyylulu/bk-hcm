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
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/cmd/woa-server/types/task"
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

	resumeReq := &task.ResumeRecycleOrderReq{
		SuborderID: []string{order.SuborderID},
	}
	err := r.recyclerIf.ResumeRecycleOrder(kt, resumeReq)
	if err != nil {
		logs.Errorf("recover failed to resume recycle order, subOrderId: %s, err: %v, rid: %s",
			order.SuborderID, err, kt.Rid)
		return err
	}
	return nil
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
