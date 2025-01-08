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

package rollingserver

import (
	"fmt"
	"time"

	rstypes "hcm/cmd/woa-server/types/rolling-server"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/querybuilder"
)

// CanApplyHost 是否可以通过滚服项目申请主机，如果不可以，会通过第二个返回值说明原因
func (l *logics) CanApplyHost(kt *kit.Kit, bizID int64, appliedCount uint, appliedType enumor.AppliedType) (bool,
	string, error) {

	// 如果是常规类型的申请，需要检验该业务已交付数+本次申请数，是否大于该业务限制的可申请额度
	if appliedType == enumor.NormalAppliedType {
		hasQuota, reason, err := l.isBizCurMonthHavingQuota(kt, bizID, appliedCount)
		if err != nil {
			logs.Errorf("determine whether the biz having quota failed, err: %v, bizID: %d, appliedCount: %d, rid: %s",
				err, bizID, appliedCount, kt.Rid)
			return false, "", err
		}
		if !hasQuota {
			logs.Errorf("biz(%d) can not apply host: reason: %s, rid: %s", bizID, reason, kt.Rid)
			return hasQuota, reason, nil
		}
	}

	// 校验当月滚服已申请数+本次申请数，是否大于全局额度
	hasQuota, reason, err := l.isSystemCurMonthHavingQuota(kt, appliedCount)
	if err != nil {
		logs.Errorf("determine whether the system having quota failed, err: %v, bizID: %d, appliedCount: %d, rid: %s",
			err, bizID, appliedCount, kt.Rid)
		return false, "", err
	}
	if !hasQuota {
		logs.Errorf("biz(%d) can not apply host: reason: %s, rid: %s", bizID, reason, kt.Rid)
		return hasQuota, reason, nil
	}

	return true, "", nil
}

// CreateAppliedRecord 创建滚服申请记录
func (l *logics) CreateAppliedRecord(kt *kit.Kit, createArr []rstypes.CreateAppliedRecordData) error {
	deviceTypes := make([]string, 0)
	for _, create := range createArr {
		deviceTypes = append(deviceTypes, create.DeviceType)
	}
	deviceTypeInfoMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return err
	}

	now := time.Now()
	records := make([]rsproto.RollingAppliedRecordCreateReq, len(createArr))
	for i, create := range createArr {
		deviceType := create.DeviceType
		deviceTypeInfo, ok := deviceTypeInfoMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, create: %+v, rid: %s", deviceType, create, kt.Rid)
			return fmt.Errorf("can not find device_type, type: %s", deviceType)
		}

		appliedRecord := rsproto.RollingAppliedRecordCreateReq{
			RequireType:   create.RequireType,
			AppliedType:   create.AppliedType,
			BkBizID:       create.BizID,
			OrderID:       create.OrderID,
			SubOrderID:    create.SubOrderID,
			Year:          now.Year(),
			Month:         int(now.Month()),
			Day:           now.Day(),
			AppliedCore:   deviceTypeInfo.CPUAmount * int64(create.Count),
			InstanceGroup: deviceTypeInfo.DeviceGroup,
			CoreType:      deviceTypeInfo.CoreType,
		}
		records[i] = appliedRecord
	}

	req := &rsproto.BatchCreateRollingAppliedRecordReq{AppliedRecords: records}
	if _, err = l.client.DataService().Global.RollingServer.CreateAppliedRecord(kt, req); err != nil {
		logs.Errorf("create rolling server applied record failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	return nil
}

// UpdateSubOrderRollingDeliveredCore 更新子单的滚服申请记录的交付cpu核心
func (l *logics) UpdateSubOrderRollingDeliveredCore(kt *kit.Kit, bizID int64, subOrderID string,
	appliedTypes []enumor.AppliedType, deviceTypeCountMap map[string]int) error {

	listReq := &rsproto.RollingAppliedRecordListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "suborder_id", Op: filter.Equal.Factory(), Value: subOrderID},
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
				&filter.AtomRule{Field: "applied_type", Op: filter.In.Factory(), Value: appliedTypes},
			},
		},
		Fields: []string{"id"},
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	listRes, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
	if err != nil {
		logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
		return err
	}
	if len(listRes.Details) == 0 {
		logs.Errorf("can not find rolling server applied record, suborder_id: %s, rid: %s", subOrderID, kt.Rid)
		return fmt.Errorf("can not find rolling server applied record, suborder_id: %s", subOrderID)
	}

	deliveredCore, err := l.GetCpuCoreSum(kt, deviceTypeCountMap)
	if err != nil {
		logs.Errorf("get cpu core failed, err: %v, deviceTypeCountMap: %v, rid: %s", err, deliveredCore, kt.Rid)
		return err
	}

	update := rsproto.RollingAppliedRecordUpdateReq{ID: listRes.Details[0].ID, DeliveredCore: &deliveredCore}
	req := &rsproto.BatchUpdateRollingAppliedRecordReq{AppliedRecords: []rsproto.RollingAppliedRecordUpdateReq{update}}
	if err = l.client.DataService().Global.RollingServer.UpdateAppliedRecord(kt, req); err != nil {
		logs.Errorf("update applied record failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	return nil
}

// GetCpuCoreSum 获取机型对应的cpu核数之和
func (l *logics) GetCpuCoreSum(kt *kit.Kit, deviceTypeCountMap map[string]int) (int64, error) {
	deviceTypes := make([]string, 0)
	for deviceType := range deviceTypeCountMap {
		deviceTypes = append(deviceTypes, deviceType)
	}
	deviceTypeInfoMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return 0, err
	}

	var deliveredCore int64 = 0
	for deviceType, count := range deviceTypeCountMap {
		deviceTypeInfo, ok := deviceTypeInfoMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return 0, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		deliveredCore += deviceTypeInfo.CPUAmount * int64(count)
	}

	return deliveredCore, nil
}

