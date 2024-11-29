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

package cvm

import (
	"encoding/json"
	"fmt"
	"strings"

	typecore "hcm/pkg/adaptor/types/core"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

// ListZiyanCmdbHost 从cc处拉取自研云主机, 支持二次拉取到云上信息
func (svc *cvmSvc) ListZiyanCmdbHost(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cscvm.CmdbHostListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 校验业务权限
	_, noPermFlag, err := handler.ListBizAuthRes(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.Cvm, Action: meta.Find})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &webserver.CloudHostListRespT[corecvm.TCloudZiyanCvmExtension]{Count: 0}, nil
	}

	// 1. 获取cmdb 业务下主机列表
	cmdbResult, err := svc.cvmLgc.GetCmdbBizHosts(cts.Kit, &cscvm.CmdbHostQueryReq{
		BkBizID:        bizID,
		Vendor:         enumor.TCloudZiyan,
		AccountID:      req.AccountID,
		Region:         req.Region,
		Zone:           req.Zone,
		CloudVpcIDs:    req.CloudVpcIDs,
		CloudSubnetIDs: req.CloudSubnetIDs,
		CloudInstIDs:   req.CloudInstIDs,
		BkSetIDs:       req.BkSetIDs,
		BkModuleIDs:    req.BkModuleIDs,
		Page:           req.Page,
	})
	if err != nil {
		logs.Errorf("fail to query cmdb biz hosts, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	resp := &webserver.CloudHostListRespT[corecvm.TCloudZiyanCvmExtension]{
		Count: cmdbResult.Count,
	}

	if req.QueryFromCloud {
		// 2. 尝试从云上获取数据
		hosts, err := svc.queryFromCloud(cts.Kit, req.AccountID, cmdbResult.Info)
		if err != nil {
			return nil, err
		}
		resp.Details = hosts
		return resp, nil
	}
	// 不从云上获取，则填充cmdb 数据
	details := slice.Map(cmdbResult.Info, func(ch cmdb.Host) *corecvm.Cvm[corecvm.TCloudZiyanCvmExtension] {
		return &corecvm.Cvm[corecvm.TCloudZiyanCvmExtension]{
			BaseCvm: corecvm.BaseCvm{
				CloudID:              ch.BkCloudInstID,
				Vendor:               enumor.TCloudZiyan,
				BkBizID:              bizID,
				AccountID:            req.AccountID,
				Region:               ch.BkCloudRegion,
				Zone:                 ch.BkCloudZone,
				PrivateIPv4Addresses: strings.Split(ch.BkHostInnerIP, ","),
				PrivateIPv6Addresses: strings.Split(ch.BkHostInnerIPv6, ","),
				PublicIPv4Addresses:  strings.Split(ch.BkHostOuterIP, ","),
				PublicIPv6Addresses:  strings.Split(ch.BkHostOuterIPv6, ","),
				CloudVpcIDs:          []string{ch.BkCloudVpcID},
				CloudSubnetIDs:       []string{ch.BkCloudSubnetID},
			},
		}
	})
	resp.Details = details
	return resp, nil
}

// QueryFromCloud 从cc处拉取自研云主机
func (svc *cvmSvc) queryFromCloud(kt *kit.Kit, accountID string, cmdbHosts []cmdb.Host) (
	hosts []*corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], err error) {

	instIds := make([]string, 0, len(cmdbHosts))
	instIDsByRegion := make(map[string][]string)
	for _, host := range cmdbHosts {
		if host.BkCloudRegion == "" {
			continue
		}
		instIds = append(instIds, host.BkCloudInstID)
		instIDsByRegion[host.BkCloudRegion] = append(instIDsByRegion[host.BkCloudRegion], host.BkCloudInstID)
	}
	hosts = make([]*corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], 0, len(cmdbHosts))
	hostMap := make(map[string]*corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], len(cmdbHosts))
	// 获取自研云账号id
	for region, cloudIDs := range instIDsByRegion {
		// 分批获取
		partIDs := slice.Split(cloudIDs, typecore.TCloudQueryLimit)
		for _, ids := range partIDs {
			csReq := &corecvm.QueryCloudCvmReq{
				AccountID: accountID,
				Vendor:    enumor.TCloudZiyan,
				Region:    region,
				CvmIDs:    ids,
				Page:      &core.BasePage{Start: 0, Limit: typecore.TCloudQueryLimit},
			}
			cloudHostsPtr, err := svc.client.HCService().TCloudZiyan.Cvm.QueryTCloudZiyanCVM(kt, csReq)
			if err != nil {
				logs.Errorf("fail to query tcloud ziyan cvm, err: %v, cloud_ids: %v, rid:%s",
					err, cloudIDs, kt.Rid)
				return nil, err
			}
			for _, vm := range cloudHostsPtr.Details {
				hostMap[vm.CloudID] = cvt.ValToPtr(vm)
			}
		}
	}

	details := make([]*corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], 0, len(instIds))
	for _, id := range instIds {
		inst, exists := hostMap[id]
		if !exists {
			logs.Warnf("cmdb host not found on cloud, cloud_inst_id: %v, rid: %s", id, kt.Rid)
			continue
		}
		details = append(details, inst)
	}
	sgCloudIDs := make([]string, 0, len(details))
	for _, detail := range details {
		sgCloudIDs = append(sgCloudIDs, detail.Extension.CloudSecurityGroupIDs...)
	}
	sgNameMap, err := svc.cvmLgc.QuerySecurityGroupNamesByCloudID(kt, enumor.TCloudZiyan, sgCloudIDs)
	if err != nil {
		logs.Errorf("fail to query security group names, err: %v, sg_cloud_ids: %v, rid: %s",
			err, sgCloudIDs, kt.Rid)
		return nil, err
	}
	for i := range details {
		for _, sgid := range details[i].Extension.CloudSecurityGroupIDs {
			details[i].Extension.SecurityGroupNames = append(details[i].Extension.SecurityGroupNames, sgNameMap[sgid])
		}
	}
	return details, nil
}

