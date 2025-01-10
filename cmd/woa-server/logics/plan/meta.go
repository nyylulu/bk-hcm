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
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// GetDemandAvailableTime get demand available time
func (c *Controller) GetDemandAvailableTime(kt *kit.Kit, expectTime time.Time) (*ptypes.DemandAvailTimeResp, error) {
	yearMonthWeek, err := c.demandTime.GetDemandYearMonthWeek(kt, expectTime)
	if err != nil {
		logs.Errorf("failed to get demand year month week, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	drMonth, err := c.demandTime.GetDemandDateRangeInMonth(kt, expectTime)
	if err != nil {
		logs.Errorf("failed to get demand date range in month, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	drWeek := c.demandTime.GetDemandDateRangeInWeek(kt, expectTime)

	return &ptypes.DemandAvailTimeResp{
		YearMonthWeek: yearMonthWeek,
		DRInWeek:      drWeek,
		DRInMonth:     drMonth,
	}, nil
}

// CreateDemandWeek create demand week
func (c *Controller) CreateDemandWeek(kt *kit.Kit, createReqs []rpproto.ResPlanWeekCreateReq) (
	*core.BatchCreateResult, error) {

	createIDs := make([]string, 0)
	for _, batch := range slice.Split(createReqs, constant.BatchOperationMaxLimit) {
		batchReq := &rpproto.ResPlanWeekBatchCreateReq{
			Weeks: batch,
		}

		rst, err := c.client.DataService().Global.ResourcePlan.BatchCreateResPlanWeek(kt, batchReq)
		if err != nil {
			logs.Errorf("failed to batch create res plan week, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		createIDs = append(createIDs, rst.IDs...)
	}

	return &core.BatchCreateResult{
		IDs: createIDs,
	}, nil
}
