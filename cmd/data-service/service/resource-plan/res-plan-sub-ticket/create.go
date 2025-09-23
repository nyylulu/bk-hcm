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

// Package resplansubticket ...
package resplansubticket

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanSubTicket create resource plan sub ticket
func (svc *service) BatchCreateResPlanSubTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanSubTicketBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.batchCreateResPlanSubTicketWithTx(cts.Kit, txn, req.SubTickets)
		if err != nil {
			logs.Errorf("failed to batch create resource plan sub ticket with tx, err: %v, rid: %s", err,
				cts.Kit.Rid)
			return nil, err
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create resource plan sub ticket failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(createIDs)
	if err != nil {
		logs.Errorf("create resource plan sub ticket but return ids type not []string, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, fmt.Errorf("create resource plan sub ticket but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *service) batchCreateResPlanSubTicketWithTx(kt *kit.Kit, txn *sqlx.Tx,
	createReqs []rpproto.ResPlanSubTicketCreateReq) ([]string, error) {

	models := make([]tablers.ResPlanSubTicketTable, len(createReqs))
	for idx, item := range createReqs {
		createT := tablers.ResPlanSubTicketTable{
			TicketID:            item.TicketID,
			SubType:             item.SubType,
			SubDemands:          item.SubDemands,
			BkBizID:             item.BkBizID,
			BkBizName:           item.BkBizName,
			OpProductID:         item.OpProductID,
			OpProductName:       item.OpProductName,
			PlanProductID:       item.PlanProductID,
			PlanProductName:     item.PlanProductName,
			VirtualDeptID:       item.VirtualDeptID,
			VirtualDeptName:     item.VirtualDeptName,
			Status:              item.Status,
			Stage:               item.Stage,
			AdminAuditStatus:    item.AdminAuditStatus,
			CrpSN:               item.CrpSN,
			CrpURL:              item.CrpURL,
			Message:             cvt.ValToPtr(""),
			SubOriginalOS:       item.SubOriginalOS,
			SubOriginalCPUCore:  item.SubOriginalCPUCore,
			SubOriginalMemory:   item.SubOriginalMemory,
			SubOriginalDiskSize: item.SubOriginalDiskSize,
			SubUpdatedOS:        item.SubUpdatedOS,
			SubUpdatedCPUCore:   item.SubUpdatedCPUCore,
			SubUpdatedMemory:    item.SubUpdatedMemory,
			SubUpdatedDiskSize:  item.SubUpdatedDiskSize,
			SubmittedAt:         item.SubmittedAt,
			Creator:             kt.User,
			Reviser:             kt.User,
		}

		models[idx] = createT
	}
	recordIDs, err := svc.dao.ResPlanSubTicket().CreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("create resource plan sub ticket failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create resource plan sub ticket failed, err: %v", err)
	}
	return recordIDs, nil
}
