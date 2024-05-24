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

package ziyan

import (
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountClient is data service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// Create account.
func (a *AccountClient) Create(kt *kit.Kit,
	request *protocloud.AccountCreateReq[protocloud.TCloudAccountExtensionCreateReq]) (*core.CreateResult, error) {

	return common.Request[protocloud.AccountCreateReq[protocloud.TCloudAccountExtensionCreateReq], core.CreateResult](
		a.client, rest.POST, kt, request, "/accounts/create")

}

// Update ...
func (a *AccountClient) Update(kt *kit.Kit, accountID string,
	request *protocloud.AccountUpdateReq[protocloud.TCloudAccountExtensionUpdateReq]) (any, error) {

	err := common.RequestNoResp[protocloud.AccountUpdateReq[protocloud.TCloudAccountExtensionUpdateReq]](
		a.client, rest.POST, kt, request, "/accounts/%s", accountID)
	return nil, err
}

// Get tcloud account detail.
func (a *AccountClient) Get(kt *kit.Kit, accountID string) (
	*protocloud.AccountGetResult[protocore.TCloudAccountExtension], error) {

	return common.Request[common.Empty, protocloud.AccountGetResult[protocore.TCloudAccountExtension]](
		a.client, rest.GET, kt, nil, "/accounts/%s", accountID)

}
