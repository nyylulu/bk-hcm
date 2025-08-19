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
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"
)

// ListCrpDemandChangeLog list crp demand change log by demand id
func (c *Controller) ListCrpDemandChangeLog(kt *kit.Kit, req *ptypes.ListDemandChangeLogReq) (
	*ptypes.ListDemandChangeLogResp, error) {

	listRules := make([]*filter.AtomRule, 0)
	listRules = append(listRules, tools.RuleEqual("demand_id", req.DemandID))

	listReq := &rpproto.DemandChangelogListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(listRules...),
			Page:   req.Page,
		},
	}

	rst, err := c.client.DataService().Global.ResourcePlan.ListDemandChangelog(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list crp demand change log, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &ptypes.ListDemandChangeLogResp{Count: rst.Count}, nil
	}

	rstDetails := make([]*ptypes.ListDemandChangeLogItem, len(rst.Details))
	for idx, tableDetail := range rst.Details {
		expectTimeStr, err := times.TransTimeStrWithLayout(strconv.Itoa(tableDetail.ExpectTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Errorf("failed to convert expect time to string, err: %v, expect time: %d, rid: %s", err,
				tableDetail.ExpectTime, kt.Rid)
			return nil, err
		}

		rstDetails[idx] = &ptypes.ListDemandChangeLogItem{
			ID:                tableDetail.ID,
			DemandId:          tableDetail.DemandID,
			ExpectTime:        expectTimeStr,
			ObsProject:        tableDetail.ObsProject,
			RegionName:        tableDetail.RegionName,
			ZoneName:          tableDetail.ZoneName,
			DeviceType:        tableDetail.DeviceType,
			ChangeCvmAmount:   tableDetail.OSChange.Decimal,
			ChangeCoreAmount:  cvt.PtrToVal(tableDetail.CpuCoreChange),
			ChangeRamAmount:   cvt.PtrToVal(tableDetail.MemoryChange),
			ChangedDiskAmount: cvt.PtrToVal(tableDetail.DiskSizeChange),
			DemandSource:      tableDetail.Type.Name(),
			TicketID:          tableDetail.TicketID,
			CrpSn:             tableDetail.CrpOrderID,
			SuborderID:        tableDetail.SuborderID,
			CreateTime:        tableDetail.CreatedAt.String(),
			Remark:            tableDetail.Remark,
		}
	}

	return &ptypes.ListDemandChangeLogResp{
		Count:   0,
		Details: rstDetails,
	}, nil
}
