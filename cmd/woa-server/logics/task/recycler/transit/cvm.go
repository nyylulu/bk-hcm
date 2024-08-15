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
	"hcm/pkg/logs"
)

// transitCvm deals with cvm transit task
func (t *Transit) transitCvm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	return t.dealTransitTask2TransitCvm(order, hosts)
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
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to run transit task, " +
			"for host id list, asset id list or ip list is empty")
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list, asset id list or ip list is empty"),
		}
		return ev
	}

	// get biz's module "待回收"
	moduleID, err := t.cc.GetBizRecycleModuleID(nil, nil, order.BizID)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to get biz %d recycle module id, "+
			"err: %v", order.BizID, err)
		if errUpdate := t.updateHostInfo(order, table.RecycleStageTransit,
			table.RecycleStatusTransitFailed); errUpdate != nil {
			logs.Errorf("failed to update recycle host info, err: %v", errUpdate)
			ev := &event.Event{
				Type:  event.TransitFailed,
				Error: fmt.Errorf("failed to update recycle host info, err: %v", errUpdate),
			}
			return ev
		}

		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to get biz %d recycle module id, err: %v", order.BizID, err),
		}
		return ev
	}

	// transfer hosts to module CR_IEG_资源服务系统专用退回中转勿改勿删 in the origin biz
	if err = t.transferHost2BizTransit(assetIds, order.BizID, moduleID, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2TransitCvm:failed, failed to transfer host "+
			"to biz's CR transit module in CMDB, err: %v", err)
		if errUpdate := t.updateHostInfo(order, table.RecycleStageTransit,
			table.RecycleStatusTransitFailed); errUpdate != nil {
			logs.Errorf("failed to update recycle host info, err: %v", errUpdate)
			ev := &event.Event{
				Type:  event.TransitFailed,
				Error: fmt.Errorf("failed to update recycle host info, err: %v", errUpdate),
			}
			return ev
		}

		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to transfer host to biz's CR transit module in CMDB, err: %v", err),
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
