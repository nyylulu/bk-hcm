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
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListShortRentalReturnedRecord list short rental returned records.
func (svc *service) ListShortRentalReturnedRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	return svc.dao.ShortRentalReturnedRecord().List(cts.Kit, opt)
}

// SumShortRentalReturnedCore sum short rental returned record returned core
func (svc *service) SumShortRentalReturnedCore(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ShortRentalReturnedRecordSumReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 构建filter表达式
	rules := make([]filter.RuleFactory, 0)
	if len(req.PhysicalDeviceFamilies) > 0 {
		rules = append(rules, tools.RuleIn("physical_device_family", req.PhysicalDeviceFamilies))
	}
	if len(req.RegionNames) > 0 {
		rules = append(rules, tools.RuleIn("region_name", req.RegionNames))
	}
	if len(req.OpProductIDs) > 0 {
		rules = append(rules, tools.RuleIn("op_product_id", req.OpProductIDs))
	}
	rules = append(rules, tools.RuleEqual("year", req.Year))
	rules = append(rules, tools.RuleEqual("month", req.Month))
	// 统计已回收核心数时，排除掉回收终止的条目
	rules = append(rules, tools.RuleNotEqual("status", enumor.ShortRentalStatusTerminate))

	expr := &filter.Expression{
		Op:    filter.And,
		Rules: rules,
	}

	result, err := svc.dao.ShortRentalReturnedRecord().SumReturnedCore(cts.Kit, expr)
	if err != nil {
		logs.Errorf("sum short rental returned record returned core failed, err: %v", err)
		return nil, err
	}

	return rpproto.ShortRentalReturnedRecordSumResult{Records: result}, nil
}
