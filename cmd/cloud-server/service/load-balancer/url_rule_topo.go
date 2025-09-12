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
	"hcm/pkg/tools/slice"
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
	ruleReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: ruleCond},
		Page:   req.Page,
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

	details, err := svc.buildUrlRuleDetail(kt, vendor, info, resp.Details)
	if err != nil {
		logs.Errorf("build url rule detail failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return &cslb.ListUrlRulesByTopologyResp{Count: int(resp.Count), Details: details}, nil
}

func (svc *lbSvc) getUrlRuleTargetCount(kt *kit.Kit, vendor enumor.Vendor, urlRules []corelb.TCloudLbUrlRule) (
	map[string]int, error) {

	if len(urlRules) == 0 {
		return nil, fmt.Errorf("url rules is empty")
	}

	// 查询规则关联的目标组关系
	ruleIDs := make([]string, 0)
	for _, rule := range urlRules {
		ruleIDs = append(ruleIDs, rule.ID)
	}

	if len(ruleIDs) == 0 {
		return make(map[string]int), nil
	}

	tgLbRels, err := svc.getTgLbRelByCond(kt, []RuleFactory{
		tools.RuleIn("listener_rule_id", ruleIDs),
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("binding_status", enumor.SuccessBindingStatus),
	})
	if err != nil {
		logs.Errorf("get tg lb rel by cond failed, err: %v, ruleIDs: %+v, rid: %s", err, ruleIDs, kt.Rid)
		return nil, err
	}

	tgIDRuleIDMap := make(map[string]string)
	tgIDs := make([]string, 0)
	for _, tgLbRel := range tgLbRels {
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
		tgIDs = append(tgIDs, tgLbRel.TargetGroupID)
	}

	if len(tgIDs) == 0 {
		return make(map[string]int), nil
	}

	targets, err := svc.getTargetByCond(kt, []RuleFactory{tools.RuleIn("target_group_id", tgIDs)})
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, tgIDs: %+v, rid: %s", err, tgIDs, kt.Rid)
		return nil, err
	}

	ruleIDTargetCountMap := make(map[string]int)
	for _, target := range targets {
		ruleID, ok := tgIDRuleIDMap[target.TargetGroupID]
		if !ok {
			continue
		}
		ruleIDTargetCountMap[ruleID]++
	}

	return ruleIDTargetCountMap, nil
}

func (svc *lbSvc) buildUrlRuleDetail(kt *kit.Kit, vendor enumor.Vendor, info *cslb.UrlRuleTopoInfo,
	urlRules []corelb.TCloudLbUrlRule) ([]cslb.UrlRuleDetail, error) {

	ruleIDTargetCountMap, err := svc.getUrlRuleTargetCount(kt, vendor, urlRules)
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
			Ip:          ip,
			LblProtocol: string(lbl.Protocol),
			LblPort:     int(lbl.Port),
			RuleUrl:     rule.URL,
			RuleDomain:  rule.Domain,
			TargetCount: ruleIDTargetCountMap[rule.ID],
			ListenerID:  lbl.CloudID,
			LbID:        lb.ID,
		}
		details = append(details, detail)
	}

	return details, nil
}

