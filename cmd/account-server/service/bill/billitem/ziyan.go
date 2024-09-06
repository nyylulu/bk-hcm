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

package billitem

import (
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/finops"
)

// fetchAccountProductInfo 根据vendor获取所有关联的数据
func (b *billItemSvc) fetchAccountProductInfo(kt *kit.Kit, vendor enumor.Vendor) (
	rootAccountMap map[string]*accountset.BaseRootAccount, mainAccountMap map[string]*accountset.BaseMainAccount,
	opProductMap map[int64]finops.OperationProduct, err error) {

	opProductMap, err = b.listOpProduct(kt)
	if err != nil {
		logs.Errorf("fail to list op product, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	mainAccounts, err := b.listMainAccount(kt, vendor)
	if err != nil {
		logs.Errorf("fail to list main account, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
		return nil, nil, nil, err
	}
	mainAccountMap = make(map[string]*accountset.BaseMainAccount, len(mainAccounts))
	for _, account := range mainAccounts {
		mainAccountMap[account.ID] = account
	}

	rootAccountMap, err = b.listRootAccount(kt, vendor)
	if err != nil {
		logs.Errorf("fail to list root account, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
		return nil, nil, nil, err
	}
	return rootAccountMap, mainAccountMap, opProductMap, nil
}

func (b *billItemSvc) listOpProduct(kt *kit.Kit) (map[int64]finops.OperationProduct, error) {

	offset := uint32(0)
	result := make(map[int64]finops.OperationProduct)
	for {
		param := &finops.ListOpProductParam{
			Page: core.BasePage{
				Start: offset,
				Limit: core.DefaultMaxPageLimit,
			},
		}
		productResult, err := b.finops.ListOpProduct(kt, param)
		if err != nil {
			return nil, err
		}
		if len(productResult.Items) == 0 {
			break
		}
		offset += uint32(core.DefaultMaxPageLimit)
		for _, product := range productResult.Items {
			result[product.OpProductId] = product
		}
	}

	return result, nil
}
