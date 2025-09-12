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
	LblProtocols   []string       `json:"lbl_protocols" validate:"omitempty"`
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
	ID          string   `json:"id"`
	LbVips      []string `json:"lb_vips"`
	LblProtocol string   `json:"lbl_protocol"`
	LblPort     int      `json:"lbl_port"`
	RuleUrl     string   `json:"rule_url"`
	RuleDomain  string   `json:"rule_domain"`
	TargetCount int      `json:"target_count"`
	CloudLblID  string   `json:"cloud_lbl_id"`
	CloudLbID   string   `json:"cloud_lb_id"`
}

// GetLbCond get lb condition
func (r *ListUrlRulesByTopologyReq) GetLbCond() []filter.RuleFactory {
	rules := make([]filter.RuleFactory, 0)

	if len(r.LbRegions) > 0 {
		rules = append(rules, tools.RuleIn("region", r.LbRegions))
	}
	if len(r.LbNetworkTypes) > 0 {
		rules = append(rules, tools.RuleIn("lb_type", r.LbNetworkTypes))
	}
	if len(r.LbIpVersions) > 0 {
		rules = append(rules, tools.RuleIn("ip_version", r.LbIpVersions))
	}
	if len(r.CloudLbIds) > 0 {
		rules = append(rules, tools.RuleIn("cloud_id", r.CloudLbIds))
	}
	if len(r.LbDomains) > 0 {
		rules = append(rules, tools.RuleIn("domain", r.LbDomains))
	}
	if len(r.LbVips) > 0 {
		rules = append(rules, tools.ExpressionOr(
			tools.RuleJsonOverlaps("private_ipv4_addresses", r.LbVips),
			tools.RuleJsonOverlaps("private_ipv6_addresses", r.LbVips),
			tools.RuleJsonOverlaps("public_ipv4_addresses", r.LbVips),
			tools.RuleJsonOverlaps("public_ipv6_addresses", r.LbVips),
		))
	}

	return rules
}

// GetLblCond get listener condition
func (r *ListUrlRulesByTopologyReq) GetLblCond() []filter.RuleFactory {
	rules := make([]filter.RuleFactory, 0)

	if len(r.LblProtocols) > 0 {
		rules = append(rules, tools.RuleIn("protocol", r.LblProtocols))
	}
	if len(r.LblPorts) > 0 {
		rules = append(rules, tools.RuleIn("port", r.LblPorts))
	}

	return rules
}

// GetRuleCond get rule condition
func (r *ListUrlRulesByTopologyReq) GetRuleCond() []filter.RuleFactory {
	rules := make([]filter.RuleFactory, 0)

	if len(r.RuleDomains) > 0 {
		rules = append(rules, tools.RuleIn("domain", r.RuleDomains))
	}
	if len(r.RuleUrls) > 0 {
		rules = append(rules, tools.RuleIn("url", r.RuleUrls))
	}

	return rules
}

// GetTargetCond get target condition
func (r *ListUrlRulesByTopologyReq) GetTargetCond() []filter.RuleFactory {
	rules := make([]filter.RuleFactory, 0)

	if len(r.TargetIps) > 0 {
		rules = append(rules, tools.RuleIn("ip", r.TargetIps))
	}
	if len(r.TargetPorts) > 0 {
		rules = append(rules, tools.RuleIn("port", r.TargetPorts))
	}

	return rules
}

// Validate validate request parameters
func (r *ListUrlRulesByTopologyReq) Validate() error {
	return nil
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
