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

package securitygroup

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	adptsgrule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// BatchCreateTCloudZiyanSGRule 批量创建自研云安全组规则
// 腾讯云安全组规则索引是一个动态的，所以每次创建需要将云上安全组规则计算一遍。
func (g *securityGroup) BatchCreateTCloudZiyanSGRule(cts *rest.Contexts) (any, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.TCloudSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.TCloudZiyan.SecurityGroup.GetSecurityGroup(cts.Kit, sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, sgID, cts.Kit.Rid)
		return nil, err
	}

	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &ziyan.SyncBaseParams{AccountID: sg.AccountID, Region: sg.Region, CloudIDs: []string{sg.ID}}
	opt := &adptsgrule.TCloudCreateOption{Region: sg.Region, CloudSecurityGroupID: sg.CloudID}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = slice.Map(req.EgressRuleSet, convertToTCloudAdaptorSGRuleCreate)
	}
	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = slice.Map(req.IngressRuleSet, convertToTCloudAdaptorSGRuleCreate)
	}

	if err := client.CreateSecurityGroupRule(cts.Kit, opt); err != nil {
		bpaasSN := errf.GetBPassSNFromErr(err)
		if len(bpaasSN) > 0 {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				sg.AccountID, sg.BkBizID, enumor.CreateSecurityGroupRule, opt, bpaasSN)
		}

		logs.Errorf("request adaptor to create tcloud ziyan security group rule failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)

		// 函数中有日志
		_, _ = g.syncZiyanSGRule(cts.Kit, syncParam)
		return nil, err
	}

	createdIds, syncErr := g.syncZiyanSGRule(cts.Kit, syncParam)
	if syncErr != nil {
		return nil, syncErr
	}
	return &core.BatchCreateResult{IDs: createdIds}, nil
}

// UpdateTCloudZiyanSGRule update tcloud ziyan security group rule.
func (g *securityGroup) UpdateTCloudZiyanSGRule(cts *rest.Contexts) (any, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcservice.TCloudSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rule, err := g.getTCloudZiyanSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &ziyan.SyncBaseParams{AccountID: rule.AccountID, Region: rule.Region,
		CloudIDs: []string{rule.SecurityGroupID},
	}
	opt := &adptsgrule.TCloudUpdateOption{
		Region: rule.Region, CloudSecurityGroupID: rule.CloudSecurityGroupID, Version: rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleSet = []adptsgrule.TCloudUpdateSpec{
			convertToTCloudSGRuleUpdate(rule.CloudPolicyIndex, req)}
	case enumor.Ingress:
		opt.IngressRuleSet = []adptsgrule.TCloudUpdateSpec{
			convertToTCloudSGRuleUpdate(rule.CloudPolicyIndex, req)}
	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}

	if err := client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		bpaasSN := errf.GetBPassSNFromErr(err)
		if len(bpaasSN) > 0 {
			sg, err := g.getSecurityGroupByID(cts, sgID)
			if err != nil {
				logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
					err, sgID, cts.Kit.Rid)
				return nil, err
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				rule.AccountID, sg.BkBizID, enumor.UpdateSecurityGroupRule, opt, bpaasSN)
		}

		logs.Errorf("request adaptor to update tcloud ziyan security group rule failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)

		_, _ = g.syncZiyanSGRule(cts.Kit, syncParam)
		return nil, err
	}

	if _, syncErr := g.syncZiyanSGRule(cts.Kit, syncParam); syncErr != nil {
		return nil, syncErr
	}
	return nil, nil
}

