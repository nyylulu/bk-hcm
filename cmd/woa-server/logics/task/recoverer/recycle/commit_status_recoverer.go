/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recoverCommittedOrder 恢复状态为RecycleStatusCommitted的回收订单
func (r *recycleRecoverer) recoverCommittedOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	ev := &event.Event{Type: event.CommitSuccess}
	count, err := r.getDetectTasksCount(kt, order.SuborderID, table.DetectStatusInit)
	if err != nil {
		logs.Errorf("failed to get detect task by suborderId and status, err: %v, status: %s, subOrderId: %s, rid: %s",
			err, table.DetectStatusInit, order.SuborderID, kt.Rid)
		ev = &event.Event{Type: event.CommitFailed, Error: err}
	}
	// 若无检测任务，则直接添加重新执行committed状态任务
	if count == 0 {
		r.recyclerIf.GetDispatcher().Add(order.SuborderID)
		return nil
	}
	logs.Infof("finish recover recycle order, subOrderId: %s, rid: %s", order.SuborderID, kt.Rid)

	task, taskCtx := r.newTask(order)
	err = task.State.UpdateState(taskCtx, ev)
	if err != nil {
		logs.Errorf("failed to update order status, err: %v, suborderId: %s, rid: %s", err, order.SuborderID, kt.Rid)
		return fmt.Errorf("failed to update order status, err: %v, suborderId: %s", err, order.SuborderID)
	}
	logs.Infof("finish to recover COMMITTED order, the first step of recycle, suborderId: %s, event: %+v, rid: %s",
		order.SuborderID, *ev, kt.Rid)
	return nil
}
