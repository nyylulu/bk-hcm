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
	_, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list url rules by topo failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list url rules by topo no auth, req: %+v, rid: %s", noPermFlag, req, cts.Kit.Rid)
		return &cslb.ListUrlRulesByTopologyResp{Count: 0, Details: make([]cslb.UrlRuleDetail, 0)}, nil
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	filters, err := svc.buildUrlRuleQueryFilter(cts.Kit, bizID, vendor, req)
	if err != nil {
		return nil, err
	}

	urlRuleList, err := svc.queryUrlRulesByFilter(cts.Kit, vendor, filters, req.Page)
	if err != nil {
		return nil, err
	}

	result, err := svc.buildUrlRuleResponse(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// buildUrlRuleQueryFilter 构建URL规则查询条件
func (svc *lbSvc) buildUrlRuleQueryFilter(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {

	conditions := req.BuildUrlRuleBaseConditions(bizID)
	lbIDs, err := svc.queryLoadBalancerIDsByConditions(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("query load balancer ids by conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
			bizID, vendor, err, kt.Rid)
		return nil, fmt.Errorf("query load balancer ids failed, err: %v", err)
	}

	if len(lbIDs) == 0 {
		return &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		}, nil
	}
	conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))

	if !req.HasListenerConditions() && !req.HasRuleConditions() && !req.HasTargetConditions() {
		return tools.ExpressionAnd(conditions...), nil
	}

	if req.HasListenerConditions() {
		listenerCloudIDs, err := svc.queryListenerCloudIDsByConditions(kt, bizID, vendor, req, lbIDs)
		if err != nil {
			logs.Errorf("query listener cloud ids by conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, fmt.Errorf("query listener cloud ids failed, err: %v", err)
		}
		if len(listenerCloudIDs) == 0 {
			return &filter.Expression{
				Op:    filter.And,
				Rules: []filter.RuleFactory{tools.RuleEqual("id", "never_match")},
			}, nil
		}
		conditions = append(conditions, tools.RuleIn("cloud_lbl_id", listenerCloudIDs))
	}

	if ruleConditions := req.BuildRuleFilter(); ruleConditions != nil {
		conditions = append(conditions, ruleConditions...)
	}
	if req.HasTargetConditions() {
		targetGroupIDs, err := svc.queryTargetGroupIDsByTargetConditions(kt, bizID, vendor, req)
		if err != nil {
			logs.Errorf("query target group ids by target conditions failed, bizID: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, fmt.Errorf("query target group ids failed, err: %v", err)
		}
		if len(targetGroupIDs) == 0 {
			return &filter.Expression{
				Op:    filter.And,
				Rules: []filter.RuleFactory{tools.RuleEqual("id", "never_match")},
			}, nil
		}
		conditions = append(conditions, tools.RuleIn("target_group_id", targetGroupIDs))
	}

	return tools.ExpressionAnd(conditions...), nil
}

// queryLoadBalancerIDsByConditions 根据负载均衡器条件查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	lbFilter := req.BuildLoadBalancerFilter(bizID, string(vendor))
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

// queryListenerCloudIDsByConditions 根据条件查询监听器云上ID，使用CLB ID作为条件减少数据量
func (svc *lbSvc) queryListenerCloudIDsByConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq, lbIDs []string) ([]string, error) {

	if len(lbIDs) == 0 {
		return []string{}, nil
	}

	listenerFilter := req.BuildListenerFilter(bizID, string(vendor), lbIDs)
	if listenerFilter == nil {
		return []string{}, nil
	}

	listenerReq := &core.ListReq{
		Filter: listenerFilter,
		Page:   core.NewDefaultBasePage(),
	}

	listenerCloudIDs := make([]string, 0)
	for {
		listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listenerReq)
		if err != nil {
			logs.Errorf("list listeners failed, bizID: %d, vendor: %s, lbIDs: %v, err: %v, rid: %s",
				bizID, vendor, lbIDs, err, kt.Rid)
			return nil, err
		}

		for _, listener := range listenerResp.Details {
			listenerCloudIDs = append(listenerCloudIDs, listener.CloudID)
		}

		if uint(len(listenerResp.Details)) < listenerReq.Page.Limit {
			break
		}

		listenerReq.Page.Start += uint32(listenerReq.Page.Limit)
	}

	return listenerCloudIDs, nil
}

