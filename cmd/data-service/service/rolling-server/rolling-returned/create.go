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
	"reflect"

	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// BatchCreateRollingReturnedRecord batch create rolling returned record
func (svc *service) BatchCreateRollingReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.BatchCreateRollingReturnedRecordReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	recordIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		records := make([]tablers.RollingReturnedRecord, 0, len(req.ReturnedRecords))
		for _, createReq := range req.ReturnedRecords {
			records = append(records, tablers.RollingReturnedRecord{
				BkBizID:          createReq.BkBizID,
				OrderID:          createReq.OrderID,
				SubOrderID:       createReq.SubOrderID,
				AppliedRecordID:  createReq.AppliedRecordID,
				MatchAppliedCore: cvt.ValToPtr(createReq.MatchAppliedCore),
				Year:             createReq.Year,
				Month:            createReq.Month,
				Day:              createReq.Day,
				ReturnedWay:      createReq.ReturnedWay,
			})
		}
		ids, err := svc.dao.RollingReturnedRecord().CreateWithTx(cts.Kit, txn, records)
		if err != nil {
			return nil, fmt.Errorf("create rolling returned record failed, err: %v", err)
		}
		if len(ids) != 1 {
			return nil, fmt.Errorf("create rolling returned record expect 1 puller IDs: %v", ids)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := recordIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create rolling returned record but return id type is not string, "+
			"id type: %v", reflect.TypeOf(recordIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
