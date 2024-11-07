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

package auth

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sys"
)

// AdaptAuthOptions convert hcm auth resource to iam action id and resources
func AdaptAuthOptions(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a == nil {
		return "", nil, errf.New(errf.InvalidParameter, "resource attribute is not set")
	}

	// skip actions do not need to relate to resources
	if a.Basic.Action == meta.SkipAction {
		return genSkipResource(a)
	}

	switch a.Basic.Type {
	case meta.Biz:
		return genBizResource(a)
	case meta.Account:
		return genAccountResource(a)
	case meta.SubAccount:
		return genSubAccountResource(a)
	case meta.Vpc:
		return genVpcResource(a)
	case meta.Subnet:
		return genSubnetResource(a)
	case meta.Disk:
		return genDiskResource(a)
	case meta.SecurityGroup:
		return genSecurityGroupResource(a)
	case meta.SecurityGroupRule:
		return genSecurityGroupRuleResource(a)
	case meta.GcpFirewallRule:
		return genGcpFirewallRuleResource(a)
	case meta.RouteTable:
		return genRouteTableResource(a)
	case meta.Route:
		return genRouteResource(a)
	case meta.RecycleBin:
		return genRecycleBinResource(a)
	case meta.Audit:
		return genAuditResource(a)
	case meta.ResPlan:
		return genResPlanResource(a)
	case meta.Cvm:
		return genCvmResource(a)
	case meta.NetworkInterface:
		return genNetworkInterfaceResource(a)
	case meta.Eip:
		return genEipResource(a)
	case meta.CloudResource:
		return genCloudResResource(a)
	case meta.Quota:
		return genProxyResourceFind(a)
	case meta.InstanceType:
		return genProxyResourceFind(a)
	case meta.CostManage:
		return genCostManageResource(a)
	case meta.BizCollection:
		return genBizCollectionResource(a)
	case meta.CloudSelectionScheme:
		return genCloudSelectionSchemeResource(a)
	case meta.CloudSelectionIdc:
		return sys.CloudSelectionRecommend, make([]client.Resource, 0), nil
	case meta.CloudSelectionBizType:
		return sys.CloudSelectionRecommend, make([]client.Resource, 0), nil
	case meta.CloudSelectionDataSource:
		return sys.CloudSelectionRecommend, make([]client.Resource, 0), nil
	case meta.ArgumentTemplate:
		return genArgumentTemplateResource(a)
	case meta.Cert:
		return genCertResource(a)
	case meta.LoadBalancer:
		return genLoadBalancerResource(a)
	case meta.Listener:
		return genListenerResource(a)
	case meta.TargetGroup:
		return genTargetGroupResource(a)
	case meta.UrlRuleAuditResType:
		return genUrlRuleResource(a)
	case meta.MainAccount:
		return genMainAccountRuleResource(a)
	case meta.RootAccount:
		return genRootAccountRuleResource(a)
	case meta.ServiceResDissolve: // 服务请求-服务-机房裁撤-菜单粒度
		return sys.ServiceResDissolve, make([]client.Resource, 0), nil
	case meta.ZiyanCvmType: // CVM机型-菜单粒度
		return sys.ZiyanCvmType, make([]client.Resource, 0), nil
	case meta.ZiyanCvmSubnet: // CVM子网-菜单粒度
		return sys.ZiyanCvmSubnet, make([]client.Resource, 0), nil
	case meta.ZiyanResShelves: // 资源上下架-菜单粒度
		return sys.ZiyanResShelves, make([]client.Resource, 0), nil
	case meta.ZiyanCvmCreate: // CVM生产-菜单粒度
		return sys.ZiyanCvmCreate, make([]client.Resource, 0), nil
	case meta.ZiyanResDissolveManage: // 机房裁撤管理-菜单粒度
		return sys.ZiyanResDissolveManage, make([]client.Resource, 0), nil
	case meta.ZiyanResInventory: // 主机库存-菜单粒度
		return sys.ZiyanResInventory, make([]client.Resource, 0), nil
	case meta.ZiYanResource: // 自研云资源的操作-业务粒度
		return genZiYanResource(a)
	case meta.ZiYanResPlan:
		return genZiYanResPlanResource(a)
	case meta.AccountBill:
		return genAccountBillResource(a)
	case meta.Application:
		return genApplicationResources(a)
	case meta.AccountBillThirdParty:
		return genAccountBillThirdPartyResource(a)
	case meta.AwsSavingsPlansCost:
		return genAwsSavingsPlansCostResource(a)
	case meta.RollingServerManage: // 平台管理-滚服管理
		return sys.RollingServerManage, make([]client.Resource, 0), nil
	case meta.TaskManagement:
		return genTaskManagementResource(a)
	case meta.Image:
		return genImageResource(a)
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm auth type: %s", a.Basic.Type)
	}
}

func genApplicationResources(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find, meta.Delete, meta.Update:
		return sys.ApplicationManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}
