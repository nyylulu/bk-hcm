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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/constant"
	bpaas "hcm/pkg/thirdparty/tencentcloud/bpaas/v20181217"

	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type clientSet struct {
	credential *common.Credential
	profile    *profile.ClientProfile
}

// 自研云客户端集
func newClientSet(s *types.BaseSecret, profile *profile.ClientProfile) *clientSet {
	return &clientSet{
		credential: common.NewCredential(s.CloudSecretID, s.CloudSecretKey),
		profile:    profile,
	}

}

// CamServiceClient tcloud ziyan sdk cam client
func (c *clientSet) CamServiceClient(region string) (*cam.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalCamEndpoint
	client, err := cam.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CvmClient tcloud ziyan sdk cvm client
func (c *clientSet) CvmClient(region string) (*cvm.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalCvmEndpoint
	client, err := cvm.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CbsClient tcloud ziyan sdk cbs client
func (c *clientSet) CbsClient(region string) (*cbs.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalCbsEndpoint
	client, err := cbs.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// VpcClient tcloud ziyan sdk vpc client
func (c *clientSet) VpcClient(region string) (*vpc.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalVpcEndpoint
	client, err := vpc.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// BillClient tcloud ziyan sdk bill client
func (c *clientSet) BillClient() (*billing.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalBillingEndpoint
	client, err := billing.NewClient(c.credential, "", c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ClbClient tcloud clb client
func (c *clientSet) ClbClient(region string) (*clb.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalClbEndpoint
	client, err := clb.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CertClient tcloud cert client
func (c *clientSet) CertClient() (*ssl.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalClbEndpoint
	client, err := ssl.NewClient(c.credential, "", c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// BPaasClient tcloud ziyan sdk bpaas client
func (c *clientSet) BPaasClient() (*bpaas.Client, error) {
	// 使用内部域名
	c.profile.HttpProfile.Endpoint = constant.InternalBPaasEndpoint
	client, err := bpaas.NewClient(c.credential, "", c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}
