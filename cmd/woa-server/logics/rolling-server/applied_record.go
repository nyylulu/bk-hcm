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
	deviceTypeCpuCoreMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return err
	}
	instGroupMap, err := l.configLogics.Device().ListInstanceGroup(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get instance group failed, err: %v, device_types: %v, rid: %s", err, deviceTypes, kt.Rid)
		return err
	}

	now := time.Now()
	records := make([]rsproto.RollingAppliedRecordCreateReq, len(createArr))
	for i, create := range createArr {
		deviceType := create.DeviceType
		deviceTypeCpuCore, ok := deviceTypeCpuCoreMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		instGroup, ok := instGroupMap[deviceType]
		if !ok {
			logs.Errorf("can not find instance group, type: %s, rid: %s", deviceType, kt.Rid)
			return fmt.Errorf("can not find instance group, type: %s", deviceType)
		}

		appliedRecord := rsproto.RollingAppliedRecordCreateReq{
			AppliedType:   create.AppliedType,
			BkBizID:       create.BizID,
			OrderID:       create.OrderID,
			SubOrderID:    create.SubOrderID,
			Year:          now.Year(),
			Month:         int(now.Month()),
			Day:           now.Day(),
			AppliedCore:   deviceTypeCpuCore.CPUAmount * int64(create.Count),
			InstanceGroup: instGroup,
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
	deviceTypeCpuCoreMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return 0, err
	}

	var deliveredCore int64 = 0
	for deviceType, count := range deviceTypeCountMap {
		deviceTypeCpuCore, ok := deviceTypeCpuCoreMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return 0, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		deliveredCore += deviceTypeCpuCore.CPUAmount * int64(count)
	}

	return deliveredCore, nil
}

// ReduceRollingCvmProdAppliedRecord 减少当月通过cvm生产的滚服交付数量
func (l *logics) ReduceRollingCvmProdAppliedRecord(kt *kit.Kit, devices []*types.MatchDeviceBrief) error {
	// 1. 获取匹配的主机的机型族以及对应的核心总数
	neededInstGroupCPUCoreMap, err := l.getRollingMatchCpuCore(kt, devices)
	if err != nil {
		logs.Errorf("get rolling server match cvm product host cpu core failed, err: %v, devices: %+v, rid: %s", err,
			devices, kt.Rid)
		return err
	}

	// 2. 获取当月cvm生产的滚服主机核心数余量，判断是否满足本次需求
	instGroups := make([]string, 0)
	for instGroup := range neededInstGroupCPUCoreMap {
		instGroups = append(instGroups, instGroup)
	}
	remainingInstGroupCPUCoreMap, err := l.getRollingCurMonthCVMProdCpuCore(kt, instGroups)
	if err != nil {
		logs.Errorf("get rolling server current month cvm product cpu core failed, err: %v, instGroups: %v, rid: %s",
			err, instGroups, kt.Rid)
		return err
	}
	for instGroup, neededCPUCore := range neededInstGroupCPUCoreMap {
		remainingCPUCore, ok := remainingInstGroupCPUCoreMap[instGroup]
		if !ok {
			logs.Errorf("the remaining core quantity of cvm production is %d, and the current demand quantity is %d, "+
				"instGroup: %s, rid: %s", remainingCPUCore, neededCPUCore, instGroup, kt.Rid)
			return fmt.Errorf("滚服当月机型族:%s剩余匹配量为%d, 当前所需匹配的核数为%d，不满足需求", instGroup, 0,
				neededCPUCore)
		}

		if remainingCPUCore < neededCPUCore {
			logs.Errorf("the remaining core quantity of cvm production is %d, and the current demand quantity is %d, "+
				"instGroup: %s, rid: %s", remainingCPUCore, neededCPUCore, instGroup, kt.Rid)
			return fmt.Errorf("滚服当月机型族:%s剩余匹配量为%d, 当前所需匹配的核数为%d，不满足需求", instGroup, 0,
				neededCPUCore)
		}
	}

	// 3. 查询cvm滚服申请记录，进行扣减
	if err = l.reduceRollingCvmProdCpuCore(kt, neededInstGroupCPUCoreMap); err != nil {
		logs.Errorf("reduce rolling server cvm product applied cpu core failed, err: %v, neededInstGroupCPUCoreMap: "+
			"%v, rid: %s", err, neededInstGroupCPUCoreMap, kt.Rid)
		return err
	}

	return nil
}

