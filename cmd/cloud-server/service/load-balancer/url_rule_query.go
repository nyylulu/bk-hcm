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
	"fmt"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// ListUrlRulesByTopo 查询URL规则信息
func (svc *lbSvc) ListUrlRulesByTopo(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cslb.ListUrlRulesByTopologyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		logs.Errorf("list url rules by topo failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list url rules by topo failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list url rules by topo no auth, req: %+v, rid: %s", req, cts.Kit.Rid)
		return nil, errf.New(errf.PermissionDenied, "no permission for list URL rules by topology")
	}

	filters, err := svc.buildUrlRuleQueryFilter(cts.Kit, bizID, vendor, req)
	if err != nil {
		logs.Errorf("build url rule query filter failed, bizID: %d, vendor: %s, err: %v, rid: %s",
			bizID, vendor, err, cts.Kit.Rid)
		return nil, err
	}

	urlRuleList, err := svc.queryUrlRulesByFilter(cts.Kit, vendor, filters, req.Page)
	if err != nil {
		logs.Errorf("query url rules by filter failed, bizID: %d, vendor: %s, err: %v, rid: %s",
			bizID, vendor, err, cts.Kit.Rid)
		return nil, err
	}

	result, err := svc.buildUrlRuleResponse(urlRuleList, cts.Kit)
	if err != nil {
		logs.Errorf("build url rule response failed, bizID: %d, vendor: %s, err: %v, rid: %s",
			bizID, vendor, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// buildUrlRuleQueryFilter 构建URL规则查询条件
func (svc *lbSvc) buildUrlRuleQueryFilter(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {
	if len(req.LblProtocol) > 0 {
		hasLayer7 := false
		for _, protocol := range req.LblProtocol {
			if protocol == "HTTP" || protocol == "HTTPS" {
				hasLayer7 = true
				break
			}
		}
		if !hasLayer7 {
			return &filter.Expression{
				Op:    filter.And,
				Rules: []filter.RuleFactory{},
			}, nil
		}
	}

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
		listenerIDs, err := svc.queryListenerIDsByConditions(kt, bizID, vendor, req, lbIDs)
		if err != nil {
			logs.Errorf("query listener ids by conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, fmt.Errorf("query listener ids failed, err: %v", err)
		}
		if len(listenerIDs) > 0 {
			conditions = append(conditions, tools.RuleIn("lbl_id", listenerIDs))
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
			Rules: []filter.RuleFactory{baseFilter, targetFilter},
		}, nil
	}

	return tools.ExpressionAnd(conditions...), nil
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

// queryRuleIDsByTargetGroupIDs 通过关联表查询规则ID
func (svc *lbSvc) queryRuleIDsByTargetGroupIDs(kt *kit.Kit, targetGroupIDs []string) ([]string, error) {
	if len(targetGroupIDs) == 0 {
		return []string{}, nil
	}

	relFilter := tools.ExpressionAnd(
		tools.RuleIn("target_group_id", targetGroupIDs),
		tools.RuleEqual("binding_status", "success"),
	)

	relReq := &core.ListReq{
		Filter: relFilter,
		Page:   core.NewDefaultBasePage(),
	}

	ruleIDMap := make(map[string]struct{})
	for {
		relResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, relReq)
		if err != nil {
			logs.Errorf("list target group listener rule relations failed, filter: %+v, err: %v, rid: %s",
				relReq.Filter, err, kt.Rid)
			return nil, err
		}

		for _, rel := range relResp.Details {
			if rel.ListenerRuleID != "" {
				ruleIDMap[rel.ListenerRuleID] = struct{}{}
			}
		}

		if uint(len(relResp.Details)) < relReq.Page.Limit {
			break
		}
		relReq.Page.Start += uint32(relReq.Page.Limit)
	}

	return maps.Keys(ruleIDMap), nil
}

// queryUrlRulesByFilter 根据条件查询URL规则
func (svc *lbSvc) queryUrlRulesByFilter(kt *kit.Kit, vendor enumor.Vendor,
	filter *filter.Expression, page *core.BasePage) (*dataproto.TCloudURLRuleListResult, error) {

	req := &core.ListReq{
		Filter: filter,
		Page:   page,
	}

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, req)
	default:
		logs.Errorf("unsupported vendor: %s, rid: %s", vendor, kt.Rid)
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
}

// buildUrlRuleResponse 构建URL规则响应
func (svc *lbSvc) buildUrlRuleResponse(urlRuleList *dataproto.TCloudURLRuleListResult,
	kt *kit.Kit) (*cslb.ListUrlRulesByTopologyResp, error) {

	result := &cslb.ListUrlRulesByTopologyResp{
		Count:   int(urlRuleList.Count),
		Details: make([]cslb.UrlRuleDetail, 0, len(urlRuleList.Details)),
	}

	if len(urlRuleList.Details) == 0 {
		return result, nil
	}

	lbIDMap := make(map[string]struct{})
	listenerIDMap := make(map[string]struct{})
	targetGroupIDMap := make(map[string]struct{})

	for _, rule := range urlRuleList.Details {
		lbIDMap[rule.LbID] = struct{}{}
		listenerIDMap[rule.LblID] = struct{}{}
		if rule.TargetGroupID != "" {
			targetGroupIDMap[rule.TargetGroupID] = struct{}{}
		}
	}

	lbIDs := maps.Keys(lbIDMap)
	listenerIDs := maps.Keys(listenerIDMap)
	targetGroupIDs := maps.Keys(targetGroupIDMap)

	lbMap, _ := svc.batchGetLoadBalancerInfo(kt, lbIDs)
	listenerMap, _ := svc.batchGetListenerInfo(kt, listenerIDs)
	targetCountMap, _ := svc.batchGetTargetCountByTargetGroupIDs(kt, targetGroupIDs)

	for _, rule := range urlRuleList.Details {
		detail, err := svc.buildUrlRuleDetail(kt, rule, lbMap, listenerMap, targetCountMap)
		if err != nil {
			logs.Errorf("build url rule detail failed, ruleID: %s, err: %v, rid: %s", rule.ID, err, kt.Rid)
			return nil, err
		}
		result.Details = append(result.Details, detail)
	}

	return result, nil
}

// buildUrlRuleDetail URL规则详情
func (svc *lbSvc) buildUrlRuleDetail(kt *kit.Kit, rule corelb.TCloudLbUrlRule,
	lbMap map[string]*corelb.BaseLoadBalancer, listenerMap map[string]*corelb.BaseListener,
	targetCountMap map[string]int) (cslb.UrlRuleDetail, error) {

	detail := cslb.UrlRuleDetail{
		ID: rule.ID,
	}

	lb, exists := lbMap[rule.LbID]
	if !exists {
		logs.Errorf("load balancer not found, lbID: %s, rid: %s", rule.LbID, kt.Rid)
		return detail, fmt.Errorf("load balancer not found, lbID: %s", rule.LbID)
	}
	detail.Ip = svc.getLoadBalancerVip(lb)
	detail.LbID = lb.ID

	listener, exists := listenerMap[rule.LblID]
	if !exists {
		logs.Errorf("listener not found, lblID: %s, rid: %s", rule.LblID, kt.Rid)
		return detail, fmt.Errorf("listener not found, lblID: %s", rule.LblID)
	}
	detail.LblProtocol = string(listener.Protocol)
	detail.LblPort = int(listener.Port)
	detail.ListenerID = listener.CloudID

	if targetCount, exists := targetCountMap[rule.TargetGroupID]; exists {
		detail.TargetCount = targetCount
	}

	detail.RuleUrl = rule.URL
	detail.RuleDomain = rule.Domain

	return detail, nil
}

// batchGetLoadBalancerInfo 获取负载均衡器信息
func (svc *lbSvc) batchGetLoadBalancerInfo(kt *kit.Kit, lbIDs []string) (map[string]*corelb.BaseLoadBalancer, error) {
	if len(lbIDs) == 0 {
		return make(map[string]*corelb.BaseLoadBalancer), nil
	}

	result := make(map[string]*corelb.BaseLoadBalancer)
	batches := slice.Split(lbIDs, int(core.DefaultMaxPageLimit))

	for _, batch := range batches {
		lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), batch)
		if err != nil {
			return nil, err
		}

		for id, lb := range lbMap {
			result[id] = &lb
		}
	}

	return result, nil
}

