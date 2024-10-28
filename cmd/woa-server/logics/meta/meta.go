/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package meta is meta logics related package.
package meta

import (
	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/maps"
)

// GetBizsByOpProd get bizs by op product.
func (l *logics) GetBizsByOpProd(kt *kit.Kit, prodID int64) ([]mtypes.Biz, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]mtypes.Biz, 0)
	for _, bizBelong := range allBizBelonging {
		if bizBelong.OpProductID == prodID {
			result = append(result, mtypes.Biz{
				BkBizID:   bizBelong.BizID,
				BkBizName: bizBelong.BizName,
			})
		}
	}

	return result, nil
}

// getCmdbAllBizBelonging get cmdb all biz belonging.
func (l *logics) getCmdbAllBizBelonging(kt *kit.Kit) ([]cmdb.SearchBizBelonging, error) {
	result := make([]cmdb.SearchBizBelonging, 0)
	batch := constant.SearchBizBelongingMaxLimit
	for start := 0; ; start += batch {
		req := &cmdb.SearchBizBelongingParams{
			Page: cmdb.SearchBizBelongingPage{
				Limit: batch,
				Start: start,
			},
		}

		resp, err := l.esbClient.Cmdb().SearchBizBelonging(kt, req)
		if err != nil {
			logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result = append(result, *resp...)

		if len(*resp) < batch {
			break
		}
	}

	return result, nil
}

// GetOpProducts get op products.
func (l *logics) GetOpProducts(kt *kit.Kit) ([]mtypes.OpProduct, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// use map to unique op product.
	opProdMap := make(map[int64]mtypes.OpProduct)
	for _, bizBelong := range allBizBelonging {
		opProdMap[bizBelong.OpProductID] = mtypes.OpProduct{
			OpProductID:   bizBelong.OpProductID,
			OpProductName: bizBelong.OpProductName,
		}
	}

	result := maps.Values(opProdMap)
	return result, nil
}

// GetPlanProducts get op products.
func (l *logics) GetPlanProducts(kt *kit.Kit) ([]mtypes.PlanProduct, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// use map to unique plan product.
	planProdMap := make(map[int64]mtypes.PlanProduct)
	for _, bizBelong := range allBizBelonging {
		planProdMap[bizBelong.PlanProductID] = mtypes.PlanProduct{
			PlanProductID:   bizBelong.PlanProductID,
			PlanProductName: bizBelong.PlanProductName,
		}
	}

	result := maps.Values(planProdMap)
	return result, nil
}
