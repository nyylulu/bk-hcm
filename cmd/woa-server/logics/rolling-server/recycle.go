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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	rstypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// CalSplitRecycleHosts 计算并匹配指定时间范围指定业务的主机Host
func (l *logics) CalSplitRecycleHosts(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost,
	allBizReturnedCpuCore, globalQuota int64) (map[string]*rstypes.RecycleHostCpuInfo, []*table.RecycleHost,
	int64, error) {

	// 1.统计该回收Host子订单中的机型、全业务已退还的CPU总核数、全局的CPU总核数
	hostCpuMap, err := l.statisRecycleHostsAndBizCpuSystemQuota(kt, hosts)
	if err != nil {
		logs.ErrorJson("statis recycle host cpu system quota failed, err: %v, bkBizID: %d, hosts: %+v, rid: %s",
			err, bkBizID, hosts, kt.Rid)
		return nil, nil, 0, err
	}

	matchRange := make([]rstypes.RecycleMatchDateRange, 0)
	// 优先匹配61-90天的
	matchRange = append(matchRange, rstypes.RecycleMatchDateRange{
		Start: rstypes.CalculateMatchSixtyDay + 1,
		End:   rstypes.CalculateMatchNinetyDay,
	})
	// 在匹配0-60天的
	matchRange = append(matchRange, rstypes.RecycleMatchDateRange{
		Start: 0,
		End:   rstypes.CalculateMatchSixtyDay,
	})
	// 最后匹配91-121天的
	matchRange = append(matchRange, rstypes.RecycleMatchDateRange{
		Start: rstypes.CalculateMatchNinetyDay + 1,
		End:   rstypes.CalculateFineEndDay,
	})

	// 是否继续滚服回收
	isContinue := false
	for _, dateRange := range matchRange {
		// 匹配滚服申请订单
		hostCpuMap, isContinue, allBizReturnedCpuCore, err = l.MatchRecycleCvmHostQuota(kt, bkBizID,
			dateRange.End, dateRange.Start, allBizReturnedCpuCore, globalQuota, hostCpuMap)
		if err != nil {
			logs.Errorf("match recycle rolling cvm host quota by bizID failed, err: %v, bkBizID: %d, dateRange：%+v, "+
				"hostCpuMap: %+v, rid: %s", err, bkBizID, dateRange, hostCpuMap, kt.Rid)
			return nil, nil, 0, err
		}

		// 记录日志方便排查问题
		logs.Infof("match recycle rolling cvm host quota by bizID success, bkBizID: %d,  globalQuota: %d, "+
			"dateRange：%+v, hostCpuMap: %+v, allBizReturnedCpuCore: %d, isContinue: %v, rid: %s", bkBizID, globalQuota,
			dateRange, hostCpuMap, allBizReturnedCpuCore, isContinue, kt.Rid)
		if !isContinue {
			break
		}
	}

	// 检查那些主机已匹配，则给主机Host设置退还方式、回收类型
	for _, host := range hosts {
		// 查询该主机是否已匹配到滚服申请单
		if hostMatchInfo, ok := hostCpuMap[host.IP]; ok && hostMatchInfo.IsMatched {
			// 设置该Host的退还方式
			host.ReturnedWay = hostMatchInfo.ReturnedWay
			host.RecycleType = table.RecycleTypeRollServer
			logs.Infof("check recycle host belongs to roll server project, bkBizID: %d, hostIP: %s, subOrderID: %s, "+
				"deviceType: %s, returnedWay: %s, cpuCore: %d, hostMatchInfo: %+v", host.BizID, host.IP,
				host.SuborderID, host.DeviceType, host.ReturnedWay, host.CpuCore, cvt.PtrToVal(hostMatchInfo))
		}
	}

	return hostCpuMap, hosts, allBizReturnedCpuCore, nil
}

