/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package operation define the operation interface
package operation

import (
	"context"
	"hcm/cmd/woa-server/logics/task/statistics"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"sort"
	"time"

	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/language"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/util"
)

// Interface operation interface
type Interface interface {
	// GetApplyStatistics get resource apply operation statistics
	GetApplyStatistics(kit *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error)
	// GetCompletionRateStatistics get completion rate statistics
	GetCompletionRateStatistics(kit *kit.Kit,
		param *types.GetCompletionRateStatReq) (*types.GetCompletionRateStatRst, error)
	// GetCompletionRateDetail 获取结单率详情统计
	GetCompletionRateDetail(kit *kit.Kit,
		param *types.GetCompletionRateDetailReq) (*types.GetCompletionRateDetailRst, error)
}

// operation provides operation statistics service
type operation struct {
	lang       language.CCLanguageIf
	statistics statistics.Interface
}

// New create a operation instance
func New(_ context.Context, clientSet *client.ClientSet) (*operation, error) {
	op := &operation{
		lang: language.NewFromCtx(language.EmptyLanguageSetting),
	}

	if clientSet != nil {
		op.statistics = statistics.New(clientSet)
	}

	return op, nil
}

// GetApplyStatistics get resource apply operation statistics
func (op *operation) GetApplyStatistics(kit *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	orderTotalStats, err := op.getOrderStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	succOrderFilter := util.CopyMap(filter, nil, nil)
	succOrderFilter["status"] = types.ApplyStatusDone
	orderSuccStats, err := op.getOrderStats(succOrderFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	osTotalStats, err := op.getDeviceStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total os statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	manualOrderList, err := op.getManualOrderList(filter)
	if err != nil {
		logs.Errorf("failed to get manual order list, err: %v", err)
		return nil, err
	}
	manualOrderFilter := util.CopyMap(filter, nil, nil)
	manualOrderFilter["suborder_id"] = mapstr.MapStr{
		pkg.BKDBIN: manualOrderList,
	}
	orderManualStats, err := op.getOrderStats(manualOrderFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply manual order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	succOsFilter := util.CopyMap(filter, nil, nil)
	succOsFilter["is_delivered"] = true
	succOsFilter["deliverer"] = "icr"
	osSuccStats, err := op.getDeviceStats(succOsFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success os statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// sort date keys
	dateKeys := make([]string, 0)
	for k := range orderTotalStats {
		dateKeys = append(dateKeys, k)
	}
	sort.Strings(dateKeys)

	rst := new(types.GetApplyStatRst)
	for _, date := range dateKeys {
		orderTotalCnt := uint(0)
		if orderTotalStat, ok := orderTotalStats[date]; ok {
			orderTotalCnt = uint(orderTotalStat.Count)
		}

		orderSuccCnt := uint(0)
		if orderSuccStat, ok := orderSuccStats[date]; ok {
			orderSuccCnt = uint(orderSuccStat.Count)
		}

		orderSuccRate := float64(0)
		if orderTotalCnt > 0 {
			orderSuccRate = float64(orderSuccCnt) / float64(orderTotalCnt)
		}

		orderManualCnt := uint(0)
		if orderManualStat, ok := orderManualStats[date]; ok {
			orderManualCnt = uint(orderManualStat.Count)
		}

		orderManualRate := float64(0)
		if orderManualCnt > 0 {
			orderManualRate = float64(orderManualCnt) / float64(orderTotalCnt)
		}

		osTotalCnt := uint(0)
		if osTotalStat, ok := osTotalStats[date]; ok {
			osTotalCnt = uint(osTotalStat.Count)
		}

		osSuccCnt := uint(0)
		if osSuccStat, ok := osSuccStats[date]; ok {
			osSuccCnt = uint(osSuccStat.Count)
		}

		osSuccRate := float64(0)
		if osTotalCnt > 0 {
			osSuccRate = float64(osSuccCnt) / float64(osTotalCnt)
		}

		applyStat := &types.ApplyStat{
			Date:            date,
			OrderTotal:      orderTotalCnt,
			OrderSucc:       orderSuccCnt,
			OrderSuccRate:   orderSuccRate,
			OrderManual:     orderManualCnt,
			OrderManualRate: orderManualRate,
			OsTotal:         osTotalCnt,
			OsSucc:          osSuccCnt,
			OsSuccRate:      osSuccRate,
		}
		rst.Info = append(rst.Info, applyStat)
	}

	return rst, nil
}

// getOrderStats get resource apply order operation statistics
func (op *operation) getOrderStats(filter map[string]interface{}, dimension types.TimeDimension) (
	map[string]metadata.StringIDCount, error) {

	format := op.getDateFormat(dimension)
	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": format,
					"date":   "$create_at"}},
			"count": map[string]interface{}{pkg.BKDBSum: 1}},
		},
		{pkg.BKDBSort: map[string]interface{}{"_id": 1}},
	}

	aggRst := make([]metadata.StringIDCount, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(context.Background(), pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get resource apply order operation statistics, err: %v", err)
		return nil, err
	}

	mapDateStat := make(map[string]metadata.StringIDCount)
	for _, stat := range aggRst {
		mapDateStat[stat.ID] = stat
	}

	return mapDateStat, nil
}

