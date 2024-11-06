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

// Package transit implements the transit module
package transit

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// TransitCvm deals with cvm transit task
func (t *Transit) TransitCvm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	kt := core.NewBackendKit()
	crpReturnHosts := make([]*table.RecycleHost, 0)
	resPoolReturnHosts := make([]*table.RecycleHost, 0)
	for _, host := range hosts {
		// 转移“回收方式”为[默认、CRP]的
		if host.ReturnedWay == "" || host.ReturnedWay == enumor.CrpReturnedWay {
			crpReturnHosts = append(crpReturnHosts, host)
			continue
		}
		// 转移“回收方式”为[资源池退还]的
		if host.ReturnedWay == enumor.ResourcePoolReturnedWay {
			resPoolReturnHosts = append(resPoolReturnHosts, host)
			continue
		}
	}

	// 记录回收转移记录日志
	logs.Infof("recycler:logics:cvm:transitCvm:start, subOrderId: %s, crpReturnHosts: %+v, resPoolReturnHosts: %+v, "+
		"rid: %s", order.SuborderID, cvt.PtrToSlice(crpReturnHosts), cvt.PtrToSlice(resPoolReturnHosts), kt.Rid)

	ev := &event.Event{
		Type:  event.TransitFailed,
		Error: fmt.Errorf("failed to transfer host to cc, no hosts to transit, subOrderID: %s", order.SuborderID),
	}

	// 转移到资源池
	if len(resPoolReturnHosts) > 0 {
		ev = t.dealTransitTask2ResourcePool(order, resPoolReturnHosts)
		logs.Infof("recycler:logics:cvm:transitCvm:resourcePool:end, transit to resource pool, subOrderId: %s, "+
			"resPoolReturnHosts: %+v, event: %+v, rid: %s", order.SuborderID, cvt.PtrToSlice(resPoolReturnHosts),
			cvt.PtrToVal(ev), kt.Rid)
	}

	if len(crpReturnHosts) == 0 {
		return ev
	}

	return t.dealTransitTask2TransitCvm(order, crpReturnHosts)
}

// dealTransitTask2Transit deal hosts transit task and transfer host to CR transit module for cvm
func (t *Transit) dealTransitTask2TransitCvm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	assetIds := make([]string, 0)
	ips := make([]string, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
		assetIds = append(assetIds, host.AssetID)
		ips = append(ips, host.IP)
	}
	if len(hostIds) == 0 || len(assetIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to run transit task, "+
			"for host id list, asset id list or ip list is empty, order.SuborderID: %s", order.SuborderID)
		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list, asset id list or ip list is empty,"+
				" order.SuborderID: %s", order.SuborderID),
		}
		return ev
	}

	// get biz's module "待回收"
	moduleID, err := t.cc.GetBizRecycleModuleID(nil, nil, order.BizID)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to get biz %d recycle module id, "+
			"err: %v, subOrderId: %s", order.BizID, err, order.SuborderID)
		if errUpdate := t.UpdateHostInfo(order, table.RecycleStageTransit,
			table.RecycleStatusTransitFailed); errUpdate != nil {
			logs.Errorf("failed to update recycle host info, err: %v, order.SuborderID: %s", errUpdate,
				order.SuborderID)
			ev := &event.Event{
				Type: event.TransitFailed,
				Error: fmt.Errorf("failed to update recycle host info, err: %v, order.SuborderID: %s", errUpdate,
					order.SuborderID),
			}
			return ev
		}

		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to get biz %d recycle module id, err: %v, order.SuborderID: %s", order.BizID, err,
				order.SuborderID),
		}
		return ev
	}

	// transfer hosts to module CR_IEG_资源服务系统专用退回中转勿改勿删 in the origin biz
	if err = t.TransferHost2BizTransit(hosts, order.BizID, moduleID, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to transfer host "+
			"to biz's CR transit module in CMDB, err: %v, order.SuborderID: %s", err, order.SuborderID)
		if errUpdate := t.UpdateHostInfo(order, table.RecycleStageTransit,
			table.RecycleStatusTransitFailed); errUpdate != nil {
			logs.Errorf("failed to update recycle host info, err: %v, order.SuborderID: %s", errUpdate,
				order.SuborderID)
			ev := &event.Event{
				Type: event.TransitFailed,
				Error: fmt.Errorf("failed to update recycle host info, err: %v, order.SuborderID: %s", errUpdate,
					order.SuborderID),
			}
			return ev
		}

		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to transfer host to biz's CR transit module in CMDB, err: %v, "+
				"order.SuborderID: %s", err, order.SuborderID),
		}
		return ev
	}

	// shield TMP alarms
	if err = t.shieldTMPAlarm(ips); err != nil {
		// add shield config may fail, ignore it
		logs.Warnf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to add shield TMP alarm config, "+
			"err: %v, ips: %v", err, ips)
	}

	// close network or shutdown
	// TODO

	return &event.Event{Type: event.TransitSuccess, Error: nil}
}

