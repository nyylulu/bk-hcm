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

// Package tcaplusapi provides the client to call tcaplus api
package tcaplusapi

import (
	"context"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// TcaplusClientInterface tcaplus api interface
type TcaplusClientInterface interface {
	// CheckTcaplus check if a host has tcaplus records or not
	CheckTcaplus(ctx context.Context, header http.Header, ip string) (*TcaplusResp, error)
}

// NewTcaplusClientInterface creates a tcaplus api instance
func NewTcaplusClientInterface(opts cc.TcaplusCli, reg prometheus.Registerer) (TcaplusClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "tcaplus api",
			servers: []string{opts.TcaplusApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &tcaplusApi{
		client: rest.NewClient(c, ""),
	}

	return client, nil
}

// tcaplusApi tcaplus api interface implementation
type tcaplusApi struct {
	client rest.ClientInterface
}

// CheckTcaplus check if a host has tcaplus records or not
func (t *tcaplusApi) CheckTcaplus(ctx context.Context, header http.Header, ip string) (*TcaplusResp, error) {
	req := &TcaplusReq{
		IpList: []string{ip},
	}

	subPath := "/app/newoms.php/webservice/host/check-ip-exist"
	params := map[string]string{
		"cmd":     "10007",
		"ip-type": "sa",
	}
	resp := new(TcaplusResp)
	err := t.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		WithParams(params).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