// ReduceRollingCvmProdAppliedRecord 减少当月通过cvm生产的滚服交付数量
func (l *logics) ReduceRollingCvmProdAppliedRecord(kt *kit.Kit, devices []*types.MatchDeviceBrief) error {
	// 1. 获取匹配的主机类型信息以及对应的核心总数
	needCoreMap, err := l.getRollingMatchCpuCore(kt, devices)
	if err != nil {
		logs.Errorf("get rolling server match cvm product host cpu core failed, err: %v, devices: %+v, rid: %s", err,
			devices, kt.Rid)
		return err
	}

	// 2. 获取当月cvm生产的滚服主机核心数余量，判断是否满足本次需求
	instGroupDeviceSizeMap := make(map[string][]enumor.CoreType)
	for instGroup, deviceSizeCoreMap := range needCoreMap {
		for deviceSize := range deviceSizeCoreMap {
			instGroupDeviceSizeMap[instGroup] = append(instGroupDeviceSizeMap[instGroup], deviceSize)
		}
	}
	remainingCoreMap, err := l.getRollingCurMonthCVMProdCpuCore(kt, instGroupDeviceSizeMap)
	if err != nil {
		logs.Errorf("get rolling server current month cvm product cpu core failed, err: %v, instGroupDeviceSizeMap: "+
			"%v, rid: %s", err, instGroupDeviceSizeMap, kt.Rid)
		return err
	}

	for instGroup, needCore := range needCoreMap {
		remainingCore, ok := remainingCoreMap[instGroup]
		if !ok {
			logs.Errorf("the remaining core quantity of cvm production is %d, and the current demand quantity is %v, "+
				"instGroup: %s, rid: %s", 0, needCore, instGroup, kt.Rid)
			return fmt.Errorf("滚服当月机型族:%s剩余匹配量为%d, 当前所需匹配的列表为%v，不满足需求", instGroup, 0,
				needCore)
		}

		commonRemaining := remainingCore[rstypes.OldVersionCoreType]

		for deviceSize, need := range needCore {
			remaining := remainingCore[deviceSize]

			if remaining+commonRemaining < need {
				logs.Errorf("the remaining core quantity of cvm production is %d, and the current demand quantity is"+
					" %d, instGroup: %s, rid: %s", remaining+commonRemaining, need, instGroup, kt.Rid)
				return fmt.Errorf("滚服当月机型族:%s,核心类型:%s剩余匹配量为%d, 当前所需匹配的核数为%d，不满足需求",
					instGroup, deviceSize, remaining+commonRemaining, need)
			}

			if remaining < need {
				commonRemaining -= need - remaining
			}
		}
	}

	// 3. 查询cvm滚服申请记录，进行扣减
	if err = l.reduceRollingCvmProdCpuCore(kt, needCoreMap); err != nil {
		logs.Errorf("reduce rolling server cvm product applied cpu core failed, err: %v, neededInstGroupCPUCoreMap: "+
			"%v, rid: %s", err, needCoreMap, kt.Rid)
		return err
	}

	return nil
}

