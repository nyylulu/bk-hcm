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

package statistics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	configModel "hcm/cmd/woa-server/model/config"
	taskModel "hcm/cmd/woa-server/model/task"
	"hcm/pkg"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"
)

// Interface statistics interface
type Interface interface {
	// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
	ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error)
}

// New 创建统计能力实例
func New() Interface {
	return &statistics{}
}

type statistics struct{}

// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
func (s *statistics) ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error) {
	if queryStart.IsZero() || queryEnd.IsZero() {
		return nil, fmt.Errorf("query start/end time can not be empty")
	}
	if queryStart.After(queryEnd) {
		return nil, fmt.Errorf("query start time can not be after end time")
	}
	// 构建查询月份集合
	monthSet, monthSlice := buildQueryMonths(queryStart, queryEnd)

	page := metadata.BasePage{
		Limit: pkg.BKNoLimit,
	}

	filter := map[string]interface{}{}
	if len(monthSlice) > 0 {
		filter["year_month"] = map[string]interface{}{pkg.BKDBIN: monthSlice}
	}
	// 根据查询区间涉及的月份，拉取配置表中对应的配置记录
	configs, err := configModel.Operation().CvmApplyOrderStatisticsConfig().FindMany(kt.Ctx, page, filter)
	if err != nil {
		return nil, fmt.Errorf("list statistics config failed: %w", err)
	}

	explicitIDs := make([]string, 0)
	timeRanges := make([]timeRangeConfig, 0)

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}

		// 再次校验是否真的落在查询月份集合
		if len(monthSet) > 0 && !monthSet[cfg.YearMonth] {
			continue
		}

		if strings.TrimSpace(cfg.SubOrderID) != "" {
			explicitIDs = append(explicitIDs, splitSubOrderIDs(cfg.SubOrderID)...)
		}

		if strings.TrimSpace(cfg.StartAt) != "" && strings.TrimSpace(cfg.EndAt) != "" {
			start, startDateOnly, err := parseConfigTime(cfg.StartAt)
			if err != nil {
				logs.Warnf("parse start_at failed, cfg_id: %s, start_at: %s, rid: %s, err: %v",
					cfg.ID, cfg.StartAt, kt.Rid, err)
				continue
			}
			end, endDateOnly, err := parseConfigTime(cfg.EndAt)
			if err != nil {
				logs.Warnf("parse end_at failed, cfg_id: %s, end_at: %s, rid: %s, err: %v",
					cfg.ID, cfg.EndAt, kt.Rid, err)
				continue
			}

			if endDateOnly {
				end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)
			}

			if startDateOnly {
				start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
			}

			if end.Before(start) {
				logs.Warnf("invalid config time range, cfg_id: %s, start_at: %s, end_at: %s, rid: %s",
					cfg.ID, cfg.StartAt, cfg.EndAt, kt.Rid)
				continue
			}

			if end.Before(queryStart) || start.After(queryEnd) {
				continue
			}

			// 只有时间段的配置先记录交集，后续统一到申请单库反查子单号
			rangeCfg := timeRangeConfig{
				start: maxTime(start, queryStart),
				end:   minTime(end, queryEnd),
			}
			timeRanges = append(timeRanges, rangeCfg)
		}
	}

	//先把显式子单号放入结果集合
	resultSet := make(map[string]struct{})
	for _, id := range explicitIDs {
		if id == "" {
			continue
		}
		resultSet[id] = struct{}{}
	}

	if len(timeRanges) > 0 {
		//用时间段的配置到申请单集合按时间反查子单号
		queryIDs, err := s.fetchSubOrderIDsFromOrders(kt, timeRanges)
		if err != nil {
			return nil, err
		}
		for _, id := range queryIDs {
			if id == "" {
				continue
			}
			resultSet[id] = struct{}{}
		}
	}

	//结果去重并排序后返回
	result := make([]string, 0, len(resultSet))
	for id := range resultSet {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
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

	filter := map[string]interface{}{
		pkg.BKDBOR: orConditions,
	}

	page := metadata.BasePage{
		Limit: pkg.BKNoLimit,
		Sort:  "suborder_id",
	}

	orders, err := taskModel.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filter)
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
	return results
}

// parseConfigTime 解析配置中的时间字符串，支持不同格式
func parseConfigTime(value string) (time.Time, bool, error) {
	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{"2006-01-02 15:04:05", false},
		{"2006-01-02 15:04", false},
		{"2006-01-02", true},
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
		key := current.Format("2006-01")
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
