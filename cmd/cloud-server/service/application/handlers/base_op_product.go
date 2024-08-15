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

package handlers

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/thirdparty/api-gateway/finops"
)

// GetOperationProductName 查询运营产品名称
func (a *BaseApplicationHandler) GetOperationProductName(opId int64) (string, error) {
	opProduct, err := a.getOperationProduct(opId)
	if err != nil {
		return "", err
	}

	return opProduct.OpProductName, nil
}

// GetOperationProductManager 查询运营产品负责人
func (a *BaseApplicationHandler) GetOperationProductManager(opId int64) (string, error) {
	opProduct, err := a.getOperationProduct(opId)
	if err != nil {
		return "", err
	}

	return opProduct.PrincipalName, nil
}

func (a *BaseApplicationHandler) getOperationProduct(opId int64) (*finops.OperationProduct, error) {
	result, err := a.FinOpsCli.ListOpProduct(a.Cts.Kit, &finops.ListOpProductParam{
		OpProductIds: []int64{opId},
		Page: core.BasePage{
			Count: false,
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) < 1 {
		return nil, fmt.Errorf("not found op product by op_product_id(%d)", opId)
	}

	opProduct := result.Items[0]
	return &opProduct, nil
}
