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

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
)

// recoverTransitedOrders 恢复状态为RecycleStatusTransiting的待回收订单
func (r *recycleRecoverer) recoverTransitedOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	ev := &event.Event{Type: event.TransitSuccess}
	hostCount, err := r.getRecycleHostsCount(kt, order.SuborderID, table.RecycleStatusTransiting)
	if err != nil {
		logs.Errorf("failed to get host count by suborderId and status, err: %v, suborderId: %s, status: %s, rid: %s",
			err, order.SuborderID, table.RecycleStatusTransiting, kt.Rid)
		ev = &event.Event{Type: event.TransitFailed, Error: err}
	}
	// 未转移主机加入队列处理
	if hostCount == 0 {
		logs.Infof("no host transit, add to queue to retransit, suborderId: %s, rid: %s", order.SuborderID, kt.Rid)
		r.recyclerIf.GetDispatcher().Add(order.SuborderID)
		return nil
	}
	// 根据恢复结果设置下一步状态
	if ev.Type == event.TransitSuccess {
		ev = r.dealTransitingOrder(kt, order)
	}

	task, taskCtx := r.newTask(order)
	if err = task.State.UpdateState(taskCtx, ev); err != nil {
		logs.Errorf("failed to update state and set next status, subOrderId: %s, err: %v, rid: %s", order.SuborderID,
			err, kt.Rid)
		return err
	}
	return nil
}

// dealTransitingOrder 恢复正在转移主机的回收订单
func (r *recycleRecoverer) dealTransitingOrder(kt *kit.Kit, order *table.RecycleOrder) *event.Event {
	// get hosts by orderId
	hosts, err := r.getRecycleHosts(kt, order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts by subOrderId, subOrderId: %s, err: %v, rid: %s", order.SuborderID,
			err, kt.Rid)
		return &event.Event{Type: event.TransitFailed, Error: err}
	}

	var ev *event.Event
	switch order.ResourceType {
	case table.ResourceTypeCvm:
		ev = r.recyclerIf.TransitCvm(order, hosts)
	case table.ResourceTypePm:
		ev = r.recoverTransitPM(kt, order, hosts)
	case table.ResourceTypeOthers:
		ev = r.recoverTransitOthers(kt, order, hosts)
	default:
		logs.Errorf("recover: transiting hosts failed, failed to deal transit task for unknown resource type, "+
			"subOrderId: %s, resourceType: %s, rid: %s", order.SuborderID, order.ResourceType, kt.Rid)
		ev = &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("recover: transiting hosts failed, failed to deal transit task for unknown resource type,"+
				" subOrderId: %s, resourceType: %s", order.SuborderID, order.ResourceType),
		}
	}
	logs.Infof("finish recover recycle order, subOrderId: %s, event: %+v, rid: %s", order.SuborderID, *ev, kt.Rid)
	return ev
}

// recoverTransitOthers recover transit CVM resource
func (r *recycleRecoverer) recoverTransitOthers(kt *kit.Kit, order *table.RecycleOrder,
	hosts []*table.RecycleHost) *event.Event {

	return r.dealRegularTransit(kt, order, hosts)
}

// recoverTransitPM recover transit PM resource
func (r *recycleRecoverer) recoverTransitPM(kt *kit.Kit, order *table.RecycleOrder,
	hosts []*table.RecycleHost) *event.Event {
	var ev *event.Event
	switch order.RecycleType {
	case table.RecycleTypeDissolve, table.RecycleTypeExpired:
		ev = r.dealDissolveTransit(kt, order, hosts)
	case table.RecycleTypeRegular:
		ev = r.dealRegularTransit(kt, order, hosts)
	default:
		logs.Errorf("failed to deal transit task for order, unknown recycle type, orderId: %s, recycleType: %s, rid: %s",
			order.SuborderID, order.RecycleType, kt.Rid)
		ev = &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order, unknown recycle type, orderId: %s, recycleType: %s",
				order.SuborderID, order.RecycleType),
		}
	}
	return ev
}

