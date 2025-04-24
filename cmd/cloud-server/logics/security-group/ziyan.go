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
	"fmt"

	ziyanlogic "hcm/cmd/cloud-server/logics/ziyan"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	apitag "hcm/pkg/api/hc-service/tag"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb"
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

// updateTCloudZiyanMgmtAttr 更新自研云安全组的云上管理属性（即标签）
func (s *securityGroup) updateTCloudZiyanMgmtAttr(kt *kit.Kit, mgmtAttrs []proto.BatchUpdateSGMgmtAttrItem,
	sgInfos map[string]cloud.BaseSecurityGroup) error {

	// 将相同的标签合并统一请求
	type tagGroup struct {
		AccountID  string
		BizID      int64
		Manager    string
		BakManager string
	}
	tagGroupMap := make(map[tagGroup][]string)
	for _, attr := range mgmtAttrs {
		sgInfo, ok := sgInfos[attr.ID]
		if !ok {
			logs.Warnf("update tcloud-ziyan tag failed, security group info not found, sg_id: %s, rid: %s",
				attr.ID, kt.Rid)
			continue
		}

		tgroup := tagGroup{
			AccountID:  sgInfo.AccountID,
			BizID:      attr.MgmtBizID,
			Manager:    attr.Manager,
			BakManager: attr.BakManager,
		}
		tagGroupMap[tgroup] = append(tagGroupMap[tgroup], attr.ID)
	}

	for tagStr, sgIDs := range tagGroupMap {
		// 生成业务标签
		tags, err := ziyanlogic.GenTagsForBizsWithManager(kt, esb.EsbClient().Cmdb(), tagStr.BizID,
			tagStr.Manager, tagStr.BakManager)
		if err != nil {
			logs.Errorf("gen tags for biz sg failed, err: %v, biz: %d, sg_ids: %v, rid: %s", err, tagStr.BizID,
				sgIDs, kt.Rid)
			return fmt.Errorf("failed to generate biz tag, err: %w", err)
		}

		resources := make([]apitag.TCloudResourceInfo, 0, len(sgIDs))
		for _, sgID := range sgIDs {
			resources = append(resources, apitag.TCloudResourceInfo{
				Region:     sgInfos[sgID].Region,
				ResType:    enumor.SecurityGroupCloudResType,
				ResCloudID: sgInfos[sgID].CloudID,
			})
		}

		tagReq := &apitag.TCloudBatchTagResRequest{
			AccountID: tagStr.AccountID,
			Resources: resources,
			Tags:      tags,
		}

		_, err = s.client.HCService().TCloudZiyan.Tag.TCloudZiyanBatchTagRes(kt, tagReq)
		if err != nil {
			logs.Errorf("failed to tag sg, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