func (l *logics) statisRecycleHostsAndBizCpuSystemQuota(kt *kit.Kit, hosts []*table.RecycleHost) (
	map[string]*rstypes.RecycleHostCpuInfo, error) {

	deviceTypes := make([]string, 0)
	hostCpuMap := make(map[string]*rstypes.RecycleHostCpuInfo, 0)
	for _, host := range hosts {
		deviceTypes = append(deviceTypes, host.DeviceType)
		hostCpuMap[host.IP] = &rstypes.RecycleHostCpuInfo{RecycleHost: host}
	}

	// 根据设备列表获取设备实例信息（CPU核数、机型族）
	deviceInstanceMap, err := l.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("split rolling recycle host, list cvm instance by device type failed, err: %v, "+
			"deviceTypes: %v, rid: %s", err, deviceTypes, kt.Rid)
		return nil, err
	}

	// 2.计算每个回收Host对应的CPU核心数
	for _, hostItem := range hostCpuMap {
		deviceInstanceInfo, ok := deviceInstanceMap[hostItem.DeviceType]
		if !ok {
			logs.Errorf("split rolling recycle host, device type not found, hostIP: %s, deviceType: %s, rid: %s",
				hostItem.IP, hostItem.DeviceType, kt.Rid)
			return nil, errf.Newf(errf.RecordNotFound, "split rolling recycle host, device type not found, "+
				"hostIP: %s, deviceType: %s", hostItem.IP, hostItem.DeviceType)
		}
		hostItem.CpuCore = deviceInstanceInfo.CPUAmount
		hostItem.DeviceGroup = deviceInstanceInfo.DeviceGroup
	}

	return hostCpuMap, nil
}

// MatchRecycleCvmHostQuota 匹配指定时间范围指定业务的主机回收额度
func (l *logics) MatchRecycleCvmHostQuota(kt *kit.Kit, bkBizID int64, startDayNum, endDayNum int,
	allBizReturnedCpuCore, globalQuota int64, hostCpuMap map[string]*rstypes.RecycleHostCpuInfo) (
	map[string]*rstypes.RecycleHostCpuInfo, bool, int64, error) {

	// 1.查询指定时间、指定业务滚服申请跟回收的列表
	appliedRecords, returnedRecordMap, err := l.listRollServerAppliedAndRecycle(kt, bkBizID, startDayNum, endDayNum)
	if err != nil {
		logs.Errorf("query rolling applied and returned records by biz failed, err: %v, bkBizID: %d, rid: %s",
			err, bkBizID, kt.Rid)
		return nil, false, 0, err
	}

	if len(appliedRecords) == 0 {
		return hostCpuMap, true, 0, nil
	}

	// 2.指定时间范围内，该业务剩余可退还的CPU总核心数
	for _, apply := range appliedRecords {
		// 该子订单已退回的CPU总核心数
		var returnedCore int64
		for _, returnedRecord := range returnedRecordMap[apply.ID] {
			returnedCore += cvt.PtrToVal(returnedRecord.MatchAppliedCore)
		}

		// 该主机申请单，是否还有剩余可退还的CPU核心数
		remainCore := cvt.PtrToVal(apply.DeliveredCore) - returnedCore

		if remainCore <= 0 {
			continue
		}

		// 寻找比剩余可退还的CPU核心数，更小的主机
		for _, hostItem := range hostCpuMap {
			// 申请单有剩余CPU核数 + 该主机尚未匹配 + 机型族相同
			if remainCore > 0 && !hostItem.IsMatched && hostItem.CpuCore <= remainCore &&
				apply.InstanceGroup == hostItem.DeviceGroup {

				// 3.计算当前业务可回收的"退还方式"，所有业务的滚服回收核数 > 全局总额度，走“资源池回收”
				returnedWay := enumor.CrpReturnedWay
				allBizReturnedCpuCore += hostItem.CpuCore
				if allBizReturnedCpuCore > globalQuota {
					returnedWay = enumor.ResourcePoolReturnedWay
				}

				hostItem.IsMatched = true
				hostItem.ReturnedWay = returnedWay
				hostItem.AppliedRecordID = apply.ID
				hostItem.MatchAppliedCore = hostItem.CpuCore
				// 扣减剩余可退还的CPU核心数
				remainCore -= hostItem.CpuCore
			}
		}
	}

	// 4.检查主机列表里面，是否还有未匹配的主机
	continueMatched := false
	for _, hostItem := range hostCpuMap {
		if !hostItem.IsMatched {
			continueMatched = true
			break
		}
	}

	hostCpuMapJson, err := json.Marshal(hostCpuMap)
	if err != nil {
		logs.Errorf("match recycle rolling cvm host quota marshal failed, err: %v, bkBizID: %d, rid: %s",
			err, bkBizID, kt.Rid)
		return nil, false, 0, err
	}
	logs.Infof("match recycle rolling cvm host quota end, bkBizID: %d, continueMatched: %v, hostCpuMap: %s, rid: %s",
		bkBizID, continueMatched, hostCpuMapJson, kt.Rid)

	return hostCpuMap, continueMatched, allBizReturnedCpuCore, nil
}

