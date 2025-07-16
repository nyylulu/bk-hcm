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
	"encoding/json"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	rslogics "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/tmpapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
)

// Host2CrTransitDefaultBatchSize once 100 hosts at most, if not specified, use 10
const Host2CrTransitDefaultBatchSize = 10

// Transit deal with device transit tasks
type Transit struct {
	cc      cmdb.Client
	tmp     tmpapi.TMPClientInterface
	rsLogic rslogics.Logics

	ctx    context.Context
	cliSet *client.ClientSet
}

// New creates a device transit station
func New(ctx context.Context, thirdCli *thirdparty.Client, cmdbCli cmdb.Client, rsLogic rslogics.Logics,
	cliSet *client.ClientSet) (*Transit, error) {

	transit := &Transit{
		cc:      cmdbCli,
		tmp:     thirdCli.Tmp,
		ctx:     ctx,
		rsLogic: rsLogic,
		cliSet:  cliSet,
	}

	return transit, nil
}

// getBizHost get biz hosts by bkBizID and bkHostIds
func (t *Transit) getBizHost(kt *kit.Kit, bkBizID int64, bkHostIds []int64) (*cmdb.ListBizHostResult, error) {

	var hosts = &cmdb.ListBizHostResult{}
	startIndex := 0
	for {
		params := &cmdb.ListBizHostParams{
			BizID: bkBizID,
			Fields: []string{
				"bk_host_id",
				"bk_asset_id",
			},
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: querybuilder.CombinedRule{
					Condition: querybuilder.ConditionAnd,
					Rules: []querybuilder.Rule{
						querybuilder.AtomRule{
							Field:    "bk_host_id",
							Operator: querybuilder.OperatorIn,
							Value:    bkHostIds,
						},
					},
				},
			},
			Page: &cmdb.BasePage{
				Start: int64(startIndex),
				Limit: pkg.BKMaxInstanceLimit,
			},
		}
		resp, err := t.cc.ListBizHost(kt, params)
		if err != nil {
			logs.Errorf("call cmdb to list biz host failed, bkBizID: %d, err: %v, params: %v", bkBizID, err, params)
			return nil, err
		}
		hosts.Info = append(hosts.Info, resp.Info...)
		if len(resp.Info) < pkg.BKMaxInstanceLimit {
			break
		}
		startIndex += pkg.BKMaxInstanceLimit
	}

	hosts.Count = int64(len(hosts.Info))
	return hosts, nil
}

// DealRecycleOrder deals with recycle order by running transit tasks
func (t *Transit) DealRecycleOrder(order *table.RecycleOrder) *event.Event {

	// init recycle host status
	stage := table.RecycleStageTransit
	status := table.RecycleStatusTransiting
	if err := t.UpdateHostInfo(order, stage, status); err != nil {
		logs.Errorf("failed to update recycle hosts, subOrderId: %s, err: %v", order.SuborderID, err)
		return &event.Event{Type: event.TransitFailed, Error: err}
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
		return t.TransitCvm(order, hosts)
	case table.ResourceTypePm:
		return t.TransitPm(order, hosts)
	case table.ResourceTypeOthers:
		return t.TransitOthers(order, hosts)
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
		Limit: pkg.BKMaxInstanceLimit,
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle hosts, err: %v", err)
		return nil, err
	}

	return insts, nil
}

// DealTransitTask2Pool deal hosts transit task and transfer host to CR resource pool
func (t *Transit) DealTransitTask2Pool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	hostIds := make([]int64, 0)
	ips := make([]string, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
		ips = append(ips, host.IP)
	}

	if len(hostIds) == 0 || len(ips) == 0 {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to run transit task, "+
			"for host id list or ip list is empty, subOrderID: %s", order.SuborderID)
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to run transit task, for host id list or ip list is empty"),
		}
		return ev
	}

	// transfer hosts to reborn-_数据待清理
	destBiz := recovertask.RebornBizId
	destModule := recovertask.DataToCleanedModule
	kt := core.NewBackendKit()
	kt.Ctx = t.ctx
	if err := t.TransferHost(kt, hostIds, order.BizID, destBiz, destModule); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to transfer host to biz %d module %d, "+
			"subOrderID: %s, err: %v", destBiz, destModule, order.SuborderID, err)
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
		logs.Warnf("recycler:logics:cvm:dealTransitTask2Pool:failed, failed to add shield TMP alarm config, "+
			"subOrderID: %s, err: %v, ips: %v", order.SuborderID, err, ips)
	}

	if errUpdate := t.UpdateHostInfo(order, table.RecycleStageDone, table.RecycleStatusDone); errUpdate != nil {
		logs.Errorf("failed to update recycle host info, subOrderID: %s, err: %v", order.SuborderID, errUpdate)
		ev := &event.Event{
			Type:  event.TransitFailed,
			Error: fmt.Errorf("failed to update recycle host info, err: %v", errUpdate),
		}
		return ev
	}

	return &event.Event{Type: event.TransitSuccess, Error: nil}
}

