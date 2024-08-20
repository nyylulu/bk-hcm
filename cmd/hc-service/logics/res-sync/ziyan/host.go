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

package ziyan

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	typescvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// Host 对比从cc获取的主机，与本地的主机的差异，进行本地主机的新增、更新、删除操作
func (cli *client) Host(kt *kit.Kit, params *SyncHostParams) (*SyncResult, error) {
	if params == nil {
		logs.Errorf("params is nil, rid: %s", kt.Rid)
		return nil, fmt.Errorf("params is nil")
	}

	if err := params.Validate(); err != nil {
		logs.Errorf("param is invalid, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ccHosts, err := cli.getHostFromCCByHostIDs(kt, params.BizID, params.HostIDs, cmdb.HostFields)
	if err != nil {
		logs.Errorf("get host from cc by host id failed, err: %v, ids: %v, rid: %s", err, params.HostIDs, kt.Rid)
		return nil, err
	}

	dbHosts, err := cli.listHostFromDBByHostIDs(kt, params.HostIDs)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, params.HostIDs, kt.Rid)
		return nil, err
	}

	if len(ccHosts) == 0 && len(dbHosts) == 0 {
		return new(SyncResult), nil
	}

	cloudHosts, err := cli.getCloudHost(kt, params.AccountID, params.BizID, ccHosts)
	if err != nil {
		logs.Errorf("get cloud host failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	addSlice, updateMap, delCloudIDs := common.Diff[cvm.Cvm[cvm.TCloudZiyanHostExtension],
		cvm.Cvm[cvm.TCloudZiyanHostExtension]](cloudHosts, dbHosts, isHostChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteHost(kt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createHost(kt, convToCreate(addSlice)); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateHost(kt, convToUpdate(updateMap)); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil

}

func (cli *client) getCloudHost(kt *kit.Kit, accountID string, bizID int64, ccHosts []cmdb.Host) (
	[]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {

	if len(ccHosts) == 0 {
		return make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0), nil
	}

	hostMap := make(map[string]cvm.Cvm[cvm.TCloudZiyanHostExtension])
	regionCloudIDMap := make(map[string][]string)
	for _, ccHost := range ccHosts {
		host := convertToHost(&ccHost, accountID, bizID)
		hostMap[host.CloudID] = host

		if ccHost.SvrSourceTypeID != cmdb.CVM {
			continue
		}

		if ccHost.BkCloudRegion == "" {
			logs.Warnf("host id(%d) region data is nil, rid: %s", ccHost.BkHostID, kt.Rid)
			continue
		}

		if _, ok := regionCloudIDMap[host.Region]; !ok {
			regionCloudIDMap[host.Region] = make([]string, 0)
		}

		regionCloudIDMap[host.Region] = append(regionCloudIDMap[host.Region], host.CloudID)
	}

	return cli.fillCloudFields(kt, accountID, regionCloudIDMap, hostMap)

}

func (cli *client) fillCloudFields(kt *kit.Kit, accountID string, regionCloudIDMap map[string][]string,
	hostMap map[string]cvm.Cvm[cvm.TCloudZiyanHostExtension]) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {

	for region, cloudIDs := range regionCloudIDMap {
		cloudVpcIDs := make([]string, 0)
		cloudSubnetIDs := make([]string, 0)
		cvms := make([]typescvm.TCloudCvm, 0)

		for _, batch := range slice.Split(cloudIDs, adcore.TCloudQueryLimit) {
			opt := &typescvm.TCloudListOption{
				Region:   region,
				CloudIDs: batch,
				Page:     &adcore.TCloudPage{Offset: 0, Limit: adcore.TCloudQueryLimit},
			}
			res, err := cli.cloudCli.ListCvm(kt, opt)
			if err != nil {
				logs.Errorf("[%s] list cvm from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
					enumor.TCloudZiyan, err, accountID, opt, kt.Rid)
				return nil, err
			}
			cvms = append(cvms, res...)

			for _, one := range res {
				cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VirtualPrivateCloud.VpcId))
				cloudSubnetIDs = append(cloudSubnetIDs, converter.PtrToVal(one.VirtualPrivateCloud.SubnetId))
			}
		}

		vpcMap, err := cli.getVpcMap(kt, accountID, region, cloudVpcIDs)
		if err != nil {
			return nil, err
		}

		subnetMap, err := cli.getSubnetMap(kt, accountID, region, cloudSubnetIDs)
		if err != nil {
			return nil, err
		}

		for _, one := range cvms {
			if _, exsit := vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)]; !exsit {
				return nil, fmt.Errorf("cvm %s can not find vpc", converter.PtrToVal(one.InstanceId))
			}

			if _, exsit := subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]; !exsit {
				return nil, fmt.Errorf("cvm %s can not find subnet", converter.PtrToVal(one.InstanceId))
			}

			cloudID := converter.PtrToVal(one.InstanceId)
			host, ok := hostMap[cloudID]
			if !ok {
				logs.Errorf("host is not exist, cloud id: %s, rid: %s", cloudID, kt.Rid)
				continue
			}

			// 补充云上字段
			host.Name = converter.PtrToVal(one.InstanceName)
			host.Zone = converter.PtrToVal(one.Placement.Zone)
			host.ImageID = converter.PtrToVal(one.ImageId)
			host.Status = converter.PtrToVal(one.InstanceState)
			host.CloudExpiredTime = converter.PtrToVal(one.ExpiredTime)
			host.CloudVpcIDs = []string{converter.PtrToVal(one.VirtualPrivateCloud.VpcId)}
			host.VpcIDs = []string{vpcMap[converter.PtrToVal(one.VirtualPrivateCloud.VpcId)].VpcID}
			host.CloudSubnetIDs = []string{converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)}
			host.SubnetIDs = []string{subnetMap[converter.PtrToVal(one.VirtualPrivateCloud.SubnetId)]}
			host.CloudCreatedTime = converter.PtrToVal(one.CreatedTime)

			hostMap[cloudID] = host
		}
	}

	res := make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0)
	for _, host := range hostMap {
		res = append(res, host)
	}

	return res, nil
}

