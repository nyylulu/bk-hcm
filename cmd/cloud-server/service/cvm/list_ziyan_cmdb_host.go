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
	"strings"

	typecore "hcm/pkg/adaptor/types/core"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ListZiyanCmdbHost 从cc处拉取自研云主机, 支持二次拉取到云上信息
func (c *cvmSvc) ListZiyanCmdbHost(cts *rest.Contexts) (any, error) {
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
		&handler.ListAuthResOption{Authorizer: c.authorizer, ResType: meta.Cvm, Action: meta.Find})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &webserver.CloudHostListRespT[corecvm.TCloudZiyanCvmExtension]{Count: 0}, nil
	}

	// 1. 获取cmdb 业务下主机列表
	cmdbResult, err := c.cvmLgc.GetCmdbBizHosts(cts.Kit, &cscvm.CmdbHostQueryReq{
		BkBizID:      bizID,
		Vendor:       enumor.TCloudZiyan,
		AccountID:    req.AccountID,
		Region:       req.Region,
		CloudInstIDs: req.CloudInstIDs,
		BkSetIDs:     req.BkSetIDs,
		BkModuleIDs:  req.BkModuleIDs,
		Page:         req.Page,
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
		hosts, err := c.queryFromCloud(cts.Kit, req.AccountID, cmdbResult.Info)
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
				PrivateIPv4Addresses: strings.Split(ch.BkHostInnerIP, ","),
				PrivateIPv6Addresses: strings.Split(ch.BkHostInnerIPv6, ","),
				PublicIPv4Addresses:  strings.Split(ch.BkHostOuterIP, ","),
				PublicIPv6Addresses:  strings.Split(ch.BkHostOuterIPv6, ","),
			},
		}
	})
	resp.Details = details
	return resp, nil
}

// QueryFromCloud 从cc处拉取自研云主机
func (c *cvmSvc) queryFromCloud(kt *kit.Kit, accountID string, cmdbHosts []cmdb.Host) (
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
			cloudHostsPtr, err := c.client.HCService().TCloudZiyan.Cvm.QueryTCloudZiyanCVM(kt, csReq)
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
	sgNameMap, err := c.cvmLgc.QuerySecurityGroupNamesByCloudID(kt, enumor.TCloudZiyan, sgCloudIDs)
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
