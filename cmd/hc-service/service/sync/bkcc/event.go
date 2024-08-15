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
	"fmt"
	"strings"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/esb/cmdb"

	etcd3 "go.etcd.io/etcd/client/v3"
)

func getCursorKey(cursorType cmdb.CursorType) string {
	return fmt.Sprintf("/hcm/event/cc/%s", cursorType)
}

func (s *Syncer) getEventCursor(kt *kit.Kit, cursorType cmdb.CursorType) (string, error) {
	key := getCursorKey(cursorType)
	resp, err := s.EtcdCli.Get(kt.Ctx, key)
	if err != nil {
		logs.Errorf("get cmdb event cursor from etcd fail, err: %v, key: %s, rid: %s", err, key, kt.Rid)
		return "", err
	}

	// 从etcd里拿不到cursor，返回空字符串，从当前时间watch
	if len(resp.Kvs) == 0 {
		logs.Warnf("can not get cmdb event cursor from etcd, key: %s, rid: %s", key, kt.Rid)
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

func (s *Syncer) setEventCursor(kt *kit.Kit, cursorType cmdb.CursorType, cursor string) error {
	key := getCursorKey(cursorType)

	leaseID, err := s.leaseOp.getLeaseID(kt, key)
	if err != nil {
		logs.Errorf("get lease id failed, err: %v, key: %s, rid: %s", err, key, kt.Rid)
		return err
	}

	if _, err = s.EtcdCli.Put(kt.Ctx, key, cursor, etcd3.WithLease(leaseID)); err != nil {
		logs.Errorf("set etcd error, err: %v, key: %s, val: %s, rid: %s", err, key, cursor, kt.Rid)
		return err
	}

	return nil
}

var allFields = make([]string, 0)

// watchCCEvent 监听cmdb事件
func (s *Syncer) watchCCEvent(sd serviced.ServiceDiscover, resType cmdb.CursorType, eventTypes []cmdb.EventType,
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
		cursor, err := s.getEventCursor(kt, resType)
		if err != nil {
			logs.Errorf("get event cursor failed, err: %v, type: %s, rid: %s", err, resType, kt.Rid)
			continue
		}
		param.Cursor = cursor

		result, err := s.EsbCli.Cmdb().ResourceWatch(kt, param)
		if err != nil {
			logs.Errorf("watch cmdb host resource failed, err: %v, req: %+v, rid: %s", err, param, kt.Rid)
			// 如果事件节点不存在，cc会返回该错误码，此时需要将cursor设置为""，从当前时间开始监听事件
			if strings.Contains(err.Error(), cmdb.CCErrEventChainNodeNotExist) {
				if err = s.setEventCursor(kt, resType, ""); err != nil {
					logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
						"", kt.Rid)
				}
			}
			continue
		}

		if !result.Watched {
			if len(result.Events) != 0 {
				newCursor := result.Events[0].Cursor
				if err = s.setEventCursor(kt, resType, newCursor); err != nil {
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
			if err = s.setEventCursor(kt, resType, newCursor); err != nil {
				logs.Errorf("set event cursor failed, err: %v, resource type: %v, val: %s, rid: %s", err, resType,
					newCursor, kt.Rid)
			}
		}
	}
}

// WatchHostEvent 监听主机事件，增量同步主机
func (s *Syncer) WatchHostEvent(sd serviced.ServiceDiscover) {
	s.watchCCEvent(sd, cmdb.HostType, []cmdb.EventType{cmdb.Create, cmdb.Update, cmdb.Delete}, cmdb.HostFields,
		s.consumeHostEvent)
}

func (s *Syncer) consumeHostEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
	if len(events) == 0 {
		return nil
	}

	idHostMap := make(map[int64]cmdb.Host)
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

		idHostMap[host.BkHostID] = *host
	}

	// 2. 查询需要创建或者更新主机所属的业务id，然后进行创建或更新
	hostIDs := make([]int64, 0)
	for id := range idHostMap {
		hostIDs = append(hostIDs, id)
	}
	hostBizID, err := s.getHostBizID(kt, hostIDs)
	if err != nil {
		logs.Errorf("get host biz id failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return err
	}
	ccHosts := make([]ccHostWithBiz, 0)
	for hostID, host := range idHostMap {
		bizID, ok := hostBizID[hostID]
		if !ok {
			logs.Errorf("can not find host biz id, host id: %d, rid: %s", hostID, kt.Rid)
			continue
		}

		ccHosts = append(ccHosts, ccHostWithBiz{host, bizID})
	}

	if err = s.upsertHost(kt, ccHosts); err != nil {
		logs.Errorf("upsert host failed, err: %v, host: %v, rid: %s", err, ccHosts, kt.Rid)
	}

	// 3. 删除需要删除的主机
	if len(deleteHostIDs) != 0 {
		if err := s.deleteHostByHostID(kt, deleteHostIDs); err != nil {
			logs.Errorf("delete host failed, err: %v, ids: %+v, rid: %s", err, deleteHostIDs, kt.Rid)
			return err
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

func (s *Syncer) upsertHost(kt *kit.Kit, ccHosts []ccHostWithBiz) error {
	if len(ccHosts) == 0 {
		return nil
	}

	hostIDs := make([]int64, 0)
	for _, host := range ccHosts {
		hostIDs = append(hostIDs, host.BkHostID)
	}

	dbHosts, err := s.listHostFromDBByHostIDs(kt, hostIDs)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return err
	}

	accountID, err := s.getTCloudZiyanAccountID(kt)
	if err != nil {
		logs.Errorf("get tcloud ziyan account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	diff, err := s.getHostDiff(accountID, ccHosts, dbHosts)
	if err != nil {
		logs.Errorf("get diff by cc host failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err = s.syncHostDiff(kt, diff); err != nil {
		logs.Errorf("sync host diff failed, err: %v, diff: %+v, rid: %s", err, diff, kt.Rid)
		return err
	}

	return nil
}

// WatchHostRelationEvent 监听主机关系事件，增量修改主机关系
func (s *Syncer) WatchHostRelationEvent(sd serviced.ServiceDiscover) {
	s.watchCCEvent(sd, cmdb.HostRelation, []cmdb.EventType{cmdb.Create}, allFields, s.consumeHostRelationEvent)
}

func (s *Syncer) consumeHostRelationEvent(kt *kit.Kit, events []cmdb.WatchEventDetail) error {
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

	dbHosts, err := s.listHostFromDBByHostIDs(kt, hostIDs)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return err
	}

	updateHosts := make([]cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension], 0)
	for _, host := range dbHosts {
		if host.Extension == nil {
			logs.ErrorJson("host extension field is nil, host: %+v, rid: %s", host, kt.Rid)
			continue
		}

		if hostBizIDMap[host.Extension.HostID] == host.BkBizID {
			continue
		}

		updateHost := cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]{
			ID:                   host.ID,
			Name:                 host.Name,
			BkBizID:              hostBizIDMap[host.Extension.HostID],
			BkCloudID:            host.BkCloudID,
			CloudVpcIDs:          host.CloudVpcIDs,
			CloudSubnetIDs:       host.CloudSubnetIDs,
			PrivateIPv4Addresses: host.PrivateIPv4Addresses,
			PrivateIPv6Addresses: host.PrivateIPv6Addresses,
			PublicIPv4Addresses:  host.PublicIPv4Addresses,
			PublicIPv6Addresses:  host.PublicIPv6Addresses,
			Extension: &cvm.TCloudZiyanHostExtension{
				HostID:          host.Extension.HostID,
				SvrSourceTypeID: host.Extension.SvrSourceTypeID,
			},
		}
		updateHosts = append(updateHosts, updateHost)
	}

	if len(updateHosts) != 0 {
		if err = s.updateHost(kt, updateHosts); err != nil {
			logs.ErrorJson("update host failed, err: %v, host: %+v, rid: %s", err, updateHosts, kt.Rid)
			return err
		}
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