func convertToHost(ccHost *cmdb.Host, accountID string, bizID int64) cvm.Cvm[cvm.TCloudZiyanHostExtension] {
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

	host := cvm.Cvm[cvm.TCloudZiyanHostExtension]{
		BaseCvm: cvm.BaseCvm{
			CloudID:              cloudID,
			Name:                 ccHost.BkHostName,
			BkBizID:              bizID,
			BkCloudID:            ccHost.BkCloudID,
			AccountID:            accountID,
			Region:               ccHost.BkCloudRegion,
			Zone:                 ccHost.BkCloudZone,
			CloudVpcIDs:          []string{ccHost.BkCloudVpcID},
			CloudSubnetIDs:       []string{ccHost.BkCloudSubnetID},
			OsName:               ccHost.BkOSName,
			PrivateIPv4Addresses: innerIpv4,
			PrivateIPv6Addresses: innerIpv6,
			PublicIPv4Addresses:  outerIpv4,
			PublicIPv6Addresses:  outerIpv6,
			MachineType:          ccHost.SvrDeviceClassName,
		},
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

func isHostChange(cloud cvm.Cvm[cvm.TCloudZiyanHostExtension], db cvm.Cvm[cvm.TCloudZiyanHostExtension]) bool {
	if db.BkBizID != cloud.BkBizID {
		return true
	}

	if db.Region != cloud.Region {
		return true
	}

	if db.Zone != cloud.Zone {
		return true
	}

	if db.AccountID != cloud.AccountID {
		return true
	}

	if db.CloudID != cloud.CloudID {
		return true
	}

	if db.Name != cloud.Name {
		return true
	}

	if !assert.IsStringSliceEqual(db.CloudVpcIDs, cloud.CloudVpcIDs) {
		return true
	}

	if !assert.IsStringSliceEqual(db.CloudSubnetIDs, cloud.CloudSubnetIDs) {
		return true
	}

	if db.CloudImageID != cloud.CloudImageID {
		return true
	}

	if db.OsName != cloud.OsName {
		return true
	}

	if db.Status != cloud.Status {
		return true
	}

	if !assert.IsStringSliceEqual(db.PrivateIPv4Addresses, cloud.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv4Addresses, cloud.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PrivateIPv6Addresses, cloud.PrivateIPv6Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv6Addresses, cloud.PublicIPv6Addresses) {
		return true
	}

	if db.MachineType != cloud.MachineType {
		return true
	}

	if db.CloudCreatedTime != cloud.CloudCreatedTime {
		return true
	}

	if db.CloudExpiredTime != cloud.CloudExpiredTime {
		return true
	}

	if db.Extension == nil || cloud.Extension == nil || db.Extension.HostID != cloud.Extension.HostID ||
		db.Extension.SvrSourceTypeID != cloud.Extension.SvrSourceTypeID {

		return true
	}

	return false
}

// RemoveHostFromCC 对比根据的主机，删除本地多余的主机
func (cli *client) RemoveHostFromCC(kt *kit.Kit, params *DelHostParams) error {
	if params == nil {
		logs.Errorf("params is nil, rid: %s", kt.Rid)
		return fmt.Errorf("params is nil")
	}

	if err := params.Validate(); err != nil {
		logs.Errorf("param is invalid, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(params.DelHostIDs) != 0 {
		return cli.deleteHostByHostID(kt, params.DelHostIDs)
	}

	return cli.removeHost(kt, params.BizID)
}

func (cli *client) deleteHostByHostID(kt *kit.Kit, hostIDs []int64) error {
	if len(hostIDs) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		deleteReq := &cloud.CvmBatchDeleteReq{Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleJsonIn("extension.bk_host_id", batch))}

		if err := cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete host failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, deleteReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (cli *client) removeHost(kt *kit.Kit, bizID int64) error {
	ccHosts, err := cli.getHostFromCCByBizID(kt, bizID, []string{"bk_host_id"})
	if err != nil {
		logs.Errorf("get host from cc failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}
	hostIDMap := make(map[int64]struct{})
	for _, host := range ccHosts {
		hostIDMap[host.BkHostID] = struct{}{}
	}

	dbHosts, err := cli.listHostFromDBByBizID(kt, bizID, []string{"id", "extension"})
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}

	delHostIDs := make([]int64, 0)
	for _, host := range dbHosts {
		if host.Extension == nil {
			logs.ErrorJson("host extension field is nil, host: %+v, rid: %s", host, kt.Rid)
			continue
		}

		if _, ok := hostIDMap[host.Extension.HostID]; !ok {
			delHostIDs = append(delHostIDs, host.Extension.HostID)
		}
	}

	if err = cli.deleteHostByHostID(kt, delHostIDs); err != nil {
		logs.Errorf("delete host by host id failed, err: %v, ids: %v, rid: %s", err, delHostIDs, kt.Rid)
		return err
	}

	return nil
}

func (cli *client) getHostFromCCByBizID(kt *kit.Kit, bizID int64, fields []string) ([]cmdb.Host, error) {
	params := &cmdb.ListBizHostParams{
		BizID:  bizID,
		Fields: fields,
		Page:   cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit), Sort: "bk_host_id"},
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: "AND",
				Rules:     []cmdb.Rule{&cmdb.AtomRule{Field: "bk_cloud_id", Operator: "equal", Value: 0}},
			},
		},
	}

	return cli.getHostsFromCC(kt, params)
}

