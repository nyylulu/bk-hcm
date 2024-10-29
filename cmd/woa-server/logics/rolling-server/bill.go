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
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	rs "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
)

// syncBillsPeriodically 每天凌晨1点计算罚金
// obs拉取的是t-1的数据，假设现在要拉取11号的数据，那么obs会在12号11点的时候进行拉取，所以我们需要在此之前准备好数据
func (l *logics) syncBillsPeriodically() {
	// 计算下一个凌晨1点的时间
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		// 如果现在已经过了1点，计算明天的1点, 并判断今天的滚服账单是否已经同步完成
		nextRun = nextRun.Add(24 * time.Hour)

		kt := kit.New()
		req := &rollingserver.RollingBillSyncReq{
			BkBizID: rollingserver.SyncAllBiz,
			Year:    now.Year(),
			Month:   int(now.Month()),
			Day:     now.Day(),
		}
		exist, err := l.isBillExist(kit.New(), req)
		if err != nil {
			logs.Errorf("%s: unable to confirm whether rolling bill exists, err: %v, req: %+v, rid: %s",
				constant.RollingServerSyncFailed, err, *req, kt.Rid)
		}
		if !exist {
			logs.Errorf("%s: can not find all biz rolling bill, year: %s, month: %+v, day: %s, rid: %s",
				constant.RollingServerSyncFailed, now.Year(), now.Month(), now.Day(), kt.Rid)
		}
	}

	// 等待直到下一个凌晨1点
	time.Sleep(time.Until(nextRun))

	for {
		kt := kit.New()
		now = time.Now()
		req := &rollingserver.RollingBillSyncReq{
			BkBizID: rollingserver.SyncAllBiz,
			Year:    now.Year(),
			Month:   int(now.Month()),
			Day:     now.Day(),
		}
		if err := l.SyncBills(kt, req); err != nil {
			logs.Errorf("sync all biz rolling bill failed, err: %v, req: %v, rid: %s", err, *req, kt.Rid)
		}

		// 计算下一个 1 点
		nextRun = nextRun.Add(24 * time.Hour)

		// 计算下次执行前的等待时间
		now = time.Now()
		if nextRun.After(now) {
			time.Sleep(time.Until(nextRun)) // 等待直到下一个1点
		}
	}
}

