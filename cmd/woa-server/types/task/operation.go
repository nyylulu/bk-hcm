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

// Package task define task operation
package task

import (
	"fmt"
	"time"

	"hcm/pkg"
	"hcm/pkg/tools/querybuilder"
)

// GetApplyStatReq get resource apply operation statistics request
type GetApplyStatReq struct {
	Start     string                    `json:"start" bson:"start"`
	End       string                    `json:"end" bson:"end"`
	Dimension TimeDimension             `json:"dimension" bson:"dimension"`
	Filter    *querybuilder.QueryFilter `json:"filter" bson:"filter"`
}

// Validate whether GetApplyStatReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *GetApplyStatReq) Validate() (errKey string, err error) {
	start, err := time.Parse(dateLayout, req.Start)
	if err != nil {
		return "start", fmt.Errorf("date format should be like %s", dateLayout)
	}

	end, err := time.Parse(dateLayout, req.End)
	if err != nil {
		return "end", fmt.Errorf("date format should be like %s", dateLayout)
	}

	if len(req.Dimension) > 0 {
		if err := req.Dimension.Validate(); err != nil {
			return "dimension", err
		}
	}

	switch req.Dimension {
	// time range limit is different for different dimension
	case DimensionDay:
		if end.After(start.AddDate(0, 0, 90)) {
			return "start,end", fmt.Errorf("time range exeeds limit 90 days")
		}
	case DimensionMonth:
		if end.After(start.AddDate(0, 24, 0)) {
			return "start,end", fmt.Errorf("time range exeeds limit 24 months")
		}
	case DimensionYear:
		if end.After(start.AddDate(3, 0, 0)) {
			return "start,end", fmt.Errorf("time range exeeds limit 3 years")
		}
	}

	if req.Filter != nil {
		if key, err := req.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if req.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (req *GetApplyStatReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})

	if req.Filter != nil {
		mgoFilter, key, err := req.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		filter = mgoFilter
	}

	timeCond := make(map[string]interface{})
	if len(req.Start) != 0 {
		startTime, err := time.Parse(dateLayout, req.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}
	if len(req.End) != 0 {
		endTime, err := time.Parse(dateLayout, req.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}
	if len(timeCond) != 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// TimeDimension statistics time dimension
type TimeDimension string

// Validate whether TimeDimension is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param TimeDimension) Validate() (err error) {
	switch param {
	case DimensionDay, DimensionMonth, DimensionYear:
	default:
		return fmt.Errorf("unkown %s dimension type", param)
	}

	return nil
}

// TimeDimension statistics time dimension
const (
	DimensionDay   TimeDimension = "DAY"
	DimensionMonth TimeDimension = "MONTH"
	DimensionYear  TimeDimension = "YEAR"
)

// GetApplyStatRst get resource apply operation statistics result
type GetApplyStatRst struct {
	Info []*ApplyStat `json:"info" bson:"info"`
}

// ApplyStat resource apply operation statistics
type ApplyStat struct {
	Date            string  `json:"date" bson:"date"`
	OrderTotal      uint    `json:"order_total" bson:"order_total"`
	OrderSucc       uint    `json:"order_succ" bson:"order_succ"`
	OrderSuccRate   float64 `json:"order_succ_rate" bson:"order_succ_rate"`
	OrderManual     uint    `json:"order_manual" bson:"order_manual"`
	OrderManualRate float64 `json:"order_manual_rate" bson:"order_manual_rate"`
	OsTotal         uint    `json:"os_total" bson:"os_total"`
	OsSucc          uint    `json:"os_succ" bson:"os_succ"`
	OsSuccRate      float64 `json:"os_succ_rate" bson:"os_succ_rate"`
}

// GetCompletionRateStatReq get completion rate statistics request
type GetCompletionRateStatReq struct {
	StartTime string `json:"start_time" bson:"start_time"`
	EndTime   string `json:"end_time" bson:"end_time"`
}

// Validate whether GetCompletionRateStatReq is valid
func (req *GetCompletionRateStatReq) Validate() (errKey string, err error) {
	startTime, err := time.Parse(dateLayout, req.StartTime)
	if err != nil {
		return "start_time", fmt.Errorf("date format should be like %s", dateLayout)
	}

	endTime, err := time.Parse(dateLayout, req.EndTime)
	if err != nil {
		return "end_time", fmt.Errorf("date format should be like %s", dateLayout)
	}

	if endTime.Before(startTime) {
		return "start_time,end_time", fmt.Errorf("end_time must be after start_time")
	}

	return "", nil
}

// GetFilter get mgo filter
func (req *GetCompletionRateStatReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})

	timeCond := make(map[string]interface{})
	if len(req.StartTime) != 0 {
		startTime, err := time.Parse(dateLayout, req.StartTime)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}
	if len(req.EndTime) != 0 {
		endTime, err := time.Parse(dateLayout, req.EndTime)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}
	if len(timeCond) != 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetCompletionRateStatRst get completion rate statistics result
type GetCompletionRateStatRst struct {
	Details []*CompletionRateStat `json:"details" bson:"details"`
}

// CompletionRateStat completion rate statistics
type CompletionRateStat struct {
	YearMonth      string  `json:"year_month" bson:"year_month"`
	CompletionRate float64 `json:"completion_rate" bson:"completion_rate"`
}

// GetCompletionRateDetailReq 获取结单率详情统计请求
type GetCompletionRateDetailReq struct {
	StartTime string `json:"start_time" bson:"start_time"` // 开始时间，格式：YYYY-MM-DD
	EndTime   string `json:"end_time" bson:"end_time"`     // 结束时间，格式：YYYY-MM-DD
}

// Validate 验证请求参数
func (req *GetCompletionRateDetailReq) Validate() (errKey string, err error) {
	startTime, err := time.Parse(dateLayout, req.StartTime)
	if err != nil {
		return "start_time", fmt.Errorf("date format should be like %s", dateLayout)
	}

	endTime, err := time.Parse(dateLayout, req.EndTime)
	if err != nil {
		return "end_time", fmt.Errorf("date format should be like %s", dateLayout)
	}

	if endTime.Before(startTime) {
		return "start_time,end_time", fmt.Errorf("end_time must be after start_time")
	}

	return "", nil
}

// GetCompletionRateDetailRst 获取结单率详情统计响应
type GetCompletionRateDetailRst struct {
	Details []*CompletionRateDetailItem `json:"details" bson:"details"`
}

// CompletionRateDetailItem 结单率详情统计项
type CompletionRateDetailItem struct {
	BkBizID        int64   `json:"bk_biz_id" bson:"bk_biz_id"`             // 业务ID
	YearMonth      string  `json:"year_month" bson:"year_month"`           // 年月，格式：YYYY-MM
	TotalOrders    int     `json:"total_orders" bson:"total_orders"`       // 总单据数
	DoneOrders     int     `json:"done_orders" bson:"done_orders"`         // 已完成单据数（stage=DONE）
	CompletionRate float64 `json:"completion_rate" bson:"completion_rate"` // 结单率（百分比），保留2位小数
}
