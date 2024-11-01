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

// Package rollingreturned ...
package rollingreturned

import (
	"fmt"

	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateRollingReturnedRecord batch update rolling returned record
func (svc *service) BatchUpdateRollingReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.BatchUpdateRollingReturnedRecordReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, updateReq := range req.ReturnedRecords {
			recordReq := &rstable.RollingReturnedRecord{
				ID: updateReq.ID,
			}
			if len(updateReq.AppliedRecordID) != 0 {
				recordReq.AppliedRecordID = updateReq.AppliedRecordID
			}
			if updateReq.MatchAppliedCore != nil {
				recordReq.MatchAppliedCore = updateReq.MatchAppliedCore
			}
			if len(updateReq.ReturnedWay) != 0 {
				recordReq.ReturnedWay = updateReq.ReturnedWay
			}
			if updateReq.Status > 0 {
				recordReq.Status = updateReq.Status
			}
			if err := svc.dao.RollingReturnedRecord().Update(
				cts.Kit, txn, tools.EqualExpression("id", updateReq.ID), recordReq); err != nil {
				return nil, fmt.Errorf("update rolling returned record failed, err: %v, id: %s", err, updateReq.ID)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
