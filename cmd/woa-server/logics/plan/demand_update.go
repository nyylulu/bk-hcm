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

package plan

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	dmtypes "hcm/pkg/dal/dao/types/meta"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// applyResPlanDemandChange apply res plan demand change.
func (c *Controller) applyResPlanDemandChange(kt *kit.Kit, ticket *TicketInfo) error {
	// call crp api to get plan order change info.
	changeDemandsOri, err := c.QueryCrpOrderChangeInfo(kt, ticket.CrpSn)
	if err != nil {
		logs.Errorf("failed to query crp order change info, err: %v, crp_sn: %s, rid: %s", err, ticket.CrpSn,
			kt.Rid)
		return err
	}
	// 需要先把key相同的预测数据聚合，避免过大的扣减在数据库中不存在
	changeDemands, err := c.aggregateDemandChangeInfo(kt, changeDemandsOri, ticket)
	if err != nil {
		logs.Errorf("failed to aggregate demand change info, err: %v, crp_sn: %s, rid: %s", err, ticket.CrpSn,
			kt.Rid)
		return err
	}

	logs.Infof("aggregate demand change info: %+v, rid: %s", cvt.PtrToSlice(changeDemands), kt.Rid)

	// changeDemand可能会在扣减时模糊匹配到同一个需求，因此需要在扣减操作生效前记录扣减量，最后统一执行
	upsertReq, updatedIDs, createLogReq, updateLogReq, err := c.prepareResPlanDemandChangeReq(kt, changeDemands, ticket)
	if err != nil {
		logs.Errorf("failed to prepare res plan demand change req, err: %v, ticket: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return err
	}

	createdIDs, err := c.BatchUpsertResPlanDemand(kt, upsertReq, updatedIDs)
	if err != nil {
		logs.Errorf("failed to batch upsert res plan demand, err: %v, ticket: %s, rid: %s", err, ticket.ID,
			kt.Rid)
		return err
	}

	// 创建预测的日志
	err = c.CreateResPlanChangelog(kt, createLogReq, createdIDs)
	if err != nil {
		// 变更记录创建失败不阻断整个预测的更新流程（且此时预测数据已变更，上游不应感知到error），因此此处仅记录warning，不返回error
		logs.Warnf("failed to create res plan demand changelog, err: %v, created demands: %v, rid: %s", err,
			createdIDs, kt.Rid)
	}
	// 更新预测的日志
	err = c.CreateResPlanChangelog(kt, updateLogReq, []string{})
	if err != nil {
		// 变更记录创建失败不阻断整个预测的更新流程（且此时预测数据已变更，上游不应感知到error），因此此处仅记录warning，不返回error
		logs.Warnf("failed to create res plan demand changelog, err: %v, updated demands: %v, rid: %s", err,
			updatedIDs, kt.Rid)
	}

	return nil
}

// CreateResPlanChangelog create resource plan demand changelog.
// If demandIDs is provided, will reset the changelog's demand id.
func (c *Controller) CreateResPlanChangelog(kt *kit.Kit, changelogReqs []rpproto.DemandChangelogCreate,
	demandIDs []string) error {
	if len(changelogReqs) == 0 {
		return nil
	}

	if len(demandIDs) > 0 && len(demandIDs) != len(changelogReqs) {
		logs.Errorf("demand ids and changelog create reqs length not equal, demand ids: %v, "+
			"changelog create reqs: %v, rid: %s", demandIDs, changelogReqs, kt.Rid)
		return fmt.Errorf("demand ids and changelog create reqs length not equal")
	}

	createDemandIDs := make([]string, len(changelogReqs))
	for idx := range changelogReqs {
		if len(demandIDs) > 0 {
			changelogReqs[idx].DemandID = demandIDs[idx]
		}
		createDemandIDs[idx] = changelogReqs[idx].DemandID
	}

	changelogReq := &rpproto.DemandChangelogCreateReq{
		Changelogs: changelogReqs,
	}

	_, err := c.client.DataService().Global.ResourcePlan.BatchCreateDemandChangelog(kt, changelogReq)
	if err != nil {
		logs.Errorf("failed to create plan demand changelog, demand ids: %v, err: %v, rid: %s", createDemandIDs,
			err, kt.Rid)
		return err
	}

	return nil
}

// prepareResPlanDemandChangeReq 准备更新资源预测表的参数
// 返回值：更新参数，被更新的资源ID，新增预测的变更日志，修改预测的变更日志
// 因新增预测的变更日志需要在更新完成后追加预测ID，因此需要和修改的变更日志分开返回
func (c *Controller) prepareResPlanDemandChangeReq(kt *kit.Kit, changeDemands []*ptypes.CrpOrderChangeInfo,
	ticket *TicketInfo) (*rpproto.ResPlanDemandBatchUpsertReq, []string, []rpproto.DemandChangelogCreate,
	[]rpproto.DemandChangelogCreate, error) {

	// needUpdateDemands记录全部扣减完成后的最终结果
	needUpdateDemands := make(map[string]*rpd.ResPlanDemandTable)
	// demandUpdatedResource记录变化量，用于记录变更日志
	demandUpdatedResource := make(map[string]*ptypes.DemandResource)

	// 从 woa_zone 获取城市/地区的中英文对照
	deviceTypeMap, err := c.GetAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, err
	}

	batchCreateReq := make([]rpproto.ResPlanDemandCreateReq, 0)
	batchCreateLogReq := make([]rpproto.DemandChangelogCreate, 0)
	for _, changeDemand := range changeDemands {
		// 因为CRP的预测和本地的不一定一致，这里需要根据聚合key获取一批通配的预测需求，在范围内调整，只保证通配范围内的总数和CRP对齐
		aggregateKey, err := changeDemand.GetAggregateKey(ticket.BkBizID, deviceTypeMap)
		if err != nil {
			logs.Errorf("failed to get aggregate key, err: %v, change demand: %+v, rid: %s", err, *changeDemand,
				kt.Rid)
			return nil, nil, nil, nil, err
		}
		localDemands, err := c.ListResPlanDemandByAggregateKey(kt, aggregateKey)
		if err != nil {
			logs.Errorf("failed to get local res plan demands, err: %v, aggregate key: %+v, change demand: %+v, "+
				"rid: %s", err, aggregateKey, *changeDemand, kt.Rid)
			return nil, nil, nil, nil, err
		}

		// 根据匹配情况决定如何调整数据库中的现有预测需求
		createReq, createLogReq, err := c.applyResPlanDemandChangeAggregate(kt, ticket, localDemands,
			changeDemand, needUpdateDemands, demandUpdatedResource)
		if err != nil {
			logs.Errorf("failed to apply res plan demand change aggregate, err: %v, change demand: %+v, rid: %s",
				err, *changeDemand, kt.Rid)
			return nil, nil, nil, nil, err
		}
		if createReq != nil {
			batchCreateReq = append(batchCreateReq, cvt.PtrToVal(createReq))
			batchCreateLogReq = append(batchCreateLogReq, cvt.PtrToVal(createLogReq))
		}
	}

	// 准备更新语句
	updatedIDs, batchUpdateReq, updateChangelogReqs, err := convUpdateResPlanDemandReqs(kt, ticket, needUpdateDemands,
		demandUpdatedResource, deviceTypeMap)
	if err != nil {
		logs.Errorf("failed to convert update res plan demand reqs, err: %v, ticket: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return nil, nil, nil, nil, err
	}

	upsertReq := &rpproto.ResPlanDemandBatchUpsertReq{
		CreateDemands: batchCreateReq,
		UpdateDemands: batchUpdateReq,
	}

	return upsertReq, updatedIDs, batchCreateLogReq, updateChangelogReqs, nil
}

func convUpdateResPlanDemandReqs(kt *kit.Kit, ticket *TicketInfo, updateDemands map[string]*rpd.ResPlanDemandTable,
	demandChangelog map[string]*ptypes.DemandResource, deviceTypeMap map[string]wdt.WoaDeviceTypeTable) (
	[]string, []rpproto.ResPlanDemandUpdateReq, []rpproto.DemandChangelogCreate, error) {

	updatedDemandIDs := make([]string, 0)
	batchUpdateReq := make([]rpproto.ResPlanDemandUpdateReq, 0)
	batchCreateChangelogReq := make([]rpproto.DemandChangelogCreate, 0)

	for demandID, updated := range updateDemands {
		deviceInfo, ok := deviceTypeMap[updated.DeviceType]
		if !ok {
			logs.Errorf("device_type: %s, not found in device_type_map, rid: %s", updated.DeviceType, kt.Rid)
			return nil, nil, nil, fmt.Errorf("device_type: %s is not found", updated.DeviceType)
		}

		updatedDemandIDs = append(updatedDemandIDs, demandID)

		// OS and memory should be calculated by cpu core
		updatedOS := decimal.NewFromInt(cvt.PtrToVal(updated.CpuCore)).Div(decimal.NewFromInt(deviceInfo.CpuCore))
		updatedMemory := updatedOS.Mul(decimal.NewFromInt(deviceInfo.Memory)).IntPart()

		updateReq := rpproto.ResPlanDemandUpdateReq{
			ID:       demandID,
			OS:       &updatedOS,
			CpuCore:  updated.CpuCore,
			Memory:   &updatedMemory,
			DiskSize: updated.DiskSize,
		}
		if kt.User == constant.BackendOperationUserKey {
			updateReq.Reviser = ticket.Applicant
		}
		batchUpdateReq = append(batchUpdateReq, updateReq)

		// 变更日志
		demandChangeRes, ok := demandChangelog[demandID]
		if !ok {
			// 更新日志无法创建不阻断整个预测的更新流程，因此此处仅记录warning，不返回error
			logs.Warnf("failed to get demand change res, demand_id: %s, rid: %s", demandID, kt.Rid)
			continue
		}

		expectTimeStr, err := times.TransTimeStrWithLayout(strconv.Itoa(updated.ExpectTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Warnf("failed to convert expect time to string, err: %v, expect time: %d, demand_id: %s, rid: %s",
				err, updated.ExpectTime, demandID, kt.Rid)
			continue
		}

		changeCpuCore := demandChangeRes.CpuCore
		changeDiskSize := demandChangeRes.DiskSize
		changeOS := decimal.NewFromInt(changeCpuCore).Div(decimal.NewFromInt(deviceInfo.CpuCore))
		changeMemory := changeOS.Mul(decimal.NewFromInt(deviceInfo.Memory)).IntPart()

		logCreateReq := rpproto.DemandChangelogCreate{
			DemandID:       demandID,
			TicketID:       ticket.ID,
			CrpOrderID:     ticket.CrpSn,
			SuborderID:     "",
			Type:           enumor.DemandChangelogTypeAdjust,
			ExpectTime:     expectTimeStr,
			ObsProject:     updated.ObsProject,
			RegionName:     updated.RegionName,
			ZoneName:       updated.ZoneName,
			DeviceType:     updated.DeviceType,
			OSChange:       &changeOS,
			CpuCoreChange:  &changeCpuCore,
			MemoryChange:   &changeMemory,
			DiskSizeChange: &changeDiskSize,
			Remark:         ticket.Remark,
		}
		batchCreateChangelogReq = append(batchCreateChangelogReq, logCreateReq)
	}

	return updatedDemandIDs, batchUpdateReq, batchCreateChangelogReq, nil
}

func (c *Controller) aggregateDemandChangeInfo(kt *kit.Kit, changeDemands []*ptypes.CrpOrderChangeInfo,
	ticket *TicketInfo) ([]*ptypes.CrpOrderChangeInfo, error) {

	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, _, err := c.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	changeDemandMap := make(map[ptypes.ResPlanDemandKey]*ptypes.CrpOrderChangeInfo)
	for _, changeDemand := range changeDemands {
		err := changeDemand.SetRegionAreaAndZoneID(zoneNameMap, regionNameMap)
		if err != nil {
			logs.Errorf("failed to set region area and zone id, err: %v, change demand: %+v, rid: %s", err,
				*changeDemand, kt.Rid)
			return nil, err
		}

		changeKey := changeDemand.GetKey(ticket.BkBizID, ticket.DemandClass)
		finalChange, ok := changeDemandMap[changeKey]
		if !ok {
			changeDemandMap[changeKey] = changeDemand
			continue
		}

		finalChange.ChangeOs = finalChange.ChangeOs.Add(changeDemand.ChangeOs)
		finalChange.ChangeCpuCore = finalChange.ChangeCpuCore + changeDemand.ChangeCpuCore
		finalChange.ChangeMemory = finalChange.ChangeMemory + changeDemand.ChangeMemory
		finalChange.ChangeDiskSize = finalChange.ChangeDiskSize + changeDemand.ChangeDiskSize
	}

	return maps.Values(changeDemandMap), nil
}

// BatchUpsertResPlanDemand batch upsert res plan demand and unlock res plans.
func (c *Controller) BatchUpsertResPlanDemand(kt *kit.Kit, upsertReq *rpproto.ResPlanDemandBatchUpsertReq,
	updatedIDs []string) ([]string, error) {

	unlockDemandIDs := make([]string, 0)
	// 批量创建和更新预测
	createdRst, err := c.client.DataService().Global.ResourcePlan.BatchUpsertResPlanDemand(kt, upsertReq)
	if err != nil {
		logs.Errorf("failed to batch upsert res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	unlockDemandIDs = append(unlockDemandIDs, createdRst.IDs...)
	unlockDemandIDs = append(unlockDemandIDs, updatedIDs...)
	if len(unlockDemandIDs) == 0 {
		return createdRst.IDs, nil
	}

	// unlock all crp demands.
	unlockReq := &rpproto.ResPlanDemandLockOpReq{
		IDs: unlockDemandIDs,
	}
	if err := c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Warnf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
	}

	return createdRst.IDs, nil
}

// applyResPlanDemandChangeAggregate apply changes according to aggregation rules.
// return the requests to create, and the needUpdateDemands map to update with a transaction.
func (c *Controller) applyResPlanDemandChangeAggregate(kt *kit.Kit, ticket *TicketInfo,
	localDemands []rpd.ResPlanDemandTable, changeDemand *ptypes.CrpOrderChangeInfo,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource) (
	*rpproto.ResPlanDemandCreateReq, *rpproto.DemandChangelogCreate, error) {

	// 追加预测，不需要模糊匹配
	if changeDemand.ChangeCpuCore > 0 {
		// 优先追加到完全匹配的预测需求上，如果找不到，直接新增
		demandKey := changeDemand.GetKey(ticket.BkBizID, ticket.DemandClass)
		matchDemands := matchDemandTableByDemandKey(kt, localDemands, demandKey)

		// 没有完全匹配的，直接新增
		if len(matchDemands) == 0 {
			createReq, createLogReq, err := convCreateResPlanDemandReqs(kt, ticket, changeDemand)
			if err != nil {
				logs.Errorf("failed to convert create res plan demand reqs, err: %v, bk_biz_id: %d, "+
					"changeDemand: %+v, rid: %s", err, ticket.BkBizID, changeDemand, kt.Rid)
				return nil, nil, err
			}

			return &createReq, &createLogReq, nil
		}

		// 以demandKey唯一键匹配到的数据，最多只可能有1条，追加
		localDemand := matchDemands[0]
		// 对已有数据的变更只记录变更量，全部变更处理完成后再统一拼装更新请求
		if _, ok := needUpdateDemands[localDemand.ID]; !ok {
			needUpdateDemands[localDemand.ID] = &localDemand
			demandUpdateRes[localDemand.ID] = &ptypes.DemandResource{
				DeviceType: localDemand.DeviceType,
				CpuCore:    0,
				DiskSize:   0,
			}
		}
		remainedCpuCore := cvt.PtrToVal(needUpdateDemands[localDemand.ID].CpuCore) + changeDemand.ChangeCpuCore
		remainedDiskSize := cvt.PtrToVal(needUpdateDemands[localDemand.ID].DiskSize) + changeDemand.ChangeDiskSize

		needUpdateDemands[localDemand.ID].CpuCore = &remainedCpuCore
		needUpdateDemands[localDemand.ID].DiskSize = &remainedDiskSize
		demandUpdateRes[localDemand.ID].CpuCore += changeDemand.ChangeCpuCore
		demandUpdateRes[localDemand.ID].DiskSize += changeDemand.ChangeDiskSize

		return nil, nil, nil
	}

	// 调减预测，在范围内随机模糊调减，优先调减加锁的
	err := c.DeductResPlanDemandAggregate(kt, localDemands, needUpdateDemands, demandUpdateRes, changeDemand)
	if err != nil {
		logs.Errorf("failed to update res plan demand aggregate, err: %v, bk_biz_id: %d, changeDemand: %+v, rid: %s",
			err, ticket.BkBizID, changeDemand, kt.Rid)
		return nil, nil, err
	}

	return nil, nil, nil
}

// QueryCrpOrderChangeInfo query crp order change info.
func (c *Controller) QueryCrpOrderChangeInfo(kt *kit.Kit, orderID string) ([]*ptypes.CrpOrderChangeInfo,
	error) {

	// init request parameter.
	queryReq := &cvmapi.PlanOrderChangeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderChangeMethod,
		},
		Params: &cvmapi.PlanOrderChangeParam{
			Page: &cvmapi.Page{
				Start: 0,
				Size:  int(core.DefaultMaxPageLimit),
			},
			OrderId: []string{orderID},
			BgName:  []string{cvmapi.CvmCbsPlanQueryBgName},
		},
	}

	// query all demands.
	result := make([]*ptypes.CrpOrderChangeInfo, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		queryReq.Params.Page.Start = start
		rst, err := c.crpCli.QueryPlanOrderChange(kt.Ctx, kt.Header(), queryReq)
		if err != nil {
			logs.Errorf("failed to query crp order change info, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if rst.Error.Code != 0 {
			logs.Errorf("failed to query crp order change info, err: %v, crp_trace: %s, rid: %s",
				rst.Error.Message, rst.TraceId, kt.Rid)
			return nil, fmt.Errorf("failed to query crp order change info, err: %s", rst.Error.Message)
		}

		for _, ele := range rst.Result.Data {
			one, err := convOrderChangeInfoFromCrpRespItem(kt, orderID, ele)
			if err != nil {
				logs.Warnf("failed to convert crp order change info, err: %v, crp_result: %+v, rid: %s", err,
					*ele, kt.Rid)
				continue
			}

			result = append(result, one)
		}

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}

// convOrderChangeInfoFromCrpRespItem convert order change info from crp response item.
func convOrderChangeInfoFromCrpRespItem(kt *kit.Kit, orderID string, item *cvmapi.PlanOrderChangeItem) (
	*ptypes.CrpOrderChangeInfo, error) {

	diskType, err := enumor.GetDiskTypeFromCrpName(item.DiskTypeName)
	if err != nil {
		logs.Errorf("failed to get disk type from crp diskTypeName, err: %v, diskTypeName: %s, rid: %s",
			err, item.DiskTypeName, kt.Rid)
		return nil, err
	}

	if err := item.PlanType.Validate(); err != nil {
		logs.Errorf("crp plan type invalid, err: %v, planType: %s, rid: %s", err, item.PlanType, kt.Rid)
		return nil, err
	}

	if err := item.ProjectName.ValidateResPlan(); err != nil {
		logs.Errorf("crp project name invalid, err: %v, project name: %s, rid: %s", err, item.ProjectName, kt.Rid)
		return nil, err
	}

	resModeCode, err := item.ResourceMode.Code()
	if err != nil {
		logs.Errorf("crp resource mode invalid, err: %v, resource mode: %s, rid: %s", err, item.ResourceMode, kt.Rid)
		return nil, err
	}

	resType := enumor.DemandResTypeCVM
	// 机型为空时为CBS类型变更，TODO 暂不支持CBS类型的变更，没有机型无法和本地数据进行匹配
	if item.InstanceModel == "" {
		resType = enumor.DemandResTypeCBS
		return nil, errors.New("cbs type is not support")
	}

	return &ptypes.CrpOrderChangeInfo{
		OrderID:        orderID,
		ExpectTime:     item.UseTime,
		ObsProject:     item.ProjectName,
		DemandResType:  resType,
		ResMode:        resModeCode,
		PlanType:       item.PlanType.GetCode(),
		RegionName:     item.CityName,
		ZoneName:       item.ZoneName,
		DeviceFamily:   item.InstanceFamily,
		DeviceClass:    item.InstanceType,
		DeviceType:     item.InstanceModel,
		CoreType:       item.CoreTypeName,
		DiskType:       diskType,
		DiskTypeName:   item.DiskTypeName,
		DiskIO:         item.InstanceIO,
		ChangeOs:       item.ChangeCvmAmount,
		ChangeCpuCore:  item.ChangeCoreAmount,
		ChangeMemory:   item.ChangeRamAmount,
		ChangeDiskSize: item.ChangedDiskAmount,
	}, nil
}

// getMetaMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap.
func (c *Controller) getMetaMaps(kt *kit.Kit) (map[string]string, map[string]dmtypes.RegionArea,
	map[string]wdt.WoaDeviceTypeTable, error) {

	// get zone id name mapping.
	zoneMap, err := c.dao.WoaZone().GetZoneMap(kt)
	if err != nil {
		logs.Errorf("get zone map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get region area mapping.
	regionAreaMap, err := c.dao.WoaZone().GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("get region area map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get device type mapping.
	deviceTypeMap, err := c.deviceTypesMap.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	return zoneMap, regionAreaMap, deviceTypeMap, nil
}

func getMetaNameMapsFromIDMap(zoneMap map[string]string, regionAreaMap map[string]dmtypes.RegionArea) (
	map[string]string, map[string]dmtypes.RegionArea) {

	zoneNameMap := make(map[string]string)
	for id, name := range zoneMap {
		zoneNameMap[name] = id
	}
	regionNameMap := make(map[string]dmtypes.RegionArea)
	for _, item := range regionAreaMap {
		regionNameMap[item.RegionName] = item
	}
	return zoneNameMap, regionNameMap
}

// ListResPlanDemandByAggregateKey list res plan demand by key.
func (c *Controller) ListResPlanDemandByAggregateKey(kt *kit.Kit, demandKey ptypes.ResPlanDemandAggregateKey) (
	[]rpd.ResPlanDemandTable, error) {

	listRules := make([]*filter.AtomRule, 0)

	startExpTime, err := times.ConvStrTimeToInt(demandKey.ExpectTimeRange.Start, constant.DateLayout)
	if err != nil {
		logs.Errorf("failed to parse month range, err: %v, month_range: %v, rid: %s", err,
			demandKey.ExpectTimeRange, kt.Rid)
		return nil, err
	}
	endExpTime, err := times.ConvStrTimeToInt(demandKey.ExpectTimeRange.End, constant.DateLayout)
	if err != nil {
		logs.Errorf("failed to parse month range, err: %v, month_range: %v, rid: %s", err,
			demandKey.ExpectTimeRange, kt.Rid)
		return nil, err
	}

	listRules = append(listRules, tools.RuleEqual("bk_biz_id", demandKey.BkBizID))
	listRules = append(listRules, tools.RuleEqual("plan_type", demandKey.PlanType))
	listRules = append(listRules, tools.RuleEqual("obs_project", demandKey.ObsProject))

	listRules = append(listRules, tools.RuleGreaterThanEqual("expect_time", startExpTime))
	listRules = append(listRules, tools.RuleLessThanEqual("expect_time", endExpTime))

	listRules = append(listRules, tools.RuleEqual("region_id", demandKey.RegionID))
	listRules = append(listRules, tools.RuleEqual("device_family", demandKey.DeviceFamily))
	listRules = append(listRules, tools.RuleEqual("core_type", demandKey.CoreType))
	listRules = append(listRules, tools.RuleEqual("demand_res_type", demandKey.ResType))

	listFilter := tools.ExpressionAnd(listRules...)

	listReq := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Filter: listFilter,
			Page: &core.BasePage{
				Start: 0,
				Limit: core.DefaultMaxPageLimit,
			},
		},
	}

	result := make([]rpd.ResPlanDemandTable, 0)
	for {
		rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list local res plan demand, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result = append(result, rst.Details...)

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return result, nil
}

// matchDemandTableByDemandKey match demand table by demand key. expect only one.
func matchDemandTableByDemandKey(kt *kit.Kit, demands []rpd.ResPlanDemandTable,
	key ptypes.ResPlanDemandKey) []rpd.ResPlanDemandTable {

	res := make([]rpd.ResPlanDemandTable, 0)

	for _, demand := range demands {
		expectDateStr, err := times.TransTimeStrWithLayout(strconv.Itoa(demand.ExpectTime), constant.DateLayoutCompact,
			constant.DateLayout)
		if err != nil {
			logs.Warnf("failed to parse demand expect time, err: %v, expect_time: %d, rid: %s", err,
				demand.ExpectTime, kt.Rid)
			continue
		}

		demandKey := ptypes.ResPlanDemandKey{
			BkBizID:       demand.BkBizID,
			DemandClass:   demand.DemandClass,
			DemandResType: demand.DemandResType,
			ResMode:       demand.ResMode,
			ObsProject:    demand.ObsProject,
			ExpectTime:    expectDateStr,
			PlanType:      demand.PlanType,
			RegionID:      demand.RegionID,
			ZoneID:        demand.ZoneID,
			DeviceType:    demand.DeviceType,
			DiskType:      demand.DiskType,
			DiskIO:        demand.DiskIO,
		}

		if demandKey == key {
			res = append(res, demand)
			break
		}
	}

	return res
}

// DeductResPlanDemandAggregate 根据聚合Key筛选一组res plan demand调减
// 只有 demandChange.ChangeCpuCore <= 0 时才可以使用这个函数，> 0 时直接走新增
// 不考虑demand key完全一致，优先选择加锁的
func (c *Controller) DeductResPlanDemandAggregate(kt *kit.Kit, aggregateDemands []rpd.ResPlanDemandTable,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource,
	demandChange *ptypes.CrpOrderChangeInfo) error {

	if demandChange.ChangeCpuCore > 0 {
		return fmt.Errorf("demand deduct cpu core should be <= 0, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
	}

	// 优先选择加锁的
	unlockedDemands := make([]rpd.ResPlanDemandTable, 0, len(aggregateDemands))
	for _, demand := range aggregateDemands {
		if demandChange.ChangeCpuCore == 0 {
			break
		}
		if cvt.PtrToVal(demand.Locked) != enumor.CrpDemandLocked {
			unlockedDemands = append(unlockedDemands, demand)
			continue
		}

		deductCpuNum := prepareDeductResPlanDemand(demand, demandChange.ChangeCpuCore, needUpdateDemands,
			demandUpdateRes)
		// deductCpuNum是正数，demandChange.ChangeCpuCore是负数
		demandChange.ChangeCpuCore += deductCpuNum
	}

	// 如果还没有扣减完，继续扣减未加锁的，但是这是预期外的，需要给一个警告
	for _, demand := range unlockedDemands {
		if demandChange.ChangeCpuCore == 0 {
			break
		}
		logs.Warnf("demand update occur on unlocked data and may result in unexpected results, update demand: %s, rid: %s",
			demand.ID, kt.Rid)

		deductCpuNum := prepareDeductResPlanDemand(demand, demandChange.ChangeCpuCore, needUpdateDemands,
			demandUpdateRes)
		// deductCpuNum是正数，demandChange.ChangeCpuCore是负数
		demandChange.ChangeCpuCore += deductCpuNum
	}

	// 现有的量不够调减
	if demandChange.ChangeCpuCore < 0 {
		logs.Errorf("failed to update res plan demand, remained demand is not enough to deduct, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
		return errors.New("remained demand is not enough to deduct")
	}

	return nil
}

// prepareDeductResPlanDemand prepare the needUpdateDemands map to deduct res plan demand.
// return is the cpu core will be deducted, always > 0.
func prepareDeductResPlanDemand(updateDemand rpd.ResPlanDemandTable, changeCpuCore int64,
	needUpdateDemands map[string]*rpd.ResPlanDemandTable, demandUpdateRes map[string]*ptypes.DemandResource) int64 {

	if _, ok := needUpdateDemands[updateDemand.ID]; !ok {
		needUpdateDemands[updateDemand.ID] = &updateDemand
		demandUpdateRes[updateDemand.ID] = &ptypes.DemandResource{
			DeviceType: updateDemand.DeviceType,
			CpuCore:    0,
			DiskSize:   0,
		}
	}
	demandCpuCore := cvt.PtrToVal(needUpdateDemands[updateDemand.ID].CpuCore)

	changeCpuCoreAbs := math.Abs(float64(changeCpuCore))
	deductCpuNum := int64(math.Min(changeCpuCoreAbs, float64(demandCpuCore)))

	// 对已有数据的变更只记录变更量，全部变更处理完成后再统一拼装更新请求
	// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
	remainedCpuCore := demandCpuCore - deductCpuNum
	needUpdateDemands[updateDemand.ID].CpuCore = &remainedCpuCore
	demandUpdateRes[updateDemand.ID].CpuCore -= deductCpuNum

	return deductCpuNum
}