// InsertReturnedHostMatched 插入需要退还的主机匹配记录
func (l *logics) InsertReturnedHostMatched(kt *kit.Kit, bkBizID int64, orderID uint64, subOrderID string,
	hosts []*table.RecycleHost, hostMatchMap map[string]*rstypes.RecycleHostCpuInfo,
	status enumor.ReturnedStatus) error {

	if hostMatchMap == nil {
		logs.Warnf("insert returned host matched skip, hostMatchMap is nil, hosts: %+v, "+
			"hostMatchMap: %+v, rid: %s", hosts, hostMatchMap, kt.Rid)
		return nil
	}

	// 按滚服申请表主键ID、机型族来分组
	appliedCpuMap := make(map[string]map[string]map[enumor.ReturnedWay]*rstypes.RecycleHostCpuInfo, 0)
	for _, host := range hosts {
		hostMatchInfo, exist := hostMatchMap[host.IP]
		if !exist || !hostMatchInfo.IsMatched {
			continue
		}
		if _, exist = appliedCpuMap[hostMatchInfo.AppliedRecordID]; !exist {
			appliedCpuMap[hostMatchInfo.AppliedRecordID] = make(
				map[string]map[enumor.ReturnedWay]*rstypes.RecycleHostCpuInfo, 0)
		}
		if _, exist = appliedCpuMap[hostMatchInfo.AppliedRecordID][host.DeviceGroup]; !exist {
			appliedCpuMap[hostMatchInfo.AppliedRecordID][host.DeviceGroup] = make(
				map[enumor.ReturnedWay]*rstypes.RecycleHostCpuInfo, 0)
		}
		if _, exist = appliedCpuMap[hostMatchInfo.AppliedRecordID][host.DeviceGroup][host.ReturnedWay]; !exist {
			appliedCpuMap[hostMatchInfo.AppliedRecordID][host.DeviceGroup][host.ReturnedWay] =
				&rstypes.RecycleHostCpuInfo{}
		}
		appliedCpuMap[hostMatchInfo.AppliedRecordID][host.DeviceGroup][host.ReturnedWay].MatchAppliedCore +=
			hostMatchInfo.CpuCore
	}

	now := time.Now()
	year, month, day := now.Year(), int(now.Month()), now.Day()

	for appliedRecordID, appliedHostList := range appliedCpuMap {
		records := make([]rsproto.RollingReturnedRecordCreateReq, 0)
		for deviceGroup, returnWayHostMap := range appliedHostList {
			for returnedWay, hostCpuItem := range returnWayHostMap {
				records = append(records, rsproto.RollingReturnedRecordCreateReq{
					BkBizID:          bkBizID,
					OrderID:          orderID,
					SubOrderID:       subOrderID,
					AppliedRecordID:  appliedRecordID,
					MatchAppliedCore: hostCpuItem.MatchAppliedCore,
					Year:             year,
					Month:            month,
					Day:              day,
					ReturnedWay:      returnedWay,
					InstanceGroup:    deviceGroup,
					Status:           status,
				})
			}
		}

		for _, partRecords := range slice.Split(records, constant.RollingServerOperateMaxLimit) {
			createReq := &rsproto.BatchCreateRollingReturnedRecordReq{
				ReturnedRecords: partRecords,
			}
			if _, err := l.client.DataService().Global.RollingServer.CreateReturnedRecord(kt, createReq); err != nil {
				logs.Errorf("insert returned host matched, create returned record failed, err: %v, bkBizID: %d, "+
					"orderID: %d, subOrderID: %s, appliedRecordID: %s, partRecords: %+v, rid: %s", err, bkBizID,
					orderID, subOrderID, appliedRecordID, partRecords, kt.Rid)
				return err
			}
			logs.Infof("insert returned host matched, create returned record success, bkBizID: %d, "+
				"orderID: %d, subOrderID: %s, appliedRecordID: %s, partRecords: %+v, rid: %s", bkBizID,
				orderID, subOrderID, appliedRecordID, partRecords, kt.Rid)
		}
	}
	return nil
}

