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

	genFunc, ok := genResourceFuncMap[a.Basic.Type]
	if !ok {
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm auth type: %s", a.Basic.Type)
	}
	return genFunc(a)
}

/**
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
case meta.AwsSavingsPlansCost:
		return genAwsSavingsPlansCostResource(a)
	case meta.RollingServerManage: // 平台管理-滚服管理
		return sys.RollingServerManage, make([]client.Resource, 0), nil
	case meta.GreenChannel: // 平台管理-小额绿通
		return sys.GreenChannel, make([]client.Resource, 0), nil
case meta.ResPlan:
		return genResPlanResource(a)
*/

type genResourceFunc func(*meta.ResourceAttribute) (client.ActionID, []client.Resource, error)

var genResourceFuncMap = map[meta.ResourceType]genResourceFunc{
	meta.Biz:                      genBizResource,
	meta.Account:                  genAccountResource,
	meta.SubAccount:               genSubAccountResource,
	meta.Vpc:                      genVpcResource,
	meta.Subnet:                   genSubnetResource,
	meta.Disk:                     genDiskResource,
	meta.SecurityGroup:            genSecurityGroupResource,
	meta.SecurityGroupRule:        genSecurityGroupRuleResource,
	meta.GcpFirewallRule:          genGcpFirewallRuleResource,
	meta.RouteTable:               genRouteTableResource,
	meta.Route:                    genRouteResource,
	meta.RecycleBin:               genRecycleBinResource,
	meta.Audit:                    genAuditResource,
	meta.Cvm:                      genCvmResource,
	meta.NetworkInterface:         genNetworkInterfaceResource,
	meta.Eip:                      genEipResource,
	meta.CloudResource:            genCloudResResource,
	meta.Quota:                    genProxyResourceFind,
	meta.InstanceType:             genProxyResourceFind,
	meta.CostManage:               genCostManageResource,
	meta.BizCollection:            genBizCollectionResource,
	meta.CloudSelectionScheme:     genCloudSelectionSchemeResource,
	meta.CloudSelectionIdc:        genCloudSelectionResource,
	meta.CloudSelectionBizType:    genCloudSelectionResource,
	meta.CloudSelectionDataSource: genCloudSelectionResource,
	meta.ArgumentTemplate:         genArgumentTemplateResource,
	meta.Cert:                     genCertResource,
	meta.LoadBalancer:             genLoadBalancerResource,
	meta.Listener:                 genListenerResource,
	meta.TargetGroup:              genTargetGroupResource,
	meta.UrlRuleAuditResType:      genUrlRuleResource,
	meta.MainAccount:              genMainAccountRuleResource,
	meta.RootAccount:              genRootAccountRuleResource,
	meta.AccountBill:              genAccountBillResource,
	meta.Application:              genApplicationResources,
	meta.AccountBillThirdParty:    genAccountBillThirdPartyResource,
	meta.Image:                    genImageResource,
	meta.TaskManagement:           genTaskManagementResource,
	meta.CosBucket:                genCosBucket,
}

func genApplicationResources(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find, meta.Delete, meta.Update:
		return sys.ApplicationManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}