// listTCloudZiyanCvmOperateHost 获取自研云主机列表&可操作状态
func (svc *cvmSvc) listTCloudZiyanCvmOperateHost(kt *kit.Kit, cvmIDs []string,
	validateStatusFunc validateOperateStatusFunc) ([]cscvm.CvmBatchOperateHostInfo, error) {

	// 根据主机ID获取主机列表
	hostIDs, hostCvmMap, err := svc.listTCloudZiyanCvmExtMapByIDs(kt, cvmIDs)
	if err != nil {
		return nil, err
	}

	// 查询cc的Topo关系
	mapHostToRel, mapModuleIdToModule, err := svc.listCmdbHostRelModule(kt, hostIDs)
	if err != nil {
		return nil, err
	}

	mapCloudIDToCvm, err := svc.mapTCloudZiyanCloudCvms(kt, hostCvmMap)
	if err != nil {
		logs.Errorf("fail to map tcloud ziyan cloud cvms, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cvmHosts := make([]cscvm.CvmBatchOperateHostInfo, 0)
	for _, host := range hostCvmMap {
		hostID := host.Extension.HostID
		moduleName := ""
		if rel, ok := mapHostToRel[hostID]; ok {
			if module, exist := mapModuleIdToModule[rel.BkModuleId]; exist {
				moduleName = module.BkModuleName
			}
		}

		cloudCvm, ok := mapCloudIDToCvm[host.CloudID]
		if !ok {
			logs.Warnf("cloud cvm not found, cloud_id: %v, host_id: %v, rid: %s", host.CloudID, hostID, kt.Rid)
			return nil, fmt.Errorf("cloud cvm not found, cloud_id: %v, host_id: %v", host.CloudID, hostID)
		}

		cvmHosts = append(cvmHosts, cscvm.CvmBatchOperateHostInfo{
			ID:                   host.ID,
			Vendor:               host.Vendor,
			AccountID:            host.AccountID,
			BkHostID:             host.Extension.HostID,
			BkHostName:           host.Name,
			CloudID:              hostCvmMap[hostID].CloudID,
			BkAssetID:            host.Extension.BkAssetID,
			PrivateIPv4Addresses: hostCvmMap[hostID].PrivateIPv4Addresses,
			PrivateIPv6Addresses: hostCvmMap[hostID].PrivateIPv6Addresses,
			PublicIPv4Addresses:  hostCvmMap[hostID].PublicIPv4Addresses,
			PublicIPv6Addresses:  hostCvmMap[hostID].PublicIPv6Addresses,
			Operator:             host.Extension.Operator,
			BkBakOperator:        host.Extension.BkBakOperator,
			DeviceType:           host.Extension.SvrDeviceClass,
			Region:               hostCvmMap[hostID].Region,
			Zone:                 hostCvmMap[hostID].Zone,
			BkOSName:             host.Extension.BkOSName,
			TopoModule:           moduleName,
			SvrSourceTypeID:      host.Extension.SvrSourceTypeID,
			Status:               hostCvmMap[hostID].Status,
			SrvStatus:            host.Extension.SrvStatus,
			OperateStatus:        validateStatusFunc(kt.User, moduleName, cloudCvm.Status, host),
		})
	}

	return cvmHosts, nil
}

// mapTCloudZiyanCloudCvms 查询云上的主机信息
func (svc *cvmSvc) mapTCloudZiyanCloudCvms(kt *kit.Kit, cvms map[int64]corecvm.Cvm[corecvm.TCloudZiyanHostExtension]) (
	map[string]corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], error) {

	// group by account, region
	mapAccountRegionToCvmCloudID := make(map[string][]string)
	for _, host := range cvms {
		key := getCombinedKey(host.AccountID, host.Region, "+")
		mapAccountRegionToCvmCloudID[key] = append(mapAccountRegionToCvmCloudID[key], host.CloudID)
	}

	result := make(map[string]corecvm.Cvm[corecvm.TCloudZiyanCvmExtension])
	for key, cloudIDs := range mapAccountRegionToCvmCloudID {
		split := strings.Split(key, "+")
		accountID, region := split[0], split[1]
		if region == "" {
			logs.Errorf("region is empty, account_id: %s, cvm_ids: %v, rid: %s", accountID, cloudIDs, kt.Rid)
			return nil, fmt.Errorf("region is empty, account_id: %s, cvm_ids: %v", accountID, cloudIDs)
		}

		for _, ids := range slice.Split(cloudIDs, typecore.TCloudQueryLimit) {
			req := &corecvm.QueryCloudCvmReq{
				Vendor:    enumor.TCloudZiyan,
				AccountID: accountID,
				Region:    region,
				CvmIDs:    ids,
				Page:      &core.BasePage{Start: 0, Limit: typecore.TCloudQueryLimit},
			}
			resp, err := svc.client.HCService().TCloudZiyan.Cvm.QueryTCloudZiyanCVM(kt, req)
			if err != nil {
				logs.Errorf("fail to query tcloud ziyan cvm, err: %v, cloud_ids: %v, rid:%s", err, cloudIDs, kt.Rid)
				return nil, err
			}
			for _, detail := range resp.Details {
				result[detail.CloudID] = detail
			}
		}
	}
	return result, nil
}

// 拼接唯一key main-sub
func getCombinedKey(main, sub, sep string) string {
	return main + sep + sub
}

// listCmdbHostRelModule 查询cc的主机列表及Topo关系
func (svc *cvmSvc) listCmdbHostRelModule(kt *kit.Kit, hostIDs []int64) (map[int64]*cmdb.HostBizRel,
	map[int64]*cmdb.ModuleInfo, error) {

	// get host topo info
	relations, err := svc.cvmLgc.GetHostTopoInfo(kt, hostIDs)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}

	bizIds := make([]int64, 0)
	mapBizToModule := make(map[int64][]int64)
	mapHostToRel := make(map[int64]*cmdb.HostBizRel)
	for _, rel := range relations {
		mapHostToRel[rel.BkHostId] = rel
		if _, ok := mapBizToModule[rel.BkBizId]; !ok {
			mapBizToModule[rel.BkBizId] = []int64{rel.BkModuleId}
			bizIds = append(bizIds, rel.BkBizId)
		} else {
			mapBizToModule[rel.BkBizId] = append(mapBizToModule[rel.BkBizId], rel.BkModuleId)
		}
	}

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := util.IntArrayUnique(moduleIds)
		moduleList, err := svc.cvmLgc.GetModuleInfo(kt, bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to cvm reset check, for get module info, err: %v, bizId: %d, "+
				"moduleIdUniq: %v, rid: %s", err, bizId, moduleIdUniq, kt.Rid)
			return nil, nil, err
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}
	// 记录日志
	hostRelJson, err := json.Marshal(mapHostToRel)
	if err != nil {
		logs.Errorf("failed to marshal mapHostToRel, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}
	moduleIdToModuleJson, err := json.Marshal(mapModuleIdToModule)
	if err != nil {
		logs.Errorf("failed to marshal mapModuleIdToModule, err: %v, hostIDs: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, nil, err
	}
	logs.Infof("list cmdb host rel module success, hostIDs: %v, mapHostToRel: %s, mapModuleIdToModule: %s, rid: %s",
		hostIDs, hostRelJson, moduleIdToModuleJson, kt.Rid)

	return mapHostToRel, mapModuleIdToModule, nil
}

// listTCloudZiyanCvmExtMapByIDs 根据主机ID获取主机列表（含扩展信息）
func (svc *cvmSvc) listTCloudZiyanCvmExtMapByIDs(kt *kit.Kit, cvmIDs []string) (
	[]int64, map[int64]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], error) {

	// 查询云主机的扩展信息
	extReq := &dataproto.CvmListReq{
		Filter: tools.ContainersExpression("id", cvmIDs),
		Page:   core.NewDefaultBasePage(),
	}
	cvmExtList := make([]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], 0)
	for {
		extResp, err := svc.client.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), extReq)
		if err != nil {
			logs.Errorf("fail to list tcloud ziyan cvm ext map, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, kt.Rid)
			return nil, nil, err
		}

		cvmExtList = append(cvmExtList, extResp.Details...)
		if len(extResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		extReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	hostIDs := make([]int64, 0)
	hostCvmMap := make(map[int64]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], 0)
	for _, item := range cvmExtList {
		if item.Extension == nil || item.Extension.HostID == 0 {
			continue
		}
		hostCvmMap[item.Extension.HostID] = item
		hostIDs = append(hostIDs, item.Extension.HostID)
	}
	hostIDs = slice.Unique(hostIDs)
	return hostIDs, hostCvmMap, nil
}
