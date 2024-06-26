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

package ziyan

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudSecurityGroupClient create a new security group api client.
func NewCloudSecurityGroupClient(client rest.ClientInterface) *SecurityGroupClient {
	return &SecurityGroupClient{
		client: client,
	}
}

// SecurityGroupClient is data service security group api client.
type SecurityGroupClient struct {
	client rest.ClientInterface
}

// BatchCreateSecurityGroup batch create security group rule.
func (cli *SecurityGroupClient) BatchCreateSecurityGroup(kt *kit.Kit, request *protocloud.
	SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]) (*core.BatchCreateResult, error) {

	return common.Request[protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension], core.BatchCreateResult](
		cli.client, rest.POST, kt, request, "/security_groups/batch/create")

}

// BatchUpdateSecurityGroup batch update security group.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroup(kt *kit.Kit,
	request *protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]) error {

	return common.RequestNoResp[protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]](
		cli.client, rest.PATCH, kt, request, "/security_groups/batch/update")

}

// GetSecurityGroup get security group.
func (cli *SecurityGroupClient) GetSecurityGroup(kt *kit.Kit, id string) (
	*corecloud.SecurityGroup[corecloud.TCloudSecurityGroupExtension], error) {

	return common.Request[common.Empty, corecloud.SecurityGroup[corecloud.TCloudSecurityGroupExtension]](
		cli.client, rest.GET, kt, nil, "/security_groups/%s", id)
}

// ListSecurityGroupExt list security group with extension.
func (cli *SecurityGroupClient) ListSecurityGroupExt(kt *kit.Kit, req *core.ListReq) (
	*protocloud.SecurityGroupExtListResult[corecloud.TCloudSecurityGroupExtension], error) {

	return common.Request[core.ListReq, protocloud.SecurityGroupExtListResult[corecloud.TCloudSecurityGroupExtension]](
		cli.client, rest.POST, kt, req, "/security_groups/list")
}

// BatchCreateSecurityGroupRule batch create security group rule.
func (cli *SecurityGroupClient) BatchCreateSecurityGroupRule(kt *kit.Kit,
	request *protocloud.TCloudSGRuleCreateReq, sgID string) (*core.BatchCreateResult, error) {

	return common.Request[protocloud.TCloudSGRuleCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, request, "/security_groups/%s/rules/batch/create", sgID)
}

// BatchUpdateSecurityGroupRule update security group rule.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroupRule(kt *kit.Kit, request *protocloud.
	TCloudSGRuleBatchUpdateReq, sgID string) error {

	return common.RequestNoResp[protocloud.TCloudSGRuleBatchUpdateReq](
		cli.client, rest.PUT, kt, request, "/security_groups/%s/rules/batch", sgID)

}

// ListSecurityGroupRule list security group rule.
func (cli *SecurityGroupClient) ListSecurityGroupRule(kt *kit.Kit,
	request *protocloud.TCloudSGRuleListReq, sgID string) (*protocloud.TCloudSGRuleListResult, error) {

	return common.Request[protocloud.TCloudSGRuleListReq, protocloud.TCloudSGRuleListResult](
		cli.client, rest.POST, kt, request, "/security_groups/%s/rules/list", sgID)
}

// BatchDeleteSecurityGroupRule delete security group rule.
func (cli *SecurityGroupClient) BatchDeleteSecurityGroupRule(kt *kit.Kit, request *protocloud.
	TCloudSGRuleBatchDeleteReq, sgID string) error {

	return common.RequestNoResp[protocloud.TCloudSGRuleBatchDeleteReq](
		cli.client, rest.DELETE, kt, request, "/security_groups/%s/rules/batch", sgID)
}
