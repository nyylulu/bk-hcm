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

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/querybuilder"
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
	if len(req.Start) == 0 {
		return "start", fmt.Errorf("start is not set")
	}

	start, err := time.Parse(dateLayout, req.Start)
	if err != nil {
		return "start", fmt.Errorf("date format should be like %s", dateLayout)
	}

	if len(req.End) == 0 {
		return "end", fmt.Errorf("end is not set")
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
			timeCond[common.BKDBGTE] = startTime
		}
	}
	if len(req.End) != 0 {
		endTime, err := time.Parse(dateLayout, req.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[common.BKDBLT] = endTime.AddDate(0, 0, 1)
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
