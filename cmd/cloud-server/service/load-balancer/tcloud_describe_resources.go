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

	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudDescribeResources ...
func (svc *lbSvc) TCloudDescribeResources(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDescribeResourcesOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Find,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("describe resources auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	account, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType,
		req.AccountID)
	if err != nil {
		// 这里校验账号是否存在，出现错误大概率是账号不存在
		logs.V(3).Errorf("fail to get account info, err: %s, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}
	switch account.Vendor {
	case enumor.TCloud:
		return svc.client.HCService().TCloud.Clb.DescribeResources(cts.Kit, req)
	case enumor.TCloudZiyan:
		return svc.client.HCService().TCloudZiyan.Clb.DescribeResources(cts.Kit, req)
	default:
		return nil, fmt.Errorf("unsupport vendor %s", account.Vendor)
	}
}

// TCloudDescribeExclusiveCluster ... TODO 支持到公有云
func (svc *lbSvc) TCloudDescribeExclusiveCluster(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDescribeExclusiveClusterReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Find,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("describe resources auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	account, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType,
		req.AccountID)
	if err != nil {
		return nil, err
	}
	switch account.Vendor {
	case enumor.TCloudZiyan:
		return svc.client.HCService().TCloudZiyan.Clb.DescribeExclusiveCluster(cts.Kit, req)
	default:
		return nil, fmt.Errorf("unsupport vendor %s", account.Vendor)
	}
}

// TCloudDescribeClusterResources 查询负载均衡集群中资源列表
func (svc *lbSvc) TCloudDescribeClusterResources(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDescribeClusterResourcesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.Account,
			Action:     meta.Find,
			ResourceID: req.AccountID,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("describe cluster resources failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	account, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		return nil, err
	}

	switch account.Vendor {
	case enumor.TCloudZiyan:
		return svc.client.HCService().TCloudZiyan.Clb.DescribeClusterResources(cts.Kit, req)
	default:
		return nil, fmt.Errorf("unsupport vendor %s", account.Vendor)
	}
}