func (cli *client) getHostFromCCByHostIDs(kt *kit.Kit, bizID int64, hostIDs []int64, fields []string) ([]cmdb.Host,
	error) {

	res := make([]cmdb.Host, 0)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		params := &cmdb.ListBizHostParams{
			BizID:  bizID,
			Fields: fields,
			Page:   cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit), Sort: "bk_host_id"},
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: &cmdb.CombinedRule{
					Condition: "AND",
					Rules: []cmdb.Rule{
						&cmdb.AtomRule{Field: "bk_cloud_id", Operator: "equal", Value: 0},
						&cmdb.AtomRule{Field: "bk_host_id", Operator: "in", Value: batch},
					},
				},
			},
		}

		hosts, err := cli.getHostsFromCC(kt, params)
		if err != nil {
			logs.Errorf("get host from cc failed, err: %v, params: %+v, rid: %s", err, params, kt.Rid)
			return nil, err
		}
		res = append(res, hosts...)
	}

	return res, nil
}

func (cli *client) getHostsFromCC(kt *kit.Kit, params *cmdb.ListBizHostParams) ([]cmdb.Host, error) {
	hosts := make([]cmdb.Host, 0)
	for {
		result, err := esb.EsbClient().Cmdb().ListBizHost(kt, params)
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

func (cli *client) listHostFromDBByBizID(kt *kit.Kit, bizID int64,
	fields []string) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {

	req := &cloud.CvmListReq{
		Field:  fields,
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan), tools.RuleEqual("bk_biz_id", bizID)),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}

	return cli.listHostFromDB(kt, req)
}

func (cli *client) listHostFromDBByHostIDs(kt *kit.Kit, hostIDs []int64) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension],
	error) {

	res := make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0)
	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		req := &cloud.CvmListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleJsonIn("extension.bk_host_id", batch)),
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
				Sort:  "id",
			},
		}
		hosts, err := cli.listHostFromDB(kt, req)
		if err != nil {
			logs.Errorf("list host from db failed ,err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		res = append(res, hosts...)
	}

	return res, nil
}

// listHostFromDB 从db中查询主机
func (cli *client) listHostFromDB(kt *kit.Kit, req *cloud.CvmListReq) ([]cvm.Cvm[cvm.TCloudZiyanHostExtension], error) {
	hosts := make([]cvm.Cvm[cvm.TCloudZiyanHostExtension], 0)
	for {
		result, err := cli.dbCli.TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), req)
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