func (g *securityGroup) getTCloudZiyanSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := g.dataCli.TCloudZiyan.SecurityGroup.ListSecurityGroupRule(cts.Kit, listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// DeleteTCloudZiyanSGRule delete tcloud ziyan security group rule.
func (g *securityGroup) DeleteTCloudZiyanSGRule(cts *rest.Contexts) (any, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := g.getTCloudZiyanSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &ziyan.SyncBaseParams{
		AccountID: rule.AccountID,
		Region:    rule.Region,
		CloudIDs:  []string{rule.SecurityGroupID},
	}
	opt := &adptsgrule.TCloudDeleteOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		Version:              rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleIndexes = []int64{rule.CloudPolicyIndex}

	case enumor.Ingress:
		opt.IngressRuleIndexes = []int64{rule.CloudPolicyIndex}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {

		bpaasSN := errf.GetBPassSNFromErr(err)
		if len(bpaasSN) > 0 {
			sg, err := g.getSecurityGroupByID(cts, sgID)
			if err != nil {
				logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
					err, sgID, cts.Kit.Rid)
				return nil, err
			}
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				rule.AccountID, sg.BkBizID, enumor.DeleteSecurityGroupRule, opt, bpaasSN)
		}

		logs.Errorf("request adaptor to delete tcloud ziyan security group rule failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)

		// 同步函数中带有日志
		_, _ = g.syncZiyanSGRule(cts.Kit, syncParam)
		return nil, err
	}

	if _, syncErr := g.syncZiyanSGRule(cts.Kit, syncParam); syncErr != nil {
		return nil, syncErr
	}

	return nil, nil
}

// syncZiyanSGRule 调用同步客户端同步云上规则，返回新增的id
func (g *securityGroup) syncZiyanSGRule(kt *kit.Kit, syncParams *ziyan.SyncBaseParams) ([]string, error) {

	syncCli, err := g.syncCli.TCloudZiyan(kt, syncParams.AccountID)
	if err != nil {
		return nil, err
	}

	syncResult, syncErr := syncCli.SecurityGroupRule(kt, syncParams, new(ziyan.SyncSGRuleOption))
	if syncErr != nil {
		logs.Errorf("sync tcloud ziyan security group failed, err: %v, params: %+v, rid: %s", err, syncParams, kt.Rid)
		return nil, syncErr
	}
	return syncResult.CreatedIds, nil
}

func (g *securityGroup) getSecurityGroupByID(cts *rest.Contexts, sgID string) (*corecloud.BaseSecurityGroup, error) {
	sgReq := &protocloud.SecurityGroupListReq{
		Filter: tools.EqualExpression("id", sgID),
		Page:   core.NewDefaultBasePage(),
	}
	sgResult, err := g.dataCli.Global.SecurityGroup.ListSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgReq)
	if err != nil {
		logs.Errorf("request dataservice list security group failed, err: %v, id: %s, rid: %s",
			err, sgID, cts.Kit.Rid)
		return nil, err
	}

	if len(sgResult.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", sgID)
	}
	sg := sgResult.Details[0]
	return &sg, nil
}

func convertToTCloudAdaptorSGRuleCreate(rule hcservice.TCloudSGRuleCreate) adptsgrule.TCloud {
	return adptsgrule.TCloud{
		Protocol:                   rule.Protocol,
		Port:                       rule.Port,
		IPv4Cidr:                   rule.IPv4Cidr,
		IPv6Cidr:                   rule.IPv6Cidr,
		CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
		Action:                     rule.Action,
		Description:                rule.Memo,
		CloudServiceID:             rule.CloudServiceID,
		CloudServiceGroupID:        rule.CloudServiceGroupID,
		CloudAddressID:             rule.CloudAddressID,
		CloudAddressGroupID:        rule.CloudAddressGroupID,
	}
}

func convertToTCloudSGRuleUpdate(ruleIdx int64, req *hcservice.TCloudSGRuleUpdateReq) adptsgrule.TCloudUpdateSpec {

	return adptsgrule.TCloudUpdateSpec{
		CloudPolicyIndex:           ruleIdx,
		Protocol:                   req.Protocol,
		Port:                       req.Port,
		IPv4Cidr:                   req.IPv4Cidr,
		IPv6Cidr:                   req.IPv6Cidr,
		CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
		Action:                     req.Action,
		Description:                req.Memo,
		CloudServiceID:             req.CloudServiceID,
		CloudServiceGroupID:        req.CloudServiceGroupID,
		CloudAddressID:             req.CloudAddressID,
		CloudAddressGroupID:        req.CloudAddressGroupID,
	}
}