// queryTargetGroupIDsByTargetConditions 根据目标条件查询目标组ID
func (svc *lbSvc) queryTargetGroupIDsByTargetConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	targetGroupFilter := req.BuildTargetGroupFilter(bizID, string(vendor))
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

	finalTargetGroupIDs := maps.Keys(finalTargetGroupIDMap)

	return finalTargetGroupIDs, nil
}

// queryUrlRulesByFilter 根据条件查询URL规则
func (svc *lbSvc) queryUrlRulesByFilter(kt *kit.Kit, vendor enumor.Vendor,
	filter *filter.Expression, page *core.BasePage) (*dataproto.TCloudURLRuleListResult, error) {

	enhancedPage := &core.BasePage{
		Start: page.Start,
		Limit: page.Limit,
	}

	if enhancedPage.Limit <= core.DefaultMaxPageLimit {
		enhancedPage.Limit = core.DefaultMaxPageLimit
	}

	result := &dataproto.TCloudURLRuleListResult{
		Count:   0,
		Details: make([]corelb.TCloudLbUrlRule, 0),
	}

	req := &core.ListReq{
		Filter: filter,
		Page:   enhancedPage,
	}

	for {
		var resp *dataproto.TCloudURLRuleListResult
		var err error

		switch vendor {
		case enumor.TCloud:
			resp, err = svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, req)
			if err != nil {
				logs.Errorf("list url rules failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
				return nil, err
			}
		default:
			logs.Errorf("unsupported vendor: %s, rid: %s", vendor, kt.Rid)
			return nil, fmt.Errorf("vendor: %s not support", vendor)
		}

		result.Count += resp.Count
		result.Details = append(result.Details, resp.Details...)

		if uint(len(resp.Details)) < req.Page.Limit {
			break
		}

		req.Page.Start += uint32(req.Page.Limit)
	}

	return result, nil
}

// buildUrlRuleResponse 构建URL规则响应
func (svc *lbSvc) buildUrlRuleResponse(kt *kit.Kit,
	urlRuleList *dataproto.TCloudURLRuleListResult) (*cslb.ListUrlRulesByTopologyResp, error) {

	result := &cslb.ListUrlRulesByTopologyResp{
		Count:   int(urlRuleList.Count),
		Details: make([]cslb.UrlRuleDetail, 0, len(urlRuleList.Details)),
	}

	if len(urlRuleList.Details) == 0 {
		return result, nil
	}

	lbIDs := make([]string, 0, len(urlRuleList.Details))
	listenerIDs := make([]string, 0, len(urlRuleList.Details))
	targetGroupIDs := make([]string, 0, len(urlRuleList.Details))

	for _, rule := range urlRuleList.Details {
		lbIDs = append(lbIDs, rule.LbID)
		listenerIDs = append(listenerIDs, rule.LblID)
		if rule.TargetGroupID != "" {
			targetGroupIDs = append(targetGroupIDs, rule.TargetGroupID)
		}
	}

	lbMap, err := svc.batchGetLoadBalancerInfo(kt, lbIDs)
	if err != nil {
		logs.Errorf("batch get load balancer info failed, err: %v, rid: %s", err, kt.Rid)
	}

	listenerMap, err := svc.batchGetListenerInfo(kt, listenerIDs)
	if err != nil {
		logs.Errorf("batch get listener info failed, err: %v, rid: %s", err, kt.Rid)
	}

	targetGroupMap, err := svc.batchGetTargetGroupInfo(kt, targetGroupIDs)
	if err != nil {
		logs.Errorf("batch get target group info failed, err: %v, rid: %s", err, kt.Rid)
	}

	targetCountMap, err := svc.batchGetTargetCountByTargetGroupIDs(kt, targetGroupIDs)
	if err != nil {
		logs.Errorf("batch get target count failed, err: %v, rid: %s", err, kt.Rid)
	}

	for _, rule := range urlRuleList.Details {
		detail := svc.buildUrlRuleDetail(rule, lbMap, listenerMap, targetGroupMap, targetCountMap)
		result.Details = append(result.Details, detail)
	}

	return result, nil
}

