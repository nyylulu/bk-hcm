/*
 *
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
	typelb "hcm/pkg/adaptor/types/load-balancer"
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
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// BatchCreateTCloudZiyanTargets 批量添加RS
func (svc *clbSvc) BatchCreateTCloudZiyanTargets(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID)
	if err != nil {
		return nil, err
	}

	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", tgID)
	}

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}
	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
	}

	// 查询Url规则列表
	ruleIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.ListenerRuleID
	})
	urlRuleReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.dataCli.TCloudZiyan.LoadBalancer.ListUrlRule(cts.Kit, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}
	rule := urlRuleList.Details[0]
	lbReq := core.ListReq{Filter: tools.EqualExpression("id", rule.LbID), Page: core.NewDefaultBasePage()}
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, &lbReq)
	if err != nil {
		logs.Errorf("fail to find ziyan load balancer for add target group, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(lbResp.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "load balancer not found")
	}

	// 调用云端批量绑定虚拟主机接口
	return svc.batchAddZiyanTargetsToGroup(cts.Kit, req, lbResp.Details[0], rule)
}

func (svc *clbSvc) batchAddZiyanTargetsToGroup(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	lbInfo corelb.BaseLoadBalancer, ruleInfo corelb.TCloudLbUrlRule) (*protolb.BatchCreateResult, error) {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	rsOpt := &typelb.TCloudRegisterTargetsOption{
		Region:         lbInfo.Region,
		LoadBalancerId: ruleInfo.CloudLbID,
	}
	for _, rsItem := range req.RsList {
		tmpRs := &typelb.BatchTarget{
			ListenerId: cvt.ValToPtr(ruleInfo.CloudLBLID),
			Port:       cvt.ValToPtr(rsItem.Port),
			Weight:     rsItem.Weight,
		}
		// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
		tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
		if ruleInfo.RuleType == enumor.Layer7RuleType {
			tmpRs.LocationId = cvt.ValToPtr(ruleInfo.CloudID)
		}
		rsOpt.Targets = append(rsOpt.Targets, tmpRs)
	}
	failIDs, err := tcloudAdpt.RegisterTargets(kt, rsOpt)
	if err != nil {
		logs.Errorf("register tcloud-ziyan target api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
		return nil, err
	}
	if len(failIDs) > 0 {
		logs.Errorf("register tcloud-ziyan target api partially failed, failLblIDs: %v, req: %+v, rsOpt: %+v, rid: %s",
			failIDs, req, rsOpt, kt.Rid)
		return nil, errf.Newf(errf.PartialFailed, "register tcloud-ziyan target failed, failListenerIDs: %v",
			failIDs)
	}

	rsIDs, err := svc.batchCreateTargetDb(kt, req, lbInfo.AccountID, req.TargetGroupID, lbInfo.Region)
	if err != nil {
		return nil, err
	}
	return &protolb.BatchCreateResult{SuccessCloudIDs: rsIDs.IDs}, nil
}

// BatchRemoveTCloudZiyanTargets 批量移除RS
func (svc *clbSvc) BatchRemoveTCloudZiyanTargets(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID)
	if err != nil {
		return nil, err
	}

	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", tgID)
	}

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
	}

	// 查询Url规则列表
	ruleIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.ListenerRuleID
	})
	urlRuleReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.dataCli.TCloudZiyan.LoadBalancer.ListUrlRule(cts.Kit, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云端批量解绑四七层后端服务接口
	return nil, svc.batchUnRegisterZiyanTargetCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchUnRegisterZiyanTargetCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	cloudLBExists := make(map[string]struct{}, 0)
	rsOpt := &typelb.TCloudRegisterTargetsOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := cloudLBExists[ruleItem.CloudLbID]; !ok {
			rsOpt.LoadBalancerId = ruleItem.CloudLbID
			cloudLBExists[ruleItem.CloudLbID] = struct{}{}
		}
		for _, rsItem := range req.RsList {
			tmpRs := &typelb.BatchTarget{
				ListenerId: cvt.ValToPtr(ruleItem.CloudLBLID),
				Port:       cvt.ValToPtr(rsItem.Port),
			}
			if ruleItem.RuleType == enumor.Layer7RuleType {
				tmpRs.LocationId = cvt.ValToPtr(ruleItem.CloudID)
			}
			// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
			tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
			rsOpt.Targets = append(rsOpt.Targets, tmpRs)
		}
		failIDs, err := tcloudAdpt.DeRegisterTargets(kt, rsOpt)
		if err != nil {
			logs.Errorf("unregister tcloud-ziyan target api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
			return err
		}
		if len(failIDs) > 0 {
			logs.Errorf("unregister tcloud-ziyan target api partially failed, failLblIDs: %v, req: %+v, rsOpt: %+v, rid: %s",
				failIDs, req, rsOpt, kt.Rid)
			return errf.Newf(errf.PartialFailed, "unregister tcloud-ziyan target failed, failListenerIDs: %v", failIDs)
		}
	}

	err = svc.batchDeleteTargetDb(kt, req, tgInfo.AccountID, tgInfo.ID)
	if err != nil {
		return err
	}
	return nil
}

// BatchModifyTCloudZiyanTargetsPort 批量修改RS端口
func (svc *clbSvc) BatchModifyTCloudZiyanTargetsPort(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID)
	if err != nil {
		return nil, err
	}

	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", tgID)
	}

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
	}

	// 查询Url规则列表
	ruleIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.ListenerRuleID
	})
	urlRuleReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.dataCli.TCloudZiyan.LoadBalancer.ListUrlRule(cts.Kit, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云端批量解绑四七层后端服务接口
	return nil, svc.batchModifyZiyanTargetPortCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchModifyZiyanTargetPortCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	rsOpt := &typelb.TCloudTargetPortUpdateOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		rsOpt.LoadBalancerId = ruleItem.CloudLbID
		rsOpt.ListenerId = ruleItem.CloudLBLID
		if ruleItem.RuleType == enumor.Layer7RuleType {
			rsOpt.LocationId = cvt.ValToPtr(ruleItem.CloudID)
		}
		for _, rsItem := range req.RsList {
			tmpRs := &typelb.BatchTarget{
				Type: cvt.ValToPtr(string(rsItem.InstType)),
				Port: cvt.ValToPtr(rsItem.Port),
			}
			// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
			tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
			rsOpt.Targets = append(rsOpt.Targets, tmpRs)
		}
		rsOpt.NewPort = cvt.PtrToVal(req.RsList[0].NewPort)
		err = tcloudAdpt.ModifyTargetPort(kt, rsOpt)
		if err != nil {
			logs.Errorf("batch modify tcloud-ziyan target port api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt,
				kt.Rid)
			return errf.Newf(errf.PartialFailed, "batch modify tcloud-ziyan target port api failed, err: %v", err)
		}
	}

	err = svc.batchUpdateTargetPortWeightDb(kt, req)
	if err != nil {
		return err
	}
	return nil
}

// BatchModifyTCloudZiyanTargetsWeight 批量修改RS权重
func (svc *clbSvc) BatchModifyTCloudZiyanTargetsWeight(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID)
	if err != nil {
		return nil, err
	}

	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", tgID)
	}

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
	}

	// 查询Url规则列表
	ruleIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.ListenerRuleID
	})
	urlRuleReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.dataCli.TCloudZiyan.LoadBalancer.ListUrlRule(cts.Kit, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud-ziyan listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 批量修改监听器绑定的后端机器的转发权重
	return nil, svc.batchModifyZiyanTargetWeightCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchModifyZiyanTargetWeightCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	rsOpt := &typelb.TCloudTargetWeightUpdateOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		rsOpt.LoadBalancerId = ruleItem.CloudLbID
		tmpWeightRule := &typelb.TargetWeightRule{
			ListenerId: cvt.ValToPtr(ruleItem.CloudLBLID),
		}
		if ruleItem.RuleType == enumor.Layer7RuleType {
			tmpWeightRule.LocationId = cvt.ValToPtr(ruleItem.CloudID)
		}
		for _, rsItem := range req.RsList {
			tmpRs := &typelb.BatchTarget{
				Type:   cvt.ValToPtr(string(rsItem.InstType)),
				Port:   cvt.ValToPtr(rsItem.Port),
				Weight: rsItem.NewWeight,
			}
			// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
			tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
			tmpWeightRule.Targets = append(tmpWeightRule.Targets, tmpRs)
			rsOpt.ModifyList = append(rsOpt.ModifyList, tmpWeightRule)
		}
		err = tcloudAdpt.ModifyTargetWeight(kt, rsOpt)
		if err != nil {
			logs.Errorf("batch modify tcloud-ziyan target port api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt,
				kt.Rid)
			return errf.Newf(errf.PartialFailed, "batch modify tcloud-ziyan target port api failed, err: %v", err)
		}
	}

	err = svc.batchUpdateTargetPortWeightDb(kt, req)
	if err != nil {
		return err
	}
	return nil
}

// ListTCloudZiyanTargetsHealth 查询目标组所在负载均衡的端口健康数据
func (svc *clbSvc) ListTCloudZiyanTargetsHealth(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudTargetHealthReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}
	if len(req.Region) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "region is required")
	}

	tcloudAdpt, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudListTargetHealthOption{
		Region:          req.Region,
		LoadBalancerIDs: req.CloudLbIDs,
	}
	healthList, err := tcloudAdpt.ListTargetHealth(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud-ziyan target health api failed, err: %v, cloudLbIDs: %v, rid: %s",
			err, req.CloudLbIDs, cts.Kit.Rid)
		return nil, err
	}

	healths := &protolb.TCloudTargetHealthResp{}
	for _, item := range healthList {
		tmpHealthInfo := protolb.TCloudTargetHealthResult{CloudLbID: cvt.PtrToVal(item.LoadBalancerId)}
		for _, lblItem := range item.Listeners {
			tmpListener := &protolb.TCloudTargetHealthLblResult{
				CloudLblID:   cvt.PtrToVal(lblItem.ListenerId),
				Protocol:     enumor.ProtocolType(cvt.PtrToVal(lblItem.Protocol)),
				ListenerName: cvt.PtrToVal(lblItem.ListenerName),
			}
			for _, ruleItem := range lblItem.Rules {
				var healthNum, unHealthNum int64
				for _, targetItem := range ruleItem.Targets {
					// 当前健康状态，true：健康，false：不健康（包括尚未开始探测、探测中、状态异常等几种状态）。
					if cvt.PtrToVal(targetItem.HealthStatus) {
						healthNum++
					} else {
						unHealthNum++
					}
				}

				if !tmpListener.Protocol.IsLayer7Protocol() {
					tmpListener.HealthCheck = &corelb.TCloudHealthCheckInfo{
						HealthNum:   cvt.ValToPtr(healthNum),
						UnHealthNum: cvt.ValToPtr(unHealthNum),
					}
					break
				} else {
					tmpListener.Rules = append(tmpListener.Rules, &protolb.TCloudTargetHealthRuleResult{
						CloudRuleID: cvt.PtrToVal(ruleItem.LocationId),
						HealthCheck: &corelb.TCloudHealthCheckInfo{
							HealthNum:   cvt.ValToPtr(healthNum),
							UnHealthNum: cvt.ValToPtr(unHealthNum),
						},
					})
				}
			}
			tmpHealthInfo.Listeners = append(tmpHealthInfo.Listeners, tmpListener)
		}
		healths.Details = append(healths.Details, tmpHealthInfo)
	}

	return healths, nil
}

// BatchModifyZiyanListenerTargetsWeight 按负载均衡批量调整监听器的RS权重
func (svc *clbSvc) BatchModifyZiyanListenerTargetsWeight(cts *rest.Contexts) (any, error) {
	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	req := new(protolb.TCloudBatchModifyRsWeightReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 过滤符合条件的RS列表
	lblRsList, err := svc.filterListenerTargetWeightList(cts.Kit, lbID, req.Details, req.NewRsWeight)
	if err != nil {
		return nil, err
	}
	// 没有需要调整权重的RS，不用处理
	if len(lblRsList) == 0 {
		logs.Infof("modify listener rs weight no call api, not need to modify rs weight, accountID: %s, lbID: %s, "+
			"cloudLbID: %s, details: %+v, rid: %s",
			req.AccountID, lbID, req.LoadBalancerCloudId, cvt.PtrToSlice(req.Details), cts.Kit.Rid)
		return &protolb.BatchCreateResult{SuccessCloudIDs: []string{"HAS-MODIFY-WEIGHT"}}, nil
	}

	cloudRuleIDs, updateRsList, err := svc.modifyTCloudZiyanListenerTargetsWeight(cts.Kit, req, lblRsList)
	if err != nil {
		return nil, err
	}

	// 更新DB中的RS权重
	rsWeightUpdateList := &protolb.TCloudBatchOperateTargetReq{LbID: lbID, RsList: updateRsList}
	err = svc.batchUpdateTargetPortWeightDb(cts.Kit, rsWeightUpdateList)
	if err != nil {
		logs.Errorf("modify listener rs weight db failed, err: %v, lbID: %s, rsWeightUpdateList: %+v, "+
			"cloudRuleIDs: %v, rid: %s", err, lbID, cvt.PtrToVal(rsWeightUpdateList), cloudRuleIDs, cts.Kit.Rid)
		return nil, err
	}

	// 记录操作日志，方便排查问题
	logs.Infof("modify listener rs weight success, lbID: %s, cloudRuleIDs: %v, req: %+v, lblRsList: %+v, rid: %s",
		lbID, cloudRuleIDs, req, cvt.PtrToSlice(lblRsList), cts.Kit.Rid)
	return &protolb.BatchCreateResult{SuccessCloudIDs: cloudRuleIDs}, nil
}

func (svc *clbSvc) modifyTCloudZiyanListenerTargetsWeight(kt *kit.Kit, req *protolb.TCloudBatchModifyRsWeightReq,
	lblRsList []*dataproto.ListBatchListenerResult) ([]string, []*dataproto.TargetBaseReq, error) {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, req.AccountID)
	if err != nil {
		logs.Errorf("modify listener rs weight tcloud-ziyan api failed, get account failed, err: %v, "+
			"accountID: %s, rid: %s", err, req.AccountID, kt.Rid)
		return nil, nil, err
	}

	cloudRuleIDs := make([]string, 0)
	updateRsList := make([]*dataproto.TargetBaseReq, 0)
	rsOpt := &typelb.TCloudTargetWeightUpdateOption{
		LoadBalancerId: req.LoadBalancerCloudId,
		Region:         req.Region,
		ModifyList:     make([]*typelb.TargetWeightRule, 0),
	}
	for _, item := range lblRsList {
		for _, rsItem := range item.RsList {
			tmpWeightRule := &typelb.TargetWeightRule{ListenerId: cvt.ValToPtr(item.CloudLblID)}
			if rsItem.RuleType == enumor.Layer7RuleType {
				tmpWeightRule.LocationId = cvt.ValToPtr(rsItem.CloudRuleID)
			}
			tmpRs := &typelb.BatchTarget{
				Type:   cvt.ValToPtr(string(rsItem.InstType)),
				Port:   cvt.ValToPtr(rsItem.Port),
				Weight: req.NewRsWeight,
			}
			// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
			tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
			tmpWeightRule.Targets = append(tmpWeightRule.Targets, tmpRs)
			rsOpt.ModifyList = append(rsOpt.ModifyList, tmpWeightRule)
			updateRsList = append(updateRsList, &dataproto.TargetBaseReq{
				ID: rsItem.ID, NewWeight: req.NewRsWeight,
			})
			cloudRuleIDs = append(cloudRuleIDs, rsItem.CloudRuleID)
		}
	}
	err = tcloudAdpt.ModifyTargetWeight(kt, rsOpt)
	if err != nil {
		logs.Errorf("modify listener rs weight tcloud-ziyan api failed, err: %v, newWeight: %d, rsOpt: %+v, rid: %s",
			err, req.NewRsWeight, rsOpt, kt.Rid)
		return nil, nil, err
	}
	return cloudRuleIDs, updateRsList, nil
}

// BatchRemoveZiyanListenerTargets 按负载均衡批量移除监听器的RS
func (svc *clbSvc) BatchRemoveZiyanListenerTargets(cts *rest.Contexts) (any, error) {
	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	req := new(protolb.TCloudBatchUnbindRsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 过滤符合条件的RS列表
	lblRsList, err := svc.filterListenerTargetList(cts.Kit, lbID, req.Details)
	if err != nil {
		return nil, err
	}
	// 没有需要移除的RS，不用处理
	if len(lblRsList) == 0 {
		logs.Infof("listener and rs no call api, has unbind rs, accountID: %s, lbID: %s, cloudLbID: %s, "+
			"details: %+v, rid: %s",
			req.AccountID, lbID, req.LoadBalancerCloudId, cvt.PtrToSlice(req.Details), cts.Kit.Rid)
		return &protolb.BatchCreateResult{SuccessCloudIDs: []string{"HAS-UNBIND-RS"}}, nil
	}

	targetIDs, cloudLblIDs, err := svc.unbindTCloudZiyanListenerTargets(cts.Kit, req, lblRsList)
	if err != nil {
		return nil, err
	}

	// 删除已解绑的RS
	for _, partIDs := range slice.Split(targetIDs, int(core.DefaultMaxPageLimit)) {
		delReq := &dataproto.LoadBalancerBatchDeleteReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", req.AccountID),
				tools.RuleIn("id", partIDs),
			),
		}
		if err = svc.dataCli.Global.LoadBalancer.BatchDeleteTarget(cts.Kit, delReq); err != nil {
			logs.Errorf("delete load balancer target failed, err: %v, partIDs: %v, rid: %s", err, partIDs, cts.Kit.Rid)
			return nil, err
		}
	}
	// 记录操作日志，方便排查问题
	logs.Infof("listener unbind rs success, lbID: %s, cloudLblIDs: %v, req: %+v, lblRsList: %+v, rid: %s",
		lbID, cloudLblIDs, req, cvt.PtrToSlice(lblRsList), cts.Kit.Rid)
	return &protolb.BatchCreateResult{SuccessCloudIDs: cloudLblIDs}, nil
}

func (svc *clbSvc) unbindTCloudZiyanListenerTargets(kt *kit.Kit, req *protolb.TCloudBatchUnbindRsReq,
	lblRsList []*dataproto.ListBatchListenerResult) ([]string, []string, error) {

	tcloudAdpt, err := svc.ad.TCloudZiyan(kt, req.AccountID)
	if err != nil {
		logs.Errorf("listener unbind rs tcloud-ziyan api failed, get account failed, err: %v, accountID: %s, rid: %s",
			err, req.AccountID, kt.Rid)
		return nil, nil, err
	}

	targetIDs := make([]string, 0)
	cloudLblIDs := make([]string, 0)
	rsOpt := &typelb.TCloudRegisterTargetsOption{
		LoadBalancerId: req.LoadBalancerCloudId,
		Region:         req.Region,
	}
	for _, item := range lblRsList {
		for _, rsItem := range item.RsList {
			tmpRs := &typelb.BatchTarget{
				ListenerId: cvt.ValToPtr(item.CloudLblID),
				Port:       cvt.ValToPtr(rsItem.Port),
			}
			if rsItem.RuleType == enumor.Layer7RuleType {
				tmpRs.LocationId = cvt.ValToPtr(rsItem.CloudRuleID)
			}
			// 对于cvm，使用InstanceId参数，其他所有类型，使用EniIp参数 --story=124323667
			tmpRs = setTargetInstanceIDAndEniIP(rsItem.InstType, rsItem.CloudInstID, rsItem.IP, tmpRs)
			rsOpt.Targets = append(rsOpt.Targets, tmpRs)
			targetIDs = append(targetIDs, rsItem.ID)
			cloudLblIDs = append(cloudLblIDs, item.CloudLblID)
		}
	}
	failIDs, err := tcloudAdpt.DeRegisterTargets(kt, rsOpt)
	if err != nil {
		logs.Errorf("listener unbind rs api failed, err: %v, rsOpt: %+v, rid: %s", err, cvt.PtrToVal(rsOpt), kt.Rid)
		return nil, nil, err
	}
	if len(failIDs) > 0 {
		logs.Errorf("listener unbind rs tcloud-ziyan api partially failed, failLblIDs: %v, req: %+v, "+
			"rsOpt: %+v, rid: %s", failIDs, cvt.PtrToVal(req), cvt.PtrToVal(rsOpt), kt.Rid)
		return nil, nil, errf.Newf(errf.PartialFailed, "unbind cloud listener target failed, failLblIDs: %v", failIDs)
	}
	return targetIDs, cloudLblIDs, nil
}

// CreateZiyanTargetGroupWithRel 创建目标组并绑定监听器，同时添加RS
func (svc *clbSvc) CreateZiyanTargetGroupWithRel(cts *rest.Contexts) (interface{}, error) {
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
		Filter: tools.EqualExpression("cloud_id", req.LoadBalancerID),
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
		Filter: tools.EqualExpression("cloud_id", req.ListenerID),
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

	targetGroupID, err := svc.createTCloudZiyanTargetGroup(cts.Kit, req, lbInfo, listenerInfo)
	if err != nil {
		logs.Errorf("create target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = svc.bindTCloudZiyanListenerToTargetGroup(cts.Kit, req, targetGroupID, lbInfo, listenerInfo)
	if err != nil {
		logs.Errorf("bind listener to target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = svc.addTCloudZiyanTargetsToTargetGroup(cts.Kit, req, targetGroupID, lbInfo)
	if err != nil {
		logs.Errorf("add targets to target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	logs.Infof("successfully created target group with rel, target group ID: %s, rid: %s", targetGroupID, cts.Kit.Rid)

	return &protolb.CreateTargetGroupWithRelResult{
		TargetGroupID: targetGroupID,
	}, nil
}

// createTCloudZiyanTargetGroup 创建TCloudZiyan目标组
func (svc *clbSvc) createTCloudZiyanTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
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
				Vendor:      enumor.TCloudZiyan,
				BkBizID:     lbInfo.BkBizID,
				HealthCheck: types.JsonField(""),
			},
		},
	}

	logs.Infof("creating target group, name: %s, protocol: %s, port: %d, rid: %s",
		targetGroupName, listenerInfo.Protocol, listenerInfo.Port, kt.Rid)

	tgResp, err := svc.dataCli.TCloudZiyan.LoadBalancer.BatchCreateTCloudTargetGroup(kt, tgCreateReq)
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

// bindTCloudZiyanListenerToTargetGroup 绑定监听器到目标组
func (svc *clbSvc) bindTCloudZiyanListenerToTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
	targetGroupID string, lbInfo corelb.BaseLoadBalancer, listenerInfo corelb.BaseListener) error {

	bindReq := &dataproto.TargetGroupListenerRelCreateReq{
		TargetGroupID:       targetGroupID,
		LblID:               listenerInfo.ID,
		ListenerRuleID:      req.ListenerRuleID,
		Vendor:              req.Vendor,
		LbID:                lbInfo.ID,
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
		err = svc.updateTCloudZiyanUrlRuleTargetGroup(kt, req.ListenerRuleID, targetGroupID)
		if err != nil {
			logs.Errorf("update url rule target group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("successfully bound listener to target group, target_group_id: %s, listener_id: %s, rid: %s",
		targetGroupID, req.ListenerID, kt.Rid)

	return nil
}

// addTCloudZiyanTargetsToTargetGroup 添加RS到目标组
func (svc *clbSvc) addTCloudZiyanTargetsToTargetGroup(kt *kit.Kit, req *protolb.CreateTargetGroupWithRelReq,
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
		logs.Errorf("batch create tcloud ziyan target failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	logs.Infof("successfully added %d targets to target group, target_group_id: %s, target_ids: %v, rid: %s",
		len(result.IDs), targetGroupID, result.IDs, kt.Rid)

	return nil
}

// updateTCloudZiyanUrlRuleTargetGroup 更新URL规则的目标组ID
func (svc *clbSvc) updateTCloudZiyanUrlRuleTargetGroup(kt *kit.Kit, ruleID, targetGroupID string) error {
	updateReq := &dataproto.TCloudUrlRuleBatchUpdateReq{
		UrlRules: []*dataproto.TCloudUrlRuleUpdate{
			{
				ID:                 ruleID,
				TargetGroupID:      targetGroupID,
				CloudTargetGroupID: targetGroupID,
			},
		},
	}

	err := svc.dataCli.TCloudZiyan.LoadBalancer.BatchUpdateTCloudUrlRule(kt, updateReq)
	if err != nil {
		logs.Errorf("update url rule target group failed, rule_id: %s, target_group_id: %s, err: %v, rid: %s",
			ruleID, targetGroupID, err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	logs.Infof("successfully updated url rule target group, rule_id: %s, target_group_id: %s, rid: %s",
		ruleID, targetGroupID, kt.Rid)

	return nil
}
