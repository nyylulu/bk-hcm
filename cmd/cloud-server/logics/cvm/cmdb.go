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
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// GetCmdbBizHosts 获取cmdb业务拓扑下主机
func (c *cvm) GetCmdbBizHosts(kt *kit.Kit, req *cscvm.CmdbHostQueryReq) (*cmdb.ListBizHostResult, error) {

	ccVendor, exists := cmdb.HcmCmdbVendorMap[req.Vendor]
	if !exists {
		return nil, errf.New(errf.InvalidParameter, "not supported vendor: %s"+string(req.Vendor))
	}
	combinedRule := &cmdb.CombinedRule{Condition: "AND",
		Rules: []cmdb.Rule{&cmdb.AtomRule{Field: "bk_cloud_vendor", Operator: cmdb.OperatorEqual, Value: ccVendor}}}
	if req.Region != "" {
		// 筛选地域
		combinedRule.Rules = append(combinedRule.Rules,
			&cmdb.AtomRule{Field: "bk_cloud_region", Operator: cmdb.OperatorEqual, Value: req.Region})
	}
	if len(req.CloudInstIDs) != 0 {
		combinedRule.Rules = append(combinedRule.Rules, &cmdb.AtomRule{
			Field: "bk_cloud_inst_id", Operator: cmdb.OperatorIn, Value: req.CloudInstIDs})
	}
	params := &cmdb.ListBizHostParams{
		BizID:              req.BkBizID,
		BkSetIDs:           req.BkSetIDs,
		BkModuleIDs:        req.BkModuleIDs,
		Fields:             cmdb.HostFields,
		Page:               req.Page,
		HostPropertyFilter: &cmdb.QueryFilter{Rule: combinedRule},
	}
	cmdbResult, err := c.esbClient.Cmdb().ListBizHost(kt, params)
	if err != nil {
		logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	return cmdbResult, nil
}
