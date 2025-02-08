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
	"context"
	"errors"
	"fmt"
	"time"

	dtime "hcm/cmd/woa-server/logics/plan/demand-time"
	mtypes "hcm/cmd/woa-server/types/meta"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	dmtypes "hcm/pkg/dal/dao/types/meta"
	rtypes "hcm/pkg/dal/dao/types/resource-plan"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// generatePenaltyBase generate penalty base on every monday.
func (c *Controller) generatePenaltyBase(ctx context.Context) {
	now := time.Now()
	logs.Infof("start to generate penalty base, time: %v", now)

	// 每周一计算上周的罚金基数
	// 计算下一个周一凌晨的时间
	nextMonday := times.GetNextMondayOfWeek(now)
	nextRunTime := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(),
		0, 0, 0, 0, nextMonday.Location())

	// 判断上周的罚金基数是否已经生成，没有需要先补上周的
	kt := core.NewBackendKit()
	days12After := now.AddDate(0, 0, 12*7)
	yearMonthWeek12After, err := c.demandTime.GetDemandYearMonthWeek(kt, days12After)
	if err != nil {
		logs.Errorf("%s: failed to get year month week, err: %v, demand_date: %s, rid: %s",
			constant.DemandPenaltyBaseGenerateFailed, err, days12After.String(), kt.Rid)
	}

	exists, err := c.isPenaltyBaseExists(kt, yearMonthWeek12After.Year, yearMonthWeek12After.YearWeek)
	if err != nil {
		logs.Errorf("%s: failed to check penalty base exists, err: %v, year: %d, week: %d, rid: %s",
			constant.DemandPenaltyBaseGenerateFailed, err, yearMonthWeek12After.Year, yearMonthWeek12After.YearWeek,
			kt.Rid)
	}

	if err == nil && !exists {
		// 补上周的罚金基数
		thisMonday := times.GetMondayOfWeek(now)
		ticketEnd := time.Date(thisMonday.Year(), thisMonday.Month(), thisMonday.Day(), 0, 0, 0, 0,
			thisMonday.Location())

		err := c.CreatePenaltyBaseFromTicket(kt, []int64{}, ticketEnd,
			c.demandTime.GetDemandDateRangeInWeek(kt, days12After), yearMonthWeek12After)
		if err != nil {
			logs.Errorf("%s: failed to create penalty base from ticket, err: %v, year_month_week: %+v, rid: %s",
				constant.DemandPenaltyBaseGenerateFailed, err, yearMonthWeek12After, kt.Rid)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 等待到下一个检查时间
		time.Sleep(time.Until(nextRunTime))

		kt = core.NewBackendKit()
		days12After = nextRunTime.AddDate(0, 0, 12*7)
		yearMonthWeek12After, err = c.demandTime.GetDemandYearMonthWeek(kt, days12After)
		if err != nil {
			logs.Errorf("%s: failed to get year month week, err: %v, demand_date: %s, rid: %s",
				constant.DemandPenaltyBaseGenerateFailed, err, days12After.String(), kt.Rid)
		}

		// 计算罚金基数
		err := c.CreatePenaltyBaseFromTicket(kt, []int64{}, nextRunTime,
			c.demandTime.GetDemandDateRangeInWeek(kt, days12After), yearMonthWeek12After)
		if err != nil {
			logs.Errorf("%s: failed to create penalty base from ticket, err: %v, year_month_week: %+v, rid: %s",
				constant.DemandPenaltyBaseGenerateFailed, err, yearMonthWeek12After, kt.Rid)
		}

		// 计算下一个检查时间
		nextRunTime = nextRunTime.AddDate(0, 0, 7)
	}
}

