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

// Package l5api is the client for l5 api
package l5api

import (
	"context"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// L5ClientInterface l5 api interface
type L5ClientInterface interface {
	// CheckL5 checks if a host has l5 sid or not
	CheckL5(ctx context.Context, header http.Header, ip string) (*L5Resp, error)
}

// NewL5ClientInterface creates a GCS api instance
func NewL5ClientInterface(opts cc.L5Cli, reg prometheus.Registerer) (L5ClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "l5 api",
			servers: []string{opts.L5ApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &l5Api{
		client: rest.NewClient(c, "/"),
	}

	return client, nil
}

// l5Api l5 api interface implementation
type l5Api struct {
	client rest.ClientInterface
}

// CheckL5 checks if a host has l5 sid or not
func (l *l5Api) CheckL5(ctx context.Context, header http.Header, ip string) (*L5Resp, error) {
	// TODO: get user from config
	req := &L5Req{
		Version:       1,
		ComponentName: "ieg-resourcerecycle",
		User:          "huibohuang",
		EventId:       20130720,
		Interface: &L5Interface{
			InterfaceName: "searchSid",
			Para: &L5Param{
				Ip: ip},
		},
	}

	subPath := "/interface/L5Interface.php"
	resp := new(L5Resp)
	err := l.client.Get().
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