func (svc *lbSvc) getUrlRuleTopoInfoByReq(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.ListUrlRulesByTopologyReq) (
	*cslb.UrlRuleTopoInfo, error) {

	filter, err := svc.buildUrlRuleQueryFilter(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("build url rule query filter failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 直接查询URL规则
	ruleReq := core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
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

	if len(resp.Details) == 0 {
		logs.Infof("no URL rules found, rid: %s", kt.Rid)
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	logs.Infof("found %d URL rules, rid: %s", len(resp.Details), kt.Rid)

	// 获取规则对应的负载均衡器和监听器信息
	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	ruleIDs := make([]string, 0)

	for _, rule := range resp.Details {
		lbIDs = append(lbIDs, rule.LbID)
		lblIDs = append(lblIDs, rule.LblID)
		ruleIDs = append(ruleIDs, rule.ID)
	}

	// 获取负载均衡器信息
	lbCond := []RuleFactory{tools.RuleIn("id", lbIDs)}
	lbMap, err := svc.getLbByCond(kt, lbCond)
	if err != nil {
		logs.Errorf("get lb by cond failed, err: %v, lbIDs: %v, rid: %s", err, lbIDs, kt.Rid)
		return nil, err
	}

	// 获取监听器信息
	lblCond := []RuleFactory{tools.RuleIn("id", lblIDs)}
	lblMap, err := svc.getLblByCond(kt, vendor, lblCond)
	if err != nil {
		logs.Errorf("get lbl by cond failed, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
		return nil, err
	}

	// 构建规则条件
	ruleCond := []RuleFactory{tools.RuleIn("id", ruleIDs)}

	return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, LblMap: lblMap, RuleCond: ruleCond}, nil
}

// buildUrlRuleQueryFilter 构建URL规则查询条件
func (svc *lbSvc) buildUrlRuleQueryFilter(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {

	conditions := []*filter.AtomRule{
		tools.RuleEqual("rule_type", enumor.Layer7RuleType),
	}
	conditions = svc.addRuleConditions(req, conditions)
	var lbIDs []string
	if req.HasLbConditions() {
		var err error
		lbIDs, err = svc.queryLoadBalancerIDsByConditions(kt, bizID, vendor, req)
		if err != nil {
			logs.Errorf("query load balancer ids by conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, fmt.Errorf("query load balancer ids failed, err: %v", err)
		}
		if len(lbIDs) > 0 {
			conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))
		}
	} else {
		lbIDs, err := svc.queryAllLoadBalancerIDsByBiz(kt, bizID, vendor)
		if err != nil {
			logs.Errorf("query all load balancer ids by biz failed, bizID: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, fmt.Errorf("query all load balancer ids failed, err: %v", err)
		}
		if len(lbIDs) > 0 {
			conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))
		}
	}

	if req.HasListenerConditions() {
		if len(lbIDs) > 0 {
			listenerIDs, err := svc.queryListenerIDsByConditions(kt, bizID, vendor, req, lbIDs)
			if err != nil {
				logs.Errorf("query listener ids by conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
					bizID, vendor, err, kt.Rid)
				return nil, fmt.Errorf("query listener ids failed, err: %v", err)
			}
			if len(listenerIDs) > 0 {
				conditions = append(conditions, tools.RuleIn("lbl_id", listenerIDs))
			}
		} else {
			logs.Infof("no load balancers found, but has listener conditions, returning empty filter, rid: %s", kt.Rid)
			return &filter.Expression{
				Op:    filter.And,
				Rules: []RuleFactory{},
			}, nil
		}
	}

	conditions, targetGroupIDs, err := svc.addTargetConditions(kt, bizID, vendor, req, conditions)
	if err != nil {
		logs.Errorf("add target conditions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(targetGroupIDs) > int(core.DefaultMaxPageLimit) {
		baseFilter := tools.ExpressionAnd(conditions...)
		targetFilter := svc.buildBatchFilter("target_group_id", targetGroupIDs)
		return &filter.Expression{
			Op:    filter.And,
			Rules: []RuleFactory{baseFilter, targetFilter},
		}, nil
	}

	return tools.ExpressionAnd(conditions...), nil
}

// addRuleConditions 添加规则相关条件
func (svc *lbSvc) addRuleConditions(req *cslb.ListUrlRulesByTopologyReq,
	conditions []*filter.AtomRule) []*filter.AtomRule {
	if ruleConditions := req.BuildRuleFilter(); ruleConditions != nil {
		conditions = append(conditions, ruleConditions...)
	}
	return conditions
}

// addTargetConditions 添加目标相关条件，需要查询目标表
func (svc *lbSvc) addTargetConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq, conditions []*filter.AtomRule) ([]*filter.AtomRule, []string, error) {
	if !req.HasTargetConditions() {
		return conditions, nil, nil
	}

	targetGroupIDs, err := svc.queryTargetGroupIDsByTargetConditions(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("query target group ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, fmt.Errorf("query target group ids failed, err: %v", err)
	}

	if len(targetGroupIDs) > 0 {
		if len(targetGroupIDs) > int(core.DefaultMaxPageLimit) {
			return conditions, targetGroupIDs, nil
		} else {
			conditions = append(conditions, tools.RuleIn("target_group_id", targetGroupIDs))
		}
	}

	return conditions, nil, nil
}

// queryLoadBalancerIDsByConditions 根据负载均衡器条件查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {
	lbFilter := req.BuildLoadBalancerFilter(bizID, vendor)
	return svc.queryLoadBalancerIDsByFilter(kt, lbFilter)
}

// queryAllLoadBalancerIDsByBiz 查询业务下所有负载均衡器ID
func (svc *lbSvc) queryAllLoadBalancerIDsByBiz(kt *kit.Kit, bizID int64, vendor enumor.Vendor) ([]string, error) {
	lbConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("bk_biz_id", bizID),
	}
	lbFilter := tools.ExpressionAnd(lbConditions...)
	return svc.queryLoadBalancerIDsByFilter(kt, lbFilter)
}

// queryLoadBalancerIDsByFilter 根据过滤器查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByFilter(kt *kit.Kit, filter *filter.Expression) ([]string, error) {
	lbReq := &core.ListReq{Filter: filter, Page: core.NewDefaultBasePage()}
	lbIDs := make([]string, 0)

	for {
		lbResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
		if err != nil {
			logs.Errorf("list load balancers failed, filter: %+v, err: %v, rid: %s",
				lbReq.Filter, err, kt.Rid)
			return nil, err
		}

		for _, lb := range lbResp.Details {
			lbIDs = append(lbIDs, lb.ID)
		}

		if uint(len(lbResp.Details)) < lbReq.Page.Limit {
			break
		}
		lbReq.Page.Start += uint32(lbReq.Page.Limit)
	}

	return lbIDs, nil
}