// DealTransitTask2Transit deal hosts transit task and transfer host to CR transit module
func (t *Transit) DealTransitTask2Transit(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
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

	kt := core.NewBackendKit()
	kt.Ctx = t.ctx
	// transfer hosts to reborn-_CR中转
	if err := t.transferHost2CrTransit(kt, hostIds, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to transfer host to "+
			"CR transit module, err: %v", err)
		if errUpdate := t.UpdateHostInfo(order, table.RecycleStageTransit,
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
	srcBizId := recovertask.RebornBizId
	srcModuleId := recovertask.CrRelayModuleId

	if err := t.TransferHost2BizTransit(kt, hosts, srcBizId, srcModuleId, order.BizID); err != nil {
		logs.Errorf("recycler:logics:cvm:dealTransitTask2Transit:failed, failed to transfer host to biz's "+
			"CR transit module in CMDB, err: %v, srcBizId: %d, bizID: %d", err, srcBizId, order.BizID)
		if errUpdate := t.UpdateHostInfo(order, table.RecycleStageTransit,
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

// TransferHost 调用cc接口转移主机
func (t *Transit) TransferHost(kt *kit.Kit, hostIds []int64, srcBizId, destBizId, destModuleId int64) error {
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

	err := t.cc.TransferHost(kt, transferReq)
	if err != nil {
		return err
	}

	return nil
}

// transferHost2CrTransit transfer hosts to CR transit module
func (t *Transit) transferHost2CrTransit(kt *kit.Kit, hostIds []int64, srcBizId int64) error {
	// transfer hosts to reborn-_CR中转
	destBiz := recovertask.RebornBizId
	destModule := recovertask.CrRelayModuleId
	return t.TransferHost(kt, hostIds, srcBizId, destBiz, destModule)
}

// countHostBizRelation get host topo info in cc 3.0
func (t *Transit) countHostBizRelation(kt *kit.Kit, hostIds []int64) (int, error) {
	req := &cmdb.HostModuleRelationParams{
		HostID: hostIds,
	}

	resp, err := t.cc.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v", err)
		return 0, err
	}

	return len(cvt.PtrToVal(resp)), nil
}

// TransferHost2BizTransit transfer hosts to given business's CR transit module in CMDB
func (t *Transit) TransferHost2BizTransit(kt *kit.Kit, hosts []*table.RecycleHost, srcBizID, srcModuleID,
	destBizId int64) error {
	hostIds := make([]int64, 0)
	for _, host := range hosts {
		hostIds = append(hostIds, host.HostID)
	}
	subOrderId := hosts[0].SuborderID
	// 查询仍在当前业务下的主机，避免部分转移重试失败
	listResult, err := t.getBizHost(kt, srcBizID, hostIds)
	if err != nil {
		logs.Errorf("failed to get biz host number, srcBizID: %d, srcModuleID: %d, destBizId: %d, subOrderId: %s, "+
			"err: %v", srcBizID, srcModuleID, destBizId, subOrderId, err)
		return err
	}
	// 若当前业务下没有主机，判断主机是否被他人转移导致查询不到
	if listResult.Count == 0 {
		hostToposCount, err := t.countHostBizRelation(kt, hostIds)
		if err != nil {
			return err
		}
		if hostToposCount == 0 {
			// 未查询到机器，机器已转移成功
			return nil
		}
		// 若可以查到主机业务信息，说明主机已被他人转移
		logs.Errorf("host is delivered by others, srcBizID: %d, srcModuleID: %d, subOrderId: %s", srcBizID, srcModuleID,
			subOrderId)
		return fmt.Errorf("host is delivered by others, srcBizID: %d, srcModuleID: %d, subOrderId: %s",
			srcBizID, srcModuleID, subOrderId)
	}

	remainAssetIds := make([]string, 0)
	for _, host := range listResult.Info {
		remainAssetIds = append(remainAssetIds, host.BkAssetID)
	}

	req := &cmdb.CrTransitReq{
		From: cmdb.CrTransitSrcInfo{
			FromBizID:    srcBizID,
			FromModuleID: srcModuleID,
		},
		To: cmdb.CrTransitDstInfo{
			ToBizID: destBizId,
		},
	}
	for _, assetBatch := range slice.Split(remainAssetIds, t.getHost2CrTransitBatchSize(kt)) {
		req.From.AssetIDs = assetBatch
		_, err := t.cc.Hosts2CrTransit(kt, req)
		if err != nil {
			logs.Errorf("recycler:logics:cvm:transferHost2BizTransit:failed, failed to transfer host to "+
				"CR transit module, err: %v, req: %+v", err, cvt.PtrToVal(req))
			return fmt.Errorf("failed to transfer host to CR transit module, err: %v", err)
		}
	}
	return nil
}

func (t *Transit) getHost2CrTransitBatchSize(kt *kit.Kit) (batchSize int) {

	batchSize = Host2CrTransitDefaultBatchSize
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRecycle),
			tools.RuleJSONEqual("config_key", constant.RecycleTransitHost2CRBatchSizeConfigKey),
		),
		Page: core.NewDefaultBasePage(),
	}
	cfgResp, err := t.cliSet.DataService().Global.GlobalConfig.List(kt, req)
	if err != nil {
		logs.Errorf("failed to get host2cr batchsize by global config, using default %d, err: %v, req: %+v, rid: %s",
			batchSize, err, cvt.PtrToVal(req), kt.Rid)
		return batchSize
	}
	if len(cfgResp.Details) == 0 {
		return batchSize
	}
	err = json.Unmarshal([]byte(cfgResp.Details[0].ConfigValue), &batchSize)
	if err != nil {
		logs.Errorf("failed to unmarshal global config host2cr value, using default %d, err: %v, raw: %v, rid: %s",
			batchSize, err, cfgResp.Details, kt.Rid)
		return batchSize
	}
	return batchSize
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

// UpdateHostInfo 更新回收主机信息
func (t *Transit) UpdateHostInfo(order *table.RecycleOrder, stage table.RecycleStage,
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
