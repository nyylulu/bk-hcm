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
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/dispatcher"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"

	"go.mongodb.org/mongo-driver/bson"
)

// getRecycleRunningOrders 根据订单状态及过期时间获取订单
func (r *recycleRecoverer) getRecycleRunningOrders(kt *kit.Kit, expireTime time.Time,
	recoverTime time.Time) ([]*table.RecycleOrder, error) {

	filter := map[string]interface{}{
		"status": bson.M{
			"$in": []table.RecycleStatus{
				table.RecycleStatusReturning,
				table.RecycleStatusDetecting,
				table.RecycleStatusCommitted,
				table.RecycleStatusTransiting,
				table.RecycleStatusReturningPlan,
			},
		},
		"create_at": mapstr.MapStr{
			"$gte": expireTime,
			"$lt":  recoverTime,
		},
	}

	page := metadata.BasePage{}
	orders, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get running recycle orders, err: %v, expireTime: %v, recoverTime: %v, rid: %s", err,
			expireTime, recoverTime, kt.Rid)
		return nil, err
	}
	return orders, nil
}

// getDetectTasksCount get count of detect task by orderId and status
func (r *recycleRecoverer) getDetectTasksCount(kt *kit.Kit, subOrderId string, status table.DetectStatus) (uint64,
	error) {

	filter := map[string]interface{}{
		"suborder_id": subOrderId,
		"status":      status,
	}
	cnt, err := dao.Set().DetectTask().CountDetectTask(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get detect task by subOrderId and status, err: %v, subOrderId: %s, status: %s, rid: %s",
			err, subOrderId, status, kt.Rid)
		return 0, err
	}
	return cnt, nil
}

// newTask create a recycle task to flow state
func (r *recycleRecoverer) newTask(order *table.RecycleOrder) (*dispatcher.Task, *dispatcher.CommonContext) {
	task := dispatcher.NewTask(order.Status)
	taskCtx := &dispatcher.CommonContext{
		Order:      order,
		Dispatcher: r.recyclerIf.GetDispatcher(),
	}

	return task, taskCtx
}

// getRecycleHosts get hosts by subOrderId
func (r *recycleRecoverer) getRecycleHosts(kt *kit.Kit, subOrderId string) ([]*table.RecycleHost, error) {
	filter := map[string]interface{}{
		"suborder_id": subOrderId,
	}
	recycleHosts := make([]*table.RecycleHost, 0)
	startIndex := 0
	for {
		page := metadata.BasePage{
			Start: startIndex,
			Limit: pkg.BKMaxInstanceLimit,
		}
		hosts, err := dao.Set().RecycleHost().FindManyRecycleHost(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get recycle hosts, err: %v, subOrderId: %s, rid: %s", err, subOrderId, kt.Rid)
			return nil, err
		}
		recycleHosts = append(recycleHosts, hosts...)
		if len(hosts) < pkg.BKMaxInstanceLimit {
			break
		}
		startIndex += pkg.BKMaxInstanceLimit
	}

	return recycleHosts, nil
}

// getRecycleHostsCount 获得回收主机数量
func (r *recycleRecoverer) getRecycleHostsCount(kt *kit.Kit, subOrderId string, status table.RecycleStatus) (uint64,
	error) {

	filter := map[string]interface{}{
		"suborder_id": subOrderId,
		"status":      status,
	}
	cnt, err := dao.Set().RecycleHost().CountRecycleHost(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get recycle host num, subOrderId: %s, err: %v, rid: %s", subOrderId, err, kt.Rid)
		return 0, err
	}

	return cnt, nil
}
