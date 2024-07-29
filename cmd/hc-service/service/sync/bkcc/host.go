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
	"strings"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/slice"
)

// FullSyncHost 全量同步cc主机
func (s *Syncer) FullSyncHost(intervalMin time.Duration, sd serviced.ServiceDiscover) {
	logs.Infof("cloud resource sync enable, syncIntervalMin: %v", intervalMin)

	for {
		if !sd.IsMaster() {
			time.Sleep(10 * time.Second)
			continue
		}

		kt := core.NewBackendKit()
		start := time.Now()
		logs.Infof("full sync host from cc start, time: %v, rid: %s", start, kt.Rid)

		if err := s.fullSyncHost(kt); err != nil {
			logs.Errorf("full sync host from cc failed, err: %v")
			time.Sleep(intervalMin)
			continue
		}

		end := time.Now()
		logs.Infof("full sync host from cc end, time: %v, cost: %v, rid: %s", end, end.Sub(start), kt.Rid)

		time.Sleep(intervalMin)
	}
}

func (s *Syncer) fullSyncHost(kt *kit.Kit) error {
	accountID, err := s.getTCloudZiyanAccountID(kt)
	if err != nil {
		logs.Errorf("get tcloud ziyan account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	bizIDs, err := s.listIEGBizIDs(kt)
	if err != nil {
		logs.Errorf("list ieg biz ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, bizID := range bizIDs {
		start := time.Now()
		logs.Infof("start sync biz(%d) host, time: %v, rid: %s", bizID, start, kt.Rid)

		if err = s.syncBizHost(kt, bizID, accountID); err != nil {
			logs.Errorf("sync biz host failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
			continue
		}

		end := time.Now()
		logs.Infof("sync biz(%d) host success, time: %v, cost: %v, rid: %s", bizID, end, end.Sub(start), kt.Rid)
	}

	return nil
}

func (s *Syncer) getTCloudZiyanAccountID(kt *kit.Kit) (string, error) {
	req := &cloud.AccountListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan)),
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}

	accounts, err := s.CliSet.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("get account failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return "", err
	}

	if len(accounts.Details) == 0 {
		logs.Errorf("can not get account, req: %+v, rid: %s", req, kt.Rid)
		return "", err
	}

	return accounts.Details[0].ID, nil
}

func (s *Syncer) syncBizHost(kt *kit.Kit, bizID int64, accountID string) error {
	ccHostInfos, err := s.getHostsFromCC(kt, bizID)
	if err != nil {
		logs.Errorf("get host from cc failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}
	ccHosts := getHostWithBizID(bizID, ccHostInfos)

	dbHosts, err := s.listHostFromDBByBizID(kt, bizID)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}

	diff, err := s.getHostDiff(accountID, ccHosts, dbHosts)
	if err != nil {
		logs.Errorf("get diff by cc host failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}

	if err = s.syncHostDiff(kt, diff); err != nil {
		logs.Errorf("sync host diff failed, err: %v, bizID: %d, diff: %+v, rid: %s", err, bizID, diff, kt.Rid)
		return err
	}

	return nil
}

func (s *Syncer) getHostsFromCC(kt *kit.Kit, bizID int64) ([]cmdb.Host, error) {
	params := &cmdb.ListBizHostParams{
		BizID:  bizID,
		Fields: cmdb.HostFields,
		Page:   cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit), Sort: "bk_host_id"},
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: "AND",
				Rules:     []cmdb.Rule{&cmdb.AtomRule{Field: "bk_cloud_id", Operator: "equal", Value: 0}},
			},
		},
	}

	hosts := make([]cmdb.Host, 0)
	for {
		result, err := s.EsbCli.Cmdb().ListBizHost(kt, params)
		if err != nil {
			logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, params, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Info...)

		if len(result.Info) < int(core.DefaultMaxPageLimit) {
			break
		}

		params.Page.Start += int64(core.DefaultMaxPageLimit)
	}

	return hosts, nil
}