func (l *logics) getRollingMatchCpuCore(kt *kit.Kit, devices []*types.MatchDeviceBrief) (map[string]int64, error) {
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
	deviceTypeCpuCoreMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return nil, err
	}
	deviceTypeInstGroupMap, err := l.configLogics.Device().ListInstanceGroup(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get instance group by device type failed, err: %v, device_types: %v, rid: %s", err, deviceTypes,
			kt.Rid)
		return nil, err
	}

	instGroupCoreSumMap := make(map[string]int64)
	for deviceType, count := range deviceTypeCountMap {
		cpuCore, ok := deviceTypeCpuCoreMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return nil, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		instGroup, ok := deviceTypeInstGroupMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return nil, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		if _, ok = instGroupCoreSumMap[instGroup]; !ok {
			instGroupCoreSumMap[instGroup] = 0
		}
		instGroupCoreSumMap[instGroup] += cpuCore.CPUAmount * int64(count)
	}

	return instGroupCoreSumMap, nil
}

func (l *logics) getRollingCurMonthCVMProdCpuCore(kt *kit.Kit, instGroups []string) (map[string]int64, error) {
	now := time.Now()
	instGroupCpuCoreMap := make(map[string]int64)
	for _, instGroup := range instGroups {
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
		}
		summary, err := l.GetCpuCoreSummary(kt, summaryReq)
		if err != nil {
			logs.Errorf("get cpu core summary failed, err: %v, req: %+v, rid: %s", err, *summaryReq, kt.Rid)
			return nil, err
		}

		instGroupCpuCoreMap[instGroup] = summary.SumDeliveredCore
	}

	return instGroupCpuCoreMap, nil
}

func (l *logics) reduceRollingCvmProdCpuCore(kt *kit.Kit, neededInstGroupCPUCoreMap map[string]int64) error {
	if len(neededInstGroupCPUCoreMap) == 0 {
		return nil
	}

	now := time.Now()
	updatedRecords := make([]rsproto.RollingAppliedRecordUpdateReq, 0)
	for instGroup, neededCPUCore := range neededInstGroupCPUCoreMap {
		listReq := &rsproto.RollingAppliedRecordListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "year", Op: filter.Equal.Factory(), Value: now.Year()},
					&filter.AtomRule{Field: "month", Op: filter.Equal.Factory(), Value: now.Month()},
					&filter.AtomRule{Field: "applied_type", Op: filter.Equal.Factory(),
						Value: enumor.CvmProduceAppliedType},
					&filter.AtomRule{Field: "instance_group", Op: filter.Equal.Factory(), Value: instGroup},
				},
			},
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
				Sort:  "created_at",
			},
		}

		records := make([]*rstable.RollingAppliedRecord, 0)
		for {
			result, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
			if err != nil {
				logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
				return err
			}
			records = append(records, result.Details...)

			if len(result.Details) < constant.BatchOperationMaxLimit {
				break
			}

			listReq.Page.Start += constant.BatchOperationMaxLimit
		}

		for _, record := range records {
			if neededCPUCore <= 0 {
				break
			}

			var deliveredCore int64 = 0
			if neededCPUCore < *record.DeliveredCore {
				deliveredCore = *record.DeliveredCore - neededCPUCore
			}
			updatedRecords = append(updatedRecords,
				rsproto.RollingAppliedRecordUpdateReq{ID: record.ID, DeliveredCore: &deliveredCore})

			neededCPUCore -= *record.DeliveredCore
		}
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
