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

// Package rollingquotaoffset ...
package rollingquotaoffset

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateRollingQuotaOffset create rolling quota offset.
func (svc *service) BatchCreateRollingQuotaOffset(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.RollingQuotaOffsetCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	quotaOffsetIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablers.RollingQuotaOffsetTable, len(req.QuotaOffsets))
		for idx, item := range req.QuotaOffsets {
			models[idx] = tablers.RollingQuotaOffsetTable{
				BkBizID:     item.BkBizID,
				BkBizName:   item.BkBizName,
				Year:        item.Year,
				Month:       item.Month,
				QuotaOffset: item.QuotaOffset,
				Creator:     cts.Kit.User,
				Reviser:     cts.Kit.User,
			}
		}
		recordIDs, err := svc.dao.RollingQuotaOffset().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create rolling quota offset failed, err: %v", err)
		}
		return recordIDs, nil
	})
	if err != nil {
		return nil, err
	}
	ids, ok := quotaOffsetIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create rolling quota offset but return ids type not []string, id type: %v",
			reflect.TypeOf(quotaOffsetIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
