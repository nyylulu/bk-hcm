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
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanDemand create resource plan demand
func (svc *service) BatchCreateResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	demandIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablers.ResPlanDemandTable, len(req.Demands))
		for idx, item := range req.Demands {
			models[idx] = tablers.ResPlanDemandTable{
				Locked:          cvt.ValToPtr(enumor.CrpDemandUnLocked),
				BkBizID:         item.BkBizID,
				BkBizName:       item.BkBizName,
				OpProductID:     item.OpProductID,
				OpProductName:   item.OpProductName,
				PlanProductID:   item.PlanProductID,
				PlanProductName: item.PlanProductName,
				VirtualDeptID:   item.VirtualDeptID,
				VirtualDeptName: item.VirtualDeptName,
				DemandClass:     item.DemandClass,
				ObsProject:      item.ObsProject,
				ExpectTime:      item.ExpectTime,
				PlanType:        item.PlanType,
				AreaID:          item.AreaID,
				AreaName:        item.AreaName,
				RegionID:        item.RegionID,
				RegionName:      item.RegionName,
				ZoneID:          item.ZoneID,
				ZoneName:        item.ZoneName,
				DeviceFamily:    item.DeviceFamily,
				DeviceClass:     item.DeviceClass,
				DeviceType:      item.DeviceType,
				CoreType:        item.CoreType,
				DiskType:        item.DiskType,
				DiskTypeName:    item.DiskTypeName,
				OS:              &types.Decimal{Decimal: item.OS},
				CpuCore:         cvt.ValToPtr(item.CpuCore),
				Memory:          cvt.ValToPtr(item.Memory),
				DiskSize:        cvt.ValToPtr(item.DiskSize),
				DiskIO:          item.DiskIO,
				Creator:         cts.Kit.User,
				Reviser:         cts.Kit.User,
			}
		}
		recordIDs, err := svc.dao.ResPlanDemand().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create resource plan demand failed, err: %v", err)
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create resource plan demand failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(demandIDs)
	if err != nil {
		logs.Errorf("create resource plan demand but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create resource plan demand but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
