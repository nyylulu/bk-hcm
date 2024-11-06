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
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
)

// ListRollingReturnedRecord list rolling returned record
func (svc *service) ListRollingReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.RollingReturnedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.RollingReturnedRecord().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	return &rsproto.RollingReturnedRecordListResult{Details: data.Details, Count: data.Count}, nil
}

// GetRollingReturnedCoreSum get rolling returned core sum
func (svc *service) GetRollingReturnedCoreSum(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.RollingReturnedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}

	result, err := svc.dao.RollingReturnedRecord().GetReturnedSumDeliveredCore(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	return result, nil
}
