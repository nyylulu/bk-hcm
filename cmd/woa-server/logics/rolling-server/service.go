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
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/config"
	rolling_server "hcm/cmd/woa-server/types/rolling-server"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/esb"
)

// Logics provides management interface for rolling server.
type Logics interface {
	GetGlobalQuotaConfig(kt *kit.Kit) (*rstable.RollingGlobalConfigTable, error)
	ListQuotaOffsetAdjustRecords(kt *kit.Kit, offsetConfigIDs []string, page *core.BasePage) (
		*rolling_server.ListQuotaOffsetsAdjustRecordsResp, error)
	ListBizsWithExistQuota(kt *kit.Kit, req *rolling_server.ListBizsWithExistQuotaReq) (
		*rolling_server.ListBizsWithExistQuotaResp, error)
	// ListBizQuotaConfigs 获取批量业务的额度信息列表，管理员使用
	ListBizQuotaConfigs(kt *kit.Kit, req *rolling_server.ListBizQuotaConfigsReq) (
		*rolling_server.ListBizQuotaConfigsResp, error)
	// ListBizBizQuotaConfigs 获取当前业务自身的额度信息，业务使用
	ListBizBizQuotaConfigs(kt *kit.Kit, bkBizID int64, req *rolling_server.ListBizBizQuotaConfigsReq) (
		*rolling_server.ListBizQuotaConfigsResp, error)
	// CreateBizQuotaConfigs 批量给业务生成当月的滚服基础额度，幂等可重复执行
	CreateBizQuotaConfigs(kt *kit.Kit, req *rolling_server.CreateBizQuotaConfigsReq) (
		*rolling_server.CreateBizQuotaConfigsResp, error)
	// CreateBizQuotaConfigsForAllBiz 批量给所有业务生成当月的滚服基础额度，此时额度从全局配置表获取
	CreateBizQuotaConfigsForAllBiz(kt *kit.Kit, quotaMonth rolling_server.QuotaMonth) (
		*rolling_server.CreateBizQuotaConfigsResp, error)
	AdjustQuotaOffsetConfigs(kt *kit.Kit, bkBizIDs []int64, adjustMonth rolling_server.AdjustMonthRange,
		quotaOffset int64) (*rolling_server.AdjustQuotaOffsetsResp, error)
	// BatchCreateQuotaOffsetConfigAudit 配额修改时创建审计记录
	BatchCreateQuotaOffsetConfigAudit(kt *kit.Kit, effectIDs []string, quotaOffset int64) error

	SyncBills(kt *kit.Kit, req *rolling_server.RollingBillSyncReq) error
	// GetCpuCoreSummary 查询滚服已交付、已退还的CPU核心数概览信息
	GetCpuCoreSummary(kt *kit.Kit, req *rolling_server.CpuCoreSummaryReq) (*rsproto.RollingCpuCoreSummaryItem, error)
	IsResPoolBiz(kt *kit.Kit, bizID int64) (bool, error)
	CanApplyHost(kt *kit.Kit, bizID int64, appliedCount uint, appliedType enumor.AppliedType) (bool, string, error)
	CreateAppliedRecord(kt *kit.Kit, createArr []rolling_server.CreateAppliedRecordData) error
	UpdateSubOrderRollingDeliveredCore(kt *kit.Kit, bizID int64, subOrderID string, appliedTypes []enumor.AppliedType,
		deviceTypeCountMap map[string]int) error
	ReduceRollingCvmProdAppliedRecord(kt *kit.Kit, devices []*types.MatchDeviceBrief) error
	GetCpuCoreSum(kt *kit.Kit, deviceTypeCountMap map[string]int) (int64, error)
	// CalSplitRecycleHosts 计算并匹配指定时间范围指定业务的主机Host
	CalSplitRecycleHosts(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost, allBizReturnedCpuCore,
		globalQuota int64) (map[string]*rolling_server.RecycleHostMatchInfo, []*table.RecycleHost, int64, error)
	// InsertReturnedHostMatched 插入需要退还的主机匹配记录
	InsertReturnedHostMatched(kt *kit.Kit, bkBizID int64, orderID uint64, subOrderID string, hosts []*table.RecycleHost,
		hostMatchMap map[string]*rolling_server.RecycleHostMatchInfo, status enumor.ReturnedStatus) error
	// UpdateReturnedStatusBySubOrderID 根据回收子订单ID更新滚服回收的状态
	UpdateReturnedStatusBySubOrderID(kt *kit.Kit, bkBizID int64, subOrderID string,
		updateLocked enumor.ReturnedStatus) error
	// ListReturnedRecordsBySubOrderID 根据回收子订单ID查询滚服回收列表
	ListReturnedRecordsBySubOrderID(kt *kit.Kit, bkBizID int64, subOrderID string) (
		[]*rstable.RollingReturnedRecord, error)
	// GetCurrentMonthAllReturnedCpuCore 获取当月所有业务回收的CPU总核心数
	GetCurrentMonthAllReturnedCpuCore(kt *kit.Kit) (int64, error)
	// GetRollingGlobalQuota 查询系统配置的全局总额度
	GetRollingGlobalQuota(kt *kit.Kit) (int64, error)
	// CheckReturnedStatusBySubOrderID 校验回收订单是否有滚服剩余额度
	CheckReturnedStatusBySubOrderID(kt *kit.Kit, orders []*table.RecycleOrder) error
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

	go rsLogics.createBaseQuotaConfigPeriodically()

	return rsLogics, nil
}
