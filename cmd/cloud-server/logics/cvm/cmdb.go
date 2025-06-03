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
	"hcm/pkg"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	cvt "hcm/pkg/tools/converter"
)

// GetCmdbBizHosts 获取cmdb业务拓扑下主机
func (c *cvm) GetCmdbBizHosts(kt *kit.Kit, req *cscvm.CmdbHostQueryReq) (*cmdb.ListBizHostResult, error) {

	var combinedRule = cmdb.CombinedRule{Condition: "AND", Rules: make([]cmdb.Rule, 0)}
	if req.Vendor != "" {
		ccVendor, exists := cmdb.HcmCmdbVendorMap[req.Vendor]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, "not supported vendor: %s"+string(req.Vendor))
		}
		combinedRule.Rules = append(combinedRule.Rules, cmdb.Equal("bk_cloud_vendor", ccVendor))
	}
	if req.Region != "" {
		// 筛选地域
		combinedRule.Rules = append(combinedRule.Rules, cmdb.Equal("bk_cloud_region", req.Region))
	}
	if req.Zone != "" {
		// 筛选可用区
		combinedRule.Rules = append(combinedRule.Rules, cmdb.Equal("bk_cloud_zone", req.Zone))
	}
	if len(req.CloudInstIDs) != 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_cloud_inst_id", req.CloudInstIDs))
	}
	if len(req.CloudVpcIDs) != 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_cloud_vpc_id", req.CloudVpcIDs))
	}
	if len(req.CloudSubnetIDs) != 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_cloud_subnet_id", req.CloudSubnetIDs))
	}
	if len(req.InnerIP) > 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_host_innerip", req.InnerIP))
	}
	if len(req.OuterIP) > 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_host_outerip", req.OuterIP))
	}
	if len(req.InnerIPv6) > 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_host_innerip_v6", req.InnerIPv6))
	}
	if len(req.OuterIPv6) > 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_host_outerip_v6", req.OuterIPv6))
	}
	if len(req.BkHostIDs) > 0 {
		combinedRule.Rules = append(combinedRule.Rules, cmdb.In("bk_host_id", req.BkHostIDs))
	}

	params := &cmdb.ListBizHostParams{
		BizID:              req.BkBizID,
		BkSetIDs:           req.BkSetIDs,
		BkModuleIDs:        req.BkModuleIDs,
		Fields:             cmdb.HostFields,
		Page:               req.Page,
		HostPropertyFilter: &cmdb.QueryFilter{Rule: combinedRule},
	}
	cmdbResult, err := c.cmdbClient.ListBizHost(kt, params)
	if err != nil {
		logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}
	return cmdbResult, nil
}

// GetHostTopoInfo get host topo info in cc 3.0
func (c *cvm) GetHostTopoInfo(kt *kit.Kit, hostIds []int64) ([]cmdb.HostTopoRelation, error) {
	req := &cmdb.HostModuleRelationParams{
		HostID: hostIds,
	}

	resp, err := c.cmdbClient.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return cvt.PtrToVal(resp), nil
}

// GetModuleInfo get module info in cc 3.0
func (c *cvm) GetModuleInfo(kit *kit.Kit, bkBizID int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleParams{
		BizID: bkBizID,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				pkg.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}
	resp, err := c.cmdbClient.SearchModule(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return resp.Info, nil
}
