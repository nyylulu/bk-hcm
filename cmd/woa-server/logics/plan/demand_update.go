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

	mtypes "hcm/cmd/woa-server/types/meta"
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
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// applyResPlanDemandChangeAggregate apply changes according to aggregation rules.
func (c *Controller) applyResPlanDemandChangeAggregate(kt *kit.Kit, ticket *TicketInfo, bizOrgRel mtypes.BizOrgRel,
	matchDemands []rpd.ResPlanDemandTable, changeDemand *ptypes.CrpOrderChangeInfo,
	deviceTypeMap map[string]wdt.WoaDeviceTypeTable) ([]string, error) {

	unlockDemandIDs := make([]string, 0)

	// 未匹配到完全一致的预测，需要根据 调增OR调减 选择新建或模糊调减数据
	if len(matchDemands) == 0 {
		// 新增数据
		if changeDemand.ChangeCpuCore > 0 {
			demandIDs, err := c.CreateResPlanDemand(kt, bizOrgRel, ticket, changeDemand)
			if err != nil {
				logs.Errorf("failed to create res plan demand, err: %v, bk_biz_id: %d, changeDemand: %+v, rid: %s",
					err, bizOrgRel.BkBizID, changeDemand, kt.Rid)
				return nil, err
			}
			unlockDemandIDs = append(unlockDemandIDs, demandIDs...)
			return unlockDemandIDs, nil
		}

		// 模糊调减
		updatedDemandIDs, err := c.DeductResPlanDemandAggregate(kt, ticket, matchDemands, changeDemand, deviceTypeMap)
		if err != nil {
			logs.Errorf("failed to update res plan demand aggregate, err: %v, bk_biz_id: %d, changeDemand: %+v, rid: %s",
				err, bizOrgRel.BkBizID, changeDemand, kt.Rid)
			return nil, err
		}
		unlockDemandIDs = append(unlockDemandIDs, updatedDemandIDs...)
		return unlockDemandIDs, nil
	}

	// 以demandKey唯一键匹配到的数据，最多只可能有1条
	// 更新数据
	err := c.UpdateResPlanDemand(kt, ticket, matchDemands[0], changeDemand)
	if err != nil {
		logs.Errorf("failed to update res plan demand, err: %v, bk_biz_id: %d, changeDemand: %+v, rid: %s",
			err, bizOrgRel.BkBizID, changeDemand, kt.Rid)
		return nil, err
	}
	unlockDemandIDs = append(unlockDemandIDs, matchDemands[0].ID)

	return unlockDemandIDs, nil
}

