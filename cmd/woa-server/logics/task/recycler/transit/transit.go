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

// Package transit implements device transit station
// which deals with resource transit tasks.
package transit

import (
	"context"
	"fmt"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/thirdparty/tmpapi"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// Transit deal with device transit tasks
type Transit struct {
	cc  cmdb.Client
	tmp tmpapi.TMPClientInterface

	ctx context.Context
}

// New creates a device transit station
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client) (*Transit, error) {
	transit := &Transit{
		cc:  esbCli.Cmdb(),
		tmp: thirdCli.Tmp,
		ctx: ctx,
	}

	return transit, nil
}

// DealRecycleOrder deals with recycle order by running transit tasks
func (t *Transit) DealRecycleOrder(order *table.RecycleOrder) *event.Event {
	// init recycle host status
	stage := table.RecycleStageTransit
	status := table.RecycleStatusTransiting
	if err := t.updateHostInfo(order, stage, status); err != nil {
		logs.Errorf("failed to update recycle hosts, order id: %s, err: %v")
		return &event.Event{Type: event.DetectFailed, Error: err}
	}

	// get hosts by order id
	hosts, err := t.getRecycleHosts(order.SuborderID)
	if err != nil {
		logs.Errorf("failed to get recycle hosts by order id: %d, err: %v", order.SuborderID, err)
		return &event.Event{Type: event.TransitFailed, Error: err}
	}

	// 记录日志
	logs.Infof("recycler:logics:cvm:DealRecycleOrder:start, subOrderID: %s, resType: %s",
		order.SuborderID, order.ResourceType)

	switch order.ResourceType {
	case table.ResourceTypeCvm:
		return t.transitCvm(order, hosts)
	case table.ResourceTypePm:
		return t.transitPm(order, hosts)
	case table.ResourceTypeOthers:
		return t.transitOthers(order, hosts)
	default:
		logs.Warnf("recycler:logics:cvm:DealRecycleOrder:failed, failed to deal transit task for order %s, "+
			"for unknown resource type %s", order.SuborderID, order.ResourceType)
		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order %s, for unknown resource type %s",
				order.SuborderID, order.ResourceType),
		}
		return ev
	}
}

func (t *Transit) getRecycleHosts(orderId string) ([]*table.RecycleHost, error) {
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

// dealTransitTask2Pool deal hosts transit task and transfer host to CR resource pool
func (t *Transit) dealTransitTask2Pool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	ips := make([]string, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
		ips = append(ips, host.IP)
	}

	if len(hostIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to run transit task, " +
			"for host id list or ip list is empty")
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list or ip list is empty"),
		}
		return ev
	}

	// transfer hosts to reborn-_数据待清理
	destBiz := int64(213)
	destModule := int64(16679)
	if err := t.transferHost(hostIds, order.BizID, destBiz, destModule); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to transfer host to biz %d module %d, "+
			"err: %v", destBiz, destModule, err)
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
			Error: fmt.Errorf("failed to transfer host to biz %d module %d, err: %v", destBiz, destModule, err),
		}
		return ev
	}

	// shield TMP alarms
	if err := t.shieldTMPAlarm(ips); err != nil {
		// add shield config may fail, ignore it
		logs.Warnf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to add shield TMP alarm config, "+
			"err: %v, ips: %v", err, ips)
	}

	if errUpdate := t.updateHostInfo(order, table.RecycleStageDone, table.RecycleStatusDone); errUpdate != nil {
		logs.Errorf("failed to update recycle host info, err: %v", errUpdate)
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to update recycle host info, err: %v", errUpdate),
		}
		return ev
	}

	return &event.Event{Type: event.TransitSuccess, Error: nil}
}

// dealTransitTask2Transit deal hosts transit task and transfer host to CR transit module
func (t *Transit) dealTransitTask2Transit(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	assetIds := make([]string, 0)
	ips := make([]string, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
		assetIds = append(assetIds, host.AssetID)
		ips = append(ips, host.IP)
	}

	if len(hostIds) == 0 || len(assetIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to run transit task, " +
			"for host id list, asset id list or ip list is empty")
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list, asset id list or ip list is empty"),
		}
		return ev
	}

	// transfer hosts to reborn-_CR中转
	if err := t.transferHost2CrTransit(hostIds, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to transfer host to "+
			"CR transit module, err: %v", err)
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
			Error: fmt.Errorf("failed to transfer host to CR transit module, err: %v", err),
		}
		return ev
	}

	// shield TMP alarms
	if err := t.shieldTMPAlarm(ips); err != nil {
		// add shield config may fail, ignore it
		logs.Warnf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to add shield TMP alarm config, "+
			"err: %v, ips: %v", err, ips)
	}

	// close network or shutdown
	// TODO

	// wait 10 seconds to avoid cmdb data mess
	time.Sleep(time.Second * 10)

	// transfer hosts from reborn-_CR中转 to destBiz-CR_IEG_资源服务系统专用退回中转勿改勿删
	// reborn biz id is 213, the id of its module "_CR中转" is 5069670
	srcBizId := int64(213)
	srcModuleId := int64(5069670)
	if err := t.transferHost2BizTransit(assetIds, srcBizId, srcModuleId, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to transfer host to biz's "+
			"CR transit module in CMDB, err: %v, srcBizId: %d, bizID: %d", err, srcBizId, order.BizID)
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

	return &event.Event{Type: event.TransitSuccess, Error: nil}
}

