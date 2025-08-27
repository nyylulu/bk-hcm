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
)

// ListUrlRulesByTopology 查询URL规则信息
func (svc *lbSvc) ListUrlRulesByTopology(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is required")
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(cslb.ListUrlRulesByTopologyReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 业务权限校验
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.LoadBalancer,
			Action: meta.Find,
		},
		BizID: bizID,
	})
	if err != nil {
		return nil, err
	}

	filters, err := svc.buildUrlRuleQueryFilter(cts.Kit, bizID, req)
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
func (svc *lbSvc) buildUrlRuleQueryFilter(kt *kit.Kit, bizID int64, req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {

	listenerIDs, err := svc.queryListenerIDsByConditions(kt, bizID, req)
	if err != nil {
		return nil, fmt.Errorf("query listener ids failed, err: %v", err)
	}

	if len(listenerIDs) == 0 {
		return tools.ExpressionAnd(tools.RuleEqual("id", "nonexistent")), nil
	}

	conditions := []*filter.AtomRule{
		tools.RuleIn("rule_type", []string{string(enumor.Layer4RuleType), string(enumor.Layer7RuleType)}),
		tools.RuleIn("lbl_id", listenerIDs),
	}

	if err = svc.addLoadBalancerConditions(kt, req, conditions); err != nil {
		return nil, err
	}

	svc.addRuleConditions(req, conditions)

	if err = svc.addTargetConditions(kt, req, conditions); err != nil {
		return nil, err
	}

	return tools.ExpressionAnd(conditions...), nil
}

// addLoadBalancerConditions 添加负载均衡器相关条件
func (svc *lbSvc) addLoadBalancerConditions(kt *kit.Kit, req *cslb.ListUrlRulesByTopologyReq, conditions []*filter.AtomRule) error {
	if len(req.LbRegions) > 0 {
		conditions = append(conditions, tools.RuleIn("region", req.LbRegions))
	}

	if len(req.LbNetworkTypes) > 0 {
		conditions = append(conditions, tools.RuleIn("lb_type", req.LbNetworkTypes))
	}

	if len(req.LbIpVersions) > 0 {
		conditions = append(conditions, tools.RuleIn("ip_version", req.LbIpVersions))
	}

	if len(req.CloudLbIds) > 0 {
		conditions = append(conditions, tools.RuleIn("cloud_lb_id", req.CloudLbIds))
	}

	if len(req.LbVips) > 0 || len(req.LbDomains) > 0 {
		lbIDs, err := svc.queryLoadBalancerIDsByConditions(kt, req)
		if err != nil {
			return fmt.Errorf("query load balancer ids failed, err: %v", err)
		}
		if len(lbIDs) > 0 {
			conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))
		}
	}

	return nil
}

// addRuleConditions 添加规则相关条件
func (svc *lbSvc) addRuleConditions(req *cslb.ListUrlRulesByTopologyReq, conditions []*filter.AtomRule) {
	if len(req.RuleDomains) > 0 {
		conditions = append(conditions, tools.RuleIn("domain", req.RuleDomains))
	}

	if len(req.RuleUrls) > 0 {
		conditions = append(conditions, tools.RuleIn("url", req.RuleUrls))
	}
}

// addTargetConditions 添加目标相关条件，需要查询目标表
func (svc *lbSvc) addTargetConditions(kt *kit.Kit, req *cslb.ListUrlRulesByTopologyReq, conditions []*filter.AtomRule) error {
	if len(req.TargetIps) == 0 && len(req.TargetPorts) == 0 {
		return nil
	}

	targetGroupIDs, err := svc.queryTargetGroupIDsByTargetConditions(kt, req)
	if err != nil {
		return fmt.Errorf("query target group ids failed, err: %v", err)
	}

	if len(targetGroupIDs) > 0 {
		conditions = append(conditions, tools.RuleIn("target_group_id", targetGroupIDs))
	}

	return nil
}

// queryLoadBalancerIDsByConditions 根据负载均衡器条件查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByConditions(kt *kit.Kit, req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {
	lbConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", req.Vendor),
		tools.RuleEqual("account_id", req.AccountID),
	}

	if len(req.LbDomains) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("domain", req.LbDomains))
	}

	if len(req.LbVips) > 0 {
		vipConditions := []*filter.AtomRule{
			tools.RuleJsonOverlaps("private_ipv4_addresses", req.LbVips),
			tools.RuleJsonOverlaps("private_ipv6_addresses", req.LbVips),
			tools.RuleJsonOverlaps("public_ipv4_addresses", req.LbVips),
			tools.RuleJsonOverlaps("public_ipv6_addresses", req.LbVips),
		}
		vipOrFilter := tools.ExpressionOr(vipConditions...)

		lbFilter, err := tools.And(tools.ExpressionAnd(lbConditions...), vipOrFilter)
		if err != nil {
			return nil, err
		}
		lbReq := &core.ListReq{
			Filter: lbFilter,
			Page:   core.NewDefaultBasePage(),
		}

		lbResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
		if err != nil {
			return nil, err
		}

		lbIDs := extractLoadBalancerIDs(lbResp.Details)

		return lbIDs, nil
	}

	lbReq := &core.ListReq{
		Filter: tools.ExpressionAnd(lbConditions...),
		Page:   core.NewDefaultBasePage(),
	}

	lbResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		return nil, err
	}

	lbIDs := extractLoadBalancerIDs(lbResp.Details)

	return lbIDs, nil
}

