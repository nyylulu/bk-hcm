/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"fmt"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/maps"
)

// RuleFactory 类型别名
type RuleFactory = filter.RuleFactory

// ListUrlRulesByTopo list url rules by topo
func (svc *lbSvc) ListUrlRulesByTopo(cts *rest.Contexts) (any, error) {
	req := new(cslb.ListUrlRulesByTopologyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	return svc.listUrlRulesByTopo(cts.Kit, bizID, vendor, req)
}

func (svc *lbSvc) listUrlRulesByTopo(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (any, error) {
	info, err := svc.getUrlRuleTopoInfoByReq(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("list url rule topo info by req failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if !info.Match {
		return &cslb.ListUrlRulesByTopologyResp{Count: 0, Details: make([]cslb.UrlRuleDetail, 0)}, nil
	}

	ruleCond := make([]RuleFactory, 0)
	ruleCond = append(ruleCond, info.RuleCond...)
	ruleFilters := req.BuildRuleFilter()
	for _, rule := range ruleFilters {
		ruleCond = append(ruleCond, rule)
	}

	page := req.Page
	if page == nil {
		page = core.NewDefaultBasePage()
	}

	if len(info.RuleCond) > 0 {

		if ruleIDRule, ok := info.RuleCond[0].(*filter.AtomRule); ok && ruleIDRule.Field == "id" {
			if ruleIDs, ok := ruleIDRule.Value.([]string); ok && len(ruleIDs) > 500 {
				page.Limit = uint(uint32(len(ruleIDs) + 100))
				logs.Infof("setting larger limit for rule query: %d, rid: %s", page.Limit, kt.Rid)
			}
		}
	}

	ruleReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: ruleCond},
		Page:   page,
	}
	resp := &cloud.TCloudURLRuleListResult{}
	switch vendor {
	case enumor.TCloud:
		resp, err = svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &ruleReq)
		if err != nil {
			logs.Errorf("get url rule failed, err: %v, req: %+v, rid: %s", err, ruleReq, kt.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
	if req.Page.Count {
		return &cslb.ListUrlRulesByTopologyResp{Count: int(resp.Count)}, nil
	}
	if len(resp.Details) == 0 {
		return &cslb.ListUrlRulesByTopologyResp{Count: 0, Details: make([]cslb.UrlRuleDetail, 0)}, nil
	}

	details, err := svc.buildUrlRuleDetail(kt, info, resp.Details)
	if err != nil {
		logs.Errorf("build url rule detail failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return &cslb.ListUrlRulesByTopologyResp{Count: int(resp.Count), Details: details}, nil
}

// getUrlRuleTopoInfoByReq 复用监听器查询的逻辑结构
func (svc *lbSvc) getUrlRuleTopoInfoByReq(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.ListUrlRulesByTopologyReq) (
	*cslb.UrlRuleTopoInfo, error) {

	// 检查监听器协议，如果是四层协议（TCP/UDP），直接返回空结果
	if len(req.LblProtocols) > 0 {
		hasLayer4Protocol := false
		hasLayer7Protocol := false
		for _, protocol := range req.LblProtocols {
			if protocol == "TCP" || protocol == "UDP" {
				hasLayer4Protocol = true
			} else if protocol == "HTTP" || protocol == "HTTPS" {
				hasLayer7Protocol = true
			}
		}
		if hasLayer4Protocol && !hasLayer7Protocol {
			return &cslb.UrlRuleTopoInfo{Match: false}, nil
		}
	}

	commonCond := make([]filter.RuleFactory, 0)
	commonCond = append(commonCond, tools.RuleEqual("bk_biz_id", bizID))
	commonCond = append(commonCond, tools.RuleEqual("vendor", vendor))
	commonCond = append(commonCond, tools.RuleEqual("account_id", req.AccountID))

	// 根据条件查询clb信息
	lbCond := make([]filter.RuleFactory, 0)
	lbCond = append(lbCond, commonCond...)
	lbCond = append(lbCond, req.GetLbCond()...)
	lbMap, err := svc.getLbByCond(kt, lbCond)
	if err != nil {
		logs.Errorf("get lb by cond failed, err: %v, lbCond: %v, rid: %s", err, lbCond, kt.Rid)
		return nil, err
	}

	if len(lbMap) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}
	lbIDs := maps.Keys(lbMap)
	reqRuleCond := req.GetRuleCond()
	reqTargetCond := req.GetTargetCond()

	// 如果请求没有规则和RS条件，需要查询监听器和规则来构建完整的拓扑信息
	if len(reqRuleCond) == 0 && len(reqTargetCond) == 0 {
		// 查询所有监听器
		lblCond := []filter.RuleFactory{tools.RuleIn("lb_id", lbIDs)}
		// 如果指定了协议，添加协议过滤条件
		if len(req.LblProtocols) > 0 {
			lblCond = append(lblCond, tools.RuleIn("protocol", req.LblProtocols))
		}
		lblMap, err := svc.getLblByCond(kt, vendor, lblCond)
		if err != nil {
			logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
			return nil, err
		}
		if len(lblMap) == 0 {
			return &cslb.UrlRuleTopoInfo{Match: false}, nil
		}
		lblIDs := maps.Keys(lblMap)

		ruleCond := []filter.RuleFactory{tools.RuleIn("lbl_id", lblIDs)}
		ruleMap, err := svc.getRuleByCond(kt, vendor, ruleCond)
		if err != nil {
			logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
			return nil, err
		}
		if len(ruleMap) == 0 {
			return &cslb.UrlRuleTopoInfo{Match: false}, nil
		}

		// 返回所有规则的ID作为查询条件
		ruleIDs := maps.Keys(ruleMap)
		finalRuleCond := []filter.RuleFactory{tools.RuleIn("id", ruleIDs)}
		return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, LblMap: lblMap, RuleCond: finalRuleCond}, nil
	}

	// 根据条件查询监听器信息
	lblCond := make([]filter.RuleFactory, 0)
	lblCond = append(lblCond, tools.RuleIn("lb_id", lbIDs))
	lblCond = append(lblCond, req.GetLblCond()...)
	lblMap, err := svc.getLblByCond(kt, vendor, lblCond)
	if err != nil {
		logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
		return nil, err
	}
	if len(lblMap) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}
	lblIDs := maps.Keys(lblMap)

	ruleCond := make([]filter.RuleFactory, 0)
	ruleCond = append(ruleCond, tools.RuleIn("lbl_id", lblIDs))
	ruleCond = append(ruleCond, reqRuleCond...)
	ruleMap, err := svc.getRuleByCond(kt, vendor, ruleCond)
	if err != nil {
		logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
		return nil, err
	}
	if len(ruleMap) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	// 如果请求中不含RS的条件，那么可以直接返回规则条件
	if len(reqTargetCond) == 0 {
		ruleIDs := maps.Keys(ruleMap)
		ruleCond := []filter.RuleFactory{tools.RuleIn("id", ruleIDs)}
		return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, LblMap: lblMap, RuleCond: ruleCond}, nil
	}

	// 根据RS条件查询，得到规则条件
	ruleIDs := maps.Keys(ruleMap)
	tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", ruleIDs),
		tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}

	ruleCond, err = svc.getRuleCondByTargetCond(kt, tgLbRelCond, reqTargetCond)
	if err != nil {
		logs.Errorf("get rule cond by target cond failed, err: %v, tgLbRelCond: %v, reqTargetCond: %v, rid: %s", err,
			tgLbRelCond, reqTargetCond, kt.Rid)
		return nil, err
	}
	if len(ruleCond) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, LblMap: lblMap, RuleCond: ruleCond}, nil
}

