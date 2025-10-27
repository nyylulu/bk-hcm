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

// Package shortrental ...
package shortrental

import (
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/config"
	demandtime "hcm/cmd/woa-server/logics/plan/demand-time"
	srtypes "hcm/cmd/woa-server/types/short-rental"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/shopspring/decimal"
)

// Logics provides management interface for short rental.
type Logics interface {
	// ListDeviceTypeFamily 获取CVM机型和物理机机型族的映射
	ListDeviceTypeFamily(kt *kit.Kit, deviceTypes []string) (map[string]string, error)
	// ListShortRentalReturnPlan 根据业务查询短租项目本月的退回计划，并将计划退回的主机分组返回
	ListShortRentalReturnPlan(kt *kit.Kit, planProductName string, opProductName string,
		hosts []*table.RecycleHost, deviceToPhysFamilyMap map[string]string) (
		map[srtypes.RecycleGroupKey][]*cvmapi.ReturnPlanItem, map[srtypes.RecycleGroupKey][]*table.RecycleHost, error)
	// ListExecutedPlanCores 获取退回计划的已执行核心数
	ListExecutedPlanCores(kt *kit.Kit, opProductID int64, recycleGroupKeys []srtypes.RecycleGroupKey) (
		map[srtypes.RecycleGroupKey]int64, error)
	// CalSplitRecycleHosts 根据短租退回计划的余量，计算并将host匹配到短租退回计划上
	CalSplitRecycleHosts(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost, recycleTypeSeq []table.RecycleType,
		allReturnedCPUCore, allPlanCPUCore decimal.Decimal) ([]*table.RecycleHost, int64, error)

	// CreateReturnedHostRecord 创建需要退还的主机匹配记录
	CreateReturnedHostRecord(kt *kit.Kit, bkBizID int64, orderID uint64, subOrderID string,
		status enumor.ShortRentalStatus) error
	// UpdateReturnedStatusBySubOrderID 根据回收子订单ID更新回收记录的状态
	UpdateReturnedStatusBySubOrderID(kt *kit.Kit, subOrderID string, updateTo enumor.ShortRentalStatus) error
}

// logics short rental logics.
type logics struct {
	sd           serviced.State
	client       *client.ClientSet
	configLogics config.Logics
	bizLogics    biz.Logics
	cmsiClient   cmsi.Client
	thirdCli     *thirdparty.Client

	demandTime demandtime.DemandTime
}

// New creates short rental logics instance.
func New(sd serviced.State, client *client.ClientSet, thirdCli *thirdparty.Client, bizLogic biz.Logics,
	cmsiCli cmsi.Client, configLogics config.Logics) (Logics, error) {

	rsLogics := &logics{
		sd:           sd,
		client:       client,
		configLogics: configLogics,
		bizLogics:    bizLogic,
		cmsiClient:   cmsiCli,
		thirdCli:     thirdCli,
		demandTime:   demandtime.NewDemandTimeFromTable(client),
	}

	return rsLogics, nil
}
