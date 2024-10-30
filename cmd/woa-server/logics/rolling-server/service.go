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

// Package rollingserver ...
package rollingserver

import (
	"hcm/cmd/woa-server/logics/config"
	rolling_server "hcm/cmd/woa-server/types/rolling-server"
	types "hcm/cmd/woa-server/types/task"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/esb"
)

// Logics provides management interface for rolling server.
type Logics interface {
	SyncBills(kt *kit.Kit, req *rolling_server.RollingBillSyncReq) error
	// GetCpuCoreSummary 查询滚服已交付、已退还的CPU核心数概览信息
	GetCpuCoreSummary(kt *kit.Kit, req *rolling_server.CpuCoreSummaryReq) (*rsproto.RollingCpuCoreSummaryItem, error)
	IsResPoolBiz(kt *kit.Kit, bizID int64) (bool, error)
	CanApplyHost(kt *kit.Kit, bizID int64, appliedCount uint, appliedType enumor.AppliedType) (bool, string, error)
	CreateAppliedRecord(kt *kit.Kit, createArr []rolling_server.CreateAppliedRecordData) error
	UpdateSubOrderRollingDeliveredCore(kt *kit.Kit, bizID int64, subOrderID string, appliedTypes []enumor.AppliedType,
		deviceTypeCountMap map[string]int) error
	ReduceRollingCvmProdAppliedRecord(kt *kit.Kit, devices []*types.MatchDeviceBrief) error
	GetCpuCoreSum(kt *kit.Kit, deviceTypeCountMap map[string]int) (uint64, error)
}

// logics rolling server logics.
type logics struct {
	sd           serviced.State
	client       *client.ClientSet
	esbClient    esb.Client
	configLogics config.Logics
}

// New creates rolling server logics instance.
func New(sd serviced.State, client *client.ClientSet, esbClient esb.Client, thirdCli *thirdparty.Client) (Logics,
	error) {
	rsLogics := &logics{
		sd:           sd,
		client:       client,
		esbClient:    esbClient,
		configLogics: config.New(thirdCli),
	}

	if cc.WoaServer().RollingServer.SyncBill {
		go rsLogics.syncBillsPeriodically()
	}

	return rsLogics, nil
}
