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

package cslb

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
)

// ListUrlRulesByTopologyReq list url rules by topology.
type ListUrlRulesByTopologyReq struct {
	AccountID      string         `json:"account_id" validate:"required"`
	LbRegions      []string       `json:"lb_regions" validate:"omitempty,max=500"`
	LbNetworkTypes []string       `json:"lb_network_types" validate:"omitempty"`
	LbIpVersions   []string       `json:"lb_ip_versions" validate:"omitempty"`
	CloudLbIds     []string       `json:"cloud_lb_ids" validate:"omitempty,max=500"`
	LbVips         []string       `json:"lb_vips" validate:"omitempty,max=500"`
	LbDomains      []string       `json:"lb_domains" validate:"omitempty,max=500"`
	LblProtocol    []string       `json:"lbl_protocol" validate:"omitempty"`
	LblPorts       []int          `json:"lbl_ports" validate:"omitempty,max=1000"`
	RuleDomains    []string       `json:"rule_domains" validate:"omitempty,max=500"`
	RuleUrls       []string       `json:"rule_urls" validate:"omitempty,max=500"`
	TargetIps      []string       `json:"target_ips" validate:"omitempty,max=5000"`
	TargetPorts    []int          `json:"target_ports" validate:"omitempty,max=500"`
	Page           *core.BasePage `json:"page" validate:"required"`
}

// ListUrlRulesByTopologyResp list url rules by topology resp.
type ListUrlRulesByTopologyResp struct {
	Count   int             `json:"count"`
	Details []UrlRuleDetail `json:"details"`
}

// UrlRuleDetail url rule detail.
type UrlRuleDetail struct {
	ID          string `json:"id"`
	Ip          string `json:"ip"`
	LblProtocol string `json:"lbl_protocol"`
	LblPort     int    `json:"lbl_port"`
	RuleUrl     string `json:"rule_url"`
	RuleDomain  string `json:"rule_domain"`
	TargetCount int    `json:"target_count"`
	ListenerID  string `json:"listener_id"`
	LbID        string `json:"lb_id"`
}

// HasLbConditions check if there are load balancer related query conditions
func (r *ListUrlRulesByTopologyReq) HasLbConditions() bool {
	return len(r.LbRegions) > 0 || len(r.LbNetworkTypes) > 0 || len(r.LbIpVersions) > 0 ||
		len(r.CloudLbIds) > 0 || len(r.LbVips) > 0 || len(r.LbDomains) > 0
}

// HasListenerConditions check if there are listener related query conditions
func (r *ListUrlRulesByTopologyReq) HasListenerConditions() bool {
	return len(r.LblProtocol) > 0 || len(r.LblPorts) > 0
}

// HasTargetConditions check if there are target related query conditions
func (r *ListUrlRulesByTopologyReq) HasTargetConditions() bool {
	return len(r.TargetIps) > 0 || len(r.TargetPorts) > 0
}

// Validate validate request parameters
func (r *ListUrlRulesByTopologyReq) Validate() error {
	return nil
}

// BuildLoadBalancerFilter build load balancer query filter
func (r *ListUrlRulesByTopologyReq) BuildLoadBalancerFilter(bizID int64, vendor enumor.Vendor) *filter.Expression {
	lbConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", r.AccountID),
		tools.RuleEqual("bk_biz_id", bizID),
	}

	// If no load balancer conditions, return only basic conditions
	if !r.HasLbConditions() {
		return tools.ExpressionAnd(lbConditions...)
	}

	if len(r.LbRegions) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("region", r.LbRegions))
	}
	if len(r.LbNetworkTypes) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("lb_type", r.LbNetworkTypes))
	}
	if len(r.LbIpVersions) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("ip_version", r.LbIpVersions))
	}
	if len(r.CloudLbIds) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("cloud_id", r.CloudLbIds))
	}
	if len(r.LbDomains) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("domain", r.LbDomains))
	}

	if len(r.LbVips) > 0 {
		vipConditions := []*filter.AtomRule{
			tools.RuleJsonOverlaps("private_ipv4_addresses", r.LbVips),
			tools.RuleJsonOverlaps("private_ipv6_addresses", r.LbVips),
			tools.RuleJsonOverlaps("public_ipv4_addresses", r.LbVips),
			tools.RuleJsonOverlaps("public_ipv6_addresses", r.LbVips),
		}
		vipOrFilter := tools.ExpressionOr(vipConditions...)
		andFilter, _ := tools.And(tools.ExpressionAnd(lbConditions...), vipOrFilter)
		return andFilter
	}

	return tools.ExpressionAnd(lbConditions...)
}

// BuildListenerFilter build listener query filter
func (r *ListUrlRulesByTopologyReq) BuildListenerFilter(bizID int64, vendor enumor.Vendor, lbIDs []string) *filter.Expression {
	if !r.HasListenerConditions() {
		return nil
	}

	listenerConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", r.AccountID),
		tools.RuleEqual("bk_biz_id", bizID),
	}

	if len(lbIDs) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("lb_id", lbIDs))
	}

	if len(r.LblProtocol) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("protocol", r.LblProtocol))
	}
	if len(r.LblPorts) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("port", r.LblPorts))
	}

	return tools.ExpressionAnd(listenerConditions...)
}

// BuildRuleFilter build rule query filter
func (r *ListUrlRulesByTopologyReq) BuildRuleFilter() []*filter.AtomRule {
	if len(r.RuleDomains) == 0 && len(r.RuleUrls) == 0 {
		return nil
	}

	var conditions []*filter.AtomRule
	if len(r.RuleDomains) > 0 {
		conditions = append(conditions, tools.RuleIn("domain", r.RuleDomains))
	}
	if len(r.RuleUrls) > 0 {
		conditions = append(conditions, tools.RuleIn("url", r.RuleUrls))
	}

	return conditions
}

// BuildTargetGroupFilter build target group query filter
func (r *ListUrlRulesByTopologyReq) BuildTargetGroupFilter(bizID int64, vendor enumor.Vendor) *filter.Expression {
	if !r.HasTargetConditions() {
		return nil
	}

	targetGroupConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", r.AccountID),
		tools.RuleEqual("bk_biz_id", bizID),
	}

	return tools.ExpressionAnd(targetGroupConditions...)
}

// BuildTargetFilter build target query filter
func (r *ListUrlRulesByTopologyReq) BuildTargetFilter(targetGroupIDs []string) *filter.Expression {
	if !r.HasTargetConditions() {
		return nil
	}

	targetConditions := []*filter.AtomRule{
		tools.RuleIn("target_group_id", targetGroupIDs),
	}

	if len(r.TargetIps) > 0 {
		targetConditions = append(targetConditions, tools.RuleIn("ip", r.TargetIps))
	}
	if len(r.TargetPorts) > 0 {
		targetConditions = append(targetConditions, tools.RuleIn("port", r.TargetPorts))
	}

	return tools.ExpressionAnd(targetConditions...)
}
