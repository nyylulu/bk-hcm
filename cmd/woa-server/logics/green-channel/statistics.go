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

// Package greenchannel ...
package greenchannel

import (
	"strings"

	model "hcm/cmd/woa-server/model/task"
	gctypes "hcm/cmd/woa-server/types/green-channel"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetCpuCoreSummary get cpu core summary.
func (l *logics) GetCpuCoreSummary(kt *kit.Kit, req *gctypes.CpuCoreSummaryReq) (*gctypes.CpuCoreSummaryResp, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate cpu core summary request, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return nil, err
	}

	filter := map[string]interface{}{
		"create_at": map[string]interface{}{
			"$gte": req.DateRange.Start.GetTime(),
			"$lte": req.DateRange.End.GetTime(),
		},
		"require_type": enumor.RequireTypeGreenChannel,
	}
	if len(req.BkBizIDs) != 0 {
		filter["bk_biz_id"] = map[string]interface{}{
			"$in": req.BkBizIDs,
		}
	}

	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":   nil,
			"count": map[string]interface{}{pkg.BKDBSum: "$success_num"}},
		},
	}

	aggRst := make([]gctypes.AggregateCount, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get apply order cpu core summary, err: %v, req: %v, rid: %s", err, pipeline, kt.Rid)
		return nil, err
	}
	var count uint64
	if len(aggRst) != 0 {
		count = aggRst[0].Count
	}

	return &gctypes.CpuCoreSummaryResp{SumDeliveredCore: count}, nil
}

// ListStatisticalRecord list statistical record.
func (l *logics) ListStatisticalRecord(kt *kit.Kit, req *gctypes.StatisticalRecordReq) (*gctypes.StatisticalRecordResp,
	error) {

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate request, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return nil, err
	}
	filter := map[string]interface{}{
		"create_at": map[string]interface{}{
			"$gte": req.DateRange.Start.GetTime(),
			"$lte": req.DateRange.End.GetTime(),
		},
		"require_type": enumor.RequireTypeGreenChannel,
	}
	if len(req.BkBizIDs) != 0 {
		filter["bk_biz_id"] = map[string]interface{}{
			"$in": req.BkBizIDs,
		}
	}

	if req.Page.Count {
		pipeline := []map[string]interface{}{
			{pkg.BKDBMatch: filter},
			{pkg.BKDBGroup: map[string]interface{}{"_id": "$bk_biz_id"}},
			{pkg.BKDBCount: "count"},
		}
		aggRst := make([]gctypes.AggregateCount, 0)
		if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &aggRst); err != nil {
			logs.Errorf("failed to get statistical record count, err: %v, req: %v, rid: %s", err, pipeline, kt.Rid)
			return nil, err
		}
		var count uint64
		if len(aggRst) != 0 {
			count = aggRst[0].Count
		}
		return &gctypes.StatisticalRecordResp{Count: count}, nil
	}

	var sumDeliveredCore, sumAppliedCore, orderCount = "sum_delivered_core", "sum_applied_core", "order_count"
	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":              "$bk_biz_id",
			"unique_order_ids": map[string]interface{}{pkg.BKDBAddToSet: "$order_id"},
			sumDeliveredCore:   map[string]interface{}{pkg.BKDBSum: "$success_num"},
			sumAppliedCore:     map[string]interface{}{pkg.BKDBSum: "$total_num"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"_id":            0,
			"bk_biz_id":      "$_id",
			orderCount:       map[string]interface{}{pkg.BKDBSize: "$unique_order_ids"},
			sumDeliveredCore: 1,
			sumAppliedCore:   1,
		}},
		{pkg.BKDBSkip: req.Page.Start},
		{pkg.BKDBLimit: req.Page.Limit},
	}
	if req.Page.Sort != "" {
		sort := make(map[string]interface{})
		split := strings.Split(req.Page.Sort, ",")
		for _, field := range split {
			switch field {
			case orderCount:
				sort[orderCount] = pkg.BKDBAsc
			case sumAppliedCore:
				sort[sumAppliedCore] = pkg.BKDBAsc
			case sumDeliveredCore:
				sort[sumDeliveredCore] = pkg.BKDBAsc
			}
		}
		pipeline = append(pipeline, map[string]interface{}{pkg.BKDBSort: sort})
	}
	aggRst := make([]gctypes.StatisticalRecordItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get statistical record details, err: %v, req: %v, rid: %s", err, pipeline, kt.Rid)
		return nil, err
	}

	return &gctypes.StatisticalRecordResp{Details: aggRst}, nil
}
