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

package bkcc

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/slice"
)

var allFields = make([]string, 0)

// watchCCEvent 监听cmdb事件
func (w *Watcher) watchCCEvent(sd serviced.ServiceDiscover, resType cmdb.CursorType, eventTypes []cmdb.EventType,
	fields []string, consumeFunc func(kt *kit.Kit, events []cmdb.WatchEventDetail) error) {

	param := &cmdb.WatchEventParams{
		EventTypes: eventTypes,
		Resource:   resType,
	}
	if len(fields) != 0 {
		param.Fields = fields
	}

	for {
		if !sd.IsMaster() {
			time.Sleep(10 * time.Second)
			continue
		}

		kt := core.NewBackendKit()
		cursor, err := w.getEventCursor(kt, resType)
		if err != nil {
			logs.Errorf("get event cursor failed, err: %v, type: %s, rid: %s", err, resType, kt.Rid)
			continue
		}
		param.Cursor = cursor

		result, err := esb.EsbClient().Cmdb().ResourceWatch(kt, param)
		if err != nil {
			logs.Errorf("watch cmdb host resource failed, err: %v, req: %+v, rid: %s", err, param, kt.Rid)
			// 如果事件节点不存在，cc会返回该错误码，此时需要将cursor设置为""，从当前时间开始监听事件
			if strings.Contains(err.Error(), cmdb.CCErrEventChainNodeNotExist) {
				if err = w.setEventCursor(kt, resType, ""); err != nil {
					logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
						"", kt.Rid)
				}
			}
			continue
		}

		if !result.Watched {
			if len(result.Events) != 0 {
				newCursor := result.Events[0].Cursor
				if err = w.setEventCursor(kt, resType, newCursor); err != nil {
					logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
						newCursor, kt.Rid)
				}
			}
			continue
		}

		if err = consumeFunc(kt, result.Events); err != nil {
			logs.Errorf("consume %s event failed, err: %+v, res: %+v, rid: %s", resType, err, result, kt.Rid)
		}

		if len(result.Events) != 0 {
			newCursor := result.Events[len(result.Events)-1].Cursor
			if err = w.setEventCursor(kt, resType, newCursor); err != nil {
				logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
					newCursor, kt.Rid)
			}
		}
	}
}

// WatchHostEvent 监听主机事件，增量同步主机
func (w *Watcher) WatchHostEvent(sd serviced.ServiceDiscover) {
	w.watchCCEvent(sd, cmdb.HostType, []cmdb.EventType{cmdb.Create, cmdb.Update, cmdb.Delete}, cmdb.HostFields,
		w.consumeHostEvent)
}

func (w *Watcher) consumeHostEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
	if len(events) == 0 {
		return nil
	}

	idHostMap := make(map[int64]struct{})
	deleteHostIDs := make([]int64, 0)

	// 1. 获取需要创建、更新、删除的主机
	for _, event := range events {
		host, err := convertHost(kt, event.Detail)
		if err != nil {
			logs.Errorf("convert host failed, err: %v, event: %+v, rid: %s", err, event, kt.Rid)
			continue
		}

		// 不需要同步非自研云的机器
		if host.BkCloudID != 0 {
			continue
		}

		if event.EventType == cmdb.Delete {
			deleteHostIDs = append(deleteHostIDs, host.BkHostID)
			delete(idHostMap, host.BkHostID)
			continue
		}

		idHostMap[host.BkHostID] = struct{}{}
	}

	// 2. 创建或更新主机
	upsertHostIDs := make([]int64, 0)
	for id := range idHostMap {
		upsertHostIDs = append(upsertHostIDs, id)
	}
	if len(upsertHostIDs) != 0 {
		if err := w.upsertHost(kt, upsertHostIDs); err != nil {
			logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, upsertHostIDs, kt.Rid)
		}
	}

	// 3. 删除需要删除的主机
	if len(deleteHostIDs) != 0 {
		if err := w.deleteHost(kt, deleteHostIDs); err != nil {
			logs.Errorf("delete host failed, err: %v, ids: %+v, rid: %s", err, deleteHostIDs, kt.Rid)
		}
	}

	return nil
}

func convertHost(kt *kit.Kit, data json.RawMessage) (*cmdb.Host, error) {
	host := &cmdb.Host{}
	if err := json.Unmarshal(data, host); err != nil {
		logs.Errorf("unmarshal host failed, err: %v, data: %v, rid: %s", err, data, kt.Rid)
		return nil, err
	}

	return host, nil
}

func (w *Watcher) upsertHost(kt *kit.Kit, upsertHostIDs []int64) error {
	if len(upsertHostIDs) == 0 {
		return nil
	}

	bizHostMap, err := getHostBizID(kt, upsertHostIDs)
	if err != nil {
		logs.Errorf("get biz host map failed, err: %v, ids: %v, rid: %s", err, upsertHostIDs, kt.Rid)
		return err
	}

	accountID, err := w.getTCloudZiyanAccountID(kt)
	if err != nil {
		logs.Errorf("get tcloud ziyan account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for bizID, hostIDs := range bizHostMap {
		for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
			req := &sync.TCloudZiyanSyncHostByCondReq{BizID: bizID, HostIDs: batch, AccountID: accountID}
			err = w.CliSet.HCService().TCloudZiyan.Cvm.SyncHostWithRelResByCond(kt.Ctx, kt.Header(), req)
			if err != nil {
				logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, batch, kt.Rid)
			}
		}
	}

	return nil
}