// batchGetListenerInfo 获取监听器信息
func (svc *lbSvc) batchGetListenerInfo(kt *kit.Kit, listenerIDs []string) (map[string]*corelb.BaseListener, error) {
	if len(listenerIDs) == 0 {
		return make(map[string]*corelb.BaseListener), nil
	}

	result := make(map[string]*corelb.BaseListener)
	batches := slice.Split(listenerIDs, int(core.DefaultMaxPageLimit))

	for _, batch := range batches {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		resp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, req)
		if err != nil {
			logs.Errorf("batch get listener info failed, listenerIDs: %v, err: %v, rid: %s",
				batch, err, kt.Rid)
			return nil, err
		}

		for _, listener := range resp.Details {
			result[listener.ID] = &listener
		}
	}

	return result, nil
}

// batchGetTargetGroupInfo 获取目标组信息
func (svc *lbSvc) batchGetTargetGroupInfo(kt *kit.Kit,
	targetGroupIDs []string) (map[string]*corelb.BaseTargetGroup, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]*corelb.BaseTargetGroup), nil
	}

	result := make(map[string]*corelb.BaseTargetGroup)
	batches := slice.Split(targetGroupIDs, int(core.DefaultMaxPageLimit))

	for _, batch := range batches {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, req)
		if err != nil {
			logs.Errorf("batch get target group info failed, targetGroupIDs: %v, err: %v, rid: %s",
				batch, err, kt.Rid)
			return nil, err
		}

		for _, targetGroup := range resp.Details {
			result[targetGroup.ID] = &targetGroup
		}
	}

	return result, nil
}