// dealTransitTask2ResourcePool 转移主机到指定的资源池(reborn业务、数据待清理模块)
func (t *Transit) dealTransitTask2ResourcePool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	assetIds := make([]string, 0)
	ips := make([]string, 0)
	isTransit := false
	for _, host := range hosts {
		// 只转移“回收方式”为[资源池退还]的
		if host.ReturnedWay != enumor.ResourcePoolReturnedWay {
			continue
		}
		isTransit = true
		hostIds = append(hostIds, host.HostID)
		assetIds = append(assetIds, host.AssetID)
		ips = append(ips, host.IP)
	}
	if !isTransit {
		logs.Warnf("recycler:logics:cvm:dealTransitTask2ResourcePool:skip, no host need to transfer pool, "+
			"suborderID: %s, hosts: %+v", order.SuborderID, cvt.PtrToSlice(hosts))
		return &event.Event{Type: event.TransitSuccess, Error: nil}
	}

	if len(hostIds) == 0 || len(assetIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2ResourcePool:failed, failed to run transit task, "+
			"for host id list, asset id list or ip list is empty, suborderID: %s", order.SuborderID)
		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list, asset id list or ip list is empty,"+
				" order.SuborderID: %s", order.SuborderID),
		}
		return ev
	}

	ev := t.dealRollingServerTransit2Pool(order, hosts)
	// 记录转移日志
	logs.Infof("recycler:logics:cvm:dealTransitTask2ResourcePool:end, suborderID: %s, hosts: %+v, ev: %+v",
		order.SuborderID, cvt.PtrToSlice(hosts), cvt.PtrToVal(ev))
	return ev
}

// dealRollingServerTransit2Pool deal rolling server hosts transit task and transfer host to CR resource pool
func (t *Transit) dealRollingServerTransit2Pool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	ips := make([]string, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
		ips = append(ips, host.IP)
	}

	if len(hostIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealRollingServerTransit2Pool:failed, failed to run transit resource pool "+
			"task,for host id list or ip list is empty, subOrderID: %s", order.SuborderID)
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to run transit resource pool task, for host id list or ip list is empty"),
		}
		return ev
	}

	// transfer hosts to reborn-_数据待清理
	destBiz := recovertask.RebornBizId
	destModule := recovertask.DataToCleanedModule
	if err := t.TransferHost(hostIds, order.BizID, destBiz, destModule); err != nil {
		logs.Errorf("recycler:logics:cvm:dealRollingServerTransit2Pool:failed, failed to transfer host to biz %d "+
			"module %d, subOrderID: %s, err: %v", destBiz, destModule, order.SuborderID, err)
		if errUpdate := t.UpdateHostInfo(order, table.RecycleStageTransit,
			table.RecycleStatusTransitFailed); errUpdate != nil {
			logs.Errorf("failed to update recycle host info, subOrderID: %s, err: %v", order.SuborderID, errUpdate)
			ev := &event.Event{
				Type:  event.TransitFailed,
				Error: fmt.Errorf("failed to update recycle host info, err: %v", errUpdate),
			}
			return ev
		}

		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to transfer host to biz %d module %d, err: %v", destBiz, destModule, err),
		}
		return ev
	}

	// shield TMP alarms
	if err := t.shieldTMPAlarm(ips); err != nil {
		// add shield config may fail, ignore it
		logs.Warnf("recycler:logics:cvm:dealRollingServerTransit2Pool:failed, failed to add shield TMP alarm config, "+
			"subOrderID: %s, err: %v, ips: %v", order.SuborderID, err, ips)
	}

	return &event.Event{Type: event.TransitSuccess, Error: nil}
}