// buildUrlRuleDetail URL规则详情
func (svc *lbSvc) buildUrlRuleDetail(rule corelb.TCloudLbUrlRule,
	lbMap map[string]*corelb.BaseLoadBalancer,
	listenerMap map[string]*corelb.BaseListener,
	targetGroupMap map[string]*corelb.BaseTargetGroup,
	targetCountMap map[string]int) cslb.UrlRuleDetail {

	detail := cslb.UrlRuleDetail{
		ID: rule.ID,
	}

	if lb, exists := lbMap[rule.LbID]; exists {
		detail.Ip = svc.getLoadBalancerVip(lb)
		detail.LbID = lb.ID
	}

	if listener, exists := listenerMap[rule.LblID]; exists {
		detail.LblProtocols = string(listener.Protocol)
		detail.LblPort = int(listener.Port)
		detail.ListenerID = listener.CloudID
	}

	if _, exists := targetGroupMap[rule.TargetGroupID]; exists {

		if targetCount, exists := targetCountMap[rule.TargetGroupID]; exists {
			detail.TargetCount = targetCount
		}
	}

	if rule.RuleType == enumor.Layer7RuleType {
		detail.RuleUrl = rule.URL
		detail.RuleDomain = rule.Domain
	}

	return detail
}

// batchGetLoadBalancerInfo 获取负载均衡器信息
func (svc *lbSvc) batchGetLoadBalancerInfo(kt *kit.Kit, lbIDs []string) (map[string]*corelb.BaseLoadBalancer, error) {
	if len(lbIDs) == 0 {
		return make(map[string]*corelb.BaseLoadBalancer), nil
	}

	lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), lbIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*corelb.BaseLoadBalancer)
	for id, lb := range lbMap {
		result[id] = &lb
	}

	return result, nil
}

// batchGetListenerInfo 获取监听器信息
func (svc *lbSvc) batchGetListenerInfo(kt *kit.Kit, listenerIDs []string) (map[string]*corelb.BaseListener, error) {
	if len(listenerIDs) == 0 {
		return make(map[string]*corelb.BaseListener), nil
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", listenerIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	listenerMap := make(map[string]*corelb.BaseListener)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, req)
		if err != nil {
			logs.Errorf("batch get listener info failed, listenerIDs: %v, err: %v, rid: %s",
				listenerIDs, err, kt.Rid)
			return nil, err
		}

		for _, listener := range resp.Details {
			listenerMap[listener.ID] = &listener
		}

		if uint(len(resp.Details)) < req.Page.Limit {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return listenerMap, nil
}

// batchGetTargetGroupInfo 获取目标组信息
func (svc *lbSvc) batchGetTargetGroupInfo(kt *kit.Kit,
	targetGroupIDs []string) (map[string]*corelb.BaseTargetGroup, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]*corelb.BaseTargetGroup), nil
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", targetGroupIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	targetGroupMap := make(map[string]*corelb.BaseTargetGroup)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, req)
		if err != nil {
			logs.Errorf("batch get target group info failed, targetGroupIDs: %v, err: %v, rid: %s",
				targetGroupIDs, err, kt.Rid)
			return nil, err
		}

		for _, targetGroup := range resp.Details {
			targetGroupMap[targetGroup.ID] = &targetGroup
		}

		if uint(len(resp.Details)) < req.Page.Limit {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return targetGroupMap, nil
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