// queryListenerIDsByConditions 根据条件查询监听器ID
func (svc *lbSvc) queryListenerIDsByConditions(kt *kit.Kit, bizID int64, req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	listenerConditions := []*filter.AtomRule{
		tools.RuleEqual("account_id", req.AccountID),
	}

	if len(req.LblProtocols) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("protocol", req.LblProtocols))
	}

	if len(req.LblPorts) > 0 {
		portInt64s := convertIntSliceToInt64Slice(req.LblPorts)
		listenerConditions = append(listenerConditions, tools.RuleIn("port", portInt64s))
	}

	listenerFilter := tools.ExpressionAnd(listenerConditions...)

	listenerReq := &core.ListReq{
		Filter: listenerFilter,
		Page:   core.NewDefaultBasePage(),
	}

	listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listenerReq)
	if err != nil {
		return nil, err
	}

	listenerIDs := extractListenerIDs(listenerResp.Details)
	for _, listener := range listenerResp.Details {
		logs.Infof("queryListenerIDsByConditions: listener ID: %s, account_id: %s, bk_biz_id: %d",
			listener.ID, listener.AccountID, listener.BkBizID)
	}

	return listenerIDs, nil
}

// queryTargetGroupIDsByTargetConditions 根据目标条件查询目标组ID
func (svc *lbSvc) queryTargetGroupIDsByTargetConditions(kt *kit.Kit, req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	targetConditions := []*filter.AtomRule{
		tools.RuleEqual("account_id", req.AccountID),
	}

	if len(req.TargetIps) > 0 {
		targetConditions = append(targetConditions, tools.RuleIn("ip", req.TargetIps))
	}

	if len(req.TargetPorts) > 0 {
		portInt64s := convertIntSliceToInt64Slice(req.TargetPorts)
		targetConditions = append(targetConditions, tools.RuleIn("port", portInt64s))
	}

	targetFilter := tools.ExpressionAnd(targetConditions...)
	targetReq := &core.ListReq{
		Filter: targetFilter,
		Page:   core.NewDefaultBasePage(),
	}

	targetResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
	if err != nil {
		return nil, err
	}

	targetGroupIDMap := make(map[string]struct{})
	for _, target := range targetResp.Details {
		if target.TargetGroupID != "" {
			targetGroupIDMap[target.TargetGroupID] = struct{}{}
		}
	}

	targetGroupIDs := make([]string, 0, len(targetGroupIDMap))
	for targetGroupID := range targetGroupIDMap {
		targetGroupIDs = append(targetGroupIDs, targetGroupID)
	}

	return targetGroupIDs, nil
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
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
}
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
	}

	if listener, exists := listenerMap[rule.LblID]; exists {
		detail.LblProtocols = string(listener.Protocol)
		detail.LblPort = int(listener.Port)
		detail.ListenerID = listener.ID
	}

	if _, exists := targetGroupMap[rule.TargetGroupID]; exists {

		if targetCount, exists := targetCountMap[rule.TargetGroupID]; exists {
			detail.TargetCount = targetCount
		} else {
			detail.TargetCount = 0
		}
	}

	if rule.RuleType == enumor.Layer7RuleType {
		detail.RuleUrl = rule.URL
		detail.RuleDomain = rule.Domain
	} else {
		detail.RuleUrl = ""
		detail.RuleDomain = ""
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

	resp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, req)
	if err != nil {
		return nil, err
	}

	listenerMap := make(map[string]*corelb.BaseListener)
	for _, listener := range resp.Details {
		listenerMap[listener.ID] = &listener
	}

	return listenerMap, nil
}

// batchGetTargetGroupInfo 获取目标组信息
func (svc *lbSvc) batchGetTargetGroupInfo(kt *kit.Kit, targetGroupIDs []string) (map[string]*corelb.BaseTargetGroup, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]*corelb.BaseTargetGroup), nil
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", targetGroupIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, req)
	if err != nil {
		return nil, err
	}

	targetGroupMap := make(map[string]*corelb.BaseTargetGroup)
	for _, targetGroup := range resp.Details {
		targetGroupMap[targetGroup.ID] = &targetGroup
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

	uniqueTargetGroupIDs := removeDuplicates(targetGroupIDs)
	targets, err := svc.getTargetByTGIDs(kt, uniqueTargetGroupIDs)
	if err != nil {
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

// convertIntSliceToInt64Slice 将int切片转换为int64切片
func convertIntSliceToInt64Slice(intSlice []int) []int64 {
	if len(intSlice) == 0 {
		return nil
	}
	int64Slice := make([]int64, len(intSlice))
	for i, v := range intSlice {
		int64Slice[i] = int64(v)
	}
	return int64Slice
}

// extractLoadBalancerIDs 从负载均衡器响应中提取ID列表
func extractLoadBalancerIDs(items []corelb.BaseLoadBalancer) []string {
	if len(items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

// extractListenerIDs 从监听器响应中提取ID列表
func extractListenerIDs(items []corelb.BaseListener) []string {
	if len(items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

// removeDuplicates 去除字符串切片中的重复元素
func removeDuplicates(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