func (l *logics) getRollingMatchCpuCore(kt *kit.Kit, devices []*types.MatchDeviceBrief) (
	map[string]map[enumor.CoreType]int64, error) {

	assetIDs := make([]string, len(devices))
	for i, device := range devices {
		assetIDs[i] = device.AssetId
	}
	deviceTypeCountMap := make(map[string]int)
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{querybuilder.AtomRule{
					Field:    "bk_asset_id",
					Operator: querybuilder.OperatorIn,
					Value:    assetIDs,
				},
				},
			},
		},
		Fields: []string{"bk_svr_device_cls_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}
	for {
		resp, err := l.esbClient.Cmdb().ListHost(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
			return nil, err
		}
		for _, host := range resp.Data.Info {
			if _, ok := deviceTypeCountMap[host.SvrDeviceClassName]; !ok {
				deviceTypeCountMap[host.SvrDeviceClassName] = 0
			}
			deviceTypeCountMap[host.SvrDeviceClassName]++
		}
		if len(resp.Data.Info) < pkg.BKMaxInstanceLimit {
			break
		}
		req.Page.Start += pkg.BKMaxInstanceLimit
	}

	deviceTypes := make([]string, 0)
	for deviceType := range deviceTypeCountMap {
		deviceTypes = append(deviceTypes, deviceType)
	}
	deviceTypeInfoMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s", err, deviceTypes,
			kt.Rid)
		return nil, err
	}

	// key为机型族，value为核心类型和核心数量的map
	deviceCoreMap := make(map[string]map[enumor.CoreType]int64)
	for deviceType, count := range deviceTypeCountMap {
		deviceTypeInfo, ok := deviceTypeInfoMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return nil, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}

		if _, ok = deviceCoreMap[deviceTypeInfo.DeviceGroup]; !ok {
			deviceCoreMap[deviceTypeInfo.DeviceGroup] = map[enumor.CoreType]int64{}
		}

		if _, ok = deviceCoreMap[deviceTypeInfo.DeviceGroup][deviceTypeInfo.CoreType]; !ok {
			deviceCoreMap[deviceTypeInfo.DeviceGroup][deviceTypeInfo.CoreType] = 0
		}

		deviceCoreMap[deviceTypeInfo.DeviceGroup][deviceTypeInfo.CoreType] += deviceTypeInfo.CPUAmount * int64(count)
	}

	return deviceCoreMap, nil
}

func (l *logics) getRollingCurMonthCVMProdCpuCore(kt *kit.Kit, instGroupDeviceSizeMap map[string][]enumor.CoreType) (
	map[string]map[enumor.CoreType]int64, error) {

	now := time.Now()
	deviceCoreMap := make(map[string]map[enumor.CoreType]int64)

	for instGroup, coreTypeArr := range instGroupDeviceSizeMap {
		needCoreTypeArr := make([]enumor.CoreType, len(coreTypeArr))
		copy(needCoreTypeArr, coreTypeArr)
		needCoreTypeArr = append(needCoreTypeArr, rstypes.OldVersionCoreType)

		for _, coreType := range needCoreTypeArr {
			summaryReq := &rstypes.CpuCoreSummaryReq{
				RollingServerDateRange: rstypes.RollingServerDateRange{
					Start: rstypes.RollingServerDateTimeItem{
						Year:  now.Year(),
						Month: int(now.Month()),
						Day:   rstypes.FirstDay,
					},
					End: rstypes.RollingServerDateTimeItem{
						Year:  now.Year(),
						Month: int(now.Month()),
						Day:   now.Day(),
					},
				},
				AppliedType:   enumor.CvmProduceAppliedType,
				InstanceGroup: instGroup,
				CoreType:      &coreType,
			}
			summary, err := l.GetCpuCoreSummary(kt, summaryReq)
			if err != nil {
				logs.Errorf("get cpu core summary failed, err: %v, req: %+v, rid: %s", err, *summaryReq, kt.Rid)
				return nil, err
			}

			if _, ok := deviceCoreMap[instGroup]; !ok {
				deviceCoreMap[instGroup] = map[enumor.CoreType]int64{}
			}

			deviceCoreMap[instGroup][coreType] = summary.SumDeliveredCore
		}

	}

	return deviceCoreMap, nil
}