func (svc *lbSvc) buildUrlRuleDetail(kt *kit.Kit, info *cslb.UrlRuleTopoInfo,
	urlRules []corelb.TCloudLbUrlRule) ([]cslb.UrlRuleDetail, error) {

	ruleIDTargetCountMap, err := svc.getUrlRuleTargetCount(kt, urlRules)
	if err != nil {
		logs.Errorf("get url rule target count failed, err: %v, urlRules: %+v, rid: %s", err, urlRules, kt.Rid)
		return nil, err
	}

	details := make([]cslb.UrlRuleDetail, 0)
	for _, rule := range urlRules {
		lb, ok := info.LbMap[rule.LbID]
		if !ok {
			logs.Errorf("lb not found, lbID: %s, rid: %s", rule.LbID, kt.Rid)
			return nil, fmt.Errorf("lb not found, lbID: %s", rule.LbID)
		}

		lbl, ok := info.LblMap[rule.LblID]
		if !ok {
			logs.Errorf("lbl not found, lblID: %s, rid: %s", rule.LblID, kt.Rid)
			return nil, fmt.Errorf("lbl not found, lblID: %s", rule.LblID)
		}

		// 获取CLB的IP地址
		ip := ""
		if len(lb.PublicIPv4Addresses) > 0 {
			ip = lb.PublicIPv4Addresses[0]
		} else if len(lb.PrivateIPv4Addresses) > 0 {
			ip = lb.PrivateIPv4Addresses[0]
		} else if len(lb.PublicIPv6Addresses) > 0 {
			ip = lb.PublicIPv6Addresses[0]
		} else if len(lb.PrivateIPv6Addresses) > 0 {
			ip = lb.PrivateIPv6Addresses[0]
		} else if lb.Domain != "" {
			ip = lb.Domain
		}

		detail := cslb.UrlRuleDetail{
			ID:          rule.ID,
			LbVips:      []string{ip},
			LblProtocol: string(lbl.Protocol),
			LblPort:     int(lbl.Port),
			RuleUrl:     rule.URL,
			RuleDomain:  rule.Domain,
			TargetCount: ruleIDTargetCountMap[rule.ID],
			CloudLblID:  lbl.CloudID,
			CloudLbID:   lb.CloudID,
		}
		details = append(details, detail)
	}

	return details, nil
}

