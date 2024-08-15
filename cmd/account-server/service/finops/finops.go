/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package finops

import (
	"net/http"

	"hcm/cmd/account-server/service/capability"
	accountserver "hcm/pkg/api/account-server"
	"hcm/pkg/api/core"
	coreop "hcm/pkg/api/core/operation-product"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/tools/slice"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	svr := &service{
		client: c.ApiClient,
		finops: c.Finops,
	}

	h := rest.NewHandler()
	h.Add("ListOpProduct", http.MethodPost, "/operation_products/list", svr.ListOpProduct)
	h.Add("GetOpProduct", http.MethodPost, "/operation_products/{op_product_id}", svr.GetOpProduct)

	h.Load(c.WebService)
}

type service struct {
	client *client.ClientSet
	finops finops.Client
}

// ListOpProduct 运营产品列表
func (svc *service) ListOpProduct(cts *rest.Contexts) (any, error) {
	req := new(accountserver.ListOpProductReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	param := &finops.ListOpProductParam{
		BgIds:          req.BgIds,
		DeptIds:        req.DeptIds,
		OpProductIds:   req.OpProductIds,
		OpProductNames: req.OpProductName,
		Page:           req.Page,
	}
	productResult, err := svc.finops.ListOpProduct(cts.Kit, param)
	if err != nil {
		logs.Errorf("fail to call finops ListOpProduct api, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	result := core.ListResultT[coreop.OperationProduct]{
		Count:   productResult.Count,
		Details: slice.Map(productResult.Items, convOpProduct),
	}
	return result, nil
}

// GetOpProduct 获取运营产品
func (svc *service) GetOpProduct(cts *rest.Contexts) (any, error) {

	opProductId, err := cts.PathParameter("op_product_id").Int64()
	if opProductId <= 0 {
		return nil, errf.New(errf.InvalidParameter, "op_product_id invalid")
	}
	param := &finops.ListOpProductParam{

		OpProductIds: []int64{opProductId},

		Page: core.BasePage{
			Count: false,
			Start: 0,
			Limit: 1,
		},
	}
	productResult, err := svc.finops.ListOpProduct(cts.Kit, param)
	if err != nil {
		logs.Errorf("fail to call finops ListOpProduct api, err: %v, op_product_id: %d, rid: %s",
			err, opProductId, cts.Kit.Rid)
		return nil, err
	}
	if len(productResult.Items) == 0 {
		return nil, errf.New(errf.RecordNotFound, "operation product not found")
	}

	return convOpProduct(productResult.Items[0]), nil
}

func convOpProduct(p finops.OperationProduct) coreop.OperationProduct {
	return coreop.OperationProduct{
		OpProductId:       p.OpProductId,
		OpProductName:     p.OpProductName,
		OpProductManagers: p.PrincipalName,
		// 没有独立的备份负责人字段
		OpProductBakManagers: "",
		PlanProductId:        p.PlProductId,
		PlanProductName:      p.PlProductName,
		BgId:                 p.BgId,
		BgName:               p.BgName,
		BgShortName:          p.BgShortName,
		DeptId:               p.DeptId,
		DeptName:             p.DeptName,
	}
}
