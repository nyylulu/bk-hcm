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
	typelb "hcm/pkg/adaptor/types/load-balancer"
	adptsg "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateTCloudZiyanSecurityGroup create tcloud ziyan security group.
func (g *securityGroup) CreateTCloudZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudCreateOption{
		Region:      req.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud ziyan security group failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{{
			CloudID:   *sg.SecurityGroupId,
			BkBizID:   req.BkBizID,
			Region:    req.Region,
			Name:      *sg.SecurityGroupName,
			Memo:      sg.SecurityGroupDesc,
			AccountID: req.AccountID,
			Extension: &corecloud.TCloudSecurityGroupExtension{
				CloudProjectID: sg.ProjectId,
			},
		}},
	}
	result, err := g.dataCli.TCloudZiyan.SecurityGroup.BatchCreateSecurityGroup(cts.Kit, createReq)
	if err != nil {

		berr := errf.GetBPassApprovalErrorf(err)
		if berr != nil {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, req.AccountID, req.BkBizID,
				enumor.CreateSecurityGroup, opt, berr)
		}

		logs.Errorf("request dataservice to create tcloud ziyan security group failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteTCloudZiyanSecurityGroup delete tcloud ziyan security group.
func (g *securityGroup) DeleteTCloudZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.TCloudZiyan.SecurityGroup.GetSecurityGroup(cts.Kit, id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudDeleteOption{
		Region:  sg.Region,
		CloudID: sg.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		berr := errf.GetBPassApprovalErrorf(err)
		if berr != nil {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli, sg.AccountID, sg.BkBizID,
				enumor.DeleteSecurityGroup, opt, berr)
		}

		logs.Errorf("request adaptor to delete tcloud ziyan security group failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateTCloudZiyanSecurityGroup update tcloud ziyan security group.
func (g *securityGroup) UpdateTCloudZiyanSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.SecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.TCloudZiyan.SecurityGroup.GetSecurityGroup(cts.Kit, id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudUpdateOption{
		CloudID:     sg.CloudID,
		Region:      sg.Region,
		Name:        req.Name,
		Description: req.Memo,
	}
	if err := client.UpdateSecurityGroup(cts.Kit, opt); err != nil {

		berr := errf.GetBPassApprovalErrorf(err)
		if berr != nil {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				sg.AccountID, sg.BkBizID, enumor.UpdateSecurityGroup, opt, berr)
		}

		logs.Errorf("request adaptor to UpdateSecurityGroup failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{{
			ID:   sg.ID,
			Name: req.Name,
			Memo: req.Memo,
		}},
	}
	if err := g.dataCli.TCloudZiyan.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit, updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s",
			err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TZiyanSGBatchAssociateCloudCvm 根据cvm云id 绑定安全组
func (g *securityGroup) TZiyanSGBatchAssociateCloudCvm(cts *rest.Contexts) (any, error) {

	req := new(proto.SecurityGroupAssociateCloudCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.getSecurityGroupByID(cts, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, req.SecurityGroupID, cts.Kit.Rid)
		return nil, err
	}
	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          req.CloudCvmIDs,
	}
	if err = client.SecurityGroupCvmBatchAssociate(cts.Kit, opt); err != nil {
		berr := errf.GetBPassApprovalErrorf(err)
		if berr != nil {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				sg.AccountID, sg.BkBizID, enumor.AssociateSecurityGroup, opt, berr)
		}
		logs.Errorf("request adaptor to tcloud ziyan security group associate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TZiyanSGBatchDisassociateCloudCvm  根据cvm云id 解绑安全组
func (g *securityGroup) TZiyanSGBatchDisassociateCloudCvm(cts *rest.Contexts) (any, error) {
	req := new(proto.SecurityGroupAssociateCloudCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.getSecurityGroupByID(cts, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud ziyan security group failed, err: %v, id: %s, rid: %s",
			err, req.SecurityGroupID, cts.Kit.Rid)
		return nil, err
	}
	client, err := g.ad.TCloudZiyan(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adptsg.TCloudBatchAssociateCvmOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		CloudCvmIDs:          req.CloudCvmIDs,
	}
	if err = client.SecurityGroupCvmBatchDisassociate(cts.Kit, opt); err != nil {

		berr := errf.GetBPassApprovalErrorf(err)
		if berr != nil {
			return nil, parseAndSaveBPaasApplication(cts.Kit, g.dataCli,
				sg.AccountID, sg.BkBizID, enumor.DisassociateSecurityGroup, opt, berr)
		}
		logs.Errorf("request adaptor to tcloud ziyan security group disassociate cvm failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// TCloudZiyanSecurityGroupAssociateLoadBalancer ...
func (g *securityGroup) TCloudZiyanSecurityGroupAssociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	req := new(hclb.TCloudSetLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据LbID查询负载均衡基本信息
	lbInfo, sgComList, err := g.getLoadBalancerInfoAndSGComRels(cts.Kit, req.LbID)
	if err != nil {
		return nil, err
	}

	sgCloudIDs, sgComReq, err := g.getUpsertSGIDsParams(cts.Kit, req, sgComList)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudSetClbSecurityGroupOption{
		Region:         lbInfo.Region,
		LoadBalancerID: lbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetLoadBalancerSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group associate lb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	if err = g.dataCli.Global.SGCommonRel.BatchUpsert(cts.Kit, sgComReq); err != nil {
		logs.Errorf("request dataservice upsert security group lb rels failed, err: %v, req: %+v, rid: %s",
			err, sgComReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// TCloudZiyanSecurityGroupDisassociateLoadBalancer ...
func (g *securityGroup) TCloudZiyanSecurityGroupDisassociateLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	req := new(hclb.TCloudDisAssociateLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据LbID查询负载均衡基本信息
	lbInfo, sgComList, err := g.getLoadBalancerInfoAndSGComRels(cts.Kit, req.LbID)
	if err != nil {
		return nil, err
	}

	allSGIDs := make([]string, 0)
	existSG := false
	for _, rel := range sgComList.Details {
		if rel.SecurityGroupID == req.SecurityGroupID {
			existSG = true
		}
		allSGIDs = append(allSGIDs, rel.SecurityGroupID)
	}
	if !existSG {
		return nil, errf.Newf(errf.RecordNotFound, "not found sg id: %s", req.SecurityGroupID)
	}

	sgMap, err := g.getSecurityGroupMap(cts.Kit, allSGIDs)
	if err != nil {
		return nil, err
	}

	// 安全组的云端ID数组
	sgCloudIDs := make([]string, 0)
	for _, sgID := range allSGIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			continue
		}
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
	}

	client, err := g.ad.TCloudZiyan(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudSetClbSecurityGroupOption{
		Region:         lbInfo.Region,
		LoadBalancerID: lbInfo.CloudID,
		SecurityGroups: sgCloudIDs,
	}
	if _, err = client.SetLoadBalancerSecurityGroups(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to tcloud security group disAssociate lb failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := buildSGCommonRelDeleteReq(
		enumor.TCloud, req.LbID, []string{req.SecurityGroupID}, enumor.LoadBalancerCloudResType)
	if err = g.dataCli.Global.SGCommonRel.BatchDelete(cts.Kit, deleteReq); err != nil {
		logs.Errorf("request dataservice tcloud delete security group lb rels failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
