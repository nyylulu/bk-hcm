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

// Package demandpenaltybase ...
package demandpenaltybase

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-penalty-base"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateDemandPenaltyBase create demand penalty base
func (svc *service) BatchCreateDemandPenaltyBase(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.DemandPenaltyBaseCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	penaltyBaseIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablers.DemandPenaltyBaseTable, len(req.PenaltyBases))
		for idx, item := range req.PenaltyBases {
			models[idx] = tablers.DemandPenaltyBaseTable{
				Year:            item.Year,
				Month:           item.Month,
				Week:            item.Week,
				YearWeek:        item.YearWeek,
				Source:          item.Source,
				BkBizID:         item.BkBizID,
				BkBizName:       item.BkBizName,
				OpProductID:     item.OpProductID,
				OpProductName:   item.OpProductName,
				PlanProductID:   item.PlanProductID,
				PlanProductName: item.PlanProductName,
				VirtualDeptID:   item.VirtualDeptID,
				VirtualDeptName: item.VirtualDeptName,
				AreaName:        item.AreaName,
				DeviceFamily:    item.DeviceFamily,
				CpuCore:         item.CpuCore,
			}
		}
		recordIDs, err := svc.dao.ResPlanDemandPenaltyBase().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create demand penalty base failed, err: %v", err)
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create demand penalty base failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(penaltyBaseIDs)
	if err != nil {
		logs.Errorf("create demand penalty base but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create demand penalty base but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
