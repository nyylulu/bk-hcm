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

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recoverReturnedOrders 恢复RecycleStatusReturning状态订单
func (r *recycleRecoverer) recoverReturnedOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	ev := &event.Event{Type: event.ReturnSuccess}
	cnt, err := r.getReturnTaskCount(kt, order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get return task count, subOrderId: %s, err: %v, rid: %s", order.SuborderID, err, kt.Rid)
		ev = &event.Event{Type: event.ReturnFailed, Error: err}
	}
	// 未创建return任务，加入queue重入return流程
	if cnt == 0 {
		r.recyclerIf.GetDispatcher().Add(order.SuborderID)
		return nil
	}
	returnTask, err := r.getReturnTask(kt, order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get return task, subOrderId: %s, err: %v, rid: %s ", order.SuborderID, err, kt.Rid)
		ev = &event.Event{Type: event.ReturnFailed, Error: err}
	}

	hosts, err := r.getRecycleHosts(kt, order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts by subOrderId: %s, err: %v, rid: %s", order.SuborderID, err, kt.Rid)
		ev = &event.Event{Type: event.ReturnFailed, Error: err}
	}

	if ev.Type == event.ReturnSuccess {
		ev = r.recoverReturnedStatus(kt, order, returnTask, hosts)
	}

	task, taskCtx := r.newTask(order)
	if err = task.State.UpdateState(taskCtx, ev); err != nil {
		logs.Errorf("failed to update state and set next status suborderId: %s, err: %v, rid: %s",
			returnTask.SuborderID, err, kt.Rid)
		return err
	}
	logs.Infof("finish recover return recycle order, subOrderId: %s, event: %+v, rid: %s", order.SuborderID, *ev,
		kt.Rid)
	return nil
}

func (r *recycleRecoverer) recoverReturnedStatus(kt *kit.Kit, order *table.RecycleOrder,
	returnTask *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {

	var ev *event.Event
	switch returnTask.Status {
	case table.ReturnStatusInit:
		ev = r.recoverReturnHosts(kt, returnTask, hosts)
	case table.ReturnStatusRunning:
		ev = r.recoverQueryReturnStatus(returnTask, hosts)
	case table.ReturnStatusSuccess:
		ev = &event.Event{Type: event.ReturnSuccess}
	case table.ReturnStatusFailed:
		ev = &event.Event{
			Type:  event.ReturnFailed,
			Error: fmt.Errorf("return task is failed, subOrderId: %s", order.SuborderID),
		}
	default:
		logs.Errorf("failed to deal return task, subOrderId: %s, unknown status: %s, rid: %s", returnTask.SuborderID,
			returnTask.Status, kt.Rid)
		ev = &event.Event{
			Type: event.ReturnFailed,
			Error: fmt.Errorf("failed to deal return task, subOrderId: %s, unknown status: %s, rid: %s",
				returnTask.SuborderID, returnTask.Status, kt.Rid),
		}
	}

	return ev
}

// RecoverQueryReturnStatus 查询return任务状态
func (r *recycleRecoverer) recoverQueryReturnStatus(returnTask *table.ReturnTask,
	hosts []*table.RecycleHost) *event.Event {

	return r.recyclerIf.QueryReturnStatus(returnTask, hosts)
}

// RecoverReturnHosts recover hosts which status is ReturnStatusInit
func (r *recycleRecoverer) recoverReturnHosts(kt *kit.Kit, returnTask *table.ReturnTask,
	hosts []*table.RecycleHost) *event.Event {

	switch returnTask.ResourceType {
	case table.ResourceTypeCvm:
		return r.recyclerIf.RecoverReturnCvm(kt, returnTask, hosts)
	case table.ResourceTypePm:
		msg := "failed to recover recycle, can not get pm device return task id"
		if err := r.recyclerIf.UpdateReturnTaskInfo(kt.Ctx, returnTask, "", table.ReturnStatusFailed, msg); err != nil {
			logs.Errorf("failed to update pm return task info, subOrderId: %s, err: %v, rid: %s", returnTask.SuborderID,
				err, kt.Rid)
			return &event.Event{Type: event.ReturnFailed, Error: err}
		}

		if errUpdate := r.recyclerIf.UpdateOrderInfo(kt, returnTask.SuborderID, recovertask.Handler, 0,
			uint(len(hosts)), 0, msg); errUpdate != nil {
			logs.Errorf("recycler: failed to update pm recycle order, subOrderId: %s, err: %v, rid: %s",
				returnTask.SuborderID, errUpdate, kt.Rid)
			return &event.Event{Type: event.ReturnFailed, Error: errUpdate}
		}

		return &event.Event{
			Type: event.ReturnFailed,
			Error: fmt.Errorf("recover recycle failed, can not get pm device return task id, subOrderId: %s",
				returnTask.SuborderID),
		}
	default:
		return &event.Event{
			Type: event.ReturnFailed,
			Error: fmt.Errorf("failed to recover recycle return order, unsupported resource type: %s, subOrderId: %s",
				returnTask.ResourceType, returnTask.SuborderID),
		}
	}
}

// getReturnTask get return task by order id
func (r *recycleRecoverer) getReturnTask(kt *kit.Kit, subOrderId string) (*table.ReturnTask, error) {
	filter := &mapstr.MapStr{
		"suborder_id": subOrderId,
	}

	returnTask, err := dao.Set().ReturnTask().GetReturnTask(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get return task for order, subOrderId: %s, err: %v, rid: %s", subOrderId, err, kt.Rid)
		return nil, err
	}
	return returnTask, nil
}

// getReturnTaskCount get count of return task by order id
func (r *recycleRecoverer) getReturnTaskCount(kt *kit.Kit, suborderID string) (uint64, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderID,
	}
	cnt, err := dao.Set().ReturnTask().CountReturnTask(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to count return task for order, subOrderId: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return 0, err
	}
	return cnt, nil
}
