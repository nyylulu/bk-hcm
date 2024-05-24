/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iamapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// IAMClientInterface iam api interface
type IAMClientInterface interface {
	// AuthVerify auth policy verify
	AuthVerify(ctx context.Context, header http.Header, req *AuthVerifyReq) (*AuthVerifyResp, error)
	// GetAuthUrl get no permission apply url
	GetAuthUrl(ctx context.Context, header http.Header, req *GetAuthUrlReq) (*GetAuthUrlResp, error)
}

// NewIAMClientInterface creates iam api instance
func NewIAMClientInterface(opts cc.IamCli, reg prometheus.Registerer) (IAMClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "iam api",
			servers: []string{opts.IAMApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &iamApi{
		client: rest.NewClient(c, ""),
	}

	return client, nil
}

// iamApi iam api interface implementation
type iamApi struct {
	client rest.ClientInterface
	opts   *cc.IamCli
}

func (i *iamApi) getAuthHeader() (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", i.opts.AppCode,
		i.opts.AppSecret, i.opts.Operator)

	return key, val
}

// AuthVerify auth policy verify
func (i *iamApi) AuthVerify(ctx context.Context, header http.Header, req *AuthVerifyReq) (*AuthVerifyResp, error) {
	subPath := "/api/v1/policy/auth"
	key, val := i.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(AuthVerifyResp)
	err := i.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetAuthUrl auth policy verify
func (i *iamApi) GetAuthUrl(ctx context.Context, header http.Header, req *GetAuthUrlReq) (*GetAuthUrlResp, error) {
	subPath := "/api/v1/open/application"
	key, val := i.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(GetAuthUrlResp)
	err := i.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