// getLoadBalancerVip 获取负载均衡器VIP
func (svc *lbSvc) getLoadBalancerVip(lb *corelb.BaseLoadBalancer) string {
	if len(lb.PublicIPv4Addresses) > 0 {
		return lb.PublicIPv4Addresses[0]
	}

	if len(lb.PublicIPv6Addresses) > 0 {
		return lb.PublicIPv6Addresses[0]
	}

	if len(lb.PrivateIPv4Addresses) > 0 {
		return lb.PrivateIPv4Addresses[0]
	}

	if len(lb.PrivateIPv6Addresses) > 0 {
		return lb.PrivateIPv6Addresses[0]
	}

	if lb.Domain != "" {
		return lb.Domain
	}

	return ""
}

// batchGetTargetCountByTargetGroupIDs 获取目标组中的RS数量
func (svc *lbSvc) batchGetTargetCountByTargetGroupIDs(kt *kit.Kit, targetGroupIDs []string) (map[string]int, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]int), nil
	}

	uniqueTargetGroupIDs := slice.Unique(targetGroupIDs)
	targets, err := svc.getTargetByTGIDs(kt, uniqueTargetGroupIDs)
	if err != nil {
		logs.Errorf("batch query targets by target group ids failed, targetGroupIDs: %v, err: %v, rid: %s",
			uniqueTargetGroupIDs, err, kt.Rid)
		return nil, fmt.Errorf("batch query targets failed, err: %v", err)
	}

	targetCountMap := make(map[string]int)
	for _, target := range targets {
		if target.TargetGroupID != "" {
			targetCountMap[target.TargetGroupID]++
		}
	}

	return targetCountMap, nil
}
