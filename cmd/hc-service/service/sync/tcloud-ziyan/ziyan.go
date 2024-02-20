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

// Package ziyan ...
package ziyan

import (
	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// InitService initial tcloud sync service
func InitService(cap *capability.Capability) {
	v := &service{
		ad:      cap.CloudAdaptor,
		cs:      cap.ClientSet,
		dataCli: cap.ClientSet.DataService(),
		syncCli: cap.ResSyncCli,
	}

	h := rest.NewHandler()
	h.Path("/vendors/tcloud-ziyan")

	h.Add("SyncSecurityGroup", "POST", "/security_groups/sync", v.SyncSecurityGroup)
	h.Add("SyncZone", "POST", "/zones/sync", v.SyncZone)
	h.Add("SyncRegion", "POST", "/regions/sync", v.SyncRegion)
	h.Add("SyncArgsTpl", "POST", "/argument_templates/sync", v.SyncArgsTpl)

	h.Load(cap.WebService)
}

type service struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	dataCli *dataservice.Client
	syncCli ressync.Interface
}

func defaultPrepare(cts *rest.Contexts, cli ressync.Interface) (*sync.TCloudSyncReq, ziyan.Interface, error) {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := cli.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, nil, err
	}

	return req, syncCli, nil
}
