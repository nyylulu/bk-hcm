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
	rstypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"
)

// GetCpuCoreSummary get cpu core summary.
func (l *logics) GetCpuCoreSummary(kt *kit.Kit, req *rstypes.CpuCoreSummaryReq) (*rsproto.RollingCpuCoreSummaryItem,
	error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate cpu core summary request, err: %s, rid: %s", err, kt.Rid)
		return nil, err
	}

	startRollDate := times.GetDataIntDate(req.Start.Year, req.Start.Month, req.Start.Day)
	endRollDate := times.GetDataIntDate(req.End.Year, req.End.Month, req.End.Day)

	var appliedRules, returnedRules []*filter.AtomRule
	appliedRules = append(appliedRules, tools.RuleGreaterThanEqual("roll_date", startRollDate))
	appliedRules = append(appliedRules, tools.RuleLessThanEqual("roll_date", endRollDate))
	if len(req.BkBizIDs) != 0 {
		appliedRules = append(appliedRules, tools.RuleIn("bk_biz_id", req.BkBizIDs))
	}
	if len(req.InstanceGroup) != 0 {
		appliedRules = append(appliedRules, tools.RuleEqual("instance_group", req.InstanceGroup))
	}
	// 回收记录的查询条件跟申请记录的一致
	returnedRules = make([]*filter.AtomRule, len(appliedRules))
	copy(returnedRules, appliedRules)
	// 只查询未终止状态的滚服回收记录
	returnedRules = append(returnedRules, tools.RuleEqual("status", enumor.NormalStatus))

	if len(req.OrderIDs) != 0 {
		appliedRules = append(appliedRules, tools.RuleIn("order_id", req.OrderIDs))
	}
	if len(req.SubOrderIDs) != 0 {
		appliedRules = append(appliedRules, tools.RuleIn("suborder_id", req.SubOrderIDs))
	}
	if len(req.AppliedType) != 0 {
		appliedRules = append(appliedRules, tools.RuleEqual("applied_type", req.AppliedType))
	}

	// 查询滚服申请记录表的总的CPU核心数-已交付
	queryFilter := tools.ExpressionAnd(appliedRules...)
	deliveredReq := &rsproto.RollingAppliedRecordListReq{
		Filter: queryFilter,
		Page:   core.NewDefaultBasePage(),
	}
	deliverCoreResp, err := l.client.DataService().Global.RollingServer.GetRollingAppliedCoreSum(kt, deliveredReq)
	if err != nil {
		logs.Errorf("list rolling server applied cpu core summary failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 查询滚服回收记录表的总的CPU核心数-已退还
	returnedReq := &rsproto.RollingReturnedRecordListReq{
		Filter: tools.ExpressionAnd(returnedRules...),
		Page:   core.NewDefaultBasePage(),
	}
	returnCoreResp, err := l.client.DataService().Global.RollingServer.GetRollingReturnedCoreSum(kt, returnedReq)
	if err != nil {
		logs.Errorf("list rolling server returned cpu core summary failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &rsproto.RollingCpuCoreSummaryItem{
		SumDeliveredCore:       deliverCoreResp.SumDeliveredCore,
		SumReturnedAppliedCore: returnCoreResp.SumReturnedAppliedCore,
	}, nil
}
