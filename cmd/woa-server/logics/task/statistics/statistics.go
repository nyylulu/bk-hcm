/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package statistics package
package statistics

import (
	"fmt"
	"strings"
	"time"

	taskModel "hcm/cmd/woa-server/model/task"
	configTypes "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao"
	daotypes "hcm/pkg/dal/dao/types"
	tablecvmapplyorderstatisticsconfig "hcm/pkg/dal/table/cvm-apply-order-statistics-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// Interface statistics interface
type Interface interface {
	// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
	ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error)
}

// New 创建统计能力实例
func New(daoSet dao.Set) Interface {
	return &statistics{
		daoSet: daoSet,
	}
}

type statistics struct {
	daoSet dao.Set
}

// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
func (s *statistics) ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error) {
	if err := validateQueryRange(queryStart, queryEnd); err != nil {
		return nil, err
	}
	// 从配置表中加载所有相关月份的配置
	monthSet, monthSlice := buildQueryMonths(queryStart, queryEnd)

	configs, err := s.loadConfigs(kt, monthSlice)
	if err != nil {
		return nil, fmt.Errorf("list statistics config failed: %w", err)
	}

	explicitIDs, timeRanges := s.extractConfigInfo(kt, configs, monthSet, queryStart, queryEnd)

	result := make([]string, 0, len(explicitIDs))
	result = append(result, explicitIDs...)

	if len(timeRanges) > 0 {
		// 从申请单中反查所有符合时间范围的子订单号
		queryIDs, err := s.fetchSubOrderIDsFromOrders(kt, timeRanges)
		if err != nil {
			return nil, err
		}
		result = append(result, queryIDs...)
	}

	return slice.Unique(result), nil
}

type timeRangeConfig struct {
	start time.Time
	end   time.Time
}

// fetchSubOrderIDsFromOrders 根据时间段配置，到申请单集合按时间反查子单号
func (s *statistics) fetchSubOrderIDsFromOrders(kt *kit.Kit, ranges []timeRangeConfig) ([]string, error) {
	if len(ranges) == 0 {
		return nil, nil
	}

	orConditions := make([]map[string]interface{}, 0, len(ranges))
	for _, tr := range ranges {
		if tr.start.After(tr.end) {
			continue
		}
		cond := map[string]interface{}{
			"create_at": map[string]interface{}{
				pkg.BKDBGTE: tr.start,
				pkg.BKDBLTE: tr.end,
			},
		}
		orConditions = append(orConditions, cond)
	}

	if len(orConditions) == 0 {
		return nil, nil
	}

	filters := map[string]interface{}{
		pkg.BKDBOR: orConditions,
	}

	page := metadata.BasePage{
		Limit: pkg.BKNoLimit,
	}

	orders, err := taskModel.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filters)
	if err != nil {
		return nil, fmt.Errorf("find apply order failed: %w", err)
	}

	ids := make([]string, 0, len(orders))
	for _, order := range orders {
		if order == nil {
			continue
		}
		if order.SubOrderId == "" {
			continue
		}
		ids = append(ids, order.SubOrderId)
	}

	return ids, nil
}

// splitSubOrderIDs 从逗号分隔的子单号字符串中提取有效子单号
func splitSubOrderIDs(ids string) []string {
	parts := strings.Split(ids, ",")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		results = append(results, part)
	}
	return slice.Unique(results)
}

// parseConfigTime 解析配置中的时间字符串，支持不同格式
func parseConfigTime(value string) (time.Time, bool, error) {
	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{constant.DateTimeLayout, false},
		{constant.DateLayout, true},
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout.layout, value, time.Local)
		if err == nil {
			return t, layout.dateOnly, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("unsupported time format: %s", value)
}

// buildQueryMonths 构建查询范围内的月份列表，格式为"YYYY-MM"
func buildQueryMonths(start, end time.Time) (map[string]bool, []string) {
	set := make(map[string]bool)
	months := make([]string, 0)

	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	last := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	for !current.After(last) {
		key := current.Format(constant.YearMonthLayout)
		set[key] = true
		months = append(months, key)
		current = current.AddDate(0, 1, 0)
	}
	return set, months
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func validateQueryRange(queryStart, queryEnd time.Time) error {
	if queryStart.IsZero() || queryEnd.IsZero() {
		return fmt.Errorf("query start/end time can not be empty")
	}
	if queryStart.After(queryEnd) {
		return fmt.Errorf("query start time can not be after end time")
	}
	return nil
}

// loadConfigs 从数据库加载所有符合月份条件的配置
func (s *statistics) loadConfigs(kt *kit.Kit, monthSlice []string) ([]*configTypes.CvmApplyOrderStatisticsConfig, error) {
	page := core.BasePage{
		Limit: pkg.BKNoLimit,
	}

	var filterExpr *filter.Expression
	if len(monthSlice) > 0 {
		rules := make([]filter.RuleFactory, 0, len(monthSlice))
		for _, month := range monthSlice {
			rules = append(rules, &filter.AtomRule{
				Field: "year_month",
				Op:    filter.Equal.Factory(),
				Value: month,
			})
		}
		filterExpr = &filter.Expression{
			Op:    filter.Or,
			Rules: rules,
		}
	} else {
		filterExpr = &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		}
	}

	listOpt := &daotypes.ListOption{
		Filter: filterExpr,
		Page:   &page,
		Fields: []string{},
	}

	result, err := s.daoSet.CvmApplyOrderStatisticsConfig().List(kt, listOpt)
	if err != nil {
		return nil, err
	}

	// 转换为 configTypes.CvmApplyOrderStatisticsConfig
	configs := make([]*configTypes.CvmApplyOrderStatisticsConfig, 0, len(result.Details))
	for _, detail := range result.Details {
		configs = append(configs, convertTableToType(&detail))
	}

	return configs, nil
}

