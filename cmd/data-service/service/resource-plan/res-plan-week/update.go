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
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-week"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateResPlanWeek update resource plan week
func (svc *service) BatchUpdateResPlanWeek(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanWeekBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := svc.batchUpdateResPlanWeekWithTx(cts.Kit, txn, req.Weeks)
		if err != nil {
			logs.Errorf("failed to batch update res plan week with tx, err: %v, rid: %v", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan week failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateResPlanWeekWithTx(kt *kit.Kit, txn *sqlx.Tx, updateReqs []rpproto.ResPlanWeekUpdateReq) (
	[]string, error) {

	for _, updateReq := range updateReqs {
		record := &tablers.ResPlanWeekTable{
			ID:    updateReq.ID,
			Start: updateReq.Start,
			End:   updateReq.End,
		}
		if updateReq.IsHoliday != nil {
			record.IsHoliday = updateReq.IsHoliday
		}

		if err := svc.dao.ResPlanWeek().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update res plan week failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	return nil, nil
}