// calcAndReportPenaltyRatioToCRP CRP每月1号凌晨出上个月的账单
// 因此每月最后7天，每天下午18:00计算当月罚金分摊比例并推送到CRP
func (c *Controller) calcAndReportPenaltyRatioToCRP(ctx context.Context) {
	now := time.Now()
	logs.Infof("start to push penalty ratio to crp, time: %v", now)

	// 每天下午18:00计算并推送当月罚金分摊比例
	// 计算下次推送的时间
	nextRunTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	if now.After(nextRunTime) {
		nextRunTime = nextRunTime.Add(time.Hour * 24)
	}

	// 首次启动直接推送一次本月的罚金分摊比例
	kt := core.NewBackendKit()
	err := c.CalcPenaltyRatioAndPush(kt, now)
	if err != nil {
		logs.Errorf("%s: failed to calc and push penalty ratio to crp, err: %v, time: %s, rid: %s",
			constant.DemandPenaltyRatioReportFailed, err, now.Format(constant.DateTimeLayout), kt.Rid)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logs.Infof("push penalty ratio to crp, next run time: %v", nextRunTime)
		// 等待到下一个检查时间
		time.Sleep(time.Until(nextRunTime))

		// CRP每月1号凌晨出上个月的账单，且每次推送都会覆盖上次推送的内容
		// 我们只需要在最后7天饱和式推送
		isLast7Day := times.IsLastNDaysOfMonth(now, 7)
		if !isLast7Day {
			// 计算下一个检查时间
			nextRunTime = nextRunTime.Add(time.Hour * 24)
			continue
		}

		kt = core.NewBackendKit()
		err := c.CalcPenaltyRatioAndPush(kt, nextRunTime)
		if err != nil {
			logs.Errorf("%s: failed to calc and push penalty ratio to crp, err: %v, time: %s, rid: %s",
				constant.DemandPenaltyRatioReportFailed, err, nextRunTime.Format(constant.DateTimeLayout), kt.Rid)
		}

		// 计算下一个检查时间
		nextRunTime = nextRunTime.Add(time.Hour * 24)
	}
}

func (c *Controller) isPenaltyBaseExists(kt *kit.Kit, year int, penaltyWeek int) (bool, error) {

	listReq := &rpproto.DemandPenaltyBaseListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("year", year),
				tools.RuleEqual("year_week", penaltyWeek),
			),
			Page: core.NewCountPage(),
		},
	}

	rst, err := c.client.DataService().Global.ResourcePlan.ListDemandPenaltyBase(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list demand penalty base, err: %v, year: %d, week: %d, rid: %s", err, year,
			penaltyWeek, kt.Rid)
		return false, err
	}

	if rst.Count > 0 {
		return true, nil
	}

	return false, nil
}

// CalcPenaltyBase 计算罚金分摊基数
func (c *Controller) CalcPenaltyBase(kt *kit.Kit, baseDay time.Time, bkBizIDs []int64) error {
	baseDayYearMonthWeek, err := c.demandTime.GetDemandYearMonthWeek(kt, baseDay)
	if err != nil {
		logs.Errorf("failed to get demand year month week, err: %v, base_day: %s, rid: %s", err,
			baseDay.String(), kt.Rid)
		return err
	}

	// 单据只看12周前的
	days12Before := baseDay.AddDate(0, 0, -12*7)
	monday12Before := times.GetMondayOfWeek(days12Before)

	err = c.CreatePenaltyBaseFromTicket(kt, bkBizIDs, monday12Before,
		c.demandTime.GetDemandDateRangeInWeek(kt, baseDay), baseDayYearMonthWeek)
	if err != nil {
		logs.Errorf("failed to create penalty base from ticket, err: %v, base_day: %s, rid: %s", err,
			baseDay.String(), kt.Rid)
		return err
	}

	return nil
}

