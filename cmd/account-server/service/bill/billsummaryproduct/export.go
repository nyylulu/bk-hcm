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
	"fmt"
	"time"

	"hcm/cmd/account-server/logics/bill/export"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/tools/slice"

	"github.com/TencentBlueKing/gopkg/conv"
)

// ExportProductSummary export product summary with options
func (s *service) ExportProductSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(bill.ProductSummaryExportReq)
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

	result, err := s.fetchProductSummary(cts, req)
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, 0, len(result))
	for _, detail := range result {
		productIDs = append(productIDs, detail.ProductID)
	}
	productMap, err := s.listOpProduct(cts.Kit, productIDs)
	if err != nil {
		logs.Errorf("list op product failed, productIDs: %v, err: %v, rid: %s", productIDs, err, cts.Kit.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(cts.Kit, generateFilename())
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	for _, header := range export.BillSummaryProductTableHeader {
		if err := writer.Write(header); err != nil {
			logs.Errorf("write header failed: %v, val: %v, rid: %s", err, header, cts.Kit.Rid)
			return nil, err
		}
	}

	table, err := toRawData(cts.Kit, result, productMap)
	if err != nil {
		logs.Errorf("convert to raw data failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := writer.WriteAll(table); err != nil {
		logs.Errorf("write data failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &bill.FileDownloadResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
}

func generateFilename() string {
	return fmt.Sprintf("bill_summary_product-%s.csv", time.Now().Format("2006-01-02-15_04_05"))
}

func toRawData(kt *kit.Kit, details []*billproto.BillSummaryProductResult,
	productMap map[int64]finops.OperationProduct) ([][]string, error) {

	result := make([][]string, 0, len(details))
	for _, detail := range details {
		row := export.BillSummaryProductTable{
			ProductID:                 conv.ToString(detail.ProductID),
			ProductName:               productMap[detail.ProductID].OpProductName,
			CurrentMonthRMBCostSynced: detail.CurrentMonthRMBCostSynced.String(),
			CurrentMonthCostSynced:    detail.CurrentMonthCostSynced.String(),
			CurrentMonthRMBCost:       detail.CurrentMonthRMBCost.String(),
			CurrentMonthCost:          detail.CurrentMonthCost.String(),
		}
		fields, err := row.GetValuesByHeader()
		if err != nil {
			logs.Errorf("get header fields failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		result = append(result, fields)
	}
	return result, nil
}

func (s *service) fetchProductSummary(cts *rest.Contexts, req *bill.ProductSummaryExportReq) (
	[]*billproto.BillSummaryProductResult, error) {

	if len(req.OpProductIDs) == 0 {
		return s.fetchAllProductSummary(cts, req)
	}

	result := make([]*billproto.BillSummaryProductResult, 0)
	for _, productIDs := range slice.Split(req.OpProductIDs, int(filter.DefaultMaxInLimit)) {
		expression := tools.ExpressionAnd(
			tools.RuleEqual("bill_year", req.BillYear),
			tools.RuleEqual("bill_month", req.BillMonth),
			tools.RuleIn("product_id", productIDs),
		)
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryProduct(cts.Kit, &core.ListReq{
			Filter: expression,
			Page:   core.NewDefaultBasePage(),
		})
		if err != nil {
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}
	if len(result) > int(req.ExportLimit) {
		result = result[:req.ExportLimit]
	}
	return result, nil
}

func (s *service) fetchAllProductSummary(cts *rest.Contexts, req *bill.ProductSummaryExportReq) (
	[]*billproto.BillSummaryProductResult, error) {

	expression := tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)

	details, err := s.client.DataService().Global.Bill.ListBillSummaryProduct(cts.Kit, &core.ListReq{
		Filter: expression,
		Page:   core.NewCountPage(),
	})
	if err != nil {
		return nil, err
	}

	limit := details.Count
	if req.ExportLimit <= limit {
		limit = req.ExportLimit
	}

	result := make([]*billproto.BillSummaryProductResult, 0, limit)
	page := core.DefaultMaxPageLimit
	for offset := uint64(0); offset < limit; offset = offset + uint64(core.DefaultMaxPageLimit) {
		if limit-offset < uint64(page) {
			page = uint(limit - offset)
		}
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryProduct(cts.Kit, &core.ListReq{
			Filter: expression,
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: page,
			},
		})
		if err != nil {
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}
	return result, nil
}