func (cli *client) deleteHost(kt *kit.Kit, cloudIDs []string) error {
	if len(cloudIDs) <= 0 {
		return nil
	}

	for _, batch := range slice.Split(cloudIDs, constant.BatchOperationMaxLimit) {
		deleteReq := &cloud.CvmBatchDeleteReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("cloud_id", batch), tools.RuleEqual("vendor", enumor.TCloudZiyan)),
		}
		if err := cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete host failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, deleteReq, kt.Rid)
			return err
		}

		logs.Infof("[%s] sync host to delete host success, count: %d, cloudIDs: %+v, rid: %s", enumor.TCloudZiyan,
			len(batch), batch, kt.Rid)
	}

	return nil
}

func convToCreate(hosts []cvm.Cvm[cvm.TCloudZiyanHostExtension]) []cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension] {
	res := make([]cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension], 0)
	for _, host := range hosts {
		res = append(res, cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension]{
			CloudID:              host.CloudID,
			Name:                 host.Name,
			BkBizID:              host.BkBizID,
			BkCloudID:            host.BkCloudID,
			AccountID:            host.AccountID,
			Region:               host.Region,
			Zone:                 host.Zone,
			CloudVpcIDs:          host.CloudVpcIDs,
			VpcIDs:               host.VpcIDs,
			CloudSubnetIDs:       host.CloudSubnetIDs,
			SubnetIDs:            host.SubnetIDs,
			CloudImageID:         host.CloudImageID,
			ImageID:              host.ImageID,
			OsName:               host.OsName,
			Memo:                 host.Memo,
			Status:               host.Status,
			PrivateIPv4Addresses: host.PrivateIPv4Addresses,
			PrivateIPv6Addresses: host.PrivateIPv6Addresses,
			PublicIPv4Addresses:  host.PublicIPv4Addresses,
			PublicIPv6Addresses:  host.PublicIPv6Addresses,
			MachineType:          host.MachineType,
			CloudCreatedTime:     host.CloudCreatedTime,
			CloudLaunchedTime:    host.CloudLaunchedTime,
			CloudExpiredTime:     host.CloudExpiredTime,
			Extension:            host.Extension,
		})
	}

	return res
}

func (cli *client) createHost(kt *kit.Kit, hosts []cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension]) error {
	if len(hosts) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hosts, constant.BatchOperationMaxLimit) {
		createReq := &cloud.CvmBatchCreateReq[cvm.TCloudZiyanHostExtension]{Cvms: batch}
		_, err := cli.dbCli.TCloudZiyan.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("create host failed, err: %v, req: %+v, rid: %s", err, createReq, kt.Rid)
			return err
		}
	}

	return nil
}

func convToUpdate(
	hosts map[string]cvm.Cvm[cvm.TCloudZiyanHostExtension]) []cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension] {

	res := make([]cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension], 0)
	for id, host := range hosts {
		res = append(res, cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]{
			ID:                   id,
			Name:                 host.Name,
			BkBizID:              host.BkBizID,
			BkCloudID:            host.BkCloudID,
			Region:               host.Region,
			Zone:                 host.Zone,
			CloudVpcIDs:          host.CloudVpcIDs,
			VpcIDs:               host.VpcIDs,
			CloudSubnetIDs:       host.CloudSubnetIDs,
			SubnetIDs:            host.SubnetIDs,
			CloudImageID:         host.CloudImageID,
			ImageID:              host.ImageID,
			OsName:               host.OsName,
			Memo:                 host.Memo,
			Status:               host.Status,
			PrivateIPv4Addresses: host.PrivateIPv4Addresses,
			PrivateIPv6Addresses: host.PrivateIPv6Addresses,
			PublicIPv4Addresses:  host.PublicIPv4Addresses,
			PublicIPv6Addresses:  host.PublicIPv6Addresses,
			MachineType:          host.MachineType,
			CloudCreatedTime:     host.CloudCreatedTime,
			CloudLaunchedTime:    host.CloudLaunchedTime,
			CloudExpiredTime:     host.CloudExpiredTime,
			Extension:            host.Extension,
		})
	}

	return res
}

func (cli *client) updateHost(kt *kit.Kit, hosts []cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]) error {
	if len(hosts) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hosts, constant.BatchOperationMaxLimit) {
		updateReq := &cloud.CvmBatchUpdateReq[cvm.TCloudZiyanHostExtension]{Cvms: batch}
		if err := cli.dbCli.TCloudZiyan.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("update host failed, err: %v, req: %+v, rid: %s", err, updateReq, kt.Rid)
			return err
		}
	}

	return nil
}
