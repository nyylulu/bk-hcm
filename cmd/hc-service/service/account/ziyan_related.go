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

package account

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// TCloudZiyanGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) TCloudZiyanGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.TCloudSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloudZiyan(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit)

}

// TCloudZiyanAccountCheck 根据传入秘钥去云上获取数据，并和传入其他数据对比，要求和云上获取数据一致
func (svc *service) TCloudZiyanAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloudZiyan(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		return nil, err
	}
	// check if cloud account info matches the hcm account detail.
	if infoBySecret.CloudSubAccountID != req.CloudSubAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubAccountID does not match the account to which the secret belongs")
	}

	if infoBySecret.CloudMainAccountID != req.CloudMainAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudMainAccountID does not match the account to which the secret belongs")
	}

	return nil, err
}

// GetTCloudZiyanNetworkAccountType ...
func (svc *service) GetTCloudZiyanNetworkAccountType(cts *rest.Contexts) (any, error) {

	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "accountID is required")
	}

	client, err := svc.ad.TCloudZiyan(cts.Kit, accountID)
	if err != nil {
		return nil, err
	}

	return client.DescribeNetworkAccountType(cts.Kit)
}
