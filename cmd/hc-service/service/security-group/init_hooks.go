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

import "hcm/pkg/rest"

func initSecurityGroupServiceHooks(svc *securityGroup, h *rest.Handler) {

	h.Add("CreateTCloudZiyanSecurityGroup", "POST",
		"/vendors/tcloud-ziyan/security_groups/create", svc.CreateTCloudZiyanSecurityGroup)
	h.Add("DeleteTCloudZiyanSecurityGroup", "DELETE",
		"/vendors/tcloud-ziyan/security_groups/{id}", svc.DeleteTCloudZiyanSecurityGroup)
	h.Add("UpdateTCloudZiyanSecurityGroup", "PATCH",
		"/vendors/tcloud-ziyan/security_groups/{id}", svc.UpdateTCloudZiyanSecurityGroup)
	h.Add("BatchCreateTCloudZiyanSGRule", "POST",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/batch/create",
		svc.BatchCreateTCloudZiyanSGRule)
	h.Add("UpdateTCloudZiyanSGRule", "PUT",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/{id}", svc.UpdateTCloudZiyanSGRule)
	h.Add("DeleteTCloudZiyanSGRule", "DELETE",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/{id}", svc.DeleteTCloudZiyanSGRule)

	h.Add("BatchUpdateZiyanSGRule", "PUT",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/batch/update", svc.BatchUpdateZiyanSGRule)

	h.Add("TZiyanSGBatchAssociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/cvms/batch", svc.TZiyanSGBatchAssociateCvm)
	h.Add("TZiyanSGBatchDisassociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/cvms/batch", svc.TZiyanSGBatchDisassociateCvm)

	h.Add("TCloudZiyanSecurityGroupAssociateLoadBalancer", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/load_balancers",
		svc.TCloudZiyanSecurityGroupAssociateLoadBalancer)
	h.Add("TCloudZiyanSecurityGroupDisassociateLoadBalancer", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/load_balancers",
		svc.TCloudZiyanSecurityGroupDisassociateLoadBalancer)
	h.Add("TCloudZiyanListSecurityGroupStatistic", "POST", "/vendors/tcloud-ziyan/security_groups/statistic",
		svc.TCloudZiyanListSecurityGroupStatistic)
	h.Add("TCloudZiyanCloneSecurityGroup", "POST", "/vendors/tcloud-ziyan/security_groups/clone",
		svc.TCloudZiyanCloneSecurityGroup)
}