// CheckReturnedStatusBySubOrderID 校验回收订单是否有滚服剩余额度
func (l *logics) CheckReturnedStatusBySubOrderID(kt *kit.Kit, orders []*table.RecycleOrder) error {
	for _, order := range orders {
		// 如果子订单不是未提交状态，跳过
		if order.Status != table.RecycleStatusUncommit {
			continue
		}

		// 如果子订单不是滚服项目，跳过
		if order.RecycleType != table.RecycleTypeRollServer {
			logs.Warnf("check returned locked status skip, order recycle type is not roll server, subOrderID: %s, rid",
				order.SuborderID, kt.Rid)
			continue
		}

		// 根据子订单查询该子订单的滚服申请、回收记录及对应的CPU核心数
		appliedRecords, returnMatchMap, returnedRecordMap, err := l.listAppliedReturnCpuCoreRecords(kt, order)
		if err != nil {
			logs.Errorf("check returned locked status failed, list applied return cpu core failed, err: %v, "+
				"subOrderID: %s, bkBizID: %d, rid: %s", err, order.SuborderID, order.BizID, kt.Rid)
			return err
		}

		for _, applyItem := range appliedRecords {
			// 该子订单已退回的CPU总核心数
			var returnedCore int64
			for _, returnedRecord := range returnedRecordMap[applyItem.ID] {
				returnedCore += cvt.PtrToVal(returnedRecord.MatchAppliedCore)
			}

			// 该主机申请单，是否还有剩余可退还的CPU核心数
			deliverCore := cvt.PtrToVal(applyItem.DeliveredCore)
			remainCore := deliverCore - returnedCore
			// 如果该主机申请单，需要退回的CPU核心数大于剩余可退还的CPU核心数，则报错
			subOrderIDMatchAppliedID := fmt.Sprintf("%s-%s", order.SuborderID, applyItem.ID)
			if returnMatchMap[subOrderIDMatchAppliedID] > remainCore {
				logs.Errorf("check returned locked status failed, has no rolling server remain quota, "+
					"subOrderID: %s, applyID: %s, deliverCore: %d, returnedCore: %d, remainCore: %d, "+
					"returnMatchMap: %+v, rid: %s", order.SuborderID, applyItem.ID, deliverCore, returnedCore,
					remainCore, returnMatchMap, kt.Rid)
				return errf.Newf(errf.RollingServerRecycleCommitCheckError, "recycle order has no remain quota")
			}
		}
	}
	return nil
}