func (s *Syncer) getHostDiff(accountID string, ccHosts []ccHostWithBiz,
	dbHosts []cvm.Cvm[cvm.TCloudZiyanHostExtension]) (*diffHost, error) {

	diff := &diffHost{}
	dbHostMap := make(map[int64]cvm.Cvm[cvm.TCloudZiyanHostExtension])
	for _, host := range dbHosts {
		if host.Extension == nil {
			diff.deleteIDs = append(diff.deleteIDs, host.ID)
			continue
		}

		dbHostMap[host.Extension.HostID] = host
	}

	hostIDs := make([]int64, len(ccHosts))
	for i, host := range ccHosts {
		hostIDs[i] = host.BkHostID
	}

	// 如果主机存在则记录需要更新的主机信息，如果不存在则记录需要添加主机
	for _, ccHost := range ccHosts {
		dbHost, ok := dbHostMap[ccHost.BkHostID]
		if !ok {
			addHost := convertToCVMCreate(&ccHost, accountID)
			diff.addHosts = append(diff.addHosts, addHost)
			continue
		}

		if isHostDiff(dbHost, &ccHost) {
			updateHost := convertToCVMUpdate(dbHost, &ccHost)
			diff.updateHosts = append(diff.updateHosts, *updateHost)
		}

		delete(dbHostMap, ccHost.BkHostID)
	}

	// 添加需要删除的主机
	for _, host := range dbHostMap {
		diff.deleteIDs = append(diff.deleteIDs, host.ID)
	}

	return diff, nil
}

func convertToCVMCreate(ccHost *ccHostWithBiz, accountID string) cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension] {
	cloudID := ccHost.BkCloudInstID
	// 当主机不存在bk_cloud_inst_id时，需要用固资号进行填充，保证cloud id唯一
	if cloudID == "" {
		cloudID = ccHost.BkAssetID
	}

	innerIpv4 := make([]string, 0)
	if len(ccHost.BkHostInnerIP) != 0 {
		innerIpv4 = splitIP(ccHost.BkHostInnerIP)
	}
	innerIpv6 := make([]string, 0)
	if len(ccHost.BkHostInnerIPv6) != 0 {
		innerIpv6 = splitIP(ccHost.BkHostInnerIPv6)
	}
	outerIpv4 := make([]string, 0)
	if len(ccHost.BkHostOuterIP) != 0 {
		outerIpv4 = splitIP(ccHost.BkHostOuterIP)
	}
	outerIpv6 := make([]string, 0)
	if len(ccHost.BkHostOuterIPv6) != 0 {
		outerIpv6 = splitIP(ccHost.BkHostOuterIPv6)
	}

	host := cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension]{
		CloudID:              cloudID,
		Name:                 ccHost.BkHostName,
		BkBizID:              ccHost.bizID,
		BkCloudID:            ccHost.BkCloudID,
		AccountID:            accountID,
		CloudVpcIDs:          []string{ccHost.BkCloudVpcID},
		CloudSubnetIDs:       []string{ccHost.BkCloudSubnetID},
		OsName:               ccHost.BkOSName,
		PrivateIPv4Addresses: innerIpv4,
		PrivateIPv6Addresses: innerIpv6,
		PublicIPv4Addresses:  outerIpv4,
		PublicIPv6Addresses:  outerIpv6,
		Extension: &cvm.TCloudZiyanHostExtension{
			HostID:          ccHost.BkHostID,
			SvrSourceTypeID: ccHost.SvrSourceTypeID,
		},
	}

	return host
}

func splitIP(ip string) []string {
	return strings.Split(ip, ",")
}

func joinIP(ips []string) string {
	return strings.Join(ips, ",")
}

func isHostDiff(dbHost cvm.Cvm[cvm.TCloudZiyanHostExtension], ccHost *ccHostWithBiz) bool {
	if dbHost.Name != ccHost.BkHostName || dbHost.BkBizID != ccHost.bizID || dbHost.BkCloudID != ccHost.BkCloudID {
		return true
	}

	for _, cloudVpcID := range dbHost.CloudVpcIDs {
		if cloudVpcID != ccHost.BkCloudVpcID {
			return true
		}
	}

	for _, cloudSubnetID := range dbHost.CloudSubnetIDs {
		if cloudSubnetID != ccHost.BkCloudSubnetID {
			return true
		}
	}

	if joinIP(dbHost.PrivateIPv4Addresses) != ccHost.BkHostInnerIP ||
		joinIP(dbHost.PrivateIPv6Addresses) != ccHost.BkHostInnerIPv6 ||
		joinIP(dbHost.PublicIPv4Addresses) != ccHost.BkHostOuterIP ||
		joinIP(dbHost.PublicIPv6Addresses) != ccHost.BkHostOuterIPv6 {

		return true
	}

	if dbHost.Extension == nil || dbHost.Extension.HostID != ccHost.BkHostID ||
		dbHost.Extension.SvrSourceTypeID != ccHost.SvrSourceTypeID {

		return true
	}

	return false
}

