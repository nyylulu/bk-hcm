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

// Package shortrentalreturnedrecord ...
package shortrentalreturnedrecord

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablesr "hcm/pkg/dal/table/short-rental"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateShortRentalReturnedRecord batch update short rental returned records.
func (svc *service) BatchUpdateShortRentalReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ShortRentalReturnedRecordBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := svc.batchUpdateShortRentalReturnedRecordWithTx(cts.Kit, txn, req.Records)
		if err != nil {
			logs.Errorf("failed to batch update short rental returned record with tx, err: %v, rid: %v",
				err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update short rental returned record failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateShortRentalReturnedRecordWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.ShortRentalReturnedRecordUpdateReq) error {

	for _, updateReq := range updateReqs {
		record := &tablesr.ShortRentalReturnedRecordTable{
			ID:      updateReq.ID,
			Reviser: kt.User,
		}
		if updateReq.Status != "" {
			record.Status = updateReq.Status
		}
		if updateReq.ReturnedCore != nil {
			record.ReturnedCore = updateReq.ReturnedCore
		}
		if updateReq.Year != nil {
			record.Year = *updateReq.Year
		}
		if updateReq.Month != nil {
			record.Month = *updateReq.Month
		}
		if updateReq.ReturnedDate != nil {
			record.ReturnedDate = *updateReq.ReturnedDate
		}

		if err := svc.dao.ShortRentalReturnedRecord().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update short rental returned record failed, err: %v, id: %d, rid: %s",
				err, updateReq.ID, kt.Rid)
			return err
		}
	}
	return nil
}
