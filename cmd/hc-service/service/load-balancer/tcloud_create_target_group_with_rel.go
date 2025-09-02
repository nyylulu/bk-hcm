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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateTargetGroupWithRel 创建目标组并绑定监听器，同时添加RS
func (svc *clbSvc) CreateTargetGroupWithRel(cts *rest.Contexts) (interface{}, error) {
	req := new(protolb.CreateTargetGroupWithRelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	logs.Infof("start create target group with rel, vendor: %s, lb_id: %s, listener_id: %s, targets count: %d, rid: %s",
		req.Vendor, req.LoadBalancerID, req.ListenerID, len(req.Targets), cts.Kit.Rid)

	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.LoadBalancerID),
		Page:   core.NewDefaultBasePage(),
	}
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, lbReq)
	if err != nil {
		logs.Errorf("list load balancer failed, lb_id: %s, err: %v, rid: %s", req.LoadBalancerID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if len(lbResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "load balancer not found, lb_id: %s", req.LoadBalancerID)
	}
	lbInfo := lbResp.Details[0]

	listenerReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.ListenerID),
		Page:   core.NewDefaultBasePage(),
	}
	listenerResp, err := svc.dataCli.Global.LoadBalancer.ListListener(cts.Kit, listenerReq)
	if err != nil {
		logs.Errorf("list listener failed, listener_id: %s, err: %v, rid: %s", req.ListenerID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if len(listenerResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "listener not found, listener_id: %s", req.ListenerID)
	}
	listenerInfo := listenerResp.Details[0]

	targetGroupID, err := svc.createTargetGroup(cts.Kit, req, lbInfo, listenerInfo)
	if err != nil {
		logs.Errorf("create target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = svc.bindListenerToTargetGroup(cts.Kit, req, targetGroupID, lbInfo, listenerInfo)
	if err != nil {
		logs.Errorf("bind listener to target group failed, err: %v, rid: %s", err, cts.Kit.Rid)

	}

	err = svc.addTargetsToTargetGroup(cts.Kit, req, targetGroupID, lbInfo)
	if err != nil {
		logs.Errorf("add targets to target group failed, err: %v, rid: %s", err, cts.Kit.Rid)

	}

	logs.Infof("successfully created target group with rel, target group ID: %s, rid: %s", targetGroupID, cts.Kit.Rid)

	return &protolb.CreateTargetGroupWithRelResult{
		TargetGroupID: targetGroupID,
		ListenerID:    req.ListenerID,
		RuleID:        req.ListenerRuleID,
		TargetsCount:  len(req.Targets),
	}, nil
}

// createTargetGroup 创建目标组
func (svc *clbSvc) createTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
	lbInfo corelb.BaseLoadBalancer, listenerInfo corelb.BaseListener) (string, error) {

	targetGroupName := fmt.Sprintf("auto_tg_%s_%d", listenerInfo.CloudID, listenerInfo.Port)
	if req.RuleType == enumor.Layer7RuleType && req.ListenerRuleID != "" {
		targetGroupName = fmt.Sprintf("auto_tg_%s_%d_%s", listenerInfo.CloudID, listenerInfo.Port, req.ListenerRuleID)
	}

	tgCreateReq := &dataproto.TCloudTargetGroupCreateReq{
		TargetGroups: []dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			{
				AccountID:   lbInfo.AccountID,
				Region:      lbInfo.Region,
				CloudVpcID:  lbInfo.CloudVpcID,
				Name:        targetGroupName,
				Protocol:    listenerInfo.Protocol,
				Port:        listenerInfo.Port,
				VpcID:       lbInfo.CloudVpcID,
				Vendor:      enumor.TCloud,
				BkBizID:     lbInfo.BkBizID,
				HealthCheck: types.JsonField(""),
			},
		},
	}

	logs.Infof("creating target group, name: %s, protocol: %s, port: %d, rid: %s",
		targetGroupName, listenerInfo.Protocol, listenerInfo.Port, kt.Rid)

	tgResp, err := svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudTargetGroup(kt, tgCreateReq)
	if err != nil {
		logs.Errorf("create target group failed, err: %v, rid: %s", err, kt.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	if len(tgResp.IDs) == 0 {
		return "", errf.New(errf.Aborted, "create target group failed, no ID returned")
	}

	targetGroupID := tgResp.IDs[0]
	logs.Infof("successfully created target group, ID: %s, name: %s, rid: %s", targetGroupID, targetGroupName, kt.Rid)

	return targetGroupID, nil
}

// bindListenerToTargetGroup 绑定监听器到目标组
func (svc *clbSvc) bindListenerToTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
	targetGroupID string, lbInfo corelb.BaseLoadBalancer, listenerInfo corelb.BaseListener) error {

	bindReq := &dataproto.TargetGroupListenerRelCreateReq{
		TargetGroupID:       targetGroupID,
		LblID:               req.ListenerID,
		ListenerRuleID:      req.ListenerRuleID,
		Vendor:              req.Vendor,
		LbID:                req.LoadBalancerID,
		CloudLbID:           lbInfo.CloudID,
		CloudLblID:          listenerInfo.CloudID,
		ListenerRuleType:    req.RuleType,
		CloudListenerRuleID: req.ListenerRuleID,
		CloudTargetGroupID:  "",
	}

	logs.Infof("binding listener to target group, target_group_id: %s, listener_id: %s, rule_id: %s, rid: %s",
		targetGroupID, req.ListenerID, req.ListenerRuleID, kt.Rid)

	_, err := svc.dataCli.Global.LoadBalancer.CreateTargetGroupListenerRel(kt, bindReq)
	if err != nil {
		logs.Errorf("create target group listener rel failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	if req.RuleType == enumor.Layer7RuleType && req.ListenerRuleID != "" {
		err = svc.updateUrlRuleTargetGroup(kt, req.ListenerRuleID, targetGroupID)
		if err != nil {
			logs.Errorf("update url rule target group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("successfully bound listener to target group, target_group_id: %s, listener_id: %s, rid: %s",
		targetGroupID, req.ListenerID, kt.Rid)

	return nil
}

// addTargetsToTargetGroup 添加RS到目标组
func (svc *clbSvc) addTargetsToTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
	targetGroupID string, lbInfo corelb.BaseLoadBalancer) error {

	targets := make([]*dataproto.TargetBaseReq, 0, len(req.Targets))
	for _, target := range req.Targets {
		targetReq := &dataproto.TargetBaseReq{
			AccountID:     lbInfo.AccountID,
			TargetGroupID: targetGroupID,
			InstType:      target.TargetType,
			Port:          target.Port,
			Weight:        target.Weight,
			Zone:          target.Zone,
		}

		switch target.TargetType {
		case enumor.EniInstType:
			targetReq.IP = target.EniIp
		case enumor.CvmInstType:
			targetReq.CloudInstID = target.CloudInstID
			targetReq.InstName = target.InstName
			targetReq.PrivateIPAddress = target.PrivateIPAddress
			targetReq.PublicIPAddress = target.PublicIPAddress
		}

		targets = append(targets, targetReq)
	}

	logs.Infof("adding %d targets to target group, target_group_id: %s, rid: %s",
		len(targets), targetGroupID, kt.Rid)

	addReq := &dataproto.TargetBatchCreateReq{
		Targets: targets,
	}
	result, err := svc.dataCli.Global.LoadBalancer.BatchCreateTCloudTarget(kt, addReq)
	if err != nil {
		logs.Errorf("batch create tcloud target failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	logs.Infof("successfully added %d targets to target group, target_group_id: %s, target_ids: %v, rid: %s",
		len(result.IDs), targetGroupID, result.IDs, kt.Rid)

	return nil
}

// updateUrlRuleTargetGroup 更新URL规则的目标组ID
func (svc *clbSvc) updateUrlRuleTargetGroup(kt *kit.Kit, ruleID, targetGroupID string) error {
	updateReq := &dataproto.TCloudUrlRuleBatchUpdateReq{
		UrlRules: []*dataproto.TCloudUrlRuleUpdate{
			{
				ID:                 ruleID,
				TargetGroupID:      targetGroupID,
				CloudTargetGroupID: targetGroupID,
			},
		},
	}

	err := svc.dataCli.TCloud.LoadBalancer.BatchUpdateTCloudUrlRule(kt, updateReq)
	if err != nil {
		logs.Errorf("update url rule target group failed, rule_id: %s, target_group_id: %s, err: %v, rid: %s",
			ruleID, targetGroupID, err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	logs.Infof("successfully updated url rule target group, rule_id: %s, target_group_id: %s, rid: %s",
		ruleID, targetGroupID, kt.Rid)

	return nil
}
