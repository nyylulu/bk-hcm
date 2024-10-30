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
	"fmt"
	"time"

	rstypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/pkg/errors"
)

func (l *logics) isBizCurMonthHavingQuota(kt *kit.Kit, bizID int64, appliedCount uint) (bool, string, error) {
	now := time.Now()
	summaryReq := &rstypes.CpuCoreSummaryReq{
		BkBizIDs: []int64{bizID},
		RollingServerDateRange: rstypes.RollingServerDateRange{
			Start: rstypes.RollingServerDateTimeItem{
				Year: now.Year(), Month: int(now.Month()), Day: rstypes.FirstDay,
			},
			End: rstypes.RollingServerDateTimeItem{
				Year: now.Year(), Month: int(now.Month()), Day: now.Day(),
			},
		},
	}
	summary, err := l.GetCpuCoreSummary(kt, summaryReq)
	if err != nil {
		logs.Errorf("get cpu core summary failed, err: %v, req: %+v, rid: %s", err, *summaryReq, kt.Rid)
		return false, "", err
	}

	bizQuota, err := l.getBizCurMonthQuota(kt, bizID)
	if err != nil {
		logs.Errorf("get biz current month quota failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return false, "", err
	}

	if int64(summary.SumDeliveredCore)+int64(appliedCount) > bizQuota {
		reason := fmt.Sprintf("业务(%d)滚服项目当前已交付%d核心，本次申请%d核心，超过本业务限制:%d核心", bizID,
			summary.SumDeliveredCore, appliedCount, bizQuota)
		return false, reason, nil
	}

	return true, "", nil
}

func (l *logics) isSystemCurMonthHavingQuota(kt *kit.Kit, appliedCount uint) (bool, string, error) {
	now := time.Now()
	summaryReq := &rstypes.CpuCoreSummaryReq{
		RollingServerDateRange: rstypes.RollingServerDateRange{
			Start: rstypes.RollingServerDateTimeItem{
				Year: now.Year(), Month: int(now.Month()), Day: rstypes.FirstDay,
			},
			End: rstypes.RollingServerDateTimeItem{
				Year: now.Year(), Month: int(now.Month()), Day: now.Day(),
			},
		},
	}
	summary, err := l.GetCpuCoreSummary(kt, summaryReq)
	if err != nil {
		logs.Errorf("get cpu core summary failed, err: %v, req: %+v, rid: %s", err, *summaryReq, kt.Rid)
		return false, "", err
	}

	listGlobalConfigReq := &rsproto.RollingGlobalConfigListReq{
		Filter: tools.AllExpression(),
		Fields: []string{"global_quota"},
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}
	config, err := l.client.DataService().Global.RollingServer.ListGlobalConfig(kt, listGlobalConfigReq)
	if err != nil {
		logs.Errorf("list rolling server global config failed, err: %v, rid: %s", err, kt.Rid)
		return false, "", err
	}
	if len(config.Details) == 0 {
		logs.Errorf("can not get rolling server global config, rid: %s", kt.Rid)
		return false, "", errors.New("can not get rolling server global config")
	}

	if int64(summary.SumDeliveredCore)+int64(appliedCount) > *config.Details[0].GlobalQuota {
		reason := fmt.Sprintf("当月滚服项目已交付%d核心，本次申请%d核心，超过本月总额度限制:%d核心",
			summary.SumDeliveredCore,
			appliedCount, *config.Details[0].GlobalQuota)
		return false, reason, nil
	}

	return true, "", nil
}

func (l *logics) getBizCurMonthQuota(kt *kit.Kit, bizID int64) (int64, error) {
	now := time.Now()
	rules := []filter.RuleFactory{
		&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
		&filter.AtomRule{Field: "year", Op: filter.Equal.Factory(), Value: now.Year()},
		&filter.AtomRule{Field: "month", Op: filter.Equal.Factory(), Value: now.Month()},
	}

	listQuotaReq := &rsproto.RollingQuotaConfigListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: rules},
		Fields: []string{"quota"},
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}

	quotaConfig, err := l.client.DataService().Global.RollingServer.ListQuotaConfig(kt, listQuotaReq)
	if err != nil {
		logs.Errorf("list rolling server quota config failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return 0, err
	}
	if len(quotaConfig.Details) == 0 {
		logs.Errorf("can not get biz rolling server quota config, bizID: %d, rid: %s", bizID, kt.Rid)
		return 0, fmt.Errorf("can not get biz rolling server quota config, bizID: %d", bizID)
	}
	quota := *quotaConfig.Details[0].Quota

	listQuotaOffsetReq := &rsproto.RollingQuotaOffsetListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: rules},
		Fields: []string{"quota_offset"},
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}
	quotaOffsetConfig, err := l.client.DataService().Global.RollingServer.ListQuotaOffset(kt, listQuotaOffsetReq)
	if err != nil {
		logs.Errorf("list rolling server quota offset config failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return 0, err
	}

	if len(quotaOffsetConfig.Details) != 0 {
		quota += *quotaOffsetConfig.Details[0].QuotaOffset
	}

	return quota, nil
}
