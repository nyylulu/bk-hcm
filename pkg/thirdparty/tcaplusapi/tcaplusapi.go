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
	"fmt"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// TcapulsCheckIPExistsMaxLength 与为了保证性能，需要把`iplist`的长度限制在200以内
const TcapulsCheckIPExistsMaxLength = 200

// TcaplusClientInterface tcaplus api interface
type TcaplusClientInterface interface {
	// CheckTcaplus check if a host has tcaplus records or not
	CheckTcaplus(kt *kit.Kit, ips []string) (*TcaplusResp, error)
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
		client: rest.NewClient(c, "/"),
	}

	return client, nil
}

// tcaplusApi tcaplus api interface implementation
type tcaplusApi struct {
	client rest.ClientInterface
}

// CheckTcaplus check if a host has tcaplus records or not, currency is limited to 500 qps
func (t *tcaplusApi) CheckTcaplus(kt *kit.Kit, ips []string) (*TcaplusResp, error) {

	if len(ips) > TcapulsCheckIPExistsMaxLength {
		return nil, fmt.Errorf("ip list length greater than limit, input length: %d, limit: %d",
			len(ips), TcapulsCheckIPExistsMaxLength)
	}

	req := &TcaplusReq{
		IpList: ips,
	}

	subPath := "/app/newoms.php/webservice/host/check-ip-exist"
	params := map[string]string{
		"cmd":     "10007",
		"ip-type": "sa",
	}
	resp := new(TcaplusResp)
	r := t.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(kt.Header()).
		WithParams(params).
		Do()
	if r.StatusCode >= 400 {
		return nil, fmt.Errorf("check tcaplus failed, http status: %s, body: %s", r.Status, r.Body)
	}

	if err := r.Into(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