// CreatePenaltyBaseFromTicket 从历史单据还原罚金分摊基数，即预测总量
// 为了简化计算，只考虑2个月内的单据，不考虑用户早于22周提预测的情况
func (c *Controller) CreatePenaltyBaseFromTicket(kt *kit.Kit, bkBizIDs []int64, ticketEnd time.Time,
	baseTimeRange times.DateRange, baseYearWeek dtime.DemandYearMonthWeek) error {

	start := time.Now()
	logs.Infof("start create penalty base from ticket, base_year_week: %+v, time: %v, rid: %s", baseYearWeek,
		start, kt.Rid)
	// 捞取时间范围内的所有订单
	ticketStart := ticketEnd.AddDate(0, -2, 0)
	listTicketRules := make([]*filter.AtomRule, 0)
	listTicketRules = append(listTicketRules,
		tools.RuleGreaterThanEqual("submitted_at", ticketStart.Format(constant.TimeStdFormat)))
	listTicketRules = append(listTicketRules,
		tools.RuleLessThanEqual("submitted_at", ticketEnd.Format(constant.TimeStdFormat)))
	if len(bkBizIDs) > 0 {
		listTicketRules = append(listTicketRules, tools.RuleIn("bk_biz_id", bkBizIDs))
	}
	listTicketFilter := tools.ExpressionAnd(listTicketRules...)
	allTickets, err := c.listAllResPlanTicket(kt, listTicketFilter)
	if err != nil {
		logs.Errorf("failed to list all res plan ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 从 woa_zone 获取大区和机型对应关系 metadata
	zoneMap, regionAreaMap, deviceTypeMap, err := c.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	_, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)
	// 从订单中捞取预测内的核心总数，按业务、大区、机型族合并
	baseCoreMap, bizOrgRelMap, err := c.calcPenaltyBaseCoreByTicket(kt, allTickets, baseTimeRange, regionNameMap,
		deviceTypeMap)
	if err != nil {
		logs.Errorf("failed to calc penalty base core by ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 本次计算为全量，为避免重复，先清理数据库中残留的数据
	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("year", baseYearWeek.Year),
			tools.RuleEqual("year_week", baseYearWeek.YearWeek),
		),
	}
	err = c.client.DataService().Global.ResourcePlan.DeleteDemandPenaltyBase(kt, deleteReq)
	if err != nil {
		logs.Errorf("failed to delete old demand penalty base, err: %v, base year: %d, base week: %d, rid: %s",
			err, baseYearWeek.Year, baseYearWeek.YearWeek, kt.Rid)
		return err
	}

	// 插入数据
	createIDs, err := c.createDemandPenaltyBase(kt, baseYearWeek, baseCoreMap, bizOrgRelMap)
	if err != nil {
		logs.Errorf("failed to create demand penalty base, err: %v, base year: %d, base week: %d, rid: %s",
			err, baseYearWeek.Year, baseYearWeek.YearWeek, kt.Rid)
		return err
	}

	end := time.Now()
	logs.Infof("end create penalty base from ticket, base_year_week: %+v, created_id: %v, time: %v, cost: %ds, rid: %s",
		baseYearWeek, createIDs, end, end.Sub(start).Seconds(), kt.Rid)
	return nil
}

func (c *Controller) createDemandPenaltyBase(kt *kit.Kit, baseYearWeek dtime.DemandYearMonthWeek,
	baseCoreMap map[ptypes.DemandPenaltyBaseKey]int64, bizOrgRelMap map[int64]mtypes.BizOrgRel) ([]string, error) {

	if len(baseCoreMap) == 0 {
		logs.Infof("no data to create penalty base, base year week: %+v, rid: %s", baseYearWeek, kt.Rid)
		return nil, nil
	}

	penaltyBaseCreateReqs := make([]rpproto.DemandPenaltyBaseCreate, 0, len(baseCoreMap))
	for key, val := range baseCoreMap {
		bizOrgRel := bizOrgRelMap[key.BkBizID]
		baseCore := max(val, 0)
		penaltyBaseCreateReqs = append(penaltyBaseCreateReqs, rpproto.DemandPenaltyBaseCreate{
			Year:            baseYearWeek.Year,
			Month:           int(baseYearWeek.Month),
			Week:            baseYearWeek.Week,
			YearWeek:        baseYearWeek.YearWeek,
			Source:          enumor.DemandPenaltyBaseSourceLocal,
			BkBizID:         key.BkBizID,
			BkBizName:       bizOrgRel.BkBizName,
			OpProductID:     bizOrgRel.OpProductID,
			OpProductName:   bizOrgRel.OpProductName,
			PlanProductID:   bizOrgRel.PlanProductID,
			PlanProductName: bizOrgRel.PlanProductName,
			VirtualDeptID:   bizOrgRel.VirtualDeptID,
			VirtualDeptName: bizOrgRel.VirtualDeptName,
			AreaName:        key.AreaName,
			DeviceFamily:    key.DeviceFamily,
			CpuCore:         &baseCore,
		})
	}
	createReq := &rpproto.DemandPenaltyBaseCreateReq{
		PenaltyBases: penaltyBaseCreateReqs,
	}
	rst, err := c.client.DataService().Global.ResourcePlan.BatchCreateDemandPenaltyBase(kt, createReq)
	if err != nil {
		logs.Errorf("failed to create demand penalty base, err: %v, base_year_week: %+v, rid: %s", err,
			baseYearWeek, kt.Rid)
		return nil, err
	}

	return rst.IDs, nil
}

// calcPenaltyBaseCoreByTicket calculate penalty base core by ticket.
func (c *Controller) calcPenaltyBaseCoreByTicket(kt *kit.Kit, tickets []rtypes.RPTicketWithStatus,
	timeRange times.DateRange, regionNameMap map[string]dmtypes.RegionArea,
	deviceTypeMap map[string]wdt.WoaDeviceTypeTable) (map[ptypes.DemandPenaltyBaseKey]int64,
	map[int64]mtypes.BizOrgRel, error) {
	baseCoreMap := make(map[ptypes.DemandPenaltyBaseKey]int64)
	bizOrgRelMap := make(map[int64]mtypes.BizOrgRel)
	for _, ticket := range tickets {
		if ticket.Status != enumor.RPTicketStatusDone {
			continue
		}
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
		bizOrgRelMap[ticket.BkBizID] = bizOrgRel
		changes, err := c.QueryCrpOrderChangeInfo(kt, ticket.CrpSn)
		if err != nil {
			logs.Errorf("failed to query crp order change info, err: %v, ticket_id: %s, crp_cn: %s, rid: %s",
				err, ticket.ID, ticket.CrpSn, kt.Rid)
			return nil, nil, err
		}
		for _, change := range changes {
			// 只关注落在时间范围内的、预测内的变更
			if change.ExpectTime < timeRange.Start || change.ExpectTime > timeRange.End {
				continue
			}
			if change.PlanType != enumor.PlanTypeCodeInPlan {
				continue
			}
			regionAreaInfo, ok := regionNameMap[change.RegionName]
			if !ok {
				logs.Errorf("failed to get region's area info, region_name: %s, rid: %s", change.RegionName,
					kt.Rid)
				return nil, nil, err
			}
			deviceTypeInfo, ok := deviceTypeMap[change.DeviceType]
			if !ok {
				logs.Errorf("failed to get device type info, device_type: %s, rid: %s", change.DeviceType,
					kt.Rid)
				return nil, nil, err
			}
			key := ptypes.DemandPenaltyBaseKey{
				BkBizID:      ticket.BkBizID,
				AreaName:     regionAreaInfo.AreaName,
				DeviceFamily: deviceTypeInfo.DeviceFamily,
			}
			baseCoreMap[key] += change.ChangeCpuCore
		}
	}
	return baseCoreMap, bizOrgRelMap, nil
}

// CalcPenaltyRatioAndPush calc penalty ratio with unexecuted cpu core
func (c *Controller) CalcPenaltyRatioAndPush(kt *kit.Kit, baseTime time.Time) error {
	start := time.Now()
	logs.Infof("start calc and push penalty ratio to crp, base_time: %v, time: %v, rid: %s", baseTime,
		start, kt.Rid)

	year, month, err := c.demandTime.GetDemandYearMonth(kt, baseTime)
	if err != nil {
		logs.Errorf("failed to get demand year month, err: %v, base_time: %s, rid: %s", err, baseTime.String(),
			kt.Rid)
		return err
	}
	// 1.获取当月预测申请基准核数
	listFilter := tools.ExpressionAnd(
		tools.RuleEqual("year", year),
		tools.RuleEqual("month", month),
	)
	penaltyBaseMap, bkBizIDs, err := c.listPenaltyBaseCore(kt, listFilter)
	if err != nil {
		logs.Errorf("failed to list penalty base core, err: %v, year: %d, month: %d, rid: %s", err, year, month,
			kt.Rid)
		return err
	}

	// 2.从主机申领记录中，获取当月预测消耗总核数
	planAppliedCore, err := c.GetBizResPlanAppliedCPUCore(kt, bkBizIDs, baseTime)
	if err != nil {
		logs.Errorf("failed to get biz res plan applied cpu core, err: %v, bk_biz_ids: %v, base_time: %s, rid: %s",
			err, bkBizIDs, baseTime, kt.Rid)
		return err
	}

	// 3.获取业务的运营产品&规划产品
	bizOrgRelMap := make(map[int64]*mtypes.BizOrgRel)
	for _, bkBizID := range bkBizIDs {
		rel, err := c.bizLogics.GetBizOrgRel(kt, bkBizID)
		if err != nil {
			logs.Errorf("failed to get biz org rel, err: %v, bk biz id: %d, rid: %s", err, bkBizID, kt.Rid)
			return err
		}
		bizOrgRelMap[bkBizID] = rel
	}

	// 4.计算运营产品下未执行到80%的核数，并收敛到规划产品
	planProductUnexecMap := make(map[int64]map[int64]int64)
	for key, baseCore := range penaltyBaseMap {
		bizOrg, ok := bizOrgRelMap[key.BkBizID]
		if !ok {
			logs.Errorf("failed to get bk_biz_id's org rel, bk_biz_id: %d, rid: %s", key.BkBizID, kt.Rid)
			return err
		}

		planProductID := bizOrg.PlanProductID
		if _, ok := planProductUnexecMap[planProductID]; !ok {
			planProductUnexecMap[planProductID] = make(map[int64]int64)
		}
		opProductID := bizOrg.OpProductID
		// 需执行量按80%计算，小数部分忽略不计
		planProductUnexecMap[planProductID][opProductID] += int64(float64(baseCore) * 0.8)
		planProductUnexecMap[planProductID][opProductID] -= planAppliedCore[key]
	}

	// 5.推送罚金比例到CRP
	err = c.pushPenaltyRatioToCRP(kt, planProductUnexecMap, fmt.Sprintf("%04d-%02d", year, month))
	if err != nil {
		logs.Errorf("failed to push penalty ratio to crp, err: %v, base_time: %s, year: %d, month: %d, rid: %s",
			err, baseTime.Format(constant.DateTimeLayout), year, month, kt.Rid)
		return err
	}

	end := time.Now()
	logs.Infof("end calc and push penalty ratio to crp, base_time: %v, time: %v, cost: %ds, rid: %s", baseTime,
		end, end.Sub(start).Seconds(), kt.Rid)
	return nil
}

// GetBizResPlanAppliedCPUCore 获取业务预测已申领的核数，按需求月维度统计，入参需提供该需求月内的任意时间
func (c *Controller) GetBizResPlanAppliedCPUCore(kt *kit.Kit, bkBizIDs []int64, baseTime time.Time) (
	map[ptypes.DemandPenaltyBaseKey]int64, error) {

	// 从 woa_zone 获取大区和机型对应关系 metadata
	_, regionAreaMap, deviceTypeMap, err := c.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	timeRange, err := c.demandTime.GetDemandDateRangeInMonth(kt, baseTime)
	if err != nil {
		logs.Errorf("failed to get demand date range in month, err: %v, base_time: %s, rid: %s", err,
			baseTime.String(), kt.Rid)
		return nil, err
	}
	startDay, endDay, err := timeRange.GetTimeDate()
	if err != nil {
		logs.Errorf("failed to parse date range, err: %v, date range: %s - %s, rid: %s", err, timeRange.Start,
			timeRange.End, kt.Rid)
		return nil, err
	}
	prodConsumePool, err := c.GetProdResConsumePoolV2(kt, bkBizIDs, startDay, endDay)
	if err != nil {
		logs.Errorf("failed to get biz resource consume pool v2, bkBizIDs: %v, err: %v, rid: %s", bkBizIDs, err,
			kt.Rid)
		return nil, err
	}
	planAppliedCore, err := convResConsumePoolToPenaltyMap(kt, prodConsumePool, regionAreaMap, deviceTypeMap)
	if err != nil {
		logs.Errorf("failed to convert res consume pool to penalty map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return planAppliedCore, nil
}

// convResConsumePoolToPenaltyMap 将 ResConsumePool 转为以 DemandPenaltyBaseKey 为 key 的 map
// 因为 ResConsumePool 精确指定了deviceType，因此在list时无法进行模糊匹配，需要进行转化后使用
func convResConsumePoolToPenaltyMap(kt *kit.Kit, pool ResPlanConsumePool, regionAreaMap map[string]dmtypes.RegionArea,
	deviceTypes map[string]wdt.WoaDeviceTypeTable) (map[ptypes.DemandPenaltyBaseKey]int64, error) {

	consumeMap := make(map[ptypes.DemandPenaltyBaseKey]int64)

	for key, cpuCore := range pool {
		if _, ok := regionAreaMap[key.RegionID]; !ok {
			logs.Errorf("failed to get region area, region id: %s, rid: %s", key.RegionID, kt.Rid)
			return nil, fmt.Errorf("failed to get region area, region id: %s", key.RegionID)
		}

		if _, ok := deviceTypes[key.DeviceType]; !ok {
			logs.Errorf("failed to get device type, device type: %s, rid: %s", key.DeviceType, kt.Rid)
			return nil, fmt.Errorf("failed to get device type, device type: %s", key.DeviceType)
		}

		penaltyKey := ptypes.DemandPenaltyBaseKey{
			BkBizID:      key.BkBizID,
			AreaName:     regionAreaMap[key.RegionID].AreaName,
			DeviceFamily: deviceTypes[key.DeviceType].DeviceFamily,
		}

		consumeMap[penaltyKey] += cpuCore
	}

	return consumeMap, nil
}

// listPenaltyBaseCore list penalty base core
func (c *Controller) listPenaltyBaseCore(kt *kit.Kit, listFilter *filter.Expression) (
	map[ptypes.DemandPenaltyBaseKey]int64, []int64, error) {

	penaltyBaseMap := make(map[ptypes.DemandPenaltyBaseKey]int64)
	bkBizIDs := make([]int64, 0)

	listReq := &rpproto.DemandPenaltyBaseListReq{
		ListReq: core.ListReq{
			Filter: listFilter,
			Page:   core.NewDefaultBasePage(),
		},
	}

	for {
		rst, err := c.client.DataService().Global.ResourcePlan.ListDemandPenaltyBase(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list demand penalty base, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		for _, detail := range rst.Details {
			baseKey := ptypes.DemandPenaltyBaseKey{
				BkBizID:      detail.BkBizID,
				AreaName:     detail.AreaName,
				DeviceFamily: detail.DeviceFamily,
			}
			penaltyBaseMap[baseKey] += cvt.PtrToVal(detail.CpuCore)
			bkBizIDs = append(bkBizIDs, detail.BkBizID)
		}

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	uqBizIDs := slice.Unique(bkBizIDs)

	return penaltyBaseMap, uqBizIDs, nil
}

// pushPenaltyRatioToCRP push penalty ratio to CRP
// planProductRatioMap: planProductID -> opProductID -> unexecutedCpuCore
// yearMonth: yearMonth, eg: 2024-11
func (c *Controller) pushPenaltyRatioToCRP(kt *kit.Kit, planProductRatioMap map[int64]map[int64]int64,
	yearMonth string) error {

	ratios := make([]cvmapi.CvmCbsPlanProductRatio, 0)

	for planProductID, opProductRatioMap := range planProductRatioMap {
		// 用量超过80%时可能会出现负数
		for opProductID, unexcutedCPUCore := range opProductRatioMap {
			if unexcutedCPUCore < 0 {
				opProductRatioMap[opProductID] = 0
			}
		}

		ratios = append(ratios, cvmapi.CvmCbsPlanProductRatio{
			// 必须提供一个空的，否则CRP接口会报空指针
			GroupDeptId:           []int64{},
			GroupPlanProductId:    []int64{planProductID},
			ProductIdPartitionMap: opProductRatioMap,
			Memo:                  "",
		})
	}

	pushReq := &cvmapi.CvmCbsPlanPenaltyRatioReportReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanPenaltyRatioReportMethod,
		},
		Params: &cvmapi.CvmCbsPlanPenaltyRatioReportParam{
			YearMonth: yearMonth,
			Data:      ratios,
		},
	}

	resp, err := c.crpCli.ReportPenaltyRatio(kt.Ctx, kt.Header(), pushReq)
	if err != nil {
		logs.Errorf("failed to report penalty ratio to crp, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to report penalty ratio to crp, code: %d, msg: %s, crp_trace: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId, kt.Rid)
		return errors.New(resp.Error.Message)
	}

	return nil
}
