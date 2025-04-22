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

package securitygroup

import (
	proto "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

func (svc *securityGroupSvc) updateTCloudZiyanSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo,
	id string) (interface{}, error) {

	req := new(proto.TCloudSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.SecurityGroupRuleAuditResType, sgBaseInfo.ID, id, updateFields)
	if err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.TCloudSGRuleUpdateReq{
		Protocol:                   req.Protocol,
		Port:                       req.Port,
		CloudServiceID:             req.CloudServiceID,
		CloudServiceGroupID:        req.CloudServiceGroupID,
		IPv4Cidr:                   req.IPv4Cidr,
		IPv6Cidr:                   req.IPv6Cidr,
		CloudAddressID:             req.CloudAddressID,
		CloudAddressGroupID:        req.CloudAddressGroupID,
		CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
		Action:                     req.Action,
		Memo:                       req.Memo,
	}
	if err := svc.client.HCService().TCloudZiyan.SecurityGroup.UpdateSecurityGroupRule(cts.Kit,
		sgBaseInfo.ID, id, updateReq); err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) batchUpdateZiyanSGRule(cts *rest.Contexts,
	sgBaseInfo *types.CloudResourceBasicInfo) (any, error) {

	req := new(proto.TCloudSGRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateReq := &hcproto.TCloudSGRuleBatchUpdateReq{
		AccountID:      sgBaseInfo.AccountID,
		EgressRuleSet:  req.EgressRuleSet,
		IngressRuleSet: req.IngressRuleSet,
	}
	err := svc.client.HCService().TCloudZiyan.SecurityGroup.BatchUpdateSecurityGroupRule(cts.Kit, sgBaseInfo.ID,
		updateReq)
	if err != nil {
		logs.Errorf("update ziyan security group rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