// getDeviceStats get resource apply delivered device operation statistics
func (op *operation) getDeviceStats(filter map[string]interface{}, dimension types.TimeDimension) (
	map[string]metadata.StringIDCount, error) {

	format := op.getDateFormat(dimension)
	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": format,
					"date":   "$create_at"}},
			"count": map[string]interface{}{pkg.BKDBSum: 1}},
		},
		{pkg.BKDBSort: map[string]interface{}{"_id": 1}},
	}

	aggRst := make([]metadata.StringIDCount, 0)
	if err := model.Operation().DeviceInfo().AggregateAll(context.Background(), pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get resource apply delivered device operation statistics, err: %v", err)
		return nil, err
	}

	mapDateStat := make(map[string]metadata.StringIDCount)
	for _, stat := range aggRst {
		mapDateStat[stat.ID] = stat
	}

	return mapDateStat, nil
}

func (op *operation) getManualOrderList(filter map[string]interface{}) ([]interface{}, error) {
	manualFilter := util.CopyMap(filter, nil, nil)
	manualFilter["deliverer"] = mapstr.MapStr{
		pkg.BKDBNE: "icr",
	}

	orderList, err := model.Operation().DeviceInfo().Distinct(context.Background(), "suborder_id", manualFilter)
	if err != nil {
		return nil, err
	}

	return orderList, nil
}

func (op *operation) getDateFormat(dimension types.TimeDimension) string {
	format := ""
	switch dimension {
	case types.DimensionDay:
		format = "%Y-%m-%d"
	case types.DimensionMonth:
		format = "%Y-%m"
	case types.DimensionYear:
		format = "%Y"
	default:
		// treat dimension as day by default
		format = "%Y-%m-%d"
	}

	return format
}