// listReturnedStatusBySubOrderID 根据子订单查询该子订单的滚服申请、回收记录及对应的CPU核心数
func (l *logics) listAppliedReturnCpuCoreRecords(kt *kit.Kit, order *table.RecycleOrder) (
	[]*rstable.RollingAppliedRecord, map[string]int64, map[string][]*rstable.RollingReturnedRecord, error) {

	returnedList, err := l.listReturnedStatusBySubOrderID(kt, order.BizID, order.SuborderID)
	if err != nil {
		logs.Errorf("update returned locked status failed, list returned record failed, err: %v, subOrderID: %s, "+
			"bkBizID: %d, rid: %s", err, order.SuborderID, order.BizID, kt.Rid)
		return nil, nil, nil, err
	}

	// 如果子订单是滚服项目，却没有滚服回收记录，则报错
	if len(returnedList) == 0 {
		logs.Errorf("update returned locked status failed, has no rolling server returned, subOrderID: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return nil, nil, nil, errf.Newf(errf.RollingServerRecycleCommitCheckError, "recycle order has no rolling server "+
			"no returned record, subOrderID: %s, rid: %s", order.SuborderID, kt.Rid)
	}

	appliedRecords := make([]*rstable.RollingAppliedRecord, 0)
	appliedRecordIDs := make([]string, 0)
	returnMatchMap := make(map[string]int64, 0)
	for _, returnedItem := range returnedList {
		listReq := &rsproto.RollingAppliedRecordListReq{
			Filter: tools.EqualExpression("id", returnedItem.AppliedRecordID),
			Page:   core.NewDefaultBasePage(),
		}
		appliedList, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling applied record failed, err: %v, id: %s, rid: %s", err,
				returnedItem.AppliedRecordID, kt.Rid)
			return nil, nil, nil, err
		}

		// 未查询到滚服申请单
		if len(appliedList.Details) == 0 {
			return nil, nil, nil, errf.Newf(errf.RollingServerRecycleCommitCheckError, "recycle order has no rolling "+
				"server no applied record, subOrderID: %s, appliedID: %s, rid: %s",
				order.SuborderID, returnedItem.AppliedRecordID, kt.Rid)
		}

		// 记录该回收ID对应的回收CPU核心数
		subOrderIDMatchAppliedID := fmt.Sprintf("%s-%s", order.SuborderID, returnedItem.AppliedRecordID)
		returnMatchMap[subOrderIDMatchAppliedID] += cvt.PtrToVal(returnedItem.MatchAppliedCore)
		appliedRecords = append(appliedRecords, appliedList.Details...)
		appliedRecordIDs = append(appliedRecordIDs, returnedItem.AppliedRecordID)
	}

	// 批量查询申请单对应的回收记录列表
	returnedRecordMap, err := l.listReturnedRecordsByAppliedIDs(kt, order.BizID, appliedRecordIDs)
	if err != nil {
		logs.Errorf("query rolling returned records by appliedIDs failed, err: %v, appliedRecordIDs: %v, "+
			"rid: %s", err, appliedRecordIDs, kt.Rid)
		return nil, nil, nil, err
	}

	return appliedRecords, returnMatchMap, returnedRecordMap, nil
}

// UpdateReturnedStatusBySubOrderID 根据回收子订单ID更新滚服回收的状态
func (l *logics) UpdateReturnedStatusBySubOrderID(kt *kit.Kit, bkBizID int64, subOrderID string,
	status enumor.ReturnedStatus) error {

	returnedList, err := l.listReturnedStatusBySubOrderID(kt, bkBizID, subOrderID)
	if err != nil {
		logs.Errorf("update returned locked status failed, list returned record failed, err: %v, subOrderID: %s, "+
			"rid: %s", err, subOrderID, kt.Rid)
		return err
	}

	updateRecords := make([]rsproto.RollingReturnedRecordUpdateReq, 0)
	for _, item := range returnedList {
		updateRecords = append(updateRecords, rsproto.RollingReturnedRecordUpdateReq{
			ID:     item.ID,
			Status: status,
		})
	}
	for _, partRecords := range slice.Split(updateRecords, constant.RollingServerOperateMaxLimit) {
		updateReq := &rsproto.BatchUpdateRollingReturnedRecordReq{
			ReturnedRecords: partRecords,
		}
		err = l.client.DataService().Global.RollingServer.UpdateReturnedRecord(kt, updateReq)
		if err != nil {
			logs.Errorf("update returned locked status failed, update returned record failed, err: %v, "+
				"subOrderID: %s, partRecords: %+v, rid: %s", err, subOrderID, partRecords, kt.Rid)
			return err
		}
	}
	return nil
}

