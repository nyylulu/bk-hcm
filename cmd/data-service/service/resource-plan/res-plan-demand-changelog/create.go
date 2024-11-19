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

// Package demandchangelog ...
package demandchangelog

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-changelog"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateDemandChangelog create demand changelog
func (svc *service) BatchCreateDemandChangelog(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.DemandChangelogCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	changelogIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablers.DemandChangelogTable, len(req.Changelogs))
		for idx, item := range req.Changelogs {
			models[idx] = tablers.DemandChangelogTable{
				DemandID:       item.DemandID,
				TicketID:       item.TicketID,
				CrpOrderID:     item.CrpOrderID,
				SuborderID:     item.SuborderID,
				Type:           item.Type,
				ExpectTime:     item.ExpectTime,
				ObsProject:     item.ObsProject,
				RegionName:     item.RegionName,
				ZoneName:       item.ZoneName,
				DeviceType:     item.DeviceType,
				OSChange:       &types.Decimal{Decimal: cvt.PtrToVal(item.OSChange)},
				CpuCoreChange:  item.CpuCoreChange,
				MemoryChange:   item.MemoryChange,
				DiskSizeChange: item.DiskSizeChange,
				Remark:         item.Remark,
			}
		}
		recordIDs, err := svc.dao.ResPlanDemandChangelog().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create demand changelog failed, err: %v", err)
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create demand changelog failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(changelogIDs)
	if err != nil {
		logs.Errorf("create demand changelog but return ids type not []string, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("create demand changelog but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
