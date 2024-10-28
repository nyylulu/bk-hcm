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

// Package tgwapi provides the client to interact with tgw api
package tgwapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// TgwClientInterface tgw api interface
type TgwClientInterface interface {
	// CheckTgw checks if a host has tgw policy or not
	CheckTgw(ctx context.Context, header http.Header, ip string) (*TgwResp, error)
	// CheckTgwNat checks if a host has tgw nat policy or not
	CheckTgwNat(ctx context.Context, header http.Header, ip string) (*TgwNatResp, error)
}

// NewTgwClientInterface creates a tgw api instance
func NewTgwClientInterface(opts cc.TGWCli, reg prometheus.Registerer) (TgwClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "tgw api",
			servers: []string{opts.TgwApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &tgwApi{
		client: rest.NewClient(c, "/"),
	}

	return client, nil
}

// tgwApi tgw api interface implementation
type tgwApi struct {
	client rest.ClientInterface
}

// CheckTgw checks if a host has tgw policy or not
func (g *tgwApi) CheckTgw(ctx context.Context, header http.Header, ip string) (*TgwResp, error) {
	param := fmt.Sprintf(`{"operator":"IEDSO","biztype":"COMMON", "rsiplist":["%s"]}`, ip)

	subPath := "/cgi-bin/fun_logic/bin/public_api/getrs.cgi"
	resp := new(TgwResp)
	err := g.client.Post().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(header).
		WithParam("data", param).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckTgwNat checks if a host has tgw nat policy or not
func (g *tgwApi) CheckTgwNat(ctx context.Context, header http.Header, ip string) (*TgwNatResp, error) {
	param := fmt.Sprintf(`{"operator":"IEDSO", "rs_ip_list":["%s"]}`, ip)

	subPath := "/cgi-bin/fun_logic/bin/public_api/get_rs_nat.cgi"
	resp := new(TgwNatResp)
	err := g.client.Post().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(header).
		WithParam("data", param).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
