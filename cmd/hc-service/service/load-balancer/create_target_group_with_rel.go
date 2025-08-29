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
	"hcm/pkg/kit"
	"time"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateTargetGroupWithRel 创建目标组并绑定监听器
func (svc *clbSvc) CreateTargetGroupWithRel(cts *rest.Contexts) (interface{}, error) {
	req := new(hcproto.CreateTargetGroupWithRelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lb, listener, rule, err := svc.getLoadBalancerListenerRule(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	targetGroupName := svc.generateTargetGroupName(rule, listener, time.Now())

	rsList := svc.convertTargetsToBaseReq(req.Targets)

	result, err := svc.createTargetGroupWithRel(cts.Kit, lb, listener, rule, rsList, targetGroupName)
	if err != nil {
		logs.Errorf("fail to create target group with rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.IDs) == 0 {
		return nil, errf.New(errf.Unknown, "create target group failed, no target group id returned")
	}

	return &hcproto.CreateTargetGroupWithRelResult{
		TargetGroupID: result.IDs[0],
	}, nil
}

// getLoadBalancerListenerRule 获取负载均衡器、监听器和规则信息
func (svc *clbSvc) getLoadBalancerListenerRule(kt *kit.Kit, req *hcproto.CreateTargetGroupWithRelReq) (
	*corelb.BaseLoadBalancer, *corelb.BaseListener, *corelb.TCloudLbUrlRule, error) {

	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, &core.ListReq{
		Filter: tools.EqualExpression("id", req.LoadBalancerID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list load balancer, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	if len(lbResp.Details) == 0 {
		return nil, nil, nil, errf.Newf(errf.RecordNotFound, "load balancer: %s not found", req.LoadBalancerID)
	}
	lb := &lbResp.Details[0]

	listenerResp, err := svc.dataCli.Global.LoadBalancer.ListListener(kt, &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", req.ListenerID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list listener, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	if len(listenerResp.Details) == 0 {
		return nil, nil, nil, errf.Newf(errf.RecordNotFound, "listener: %s not found", req.ListenerID)
	}
	listener := &listenerResp.Details[0]

	if req.ListenerRuleID == "" {
		return lb, listener, nil, nil
	}

	ruleResp, err := svc.dataCli.TCloud.LoadBalancer.ListUrlRule(kt, &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", req.ListenerRuleID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list url rule, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	if len(ruleResp.Details) == 0 {
		return nil, nil, nil, errf.Newf(errf.RecordNotFound, "url rule: %s not found", req.ListenerRuleID)
	}
	rule := &ruleResp.Details[0]

	return lb, listener, rule, nil
}

// convertTargetsToBaseReq 转换Target到TargetBaseReq
func (svc *clbSvc) convertTargetsToBaseReq(targets []*hcproto.RegisterTarget) []*dataproto.TargetBaseReq {
	rsList := make([]*dataproto.TargetBaseReq, 0, len(targets))
	for _, target := range targets {
		rs := &dataproto.TargetBaseReq{
			InstType:    target.TargetType,
			Port:        target.Port,
			Weight:      target.Weight,
			CloudInstID: target.CloudInstID,
			IP: func() string {
				if len(target.PrivateIPAddress) > 0 {
					return target.PrivateIPAddress[0]
				}
				return ""
			}(),
		}
		rsList = append(rsList, rs)
	}
	return rsList
}

// createTargetGroupWithRel 创建目标组并绑定监听器
func (svc *clbSvc) createTargetGroupWithRel(kt *kit.Kit, lb *corelb.BaseLoadBalancer,
	listener *corelb.BaseListener, rule *corelb.TCloudLbUrlRule,
	rsList []*dataproto.TargetBaseReq, targetGroupName string) (*core.BatchCreateResult, error) {

	isLayer4 := rule == nil

	tgCreate := dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{
		TargetGroup: dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			Name:            targetGroupName,
			Vendor:          enumor.TCloud,
			AccountID:       lb.AccountID,
			BkBizID:         lb.BkBizID,
			Region:          lb.Region,
			Protocol:        listener.Protocol,
			Port:            listener.Port,
			TargetGroupType: enumor.LocalTargetGroupType,
			Weight:          0,
			CloudVpcID:      lb.CloudVpcID,
			Memo:            converter.ValToPtr(svc.generateTargetGroupMemo(isLayer4, rule)),
			RsList:          rsList,
		},
		LbID:          lb.ID,
		CloudLbID:     lb.CloudID,
		LblID:         listener.ID,
		CloudLblID:    listener.CloudID,
		BindingStatus: enumor.SuccessBindingStatus,
	}
	if isLayer4 {
		tgCreate.ListenerRuleID = ""
		tgCreate.CloudListenerRuleID = ""
		tgCreate.ListenerRuleType = enumor.Layer4RuleType
	} else {
		tgCreate.ListenerRuleID = rule.ID
		tgCreate.CloudListenerRuleID = rule.CloudID
		tgCreate.ListenerRuleType = rule.RuleType
	}

	tgCreateReq := &dataproto.TCloudBatchCreateTgWithRelReq{
		TargetGroups: []dataproto.CreateTargetGroupWithRel[corelb.TCloudTargetGroupExtension]{tgCreate},
	}

	return svc.dataCli.TCloud.LoadBalancer.BatchCreateTargetGroupWithRel(kt, tgCreateReq)
}

// generateTargetGroupMemo 生成目标组备注信息
func (svc *clbSvc) generateTargetGroupMemo(isLayer4 bool, rule *corelb.TCloudLbUrlRule) string {
	if isLayer4 {
		return "auto created for layer4 listener"
	}
	return fmt.Sprintf("auto created for rule %s", rule.CloudID)
}

// generateTargetGroupName 生成目标组名称
func (svc *clbSvc) generateTargetGroupName(rule *corelb.TCloudLbUrlRule, listener *corelb.BaseListener, now time.Time) string {
	if rule == nil {
		// 四层监听器：使用监听器ID
		return fmt.Sprintf("auto_tg_%s_%s", listener.CloudID, now.Format("20060102150405"))
	}
	return fmt.Sprintf("auto_tg_%s_%s", rule.CloudID, now.Format("20060102150405"))
}