// convertTableToType
func convertTableToType(table *tablecvmapplyorderstatisticsconfig.CvmApplyOrderStatisticsConfigTable) *configTypes.CvmApplyOrderStatisticsConfig {
	// types.Time 是字符串类型，需要解析为 time.Time
	createdAt := time.Time{}
	updatedAt := time.Time{}
	if len(table.CreatedAt) > 0 {
		if t, err := time.Parse(constant.TimeStdFormat, string(table.CreatedAt)); err == nil {
			createdAt = t
		}
	}
	if len(table.UpdatedAt) > 0 {
		if t, err := time.Parse(constant.TimeStdFormat, string(table.UpdatedAt)); err == nil {
			updatedAt = t
		}
	}

	return &configTypes.CvmApplyOrderStatisticsConfig{
		ID:         table.ID,
		YearMonth:  table.YearMonth,
		BkBizID:    table.BkBizID,
		SubOrderID: table.SubOrderID,
		StartAt:    table.StartAt,
		EndAt:      table.EndAt,
		Memo:       table.Memo,
		Extension:  table.Extension,
		Creator:    table.Creator,
		Reviser:    table.Reviser,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// extractConfigInfo 从配置中提取显式子单号和时间范围配置
func (s *statistics) extractConfigInfo(kt *kit.Kit, configs []*configTypes.CvmApplyOrderStatisticsConfig,
	monthSet map[string]bool, queryStart, queryEnd time.Time) ([]string, []timeRangeConfig) {

	explicitIDs := make([]string, 0)
	timeRanges := make([]timeRangeConfig, 0)

	for _, cfg := range configs {
		if !s.monthMatched(cfg, monthSet) {
			continue
		}
		explicitIDs = append(explicitIDs, splitSubOrderIDs(cfg.SubOrderID)...)

		rangeCfg, ok := s.buildTimeRange(kt, cfg, queryStart, queryEnd)
		if ok {
			timeRanges = append(timeRanges, rangeCfg)
		}
	}

	return explicitIDs, timeRanges
}

func (s *statistics) monthMatched(cfg *configTypes.CvmApplyOrderStatisticsConfig, monthSet map[string]bool) bool {
	if cfg == nil {
		return false
	}
	if len(monthSet) == 0 {
		return true
	}
	return monthSet[cfg.YearMonth]
}

// buildTimeRange 构建配置的时间范围，确保不超出查询范围
func (s *statistics) buildTimeRange(kt *kit.Kit, cfg *configTypes.CvmApplyOrderStatisticsConfig,
	queryStart, queryEnd time.Time) (timeRangeConfig, bool) {

	if cfg == nil {
		return timeRangeConfig{}, false
	}
	if strings.TrimSpace(cfg.StartAt) == "" || strings.TrimSpace(cfg.EndAt) == "" {
		return timeRangeConfig{}, false
	}

	start, startDateOnly, err := parseConfigTime(cfg.StartAt)
	if err != nil {
		logs.Warnf("parse start_at failed, cfg_id: %s, start_at: %s, err: %v, rid: %s", cfg.ID, cfg.StartAt, err, kt.Rid)
		return timeRangeConfig{}, false
	}
	end, endDateOnly, err := parseConfigTime(cfg.EndAt)
	if err != nil {
		logs.Warnf("parse end_at failed, cfg_id: %s, end_at: %s, err: %v, rid: %s", cfg.ID, cfg.EndAt, err, kt.Rid)
		return timeRangeConfig{}, false
	}

	if endDateOnly {
		end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}
	if startDateOnly {
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	}

	if end.Before(start) {
		logs.Warnf("invalid config time range, cfg_id: %s, start_at: %s, end_at: %s, rid: %s", cfg.ID, cfg.StartAt, cfg.EndAt, kt.Rid)
		return timeRangeConfig{}, false
	}

	if end.Before(queryStart) || start.After(queryEnd) {
		return timeRangeConfig{}, false
	}

	return timeRangeConfig{
		start: maxTime(start, queryStart),
		end:   minTime(end, queryEnd),
	}, true
}