func (t *Transit) transferHost(hostIds []int64, srcBizId, destBizId, destModuleId int64) error {
	transferReq := &cmdb.TransferHostReq{
		From: cmdb.TransferHostSrcInfo{
			FromBizID: srcBizId,
			HostIDs:   hostIds,
		},
		To: cmdb.TransferHostDstInfo{
			ToBizID:    destBizId,
			ToModuleID: destModuleId,
		},
	}

	resp, err := t.cc.TransferHost(nil, nil, transferReq)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		return fmt.Errorf("failed to transfer host to target business, code: %d, msg: %s", resp.Code, resp.ErrMsg)
	}
	return nil
}

// transferHost2CrTransit transfer hosts to CR transit module
func (t *Transit) transferHost2CrTransit(hostIds []int64, srcBizId int64) error {
	// transfer hosts to reborn-_CR中转
	destBiz := int64(213)
	destModule := int64(5069670)
	return t.transferHost(hostIds, srcBizId, destBiz, destModule)
}

// transferHost2BizTransit transfer hosts to given business's CR transit module in CMDB
func (t *Transit) transferHost2BizTransit(assetIds []string, srcBizID, srcModuleID, destBizId int64) error {
	// once 10 hosts at most
	maxNum := 10
	begin := 0
	end := begin
	length := len(assetIds)

	for begin < length {
		end += maxNum
		if end > length {
			end = length
		}

		req := &cmdb.CrTransitReq{
			From: cmdb.CrTransitSrcInfo{
				FromBizID:    srcBizID,
				FromModuleID: srcModuleID,
				AssetIDs:     assetIds[begin:end],
			},
			To: cmdb.CrTransitDstInfo{
				ToBizID: destBizId,
			},
		}

		resp, err := t.cc.Hosts2CrTransit(nil, nil, req)
		if err != nil {
			logs.Errorf("recycler:logics:cvm:transferHost2BizTransit:failed, failed to transfer host to "+
				"CR transit module, err: %v, req: %+v", err, cvt.PtrToVal(req))
			return fmt.Errorf("failed to transfer host to CR transit module, err: %v", err)
		}

		if resp.Result == false || resp.Code != 0 {
			logs.Errorf("recycler:logics:cvm:transferHost2BizTransit:failed, failed to transfer host to "+
				"CR transit module, code: %d, msg: %s", resp.Code, resp.ErrMsg)
			return fmt.Errorf("failed to transfer host to CR transit module, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		}

		begin = end
	}

	return nil
}

// shieldTMPAlarm add shield TMP alarm config
func (t *Transit) shieldTMPAlarm(ips []string) error {
	shieldStart := time.Now().Format("2006-01-02 15:04")
	shieldEnd := time.Now().Add(time.Hour * 4).Format("2006-01-02 15:04")

	// once 100 hosts at most
	maxNum := 100
	length := len(ips)
	for i := 0; i < length; i += maxNum {
		begin := i
		end := i + maxNum
		if end > length {
			end = length
		}

		req := &tmpapi.AddShieldReq{
			Method: tmpapi.AddShieldMethod,
			Params: &tmpapi.AddShieldParams{
				Ip:          ips[begin:end],
				Operator:    tmpapi.OperatorCr,
				OIp:         cc.WoaServer().Network.BindIP,
				Reason:      "",
				ShieldStart: shieldStart,
				ShieldEnd:   shieldEnd,
			},
		}

		resp, err := t.tmp.AddShieldConfig(nil, nil, req)
		if err != nil {
			// add shield config may fail, ignore it
			logs.Warnf("failed to add shield TMP alarm config, err: %v", err)
			continue
		}

		if resp.Code != 0 {
			// add shield config may fail, ignore it
			logs.Warnf("failed to add shield TMP alarm config, code: %d, msg: %s", resp.Code, resp.Msg)
		}
	}

	return nil
}

func (t *Transit) updateHostInfo(order *table.RecycleOrder, stage table.RecycleStage,
	status table.RecycleStatus) error {

	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	now := time.Now()
	update := mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": now,
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, order id: %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