func (l *logics) reduceRollingCvmProdCpuCore(kt *kit.Kit, needCoreMap map[string]map[enumor.CoreType]int64) error {
	if len(needCoreMap) == 0 {
		return nil
	}

	now := time.Now()
	updatedRecords := make([]rsproto.RollingAppliedRecordUpdateReq, 0)
	for instGroup, coreTypeCoreNumMap := range needCoreMap {
		coreTypeArr := make([]enumor.CoreType, 0)
		for coreType := range coreTypeCoreNumMap {
			coreTypeArr = append(coreTypeArr, coreType)
		}
		coreTypeArr = append(coreTypeArr, rstypes.OldVersionCoreType)

		listReq := &rsproto.RollingAppliedRecordListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "year", Op: filter.Equal.Factory(), Value: now.Year()},
					&filter.AtomRule{Field: "month", Op: filter.Equal.Factory(), Value: now.Month()},
					&filter.AtomRule{Field: "applied_type", Op: filter.Equal.Factory(),
						Value: enumor.CvmProduceAppliedType},
					&filter.AtomRule{Field: "instance_group", Op: filter.Equal.Factory(), Value: instGroup},
					&filter.AtomRule{Field: "core_type", Op: filter.In.Factory(), Value: coreTypeArr},
				},
			},
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
				Sort:  "created_at",
			},
		}

		commonRecords := make([]*rstable.RollingAppliedRecord, 0)
		specialRecords := make(map[enumor.CoreType][]*rstable.RollingAppliedRecord)
		for {
			result, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
			if err != nil {
				logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
				return err
			}

			for _, record := range result.Details {
				if record.CoreType == rstypes.OldVersionCoreType {
					commonRecords = append(commonRecords, record)
					continue
				}

				if _, ok := specialRecords[record.CoreType]; !ok {
					specialRecords[record.CoreType] = make([]*rstable.RollingAppliedRecord, 0)
				}
				specialRecords[record.CoreType] = append(specialRecords[record.CoreType], record)
			}

			if len(result.Details) < constant.BatchOperationMaxLimit {
				break
			}

			listReq.Page.Start += constant.BatchOperationMaxLimit
		}

		subUpdatedRecords, err := l.calculateUpdateRecord(specialRecords, commonRecords, coreTypeCoreNumMap)
		if err != nil {
			logs.Errorf("calculate update record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return err
		}
		updatedRecords = append(updatedRecords, subUpdatedRecords...)
	}

	if len(updatedRecords) == 0 {
		return nil
	}

	req := &rsproto.BatchUpdateRollingAppliedRecordReq{AppliedRecords: updatedRecords}
	if err := l.client.DataService().Global.RollingServer.UpdateAppliedRecord(kt, req); err != nil {
		logs.Errorf("update rolling applied record failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	return nil
}

// calculateUpdateRecord 计算需要更新滚服申请记录，第一个参数为指定大小核心的记录，第二个参数为通用的未指定大小核心的记录，
// 第三个参数为需要进行匹配计算的核心数
func (l *logics) calculateUpdateRecord(specialRecords map[enumor.CoreType][]*rstable.RollingAppliedRecord,
	commonRecords []*rstable.RollingAppliedRecord, coreTypeCoreNumMap map[enumor.CoreType]int64) (
	[]rsproto.RollingAppliedRecordUpdateReq, error) {

	updatedRecordMap := make(map[string]int64)
	for coreType, needCore := range coreTypeCoreNumMap {
		// 优先匹配指定大小核心的记录
		for _, record := range specialRecords[coreType] {
			if needCore <= 0 {
				break
			}

			var deliveredCore int64 = 0
			if needCore < *record.DeliveredCore {
				deliveredCore = *record.DeliveredCore - needCore
			}
			updatedRecordMap[record.ID] = deliveredCore

			needCore -= *record.DeliveredCore
		}

		if needCore == 0 {
			continue
		}

		// 如果还有没匹配的核心数，那么匹配通用的记录
		for i, record := range commonRecords {
			if needCore <= 0 {
				break
			}

			if *record.DeliveredCore == 0 {
				continue
			}

			var deliveredCore int64 = 0
			if needCore < *record.DeliveredCore {
				deliveredCore = *record.DeliveredCore - needCore
			}
			updatedRecordMap[record.ID] = deliveredCore

			needCore -= *record.DeliveredCore
			commonRecords[i].DeliveredCore = &deliveredCore
		}
	}

	updatedRecords := make([]rsproto.RollingAppliedRecordUpdateReq, 0)
	for id, deliveredCore := range updatedRecordMap {
		updatedRecords = append(updatedRecords,
			rsproto.RollingAppliedRecordUpdateReq{ID: id, DeliveredCore: &deliveredCore})
	}

	return updatedRecords, nil
}
