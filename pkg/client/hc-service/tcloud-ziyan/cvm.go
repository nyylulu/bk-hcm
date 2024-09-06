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
	"context"
	"net/http"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCvmClient create a new cvm api client.
func NewCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// CvmClient is hc service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// QueryTCloudZiyanCVM  查询云上cvm
func (cli *CvmClient) QueryTCloudZiyanCVM(kt *kit.Kit, request *cvm.QueryCloudCvmReq) (
	*core.ListResultT[cvm.Cvm[cvm.TCloudZiyanCvmExtension]], error) {

	return common.Request[cvm.QueryCloudCvmReq, core.ListResultT[cvm.Cvm[cvm.TCloudZiyanCvmExtension]]](cli.client,
		rest.POST, kt, request, "/cvms/query")
}

// SyncHostWithRelResource sync host with rel resource.
func (cli *CvmClient) SyncHostWithRelResource(ctx context.Context, h http.Header,
	request *sync.TCloudZiyanSyncHostReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/hosts/with/relation_resources/sync").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// SyncHostWithRelResByCond sync host with rel resource by conditon.
func (cli *CvmClient) SyncHostWithRelResByCond(ctx context.Context, h http.Header,
	request *sync.TCloudZiyanSyncHostByCondReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/hosts/with/relation_resources/by_condition/sync").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// DeleteHostByCond delete host by condition.
func (cli *CvmClient) DeleteHostByCond(ctx context.Context, h http.Header,
	request *sync.TCloudZiyanDelHostByCondReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/hosts/by_condition/delete").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}
