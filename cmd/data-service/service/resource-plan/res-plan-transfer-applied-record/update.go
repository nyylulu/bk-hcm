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

// Package transferappliedrecord ...
package transferappliedrecord

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablerptar "hcm/pkg/dal/table/resource-plan/res-plan-transfer-applied-record"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateTransferAppliedRecord update resource plan transfer applied record
func (svc *service) BatchUpdateTransferAppliedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.TransferAppliedRecordBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := svc.batchUpdateTransferAppliedRecordWithTx(cts.Kit, txn, req.Records)
		if err != nil {
			logs.Errorf("failed to batch update res plan transfer applied record with tx, err: %v, rid: %v",
				err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan transfer applied record failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateTransferAppliedRecordWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.TransferAppliedRecordUpdateReq) ([]string, error) {

	for _, updateReq := range updateReqs {
		record := &tablerptar.ResPlanTransferAppliedRecordTable{
			ID:      updateReq.ID,
			Reviser: kt.User,
		}
		if updateReq.TechnicalClass != "" {
			record.TechnicalClass = updateReq.TechnicalClass
		}
		if updateReq.ObsProject != "" {
			record.ObsProject = updateReq.ObsProject
		}
		if updateReq.ExpectedCore != nil {
			record.ExpectedCore = updateReq.ExpectedCore
		}
		if updateReq.AppliedCore != nil {
			record.AppliedCore = updateReq.AppliedCore
		}

		if err := svc.dao.ResPlanTransferAppliedRecord().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update res plan transfer applied record failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	return nil, nil
}
