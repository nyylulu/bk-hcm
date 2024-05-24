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

package hcziyancli

import (
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountClient is hc service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// Check 联通性和云上字段匹配校验
func (a *AccountClient) Check(kt *kit.Kit, request *hsaccount.TCloudAccountCheckReq) error {

	return common.RequestNoResp[hsaccount.TCloudAccountCheckReq](a.client, rest.POST, kt, request, "/accounts/check")
}

// GetBySecret get account info by secret
func (a *AccountClient) GetBySecret(kt *kit.Kit, request *cloud.TCloudSecret) (*cloud.TCloudInfoBySecret, error) {

	return common.Request[cloud.TCloudSecret, cloud.TCloudInfoBySecret](a.client, rest.POST, kt, request,
		"/accounts/secret")
}