func (w *Watcher) deleteHost(kt *kit.Kit, hostIDs []int64) error {
	if len(hostIDs) == 0 {
		return nil
	}

	accountID, err := w.getTCloudZiyanAccountID(kt)
	if err != nil {
		logs.Errorf("get tcloud ziyan account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		req := &sync.TCloudZiyanDelHostByCondReq{HostIDs: batch, AccountID: accountID}
		if err = w.CliSet.HCService().TCloudZiyan.Cvm.DeleteHostByCond(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("delete host failed, err: %v, ids: %+v, rid: %s", err, batch, kt.Rid)
		}
	}

	return nil
}

func (w *Watcher) getTCloudZiyanAccountID(kt *kit.Kit) (string, error) {
	req := &cloud.AccountListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan)),
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}

	accounts, err := w.CliSet.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("get account failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return "", err
	}

	if len(accounts.Details) == 0 {
		logs.Errorf("can not get account, req: %+v, rid: %s", req, kt.Rid)
		return "", errors.New("can not get tcloud ziyan account")
	}

	return accounts.Details[0].ID, nil
}

func getHostBizID(kt *kit.Kit, hostIDs []int64) (map[int64][]int64, error) {
	if len(hostIDs) == 0 {
		return make(map[int64][]int64), nil
	}

	hostBizIDMap := make(map[int64]int64)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		req := &cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := esb.EsbClient().Cmdb().FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("fail to find cmdb topo relation, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, relation := range *relationRes {
			hostBizIDMap[relation.HostID] = relation.BizID
		}
	}

	bizHostMap := make(map[int64][]int64)
	for hostID, bizID := range hostBizIDMap {
		if _, ok := bizHostMap[bizID]; !ok {
			bizHostMap[bizID] = make([]int64, 0)
		}

		bizHostMap[bizID] = append(bizHostMap[bizID], hostID)
	}

	return bizHostMap, nil
}

// WatchHostRelationEvent 监听主机关系事件，增量修改主机关系
func (w *Watcher) WatchHostRelationEvent(sd serviced.ServiceDiscover) {
	w.watchCCEvent(sd, cmdb.HostRelation, []cmdb.EventType{cmdb.Create}, allFields, w.consumeHostRelationEvent)
}

func (w *Watcher) consumeHostRelationEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
	if len(events) == 0 {
		return nil
	}

	hostBizIDMap := make(map[int64]int64)
	hostIDs := make([]int64, 0)
	for _, event := range events {
		relation, err := convertHostRelation(kt, event.Detail)
		if err != nil {
			logs.Errorf("convert host relation failed, err: %v, event: %+v, rid: %s", err, event, kt.Rid)
			continue
		}

		if _, ok := hostBizIDMap[relation.HostID]; !ok {
			hostIDs = append(hostIDs, relation.HostID)
		}

		hostBizIDMap[relation.HostID] = relation.BizID
	}

	dbHosts, err := w.listHostFromDB(kt, hostIDs)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return err
	}

	updateHostIDs := make([]int64, 0)
	for _, host := range dbHosts {
		if host.Extension == nil {
			logs.ErrorJson("host extension field is nil, host: %+v, rid: %s", host, kt.Rid)
			continue
		}

		if hostBizIDMap[host.Extension.HostID] == host.BkBizID {
			continue
		}

		updateHostIDs = append(updateHostIDs, host.Extension.HostID)
	}

	if len(updateHostIDs) == 0 {
		return nil
	}

	if err = w.upsertHost(kt, updateHostIDs); err != nil {
		logs.Errorf("upsert host failed, err: %v, hostIDs: %v, rid: %s", err, updateHostIDs, kt.Rid)
	}

	return nil
}

func convertHostRelation(kt *kit.Kit, data json.RawMessage) (*cmdb.HostTopoRelation, error) {
	relation := &cmdb.HostTopoRelation{}
	if err := json.Unmarshal(data, relation); err != nil {
		logs.Errorf("unmarshal host relation failed, err: %v, data: %v, rid: %s", err, data, kt.Rid)
		return nil, err
	}

	return relation, nil
}

func (w *Watcher) listHostFromDB(kt *kit.Kit, hostIDs []int64) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {
	req := &cloud.CvmListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleJsonIn("extension.bk_host_id", hostIDs)),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}

	hosts := make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0)
	for {
		result, err := w.CliSet.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.ErrorJson("[%s] request dataservice to list cvm failed, err: %v, req: %v, rid: %s", enumor.TCloudZiyan,
				err, req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Details...)

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return hosts, nil
}
