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

// Package resplanweek ...
package resplanweek

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-week"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanWeek create resource plan week
func (svc *service) BatchCreateResPlanWeek(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanWeekBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	weekIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.batchCreateResPlanWeekWithTx(cts.Kit, txn, req.Weeks)
		if err != nil {
			logs.Errorf("failed to batch create resource plan week with tx, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create resource plan week failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(weekIDs)
	if err != nil {
		logs.Errorf("create resource plan week but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create resource plan week but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *service) batchCreateResPlanWeekWithTx(kt *kit.Kit, txn *sqlx.Tx, createReqs []rpproto.ResPlanWeekCreateReq) (
	[]string, error) {

	models := make([]tablers.ResPlanWeekTable, len(createReqs))
	for idx, item := range createReqs {
		createT := tablers.ResPlanWeekTable{
			Year:      item.Year,
			Month:     item.Month,
			YearWeek:  item.YearWeek,
			Start:     item.Start,
			End:       item.End,
			IsHoliday: item.IsHoliday,
		}

		models[idx] = createT
	}
	recordIDs, err := svc.dao.ResPlanWeek().CreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("create resource plan week failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create resource plan week failed, err: %v", err)
	}
	return recordIDs, nil
}
