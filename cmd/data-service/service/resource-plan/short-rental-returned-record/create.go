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
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablesr "hcm/pkg/dal/table/short-rental"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateShortRentalReturnedRecord batch create short rental returned records.
func (svc *service) BatchCreateShortRentalReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ShortRentalReturnedRecordBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	newIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.batchCreateShortRentalReturnedRecordWithTx(cts.Kit, txn, req.Records)
		if err != nil {
			logs.Errorf("batch create short rental returned record failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return recordIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, err := util.GetStrSliceByInterface(newIDs)
	if err != nil {
		logs.Errorf("create short rental returned record but return ids type invalid, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create short rental returned record but return ids type invalid, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *service) batchCreateShortRentalReturnedRecordWithTx(kt *kit.Kit, txn *sqlx.Tx,
	createReqs []rpproto.ShortRentalReturnedRecordCreateReq) ([]string, error) {

	models := make([]tablesr.ShortRentalReturnedRecordTable, len(createReqs))
	for idx, item := range createReqs {
		models[idx] = tablesr.ShortRentalReturnedRecordTable{
			BkBizID:              item.BkBizID,
			BkBizName:            item.BkBizName,
			OpProductID:          item.OpProductID,
			OpProductName:        item.OpProductName,
			PlanProductID:        item.PlanProductID,
			PlanProductName:      item.PlanProductName,
			VirtualDeptID:        item.VirtualDeptID,
			VirtualDeptName:      item.VirtualDeptName,
			OrderID:              item.OrderID,
			SuborderID:           item.SuborderID,
			Year:                 item.Year,
			Month:                item.Month,
			ReturnedDate:         item.ReturnedDate,
			PhysicalDeviceFamily: item.PhysicalDeviceFamily,
			RegionID:             item.RegionID,
			RegionName:           item.RegionName,
			Status:               item.Status,
			ReturnedCore:         item.ReturnedCore,
			Creator:              kt.User,
			Reviser:              kt.User,
		}
	}

	recordIDs, err := svc.dao.ShortRentalReturnedRecord().CreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("create short rental returned record failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create short rental returned record failed, err: %w", err)
	}
	return recordIDs, nil
}