// SyncBills sync rolling server bills
func (l *logics) SyncBills(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) error {
	resPoolBizMap, err := l.listResPoolBizIDs(kt)
	if err != nil {
		logs.Errorf("list rolling resource pool business failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if req.BkBizID != rollingserver.SyncAllBiz {
		if _, ok := resPoolBizMap[req.BkBizID]; ok {
			logs.Infof("skip resource pool business rolling bill sync, bizID: %d, rid: %s", req.BkBizID, kt.Rid)
			return nil
		}

		if err := l.syncBizBills(kt, req); err != nil {
			logs.Errorf("sync biz rolling bill failed, err: %v, bizID: %d, rid: %s", err, req.BkBizID, kt.Rid)
			return err
		}
	}

	start := time.Now()
	logs.Infof("--- start sync all biz rolling bill, start time: %v, rid: %s ---", start, kt.Rid)

	bizIDs, err := l.listIEGBizIDs(kt)
	if err != nil {
		logs.Errorf("list ieg biz ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, bizID := range bizIDs {
		if _, ok := resPoolBizMap[bizID]; ok {
			logs.Infof("skip resource pool business rolling bill sync, bizID: %d, rid: %s", bizID, kt.Rid)
			continue
		}

		subReq := &rollingserver.RollingBillSyncReq{BkBizID: bizID, Year: req.Year, Month: req.Month, Day: req.Day}
		if err = l.syncBizBills(kt, subReq); err != nil {
			logs.Errorf("%s:sync biz rolling bill failed, err: %v, bizID: %d, rid: %s",
				constant.RollingServerSyncFailed, err, bizID, kt.Rid)
			continue
		}
	}

	end := time.Now()
	logs.Infof("--- end sync all biz rolling bill, end time: %v, cost: %v, rid: %s ---", end, end.Sub(start), kt.Rid)

	return nil
}

func (l *logics) syncBizBills(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) error {
	start := time.Now()
	logs.Infof("--- start sync biz(%d) rolling bill, start time: %v, rid: %s ---", req.BkBizID, start, kt.Rid)

	if err := req.Validate(); err != nil {
		logs.Errorf("validate rolling server bills param failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1.查询需要计算罚金的滚服申请记录表数据
	appliedRecords, err := l.findBillAppliedRecords(kt, req)
	if err != nil {
		logs.Errorf("find rolling applied records failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	// 2.根据step1里的滚服申请记录的唯一标识，匹配滚服回收执行记录表里的数据，得到该子单目前对应的退还记录
	appliedRecordIDs := make([]string, len(appliedRecords))
	for i, appliedRecord := range appliedRecords {
		appliedRecordIDs[i] = appliedRecord.ID
	}
	returnedRecordMap, err := l.findBillReturnedRecords(kt, appliedRecordIDs)
	if err != nil {
		logs.Errorf("find rolling returned records failed, err: %v, applied records: %v, rid: %s", err,
			appliedRecordIDs, kt.Rid)
		return err
	}

	// 3.聚合step1和step2关联的数据，如果已交付数 > 已退回数，将交付和回收核心数等信息存储在滚服罚金明细中
	if err = l.addFineDetail(kt, req, appliedRecords, returnedRecordMap); err != nil {
		logs.Errorf("add rolling fine detail failed, err: %v, bizID: %d, rid: %s", err, req.BkBizID, kt.Rid)
		return err
	}

	// 4.以单业务纬度聚合数据，计算出某个业务单天的罚金，存储到obs滚服账单表中
	if err = l.calculateBill(kt, req); err != nil {
		logs.Errorf("calculate rolling bill failed, err: %v, bizID: %d, rid: %s", err, req.BkBizID, kt.Rid)
		return err
	}

	end := time.Now()
	logs.Infof("--- end sync biz(%d) rolling bill, end time: %v, cost: %v, rid: %s ---", req.BkBizID, end,
		end.Sub(start), kt.Rid)

	return nil
}

// findBillAppliedRecords 查询需要计算罚金的滚服申请记录表数据
func (l *logics) findBillAppliedRecords(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) (
	[]*rs.RollingAppliedRecord, error) {

	startYear, startMonth, startDay := subDay(req.Year, req.Month, req.Day, rollingserver.CalculateFineEndDay)
	endYear, endMonth, endDay := subDay(req.Year, req.Month, req.Day, rollingserver.CalculateFineStartDay)

	records := make([]*rs.RollingAppliedRecord, 0)
	listReq := &rsproto.RollingAppliedRecordListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: req.BkBizID},
				&filter.AtomRule{Field: "applied_type", Op: filter.Equal.Factory(), Value: enumor.NormalAppliedType},
				// 大于等于该日期的申请记录
				&filter.AtomRule{Field: "year", Op: filter.GreaterThanEqual.Factory(), Value: startYear},
				&filter.AtomRule{Field: "month", Op: filter.GreaterThanEqual.Factory(), Value: startMonth},
				&filter.AtomRule{Field: "day", Op: filter.GreaterThanEqual.Factory(), Value: startDay},
				// 小于等于该日期的申请记录
				&filter.AtomRule{Field: "year", Op: filter.LessThanEqual.Factory(), Value: endYear},
				&filter.AtomRule{Field: "month", Op: filter.LessThanEqual.Factory(), Value: endMonth},
				&filter.AtomRule{Field: "day", Op: filter.LessThanEqual.Factory(), Value: endDay},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}

	for {
		result, err := l.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return nil, err
		}
		records = append(records, result.Details...)

		if len(result.Details) < constant.BatchOperationMaxLimit {
			break
		}

		listReq.Page.Start += constant.BatchOperationMaxLimit
	}

	return records, nil
}

func subDay(year, month, day int, subDay int) (int, int, int) {
	originalDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	pastDate := originalDate.AddDate(0, 0, -subDay)

	return pastDate.Year(), int(pastDate.Month()), pastDate.Day()
}

// findBillReturnedRecords 查询需要进行账单计算的主机回收记录
func (l *logics) findBillReturnedRecords(kt *kit.Kit, appliedRecordIDs []string) (
	map[string][]*rs.RollingReturnedRecord, error) {

	recordMap := make(map[string][]*rs.RollingReturnedRecord)

	for _, ids := range slice.Split(appliedRecordIDs, constant.BatchOperationMaxLimit) {
		listReq := &rsproto.RollingReturnedRecordListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "applied_record_id", Op: filter.In.Factory(), Value: ids},
				},
			},
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
			},
		}

		for {
			result, err := l.client.DataService().Global.RollingServer.ListReturnedRecord(kt, listReq)
			if err != nil {
				logs.Errorf("list rolling returned record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
				return nil, err
			}

			for _, record := range result.Details {
				if _, ok := recordMap[record.AppliedRecordID]; !ok {
					recordMap[record.AppliedRecordID] = make([]*rs.RollingReturnedRecord, 0)
				}

				recordMap[record.AppliedRecordID] = append(recordMap[record.AppliedRecordID], record)
			}

			if len(result.Details) < constant.BatchOperationMaxLimit {
				break
			}

			listReq.Page.Start += constant.BatchOperationMaxLimit
		}
	}

	return recordMap, nil
}

