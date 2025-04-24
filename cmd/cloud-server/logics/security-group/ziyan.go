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

package securitygroup

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
)

// listTCloudZiyanSecurityGroupRulesByCloudTargetSGID 查询来源为安全组的 安全组规则,
// 返回 map[安全组ID]安全组规则ID列表
func listTCloudZiyanSecurityGroupRulesByCloudTargetSGID(kt *kit.Kit, cli *dataservice.Client, sgID string) (
	map[string][]string, error) {

	sgCloudIDToRuleIDs := make(map[string][]string)
	listReq := &dataproto.TCloudSGRuleListReq{
		Filter: tools.ExpressionAnd(tools.RuleNotEqual("cloud_target_security_group_id", "")),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		resp, err := cli.TCloudZiyan.SecurityGroup.ListSecurityGroupRule(kt, listReq, sgID)
		if err != nil {
			return nil, err
		}
		for _, rule := range resp.Details {
			if rule.CloudTargetSecurityGroupID != nil {
				cloudID := *rule.CloudTargetSecurityGroupID
				sgCloudIDToRuleIDs[cloudID] = append(sgCloudIDToRuleIDs[cloudID], rule.ID)
			}
		}
		if len(resp.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return sgCloudIDToRuleIDs, nil
}
