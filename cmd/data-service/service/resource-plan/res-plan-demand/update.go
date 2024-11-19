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

// Package resplandemand ...
package resplandemand

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateResPlanDemand update resource plan demand
func (svc *service) BatchUpdateResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, updateReq := range req.Demands {
			record := &tablers.ResPlanDemandTable{
				ID:              updateReq.ID,
				OpProductID:     updateReq.OpProductID,
				OpProductName:   updateReq.OpProductName,
				PlanProductID:   updateReq.PlanProductID,
				PlanProductName: updateReq.PlanProductName,
				VirtualDeptID:   updateReq.VirtualDeptID,
				VirtualDeptName: updateReq.VirtualDeptName,
				CoreType:        updateReq.CoreType,
				Reviser:         cts.Kit.User,
			}
			if updateReq.OS != nil {
				record.OS = &types.Decimal{Decimal: cvt.PtrToVal(updateReq.OS)}
			}
			if updateReq.CpuCore != nil {
				record.CpuCore = updateReq.CpuCore
			}
			if updateReq.Memory != nil {
				record.Memory = updateReq.Memory
			}
			if updateReq.DiskSize != nil {
				record.DiskSize = updateReq.DiskSize
			}

			if err := svc.dao.ResPlanDemand().UpdateWithTx(cts.Kit, txn,
				tools.EqualExpression("id", updateReq.ID), record); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// LockResPlanDemand lock resource plan demand
func (svc *service) LockResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandLockOpReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.ResPlanDemand().ExamineAndLockAllRPDemand(cts.Kit, req.IDs); err != nil {
		logs.Errorf("lock res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UnlockResPlanDemand unlock resource plan demand
func (svc *service) UnlockResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandLockOpReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.ResPlanDemand().UnlockAllResPlanDemand(cts.Kit, req.IDs); err != nil {
		logs.Errorf("unlock res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