// applyResPlanDemandChange apply res plan demand change.
func (c *Controller) applyResPlanDemandChange(kt *kit.Kit, ticket *TicketInfo) error {
	// call crp api to get plan order change info.
	changeDemands, err := c.QueryCrpOderChangeInfo(kt, ticket.CrpSn)
	if err != nil {
		logs.Errorf("failed to query crp order change info, err: %v, crp_sn: %s, rid: %s", err, ticket.CrpSn, kt.Rid)
		return err
	}

	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, deviceTypeMap, err := c.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	// upsert crp demand id and biz relation.
	bizOrgRel := mtypes.BizOrgRel{
		BkBizID:         ticket.BkBizID,
		BkBizName:       ticket.BkBizName,
		OpProductID:     ticket.OpProductID,
		OpProductName:   ticket.OpProductName,
		PlanProductID:   ticket.PlanProductID,
		PlanProductName: ticket.PlanProductName,
		VirtualDeptID:   ticket.VirtualDeptID,
		VirtualDeptName: ticket.VirtualDeptName,
	}

	unlockDemandIDs := make([]string, 0)
	for _, changeDemand := range changeDemands {
		err := changeDemand.SetRegionAreaAndZoneID(zoneNameMap, regionNameMap)
		if err != nil {
			logs.Warnf("failed to set region area and zone id, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		// 因为CRP的预测和本地的不一定一致，这里需要根据聚合key获取一批通配的预测需求，在范围内调整，只保证通配范围内的总数和CRP对齐
		aggregateKey, err := changeDemand.GetAggregateKey(ticket.BkBizID, deviceTypeMap)
		if err != nil {
			logs.Warnf("failed to get aggregate key, err: %v, rid: %s", err, kt.Rid)
			continue
		}
		localDemands, err := c.ListResPlanDemandByAggregateKey(kt, aggregateKey)
		if err != nil {
			logs.Warnf("failed to get local res plan demands, err: %v, aggregate key: %+v, rid: %s", err,
				aggregateKey, kt.Rid)
			continue
		}

		// 优先调整完全匹配的预测需求，如果找不到，调减时随机挑选几条调减（优先调整加锁的），避免调出负数
		demandKey := changeDemand.GetKey(ticket.BkBizID, ticket.DemandClass)
		matchDemands := matchDemandTableByDemandKey(kt, localDemands, demandKey)

		// 根据匹配情况决定如何调整数据库中的现有预测需求
		changeDemandIDs, err := c.applyResPlanDemandChangeAggregate(kt, ticket, bizOrgRel, matchDemands,
			changeDemand, deviceTypeMap)
		if err != nil {
			logs.Warnf("failed to apply res plan demand change aggregate, err: %v, rid: %s", err, kt.Rid)
			continue
		}
		unlockDemandIDs = append(unlockDemandIDs, changeDemandIDs...)
	}

	if len(unlockDemandIDs) == 0 {
		return nil
	}

	// unlock all crp demands.
	unlockReq := &rpproto.ResPlanDemandLockOpReq{
		IDs: unlockDemandIDs,
	}
	if err = c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// QueryCrpOderChangeInfo query crp order change info.
func (c *Controller) QueryCrpOderChangeInfo(kt *kit.Kit, orderID string) ([]*ptypes.CrpOrderChangeInfo,
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
	// 机型为空时为CBS类型变更
	if item.InstanceModel == "" {
		resType = enumor.DemandResTypeCBS
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
	deviceTypeMap, err := c.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
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
func (c *Controller) DeductResPlanDemandAggregate(kt *kit.Kit, ticket *TicketInfo,
	aggregateDemands []rpd.ResPlanDemandTable, demandChange *ptypes.CrpOrderChangeInfo,
	deviceTypes map[string]wdt.WoaDeviceTypeTable) ([]string, error) {

	if demandChange.ChangeCpuCore > 0 {
		return nil, fmt.Errorf("demand deduct cpu core should be <= 0, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
	}

	updatedDemandIDs := make([]string, 0)
	updateReqs := make([]rpproto.ResPlanDemandUpdateReq, 0)
	changelogCreateReqs := make([]rpproto.DemandChangelogCreate, 0)

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

		demandUpdateReq, changelogCreateReq := convDeductNumberAndUpdateReq(kt, demand, demandChange, ticket,
			deviceTypes)
		updateReqs = append(updateReqs, demandUpdateReq)
		changelogCreateReqs = append(changelogCreateReqs, changelogCreateReq)
		updatedDemandIDs = append(updatedDemandIDs, demand.ID)
	}

	// 如果还没有扣减完，继续扣减未加锁的，但是这是预期外的，需要给一个警告
	for _, demand := range unlockedDemands {
		if demandChange.ChangeCpuCore == 0 {
			break
		}

		demandUpdateReq, changelogCreateReq := convDeductNumberAndUpdateReq(kt, demand, demandChange, ticket,
			deviceTypes)
		updateReqs = append(updateReqs, demandUpdateReq)
		logs.Warnf("demand update occur on unlocked data and may result in unexpected results, update demand: %s, rid: %s",
			demand.ID, kt.Rid)
		changelogCreateReqs = append(changelogCreateReqs, changelogCreateReq)
		updatedDemandIDs = append(updatedDemandIDs, demand.ID)
	}

	// 现有的量不够调减
	if demandChange.ChangeCpuCore < 0 {
		logs.Errorf("failed to update res plan demand, remained demand is not enough to deduct, change_info: %+v, rid: %s",
			*demandChange, kt.Rid)
		return updatedDemandIDs, errors.New("remained demand is not enough to deduct")
	}

	if len(updateReqs) == 0 {
		return updatedDemandIDs, nil
	}

	batchUpdateReq := &rpproto.ResPlanDemandBatchUpdateReq{
		Demands: updateReqs,
	}
	err := c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemand(kt, batchUpdateReq)
	if err != nil {
		logs.Errorf("failed to update res plan demand, err: %v, rid: %s", err, kt.Rid)
		return updatedDemandIDs, err
	}

	// 更新日志
	changelogReq := &rpproto.DemandChangelogCreateReq{
		Changelogs: changelogCreateReqs,
	}

	_, err = c.client.DataService().Global.ResourcePlan.BatchCreateDemandChangelog(kt, changelogReq)
	if err != nil {
		logs.Warnf("failed to batch create plan demand changelog, err: %v, rid: %s", err, kt.Rid)
	}

	return updatedDemandIDs, nil
}

// convDeductNumberAndUpdateReq 拼装扣减更新请求
// 注意该方法会改变传入的demandChange数据，直接扣减掉对应的CPU数量
func convDeductNumberAndUpdateReq(kt *kit.Kit, demand rpd.ResPlanDemandTable, demandChange *ptypes.CrpOrderChangeInfo,
	ticket *TicketInfo, deviceTypes map[string]wdt.WoaDeviceTypeTable) (
	rpproto.ResPlanDemandUpdateReq, rpproto.DemandChangelogCreate) {

	// deductCpuNum是正数，demandChange.ChangeCpuCore是负数
	changeCpuCoreAbs := math.Abs(float64(demandChange.ChangeCpuCore))
	deductCpuNum := int64(math.Min(changeCpuCoreAbs, float64(cvt.PtrToVal(demand.CpuCore))))
	demandChange.ChangeCpuCore += deductCpuNum

	// 换算调减的OS、memory数量
	deductOSNum := decimal.NewFromInt(deductCpuNum).Div(decimal.NewFromInt(deviceTypes[demand.DeviceType].CpuCore))
	deductMemoryNum := deductOSNum.Mul(decimal.NewFromInt(deviceTypes[demand.DeviceType].Memory)).IntPart()

	remainedOS := demand.OS.Decimal.Sub(deductOSNum)
	remainedCpuCore := cvt.PtrToVal(demand.CpuCore) - deductCpuNum
	remainedMemory := cvt.PtrToVal(demand.Memory) - deductMemoryNum

	demandUpdateReq := rpproto.ResPlanDemandUpdateReq{
		ID:      demand.ID,
		OS:      &remainedOS,
		CpuCore: &remainedCpuCore,
		Memory:  &remainedMemory,

		// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
		DiskSize: demand.DiskSize,
	}
	if kt.User == constant.BackendOperationUserKey {
		demandUpdateReq.Reviser = ticket.Applicant
	}

	// 记录更新日志
	expectDateStr, err := times.TransTimeStrWithLayout(strconv.Itoa(demand.ExpectTime), constant.DateLayoutCompact,
		constant.DateLayout)
	if err != nil {
		logs.Warnf("failed to parse demand expect time, err: %v, expect_time: %d, rid: %s", err,
			demand.ExpectTime, kt.Rid)
		expectDateStr = ""
	}
	osChange := deductOSNum.Neg()
	cpuCoreChange := -deductCpuNum
	memoryChange := -deductMemoryNum
	diskSizeChange := int64(0)
	changelogCreateReq := rpproto.DemandChangelogCreate{
		DemandID:       demand.ID,
		TicketID:       ticket.ID,
		CrpOrderID:     ticket.CrpSn,
		SuborderID:     "",
		Type:           enumor.DemandChangelogTypeAdjust,
		ExpectTime:     expectDateStr,
		ObsProject:     demand.ObsProject,
		RegionName:     demand.RegionName,
		ZoneName:       demand.ZoneName,
		DeviceType:     demand.DeviceType,
		OSChange:       &osChange,
		CpuCoreChange:  &cpuCoreChange,
		MemoryChange:   &memoryChange,
		DiskSizeChange: &diskSizeChange,
		Remark:         ticket.Remark,
	}

	return demandUpdateReq, changelogCreateReq
}