// listReturnedStatusBySubOrderID 根据回收子订单ID获取滚服回收列表
func (l *logics) listReturnedStatusBySubOrderID(kt *kit.Kit, bkBizID int64, subOrderID string) (
	[]*rstable.RollingReturnedRecord, error) {

	query := &rsproto.RollingReturnedRecordListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("suborder_id", subOrderID),
		),
		Page: core.NewDefaultBasePage(),
	}
	returnedList := make([]*rstable.RollingReturnedRecord, 0)
	for {
		returnedRecords, err := l.client.DataService().Global.RollingServer.ListReturnedRecord(kt, query)
		if err != nil {
			logs.Errorf("list returned locked status failed, list returned record failed, err: %v, subOrderID: %s, "+
				"rid: %s", err, subOrderID, kt.Rid)
			return nil, err
		}
		if returnedRecords == nil || len(returnedRecords.Details) == 0 {
			logs.Warnf("list returned locked status skip, list returned record empty, err: %v, subOrderID: %s, "+
				"bkBizID: %d, rid: %s", err, subOrderID, bkBizID, kt.Rid)
			return nil, nil
		}

		returnedList = append(returnedList, returnedRecords.Details...)
		if len(returnedRecords.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		query.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return returnedList, nil
}

// listRollServerAppliedAndRecycle 查询指定时间、指定业务滚服申请跟回收的列表
func (l *logics) listRollServerAppliedAndRecycle(kt *kit.Kit, bkBizID int64, startDayNum, endDayNum int) (
	[]*rstable.RollingAppliedRecord, map[string][]*rstable.RollingReturnedRecord, error) {

	now := time.Now()
	year, month, day := now.Year(), int(now.Month()), now.Day()

	// 1.查询当前时间之前121天的滚服申请记录
	appliedRecords, err := l.listAppliedRecordsByDate(kt, bkBizID, year, month, day, startDayNum, endDayNum)
	if err != nil {
		logs.Errorf("query rolling applied records by date failed, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, nil, err
	}

	// 2.根据step1里的滚服申请记录的唯一标识，匹配滚服回收执行记录表里的数据，得到该子订单单目前对应的退还记录
	appliedRecordIDs := make([]string, len(appliedRecords))
	for i, appliedRecord := range appliedRecords {
		appliedRecordIDs[i] = appliedRecord.ID
	}
	returnedRecordMap, err := l.listReturnedRecordsByAppliedIDs(kt, bkBizID, appliedRecordIDs)
	if err != nil {
		logs.Errorf("query rolling returned records by appliedIDs failed, err: %v, appliedRecords: %v, rid: %s", err,
			appliedRecordIDs, kt.Rid)
		return nil, nil, err
	}
	return appliedRecords, returnedRecordMap, nil
}

// listAppliedRecordsByDate 查询指定时间之前N天的滚服申请记录
func (l *logics) listAppliedRecordsByDate(kt *kit.Kit, bkBizID int64, year, month, day, startDayNum, endDayNum int) (
	[]*rstable.RollingAppliedRecord, error) {

	// 查询121天内，该业务的申请记录
	startYear, startMonth, startDay := subDay(year, month, day, startDayNum)
	startRollDate := times.GetDataIntDate(startYear, startMonth, startDay)
	endYear, endMonth, endDay := subDay(year, month, day, endDayNum)
	endRollDate := times.GetDataIntDate(endYear, endMonth, endDay)
	records := make([]*rstable.RollingAppliedRecord, 0)
	listReq := &rsproto.RollingAppliedRecordListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bkBizID},
				&filter.AtomRule{Field: "applied_type", Op: filter.Equal.Factory(), Value: enumor.NormalAppliedType},
				// 大于等于该日期的回收记录
				&filter.AtomRule{Field: "roll_date", Op: filter.GreaterThanEqual.Factory(), Value: startRollDate},
				// 小于等于该日期的回收记录
				&filter.AtomRule{Field: "roll_date", Op: filter.LessThanEqual.Factory(), Value: endRollDate},
			},
		},
		Page: core.NewDefaultBasePage(),
	}

	for {
		result, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return nil, err
		}
		records = append(records, result.Details...)

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return records, nil
}