// GetCompletionRateStatistics get completion rate statistics
func (op *operation) GetCompletionRateStatistics(kit *kit.Kit,
	param *types.GetCompletionRateStatReq) (*types.GetCompletionRateStatRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get completion rate statistics, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	startTime, err := time.Parse(constant.DateLayout, param.StartTime)
	if err != nil {
		logs.Errorf("failed to parse start_time, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	endTime, err := time.Parse(constant.DateLayout, param.EndTime)
	if err != nil {
		logs.Errorf("failed to parse end_time, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	var excludeSuborderIDs []string
	if op.statistics != nil {
		excludeSuborderIDs, err = op.statistics.ListExcludedSubOrderIDs(kit, startTime, endTime)
		if err != nil {
			logs.Errorf("failed to get exclude suborder ids for completion rate statistics, err: %v, rid: %s",
				err, kit.Rid)
			return nil, err
		}
	}

	if len(excludeSuborderIDs) > 0 {
		suborderFilter, ok := filter["suborder_id"].(map[string]interface{})
		if !ok || suborderFilter == nil {
			suborderFilter = make(map[string]interface{})
		}
		suborderFilter[pkg.BKDBNIN] = excludeSuborderIDs
		filter["suborder_id"] = suborderFilter
	}

	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{"$addFields": map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at"}},
			"is_done": map[string]interface{}{
				"$cond": []interface{}{
					map[string]interface{}{"$eq": []interface{}{"$stage", "DONE"}},
					1,
					0,
				},
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":         "$year_month",
			"total_count": map[string]interface{}{pkg.BKDBSum: 1},
			"done_count":  map[string]interface{}{pkg.BKDBSum: "$is_done"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"year_month": "$_id",
			"completion_rate": map[string]interface{}{
				"$round": []interface{}{
					map[string]interface{}{
						"$multiply": []interface{}{
							map[string]interface{}{
								"$divide": []interface{}{
									"$done_count",
									map[string]interface{}{
										"$cond": []interface{}{
											map[string]interface{}{"$eq": []interface{}{"$total_count", 0}},
											1,
											"$total_count",
										},
									},
								},
							},
							100,
						},
					},
					2,
				},
			},
		}},
		{pkg.BKDBSort: map[string]interface{}{"year_month": 1}},
	}

	aggRst := make([]struct {
		YearMonth      string  `bson:"year_month"`
		CompletionRate float64 `bson:"completion_rate"`
	}, 0)

	if err := model.Operation().ApplyOrder().AggregateAll(kit.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get completion rate statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetCompletionRateStatRst{
		Details: make([]*types.CompletionRateStat, 0, len(aggRst)),
	}

	for _, stat := range aggRst {
		rst.Details = append(rst.Details, &types.CompletionRateStat{
			YearMonth:      stat.YearMonth,
			CompletionRate: stat.CompletionRate,
		})
	}

	return rst, nil
}

// GetCompletionRateDetail 获取结单率详情统计
func (op *operation) GetCompletionRateDetail(kit *kit.Kit,
	param *types.GetCompletionRateDetailReq) (*types.GetCompletionRateDetailRst, error) {
	// 解析时间范围
	startTime, err := time.Parse(constant.DateLayout, param.StartTime)
	if err != nil {
		logs.Errorf("failed to parse start_time, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	endTime, err := time.Parse(constant.DateLayout, param.EndTime)
	if err != nil {
		logs.Errorf("failed to parse end_time, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// 结束时间需要加1天，因为查询条件是 $lt（小于）不包含当天
	endTime = endTime.AddDate(0, 0, 1)

	var excludeSuborderIDs []string
	if op.statistics != nil {
		excludeSuborderIDs, err = op.statistics.ListExcludedSubOrderIDs(kit, startTime, endTime)
		if err != nil {
			logs.Errorf("failed to get exclude suborder ids for completion rate detail, err: %v, rid: %s",
				err, kit.Rid)
			return nil, err
		}
	}
	// 构建基础过滤条件
	baseFilter := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: startTime,
			pkg.BKDBLT:  endTime,
		},
	}

	if len(excludeSuborderIDs) > 0 {
		baseFilter["suborder_id"] = map[string]interface{}{
			pkg.BKDBNIN: excludeSuborderIDs,
		}
	}

	// 构建聚合管道
	pipeline := []map[string]interface{}{
		// 过滤时间范围 + 排除特定订单
		{pkg.BKDBMatch: baseFilter},
		// 提取年月信息
		{
			"$addFields": map[string]interface{}{
				"year_month": map[string]interface{}{
					"$dateToString": map[string]interface{}{
						"format": "%Y-%m",
						"date":   "$create_at",
					},
				},
			},
		},
		// 添加字段，标记已完成单据
		{
			"$addFields": map[string]interface{}{
				"is_done": map[string]interface{}{
					"$cond": []interface{}{
						map[string]interface{}{
							"$and": []interface{}{
								map[string]interface{}{"$eq": []interface{}{"$stage", types.TicketStageDone}},
								map[string]interface{}{"$eq": []interface{}{"$status", types.ApplyStatusDone}},
							},
						},
						1,
						0,
					},
				},
			},
		},
		// 按业务ID和月份分组统计
		{
			pkg.BKDBGroup: map[string]interface{}{
				"_id": map[string]interface{}{
					"bk_biz_id":  "$bk_biz_id",
					"year_month": "$year_month",
				},
				"total_orders": map[string]interface{}{pkg.BKDBSum: 1},          // 总单据数
				"done_orders":  map[string]interface{}{pkg.BKDBSum: "$is_done"}, // 已完成单据数
			},
		},
		//计算结单率
		{
			"$addFields": map[string]interface{}{
				"completion_rate": map[string]interface{}{
					"$cond": []interface{}{
						map[string]interface{}{"$eq": []interface{}{"$total_orders", 0}},
						0.0,
						map[string]interface{}{
							"$multiply": []interface{}{
								map[string]interface{}{
									"$divide": []interface{}{"$done_orders", "$total_orders"},
								},
								100,
							},
						},
					},
				},
			},
		},
		// 格式化输出
		{
			pkg.BKDBProject: map[string]interface{}{
				"_id":          0,
				"bk_biz_id":    "$_id.bk_biz_id",
				"year_month":   "$_id.year_month",
				"total_orders": "$total_orders",
				"done_orders":  "$done_orders",
				"completion_rate": map[string]interface{}{
					"$round": []interface{}{"$completion_rate", 2},
				},
			},
		},
		// 按结单率降序排序，相同则按业务ID和月份升序
		{
			pkg.BKDBSort: map[string]interface{}{
				"completion_rate": -1,
				"bk_biz_id":       1,
				"year_month":      1,
			},
		},
	}

	// 执行聚合查询
	aggRst := make([]*types.CompletionRateDetailItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kit.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get completion rate detail statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return &types.GetCompletionRateDetailRst{
		Details: aggRst,
	}, nil
}