func (l *logics) addFineDetail(kt *kit.Kit, req *rollingserver.RollingBillSyncReq,
	appliedRecords []*rs.RollingAppliedRecord, returnedRecordMap map[string][]*rs.RollingReturnedRecord) error {

	fineDetails := make([]rsproto.RollingFineDetailCreateReq, 0)
	existFineDetailMap, err := l.getExistFineDetail(kt, req)
	if err != nil {
		logs.Errorf("get exist fine detail failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}
	unitPrice, err := l.getRollingUnitPrice(kt)
	if err != nil {
		logs.Errorf("get rolling server unit price failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, apply := range appliedRecords {
		key := getFineDetailUniqueKey(apply.Year, apply.Month, apply.Day, apply.ID)
		if _, ok := existFineDetailMap[key]; ok {
			logs.Infof("rolling fine detail exist, key: %s, bizID: %s, rid: %s", key, req.BkBizID, kt.Rid)
			continue
		}

		var returnedCore uint64
		for _, returnedRecord := range returnedRecordMap[apply.ID] {
			returnedCore += *returnedRecord.MatchAppliedCore
		}

		if *apply.DeliveredCore > returnedCore {
			detail := rsproto.RollingFineDetailCreateReq{
				BkBizID:         apply.BkBizID,
				AppliedRecordID: apply.ID,
				OrderID:         apply.OrderID,
				SubOrderID:      apply.SubOrderID,
				Year:            apply.Year,
				Month:           apply.Month,
				Day:             apply.Day,
				DeliveredCore:   *apply.DeliveredCore,
				ReturnedCore:    returnedCore,
				Fine:            unitPrice.Mul(decimal.NewFromUint64(*apply.DeliveredCore - returnedCore)),
			}
			fineDetails = append(fineDetails, detail)
		}

		existFineDetailMap[key] = struct{}{}
	}

	for _, details := range slice.Split(fineDetails, constant.BatchOperationMaxLimit) {
		createReq := &rsproto.BatchCreateRollingFineDetailReq{FineDetails: details}
		if _, err = l.client.DataService().Global.RollingServer.BatchCreateFineDetail(kt, createReq); err != nil {
			logs.Errorf("batch create fine failed, err: %v, bizID: %d, rid: %s", err, req.BkBizID, kt.Rid)
			return err
		}
	}

	return nil
}

func (l *logics) getExistFineDetail(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) (map[string]struct{}, error) {
	details, err := l.getFineDetail(kt, req.BkBizID, req.Year, req.Month, req.Day, req.Day)
	if err != nil {
		logs.Errorf("get rolling fine detail failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return nil, err
	}

	existMap := make(map[string]struct{})
	for _, detail := range details {
		existMap[getFineDetailUniqueKey(detail.Year, detail.Month, detail.Day, detail.AppliedRecordID)] = struct{}{}
	}

	return existMap, nil
}

func (l *logics) getFineDetail(kt *kit.Kit, bizID int64, year, month, startDay, endDay int) (
	[]*rs.RollingFineDetailTable, error) {

	listReq := &rsproto.RollingFineDetailListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
				&filter.AtomRule{Field: "year", Op: filter.Equal.Factory(), Value: year},
				&filter.AtomRule{Field: "month", Op: filter.Equal.Factory(), Value: month},
				&filter.AtomRule{Field: "day", Op: filter.GreaterThanEqual.Factory(), Value: startDay},
				&filter.AtomRule{Field: "day", Op: filter.LessThanEqual.Factory(), Value: endDay},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}

	details := make([]*rs.RollingFineDetailTable, 0)
	for {
		result, err := l.client.DataService().Global.RollingServer.ListFineDetail(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling applied record failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return nil, err
		}

		details = append(details, result.Details...)
		if len(result.Details) < constant.BatchOperationMaxLimit {
			break
		}

		listReq.Page.Start += constant.BatchOperationMaxLimit
	}

	return details, nil
}

func getFineDetailUniqueKey(year, month, day int, appliedRecordID string) string {
	return fmt.Sprintf("year:%d month:%d day:%d applied_record_id:%s", year, month, day, appliedRecordID)
}

func (l *logics) calculateBill(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) error {
	exist, err := l.isBillExist(kt, req)
	if err != nil {
		logs.Errorf("find rolling bill failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}
	if exist {
		logs.Infof("rolling bill exist, bizID: %s, year: %d, month: %d, day: %d, rid: %s", req.BkBizID, req.Year,
			req.Month, req.Day, kt.Rid)
		return nil
	}

	details, err := l.getFineDetail(kt, req.BkBizID, req.Year, req.Month, rollingserver.FirstDay, req.Day)
	if err != nil {
		logs.Errorf("get rolling fine detail failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	var deliveredCore uint64 = 0
	var returnedCore uint64 = 0
	amount := decimal.NewFromFloat(0)
	amountInCurrentDate := decimal.NewFromFloat(0)
	for _, detail := range details {
		amount.Add(detail.Fine)
		if detail.Day == req.Day {
			amountInCurrentDate.Add(detail.Fine)
			deliveredCore += detail.DeliveredCore
			returnedCore += detail.ReturnedCore
		}
	}

	bizReq := &cmdb.SearchBizBelongingParams{
		BizIDs: []int64{req.BkBizID},
	}
	resp, err := l.esbClient.Cmdb().SearchBizBelonging(kt, bizReq)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return errors.New("search biz belonging, but resp is empty or len resp != 1")
	}
	bizBelong := (*resp)[0]

	bill := rsproto.RollingBillCreateReq{
		BkBizID:             req.BkBizID,
		DeliveredCore:       deliveredCore,
		ReturnedCore:        returnedCore,
		NotReturnedCore:     deliveredCore - returnedCore,
		Year:                req.Year,
		Month:               req.Month,
		Day:                 req.Day,
		DataDate:            getObsDataDate(req.Year, req.Month, req.Day),
		ProductID:           bizBelong.OpProductID,
		BusinessSetID:       bizBelong.Bs1NameID,
		BusinessSetName:     bizBelong.Bs1Name,
		CityID:              rollingserver.DefaultCityID,
		BusinessID:          bizBelong.Bs2NameID,
		BusinessName:        bizBelong.Bs2Name,
		BusinessModID:       rollingserver.DefaultBusinessModID,
		BusinessModName:     rollingserver.DefaultBusinessModName,
		PlatformID:          rollingserver.PlatformID,
		ResClassID:          rollingserver.ResClassID,
		Amount:              amount.InexactFloat64(),
		AmountInCurrentDate: amountInCurrentDate.InexactFloat64(),
	}
	createReq := &rsproto.BatchCreateRollingBillReq{Bills: []rsproto.RollingBillCreateReq{bill}}

	if _, err = l.client.DataService().Global.RollingServer.BatchCreateBill(kt, createReq); err != nil {
		logs.Errorf("create rolling bill failed, err: %v, bizID: %d, year: %d, month: %d, day: %d, rid: %s", err,
			req.BkBizID, req.Year, req.Month, req.Day, kt.Rid)
		return err
	}

	return nil
}

// getObsDataDate 如：year:2021, month：1, day: 2 => 20210102
func getObsDataDate(year, month, day int) string {
	date := year*10000 + month*100 + day
	return fmt.Sprintf("%d", date)
}

func (l *logics) isBillExist(kt *kit.Kit, req *rollingserver.RollingBillSyncReq) (bool, error) {
	listReq := &rsproto.RollingBillListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "year", Op: filter.Equal.Factory(), Value: req.Year},
				&filter.AtomRule{Field: "month", Op: filter.Equal.Factory(), Value: req.Month},
				&filter.AtomRule{Field: "day", Op: filter.Equal.Factory(), Value: req.Day},
			},
		},
		Fields: []string{"id", "bk_biz_id"},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	if req.BkBizID != rollingserver.SyncAllBiz {
		listReq.Filter.Rules = append(listReq.Filter.Rules,
			&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: req.BkBizID})
	}

	existBizDetailMap := make(map[int64]struct{})
	for {
		result, err := l.client.DataService().Global.RollingServer.ListBill(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling bills failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return false, err
		}

		for _, detail := range result.Details {
			existBizDetailMap[detail.BkBizID] = struct{}{}
		}

		if len(result.Details) < constant.BatchOperationMaxLimit {
			break
		}

		listReq.Page.Start += constant.BatchOperationMaxLimit
	}

	if req.BkBizID != rollingserver.SyncAllBiz {
		return len(existBizDetailMap) != 0, nil
	}

	bizIDs, err := l.listIEGBizIDs(kt)
	if err != nil {
		logs.Errorf("list ieg biz ids failed, err: %v, rid: %s", err, kt.Rid)
		return false, nil
	}
	resPoolBizMap, err := l.listResPoolBizIDs(kt)
	if err != nil {
		logs.Errorf("list rolling resource pool business failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	exist := true
	for _, bizID := range bizIDs {
		if _, ok := resPoolBizMap[bizID]; ok {
			continue
		}

		if _, ok := existBizDetailMap[bizID]; !ok {
			exist = false
			logs.Errorf("can not find biz rolling bill, bizID: %d, year: %d, month: %d, day: %d, rid: %s", bizID,
				req.Year, req.Month, req.Day, kt.Rid)
		}
	}

	return exist, nil
}

func (l *logics) listIEGBizIDs(kt *kit.Kit) ([]int64, error) {
	req := &cmdb.SearchBizReq{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_operate_dept_id",
						Operator: querybuilder.OperatorEqual,
						Value:    rollingserver.IEGOperateDeptID,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id"},
	}
	resp, err := l.esbClient.Cmdb().SearchBiz(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("search business from cc failed, err: %v, param:%+v, rid: %s", err, req, kt.Rid)
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	bizIDs := make([]int64, 0)
	for _, biz := range resp.Data.Info {
		bizIDs = append(bizIDs, biz.BkBizId)
	}

	return bizIDs, nil
}

func (l *logics) listResPoolBizIDs(kt *kit.Kit) (map[int64]struct{}, error) {
	listReq := &rsproto.ResourcePoolBusinessListReq{
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}

	bizMap := make(map[int64]struct{})
	for {
		result, err := l.client.DataService().Global.RollingServer.ListResPoolBiz(kt, listReq)
		if err != nil {
			logs.Errorf("list rolling resource pool business failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
			return nil, err
		}

		for _, biz := range result.Details {
			bizMap[biz.BkBizID] = struct{}{}
		}

		if len(result.Details) < constant.BatchOperationMaxLimit {
			break
		}

		listReq.Page.Start += constant.BatchOperationMaxLimit
	}

	return bizMap, nil
}

func (l *logics) getRollingUnitPrice(kt *kit.Kit) (*types.Decimal, error) {
	listReq := &rsproto.RollingGlobalConfigListReq{
		Fields: []string{"unit_price"},
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	result, err := l.client.DataService().Global.RollingServer.ListGlobalConfig(kt, listReq)
	if err != nil {
		logs.Errorf("list rolling resource pool business failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		logs.Errorf("can not find rolling global config, req: %+v, rid:%s", *listReq, kt.Rid)
		return nil, errors.New("can not find rolling global config")
	}

	return result.Details[0].UnitPrice, nil
}
