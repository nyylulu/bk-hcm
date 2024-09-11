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

package billsummaryproduct

import (
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/tools/slice"
)

// ListProductSummary list product summary with options
func (s *service) ListProductSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(bill.ProductSummaryListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	var expression = tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)
	if len(req.OpProductIDs) > 0 {
		expression, err = tools.And(expression, tools.RuleIn("product_id", req.OpProductIDs))
		if err != nil {
			return nil, err
		}
	}

	summary, err := s.client.DataService().Global.Bill.ListBillSummaryProduct(cts.Kit, &core.ListReq{
		Filter: expression,
		Page:   req.Page,
	})
	if err != nil {
		return nil, err
	}
	if len(summary.Details) == 0 {
		return summary, nil
	}

	// 补全 product_name
	productIDs := make([]int64, 0, len(summary.Details))
	for _, detail := range summary.Details {
		productIDs = append(productIDs, detail.ProductID)
	}
	productMap, err := s.listOpProduct(cts.Kit, productIDs)
	if err != nil {
		return nil, err
	}

	for _, detail := range summary.Details {
		detail.ProductName = productMap[detail.ProductID].OpProductName
	}

	return summary, nil
}

func (s *service) listOpProduct(kt *kit.Kit, ids []int64) (map[int64]finops.OperationProduct, error) {
	ids = slice.Unique(ids)
	result := make(map[int64]finops.OperationProduct, len(ids))
	for _, tmpIds := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		param := &finops.ListOpProductParam{
			OpProductIds: tmpIds,
			Page:         *core.NewDefaultBasePage(),
		}

		productResult, err := s.finops.ListOpProduct(kt, param)
		if err != nil {
			return nil, err
		}
		for _, product := range productResult.Items {
			result[product.OpProductId] = product
		}
	}

	return result, nil
}