// listReturnedRecordsByAppliedIDs 查询指定滚服申请ID的回收记录
func (l *logics) listReturnedRecordsByAppliedIDs(kt *kit.Kit, bkBizID int64, appliedRecordIDs []string) (
	map[string][]*rstable.RollingReturnedRecord, error) {

	recordMap := make(map[string][]*rstable.RollingReturnedRecord)
	for _, ids := range slice.Split(appliedRecordIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &rsproto.RollingReturnedRecordListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bkBizID},
					&filter.AtomRule{Field: "applied_record_id", Op: filter.In.Factory(), Value: ids},
					&filter.AtomRule{Field: "status", Op: filter.Equal.Factory(), Value: enumor.NormalStatus},
				},
			},
			Page: core.NewDefaultBasePage(),
		}

		for {
			result, err := l.client.DataService().Global.RollingServer.ListReturnedRecord(kt, listReq)
			if err != nil {
				logs.Errorf("list rolling returned record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
				return nil, err
			}

			for _, record := range result.Details {
				if _, ok := recordMap[record.AppliedRecordID]; !ok {
					recordMap[record.AppliedRecordID] = make([]*rstable.RollingReturnedRecord, 0)
				}
				recordMap[record.AppliedRecordID] = append(recordMap[record.AppliedRecordID], record)
			}

			if len(result.Details) < int(core.DefaultMaxPageLimit) {
				break
			}

			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return recordMap, nil
}

// GetAllReturnedCpuCore 获取指定时间内所有业务回收的CPU总核心数
func (l *logics) GetAllReturnedCpuCore(kt *kit.Kit) (int64, error) {
	now := time.Now()
	year, month, day := now.Year(), int(now.Month()), now.Day()

	// 查询121天内，所有滚服回收记录已退回的CPU总核心数
	startYear, startMonth, startDay := subDay(year, month, day, rstypes.CalculateFineEndDay)
	startRollDate := times.GetDataIntDate(startYear, startMonth, startDay)
	endRollDate := times.GetDataIntDate(year, month, day)

	listReq := &rsproto.RollingReturnedRecordListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				// 大于等于该日期的回收记录
				&filter.AtomRule{Field: "roll_date", Op: filter.GreaterThanEqual.Factory(), Value: startRollDate},
				// 小于等于该日期的回收记录
				&filter.AtomRule{Field: "roll_date", Op: filter.LessThanEqual.Factory(), Value: endRollDate},
				tools.RuleEqual("status", enumor.NormalStatus),
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := l.client.DataService().Global.RollingServer.GetRollingReturnedCoreSum(kt, listReq)
	if err != nil {
		logs.Errorf("get rolling returned core sum match_applied_core failed, err: %v, startRollDate: %d, "+
			"endRollDate: %d, rid: %s", err, startRollDate, endRollDate, kt.Rid)
		return 0, err
	}
	return result.SumReturnedAppliedCore, nil
}

// GetRollingGlobalQuota 查询系统配置的全局总额度
func (l *logics) GetRollingGlobalQuota(kt *kit.Kit) (int64, error) {
	systemGlobalConfig, err := l.getRollingGlobalConfig(kt, []string{"global_quota"})
	if err != nil {
		logs.Errorf("query rolling recycle global quota failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	// 全局总额度尚未配置
	globalQuota := cvt.PtrToVal(systemGlobalConfig.GlobalQuota)
	if globalQuota <= 0 {
		return 0, errf.Newf(errf.RecordNotFound, "rolling global quota has not config")
	}
	return globalQuota, nil
}

func (l *logics) getRollingGlobalConfig(kt *kit.Kit, fields []string) (rstable.RollingGlobalConfigTable, error) {
	listReq := &rsproto.RollingGlobalConfigListReq{
		Filter: tools.AllExpression(),
		Fields: fields,
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}
	result, err := l.client.DataService().Global.RollingServer.ListGlobalConfig(kt, listReq)
	if err != nil {
		logs.Errorf("list rolling global config failed, err: %v, fields: %v, rid: %s", err, fields, kt.Rid)
		return rstable.RollingGlobalConfigTable{}, err
	}

	if len(result.Details) == 0 {
		logs.Errorf("can not find rolling global config, fields: %v, rid:%s", fields, kt.Rid)
		return rstable.RollingGlobalConfigTable{}, errors.New("can not find rolling global config")
	}

	return result.Details[0], nil
}

// ListReturnedRecordsBySubOrderID 根据回收子订单ID查询滚服回收的记录
func (l *logics) ListReturnedRecordsBySubOrderID(kt *kit.Kit, bkBizID int64, subOrderID string) (
	[]*rstable.RollingReturnedRecord, error) {

	query := &rsproto.RollingReturnedRecordListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("suborder_id", subOrderID),
			tools.RuleNotEqual("status", enumor.TerminateStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	returnedRecords, err := l.client.DataService().Global.RollingServer.ListReturnedRecord(kt, query)
	if err != nil {
		logs.Errorf("list rolling returned record failed, err: %v, bkBizID: %d, subOrderID: %s, rid: %s",
			err, bkBizID, subOrderID, kt.Rid)
		return nil, err
	}
	if returnedRecords == nil || len(returnedRecords.Details) == 0 {
		return nil, nil
	}

	return returnedRecords.Details, nil
}