// queryListenerIDsByConditions 根据条件查询监听器本地ID
func (svc *lbSvc) queryListenerIDsByConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq, lbIDs []string) ([]string, error) {

	listenerFilter := req.BuildListenerFilter(bizID, vendor, lbIDs)
	if listenerFilter == nil {
		return []string{}, nil
	}

	listenerReq := &core.ListReq{
		Filter: listenerFilter,
		Page:   core.NewDefaultBasePage(),
	}

	listenerIDs := make([]string, 0)
	for {
		listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listenerReq)
		if err != nil {
			logs.Errorf("list listeners failed, bizID: %d, vendor: %s, lbIDs: %v, err: %v, rid: %s",
				bizID, vendor, lbIDs, err, kt.Rid)
			return nil, err
		}
		for _, listener := range listenerResp.Details {
			listenerIDs = append(listenerIDs, listener.ID)
		}
		if uint(len(listenerResp.Details)) < listenerReq.Page.Limit {
			break
		}

		listenerReq.Page.Start += uint32(listenerReq.Page.Limit)
	}

	return listenerIDs, nil
}

// queryTargetGroupIDsByTargetConditions 根据目标条件查询目标组ID
func (svc *lbSvc) queryTargetGroupIDsByTargetConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	targetGroupFilter := req.BuildTargetGroupFilter(bizID, vendor)
	if targetGroupFilter == nil {
		return []string{}, nil
	}

	targetGroupReq := &core.ListReq{
		Filter: targetGroupFilter,
		Page:   core.NewDefaultBasePage(),
	}

	targetGroupIDs := make([]string, 0)
	for {
		targetGroupResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, targetGroupReq)
		if err != nil {
			logs.Errorf("list target groups failed, filter: %+v, err: %v, rid: %s",
				targetGroupReq.Filter, err, kt.Rid)
			return nil, err
		}
		for _, targetGroup := range targetGroupResp.Details {
			targetGroupIDs = append(targetGroupIDs, targetGroup.ID)
		}
		if uint(len(targetGroupResp.Details)) < targetGroupReq.Page.Limit {
			break
		}
		targetGroupReq.Page.Start += uint32(targetGroupReq.Page.Limit)
	}

	if len(targetGroupIDs) == 0 {
		return []string{}, nil
	}

	targetFilter := req.BuildTargetFilter(targetGroupIDs)
	if targetFilter == nil {
		return targetGroupIDs, nil
	}

	targetReq := &core.ListReq{
		Filter: targetFilter,
		Page:   core.NewDefaultBasePage(),
	}

	finalTargetGroupIDMap := make(map[string]struct{})
	for {
		targetResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
		if err != nil {
			logs.Errorf("list targets failed, filter: %+v, err: %v, rid: %s",
				targetReq.Filter, err, kt.Rid)
			return nil, err
		}
		for _, target := range targetResp.Details {
			if target.TargetGroupID != "" {
				finalTargetGroupIDMap[target.TargetGroupID] = struct{}{}
			}
		}
		if uint(len(targetResp.Details)) < targetReq.Page.Limit {
			break
		}
		targetReq.Page.Start += uint32(targetReq.Page.Limit)
	}

	return maps.Keys(finalTargetGroupIDMap), nil
}

// buildBatchFilter 构建批量过滤条件
func (svc *lbSvc) buildBatchFilter(field string, ids []string) *filter.Expression {
	if len(ids) > int(core.DefaultMaxPageLimit) {
		batches := slice.Split(ids, int(core.DefaultMaxPageLimit))
		batchConditions := make([]*filter.AtomRule, 0, len(batches))
		for _, batch := range batches {
			batchConditions = append(batchConditions, tools.RuleIn(field, batch))
		}
		return tools.ExpressionOr(batchConditions...)
	}
	return tools.ExpressionAnd(tools.RuleIn(field, ids))
}

func (svc *lbSvc) getRuleCondByTargetCond(kt *kit.Kit, tgLbRelCond []RuleFactory,
	reqTargetCond *filter.Expression) ([]RuleFactory, error) {

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
	if reqTargetCond != nil {
		targetCond = append(targetCond, reqTargetCond.Rules...)
	}
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
