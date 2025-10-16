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

package loadbalancer

import (
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

func (svc *lbSvc) listUrlRuleByCondForTCloudZiyan(kt *kit.Kit, req cslb.TargetQueryLine, lblIDs []string) ([]string,
	map[string]urlRuleInfo, error) {
	ruleType := enumor.Layer4RuleType
	if req.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
	}
	queryRules := []*filter.AtomRule{
		tools.RuleEqual("rule_type", ruleType),
	}
	if len(req.Domains) > 0 {
		queryRules = append(queryRules, tools.RuleIn("domain", req.Domains))
	}
	if len(req.Urls) > 0 {
		queryRules = append(queryRules, tools.RuleIn("url", req.Urls))
	}
	result := make([]string, 0)
	ruleMap := make(map[string]urlRuleInfo)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				append(queryRules,
					tools.RuleIn("lbl_id", batch),
				)...,
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			resp, err := svc.client.DataService().TCloudZiyan.LoadBalancer.ListUrlRule(kt, listReq)
			if err != nil {
				return nil, nil, err
			}
			for _, detail := range resp.Details {
				result = append(result, detail.ID)
				ruleMap[detail.ID] = urlRuleInfo{
					domain:     detail.Domain,
					url:        detail.URL,
					lblID:      detail.LblID,
					cloudLblID: detail.CloudLBLID,
					cloudLBID:  detail.CloudLbID,
				}
			}
			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return result, ruleMap, nil
}

func (svc *lbSvc) listUrlRuleMapByIDsForTCloudZiyan(kt *kit.Kit, ids []string) (map[string]urlRuleInfo, error) {
	result := make(map[string]urlRuleInfo, 0)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().TCloudZiyan.LoadBalancer.ListUrlRule(kt, listReq)
		if err != nil {
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = urlRuleInfo{
				domain:     detail.Domain,
				url:        detail.URL,
				lblID:      detail.LblID,
				cloudLblID: detail.CloudLBLID,
				cloudLBID:  detail.CloudLbID,
			}
		}
	}
	return result, nil
}
