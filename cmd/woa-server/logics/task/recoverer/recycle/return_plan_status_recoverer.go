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
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recoverReturningPlanOrder 恢复RecycleStatusReturningPlan状态订单
func (r *recycleRecoverer) recoverReturningPlanOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	ev := &event.Event{Type: event.ReturnSuccess}

	task, taskCtx := r.newTask(order)
	if err := task.State.Execute(taskCtx); err != nil {
		logs.Errorf("failed to execute state suborderId: %s, err: %v, rid: %s", order.SuborderID, err, kt.Rid)
		return err
	}
	logs.Infof("finish recover return recycle order, subOrderId: %s, event: %+v, rid: %s", order.SuborderID, *ev,
		kt.Rid)
	return nil
}
