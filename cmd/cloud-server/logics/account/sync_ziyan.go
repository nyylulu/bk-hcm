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

package account

import (
	tziyan "hcm/cmd/cloud-server/service/sync/tcloud-ziyan"
	"hcm/pkg/api/core"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
)

func init() {
	availableVendorSyncer = append(availableVendorSyncer, newTCloudZiyanSyncer())
	vendorSyncerMap[enumor.TCloudZiyan] = newTCloudZiyanSyncer()
}

func newTCloudZiyanSyncer() tcloudZiyanSyncer {
	return tcloudZiyanSyncer{generalSyncer{vendor: enumor.TCloudZiyan}}
}

// tcloudSyncer ...
type tcloudZiyanSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t tcloudZiyanSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.TCloudZiyan.Region.ListRegion(kt, req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// SyncAllResource ...
func (t tcloudZiyanSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &tziyan.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return tziyan.SyncAllResource(kt, cli, opt)
}
