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
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-changelog"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateDemandChangelog update demand changelog
func (svc *service) BatchUpdateDemandChangelog(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.DemandChangelogBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, updateReq := range req.Changelogs {
			record := &tablers.DemandChangelogTable{
				ID:         updateReq.ID,
				DemandID:   updateReq.DemandID,
				TicketID:   updateReq.TicketID,
				CrpOrderID: updateReq.CrpOrderID,
				SuborderID: updateReq.SuborderID,
				Type:       updateReq.Type,
				ObsProject: updateReq.ObsProject,
				RegionName: updateReq.RegionName,
				ZoneName:   updateReq.ZoneName,
				DeviceType: updateReq.DeviceType,
				Remark:     updateReq.Remark,
			}
			// 把字符串类型的[期望交付时间转]为符合格式的Int类型
			if len(updateReq.ExpectTime) > 0 {
				expectTimeInt, err := times.ConvStrTimeToInt(updateReq.ExpectTime, constant.DateLayout)
				if err != nil {
					return nil, errf.NewFromErr(errf.InvalidParameter, err)
				}
				record.ExpectTime = expectTimeInt
			}
			if updateReq.OSChange != nil {
				record.OSChange = &types.Decimal{Decimal: cvt.PtrToVal(updateReq.OSChange)}
			}
			if updateReq.CpuCoreChange != nil {
				record.CpuCoreChange = updateReq.CpuCoreChange
			}
			if updateReq.MemoryChange != nil {
				record.MemoryChange = updateReq.MemoryChange
			}
			if updateReq.DiskSizeChange != nil {
				record.DiskSizeChange = updateReq.DiskSizeChange
			}

			if err := svc.dao.ResPlanDemandChangelog().UpdateWithTx(cts.Kit, txn,
				tools.EqualExpression("id", updateReq.ID), record); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update demand changelog failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
