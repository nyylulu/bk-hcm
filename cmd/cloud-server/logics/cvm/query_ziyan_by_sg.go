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
	typecore "hcm/pkg/adaptor/types/core"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// queryTCloudZiyanCvmBySGID 查询自研云安全组下的主机
func (c *cvm) queryTCloudZiyanCvmBySGIDWithSGName(kt *kit.Kit, bizID int64, sgInfo *types.CloudResourceBasicInfo) (
	*core.ListResultT[corecvm.Cvm[corecvm.TCloudZiyanCvmExtension]], error) {

	cvmList, err := c.queryTCloudZiyanCvmBySGID(kt, bizID, sgInfo)
	if err != nil {
		return nil, err
	}
	sgCloudIDs := make([]string, 0, len(cvmList))
	for _, cvm := range cvmList {
		sgCloudIDs = append(sgCloudIDs, cvm.Extension.CloudSecurityGroupIDs...)
	}
	sgNameMap, err := c.QuerySecurityGroupNamesByCloudID(kt, enumor.TCloudZiyan, sgCloudIDs)
	if err != nil {
		return nil, err
	}
	for i, cvm := range cvmList {
		for _, cloudSGID := range cvm.Extension.CloudSecurityGroupIDs {
			cvmList[i].Extension.SecurityGroupNames = append(
				cvmList[i].Extension.SecurityGroupNames, sgNameMap[cloudSGID])
		}
	}
	return &core.ListResultT[corecvm.Cvm[corecvm.TCloudZiyanCvmExtension]]{
		Count: uint64(len(cvmList)), Details: cvmList}, nil
}

// QuerySecurityGroupNamesByCloudID 根据云id查询安全组名字
func (c *cvm) QuerySecurityGroupNamesByCloudID(kt *kit.Kit, vendor enumor.Vendor,
	sgCloudIds []string) (map[string]string, error) {

	sgCloudIds = slice.Unique(sgCloudIds)

	parts := slice.Split(sgCloudIds, int(core.DefaultMaxPageLimit))
	sgNameMap := make(map[string]string)
	for _, partSGCloudIDs := range parts {
		querySGReq := &cloud.SecurityGroupListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
					filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: partSGCloudIDs},
				},
			},
			Page: core.NewDefaultBasePage(),
		}
		sgList, err := c.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), querySGReq)
		if err != nil {
			logs.Errorf("fail to query security group by cloud ids, err: %v, sg_ids: %v, rid: %s",
				err, partSGCloudIDs, kt.Rid)
			return nil, err
		}
		for _, sg := range sgList.Details {
			sgNameMap[sg.CloudID] = sg.Name
		}
	}
	return sgNameMap, nil
}

// queryTCloudZiyanCvmBySGID 查询自研云安全组下的主机
func (c *cvm) queryTCloudZiyanCvmBySGID(kt *kit.Kit, bizID int64, sgInfo *types.CloudResourceBasicInfo) (
	[]corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], error) {

	// 查询云上id
	securityGroup, err := c.client.DataService().TCloudZiyan.SecurityGroup.GetSecurityGroup(kt, sgInfo.ID)
	if err != nil {
		logs.Errorf("fail to query security group, err: %v, sg_id: %s, rid: %s", err, sgInfo.ID, kt.Rid)
		return nil, err
	}
	// 1. 根据安全组id 去云上查询主机
	var req = &corecvm.QueryCloudCvmReq{
		Vendor:    enumor.TCloudZiyan,
		AccountID: securityGroup.AccountID,
		Region:    securityGroup.Region,
		SGIDs:     []string{securityGroup.CloudID},
		Page:      &core.BasePage{Limit: typecore.TCloudQueryLimit},
	}
	allBizCvm := make([]corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], 0)
	offset := uint32(0)
	// 全量返回
	for {
		req.Page.Start = offset
		cloudCvms, err := c.client.HCService().TCloudZiyan.Cvm.QueryTCloudZiyanCVM(kt, req)
		if err != nil {
			logs.Errorf("fail to query cvm by sg id, err: %v, sg_cloud_id: %v, rid: %s",
				err, securityGroup.CloudID, kt.Rid)
			return nil, err
		}

		if len(cloudCvms.Details) == 0 {
			break
		}
		offset += uint32(len(cloudCvms.Details))

		bizCvms, err := c.filterCmdbBizHost(kt, bizID, cloudCvms.Details)
		if err != nil {
			return nil, err
		}
		allBizCvm = append(allBizCvm, bizCvms...)
	}
	return allBizCvm, nil
}

func (c *cvm) filterCmdbBizHost(kt *kit.Kit, bizID int64, cloudCvms []corecvm.Cvm[corecvm.TCloudZiyanCvmExtension]) (
	bizCvmList []corecvm.Cvm[corecvm.TCloudZiyanCvmExtension], err error) {

	//  资源下直接返回所有cvm
	if bizID == constant.UnassignedBiz {
		return cloudCvms, nil
	}

	// 否则去cmdb 上过滤
	cloudCvmMap := map[string]corecvm.Cvm[corecvm.TCloudZiyanCvmExtension]{}
	for _, detail := range cloudCvms {
		cloudCvmMap[detail.CloudID] = detail
	}

	cmdbReq := &cscvm.CmdbHostQueryReq{
		BkBizID:      bizID,
		Vendor:       enumor.TCloudZiyan,
		CloudInstIDs: maps.Keys(cloudCvmMap),
		Page:         &cmdb.BasePage{Limit: typecore.TCloudQueryLimit, Start: 0},
	}
	cmdbResult, err := c.GetCmdbBizHosts(kt, cmdbReq)

	if err != nil {
		logs.Errorf("call cmdb to list biz host by sg id failed, err: %v, req: %+v, rid: %s", err, cmdbReq, kt.Rid)
		return nil, err
	}
	// filter biz cvm
	for _, host := range cmdbResult.Info {
		bizCvmList = append(bizCvmList, cloudCvmMap[host.BkCloudInstID])
	}
	return bizCvmList, nil
}