// getUrlRuleTargetCount 获取规则的目标数量
func (svc *lbSvc) getUrlRuleTargetCount(kt *kit.Kit, rules []corelb.TCloudLbUrlRule) (map[string]int, error) {
	if len(rules) == 0 {
		return make(map[string]int), nil
	}

	tgIDs := make([]string, 0)
	ruleIDTgIDMap := make(map[string]string)
	for _, rule := range rules {
		if rule.TargetGroupID != "" {
			tgIDs = append(tgIDs, rule.TargetGroupID)
			ruleIDTgIDMap[rule.ID] = rule.TargetGroupID
		}
	}

	if len(tgIDs) == 0 {
		ruleTargetCountMap := make(map[string]int)
		for _, rule := range rules {
			ruleTargetCountMap[rule.ID] = 0
		}
		return ruleTargetCountMap, nil
	}

	// 查询目标
	targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", tgIDs)}
	targets, err := svc.getTargetByCond(kt, targetCond)
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
		return nil, err
	}

	// 计算每个规则的目标数量
	ruleTargetCountMap := make(map[string]int)
	for _, rule := range rules {
		ruleTargetCountMap[rule.ID] = 0
	}

	// 统计每个目标组的目标数量
	tgIDTargetCountMap := make(map[string]int)
	for _, target := range targets {
		tgIDTargetCountMap[target.TargetGroupID]++
	}

	// 将目标组的目标数量分配给对应的规则
	for ruleID, tgID := range ruleIDTgIDMap {
		ruleTargetCountMap[ruleID] = tgIDTargetCountMap[tgID]
	}

	return ruleTargetCountMap, nil
}

// getRuleCondByTargetCond 复用监听器查询的 getLblCondByTargetCond 逻辑
func (svc *lbSvc) getRuleCondByTargetCond(kt *kit.Kit, tgLbRelCond []RuleFactory,
	reqTargetCond []filter.RuleFactory) ([]RuleFactory, error) {

	// 根据条件查询clb和目标组关系
	tgLbRels, err := svc.getTgLbRelByCond(kt, tgLbRelCond)
	if err != nil {
		logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
		return nil, err
	}
	if len(tgLbRels) == 0 {
		return make([]RuleFactory, 0), nil
	}

	tgIDMap := make(map[string]struct{})
	tgIDRuleIDMap := make(map[string]string)
	for _, tgLbRel := range tgLbRels {
		tgIDMap[tgLbRel.TargetGroupID] = struct{}{}
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
	}

	// 根据条件查询RS
	targetCond := []RuleFactory{tools.RuleIn("target_group_id", maps.Keys(tgIDMap))}
	targetCond = append(targetCond, reqTargetCond...)
	targets, err := svc.getTargetByCond(kt, targetCond)
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
		return nil, err
	}
	if len(targets) == 0 {
		return make([]RuleFactory, 0), nil
	}

	// 根据RS反向推出匹配的规则条件
	ruleIDMap := make(map[string]struct{})
	for _, target := range targets {
		ruleID, ok := tgIDRuleIDMap[target.TargetGroupID]
		if !ok {
			logs.Errorf("use target group id not found rule, tgID: %s, rid: %s", target.TargetGroupID, kt.Rid)
			return nil, fmt.Errorf("use target group not found rule, tgID: %s", target.TargetGroupID)
		}
		ruleIDMap[ruleID] = struct{}{}
	}
	ruleIDs := maps.Keys(ruleIDMap)
	if len(ruleIDs) == 0 {
		return make([]RuleFactory, 0), nil
	}

	return []RuleFactory{tools.RuleIn("id", ruleIDs)}, nil
}