// dealDissolveTransit recover transit PM resource which is dissolve or expired
func (r *recycleRecoverer) dealDissolveTransit(kt *kit.Kit, order *table.RecycleOrder,
	hosts []*table.RecycleHost) *event.Event {

	if len(hosts) == 0 {
		logs.Errorf("failed to get host, subOrderId: %s, rid: %s", order.SuborderID, kt.Rid)
		return &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to get host, subOrderId: %s", order.SuborderID),
		}
	}
	hostIds := make([]int64, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
	}
	findRelReq := &cmdb.HostModuleRelationParams{HostID: hostIds}
	resp, err := r.cmdbCli.FindHostBizRelations(kt, findRelReq)
	if err != nil {
		logs.Errorf("recycle: failed to get biz host relation, subOrderId: %s, err: %v, rid: %s", order.SuborderID, err,
			kt.Rid)
		return &event.Event{Type: event.TransitFailed, Error: err}
	}

	// 机器查询不到，说明转移成功
	rels := converter.PtrToVal(resp)
	if len(rels) == 0 {
		return &event.Event{Type: event.TransitSuccess}
	}

	currentHostBiz := rels[0].BizID
	currentHostModule := rels[0].BkModuleID
	// 未转移到reborn业务或cr_中转模块
	if currentHostBiz == order.BizID && currentHostModule != recovertask.CrRelayModuleId {
		if ev := r.recyclerIf.DealTransitTask2Transit(order, hosts); ev.Type == event.TransitSuccess {
			return ev
		}
		logs.Errorf("failed to deal transit task for order, subOrderId: %s, bizId: %d, rid: %s", order.SuborderID,
			currentHostBiz, kt.Rid)
		return &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order, subOrderId: %s, bizId: %d", order.SuborderID,
				currentHostBiz),
		}
	}

	srcBizId := recovertask.RebornBizId
	srcModuleId := recovertask.CrRelayModuleId
	if currentHostBiz != srcBizId || currentHostModule != srcModuleId {
		return &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order, subOrderId: %s, bizId %d", order.SuborderID,
				rels[0].BizID),
		}
	}

	if err = r.recyclerIf.TransferHost2BizTransit(kt, hosts, srcBizId, srcModuleId, order.BizID); err != nil {
		logs.Errorf("failed to transfer host to biz, err: %v, subOrderId: %s, rid: %v", err, order.SuborderID, kt.Rid)
		return &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order, subOrderId: %s, bizId %d", order.SuborderID,
				rels[0].BizID),
		}
	}
	return &event.Event{Type: event.TransitSuccess}
}

func (r *recycleRecoverer) dealRegularTransit(kt *kit.Kit, order *table.RecycleOrder,
	hosts []*table.RecycleHost) *event.Event {

	hostIds := make([]int64, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
	}
	req := &cmdb.HostModuleRelationParams{HostID: hostIds}
	resp, err := r.cmdbCli.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get biz host relation, err: %v, subOrderId: %s, rid: %s", err, order.SuborderID, kt.Rid)
		return &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to get biz host relation, err: %v, subOrderId: %s", err, order.SuborderID),
		}
	}
	// PM类型常规机器转移到reborn_数据待清理模块（cc可查询）
	rels := converter.PtrToVal(resp)
	if len(rels) != 1 {
		logs.Errorf("failed to get biz host relation, err: %v, subOrderId: %s, rid: %v", err, order.SuborderID, kt.Rid)
		return &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to get biz host relation, err: %v, subOrderId: %s", err, order.SuborderID),
		}
	}
	currentHostBiz := rels[0].BizID
	currentHostModule := rels[0].BkModuleID
	switch currentHostBiz {
	case recovertask.RebornBizId:
		// 若当前在reborn业务下的DataPendingClean模块
		if currentHostModule == recovertask.DataPendingClean {
			errUpdate := r.recyclerIf.UpdateHostInfo(order, table.RecycleStageDone, table.RecycleStatusDone)
			if errUpdate != nil {
				logs.Errorf("failed to update recycle host info, err: %v, subOrderId: %s, rid: %s", errUpdate,
					order.SuborderID, kt.Rid)
				return &event.Event{
					Type: event.TransitFailed,
					Error: fmt.Errorf("failed to update recycle host info, err: %v, subOrderId: %s", errUpdate,
						order.SuborderID),
				}
			}
			return &event.Event{Type: event.TransitSuccess}
		}
		// 若原业务是reborn且不在DataPendingClean模块，则进行转移
		if order.BizID == recovertask.RebornBizId {
			return r.recyclerIf.DealTransitTask2Pool(order, hosts)
		}
		// 若原业务不是reborn，但当前在reborn业务下非DataPendingClean模块，机器被他人转移
		logs.Errorf("host is delivered by others, subOrderId: %s, module: %d, rid: %s", order.SuborderID,
			currentHostModule, kt.Rid)
		return &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("host is delivered by others, subOrderId: %s, module: %d", order.SuborderID,
				currentHostModule),
		}
	case order.BizID:
		return r.recyclerIf.DealTransitTask2Pool(order, hosts)
	default:
		logs.Errorf("unknown biz id, subOrderId: %s, bizId: %d, rid: %s", order.SuborderID, currentHostBiz, kt.Rid)
		return &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("unknown biz id, subOrderId: %s, bizId: %d", order.SuborderID, currentHostBiz),
		}
	}

}