func convertToCVMUpdate(dbHost cvm.Cvm[cvm.TCloudZiyanHostExtension],
	ccHost *ccHostWithBiz) *cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension] {

	innerIpv4 := make([]string, 0)
	if len(ccHost.BkHostInnerIP) != 0 {
		innerIpv4 = splitIP(ccHost.BkHostInnerIP)
	}
	innerIpv6 := make([]string, 0)
	if len(ccHost.BkHostInnerIPv6) != 0 {
		innerIpv6 = splitIP(ccHost.BkHostInnerIPv6)
	}
	outerIpv4 := make([]string, 0)
	if len(ccHost.BkHostOuterIP) != 0 {
		outerIpv4 = splitIP(ccHost.BkHostOuterIP)
	}
	outerIpv6 := make([]string, 0)
	if len(ccHost.BkHostOuterIPv6) != 0 {
		outerIpv6 = splitIP(ccHost.BkHostOuterIPv6)
	}

	return &cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]{
		ID:                   dbHost.ID,
		Name:                 ccHost.BkHostName,
		BkBizID:              ccHost.bizID,
		BkCloudID:            ccHost.BkCloudID,
		CloudVpcIDs:          []string{ccHost.BkCloudVpcID},
		CloudSubnetIDs:       []string{ccHost.BkCloudSubnetID},
		PrivateIPv4Addresses: innerIpv4,
		PrivateIPv6Addresses: innerIpv6,
		PublicIPv4Addresses:  outerIpv4,
		PublicIPv6Addresses:  outerIpv6,
		Extension: &cvm.TCloudZiyanHostExtension{
			HostID:          ccHost.BkHostID,
			SvrSourceTypeID: ccHost.SvrSourceTypeID,
		},
	}
}

func (s *Syncer) listHostFromDBByBizID(kt *kit.Kit, bizID int64) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {
	req := &cloud.CvmListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan), tools.RuleEqual("bk_biz_id", bizID)),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}

	return s.listHostFromDB(kt, req)
}

func (s *Syncer) listHostFromDBByHostIDs(kt *kit.Kit, hostIDs []int64) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension],
	error) {

	req := &cloud.CvmListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleJsonIn("extension.bk_host_id", hostIDs)),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}

	return s.listHostFromDB(kt, req)
}

// listHostFromDB 从db中查询主机
func (s *Syncer) listHostFromDB(kt *kit.Kit, req *cloud.CvmListReq) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {
	hosts := make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0)
	for {
		result, err := s.CliSet.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
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

func (s *Syncer) syncHostDiff(kt *kit.Kit, diff *diffHost) error {
	if diff == nil {
		return nil
	}

	if len(diff.addHosts) != 0 {
		if err := s.addHost(kt, diff.addHosts); err != nil {
			logs.Errorf("add host failed, err: %v, host: %+v, rid: %s", err, diff.addHosts, kt.Rid)
			return err
		}
	}

	if len(diff.updateHosts) != 0 {
		if err := s.updateHost(kt, diff.updateHosts); err != nil {
			logs.Errorf("update host failed, err: %v, host: %+v, rid: %s", err, diff.updateHosts, kt.Rid)
			return err
		}
	}

	if len(diff.deleteIDs) != 0 {
		if err := s.deleteHostByID(kt, diff.deleteIDs); err != nil {
			logs.Errorf("delete host failed, err: %v, ids: %+v, rid: %s", err, diff.deleteIDs, kt.Rid)
			return err
		}
	}

	return nil
}

func (s *Syncer) addHost(kt *kit.Kit, hosts []cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension]) error {
	if len(hosts) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hosts, constant.BatchOperationMaxLimit) {
		createReq := &cloud.CvmBatchCreateReq[cvm.TCloudZiyanHostExtension]{Cvms: batch}
		_, err := s.CliSet.DataService().TCloudZiyan.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("create host failed, err: %v, req: %+v, rid: %s", err, createReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (s *Syncer) updateHost(kt *kit.Kit, hosts []cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]) error {
	if len(hosts) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hosts, constant.BatchOperationMaxLimit) {
		updateReq := &cloud.CvmBatchUpdateReq[cvm.TCloudZiyanHostExtension]{Cvms: batch}
		if err := s.CliSet.DataService().TCloudZiyan.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("update host failed, err: %v, req: %+v, rid: %s", err, updateReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (s *Syncer) deleteHostByID(kt *kit.Kit, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for _, batch := range slice.Split(ids, constant.BatchOperationMaxLimit) {
		deleteReq := &cloud.CvmBatchDeleteReq{Filter: tools.ContainersExpression("id", batch)}
		if err := s.CliSet.DataService().Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete host failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, deleteReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (s *Syncer) deleteHostByHostID(kt *kit.Kit, hostIDs []int64) error {
	if len(hostIDs) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		deleteReq := &cloud.CvmBatchDeleteReq{Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleJsonIn("extension.bk_host_id", batch))}

		if err := s.CliSet.DataService().Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete host failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, deleteReq, kt.Rid)
			return err
		}
	}

	return nil
}
