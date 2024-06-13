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

// Package tmpapi is a client for Third Party API
package tmpapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// TMPClientInterface TMP api interface
type TMPClientInterface interface {
	// CheckTMP returns if a host pass TMP alarm policy check or not
	CheckTMP(ctx context.Context, header http.Header, ip string) ([]interface{}, error)
	// AddShieldConfig add shield TMP alarm config
	AddShieldConfig(ctx context.Context, header http.Header, req *AddShieldReq) (*AddShieldResp, error)
}

// NewTMPClientInterface creates a TMP api instance
func NewTMPClientInterface(opts cc.TmpCli, reg prometheus.Registerer) (TMPClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "tmp api",
			servers: []string{opts.TMPApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	tmp := &tmpApi{
		client: rest.NewClient(c, "/"),
	}

	return tmp, nil
}

// tmpApi TMP api interface implementation
type tmpApi struct {
	client rest.ClientInterface
}

// CheckTMP returns if a host pass TMP alarm policy check or not
func (t *tmpApi) CheckTMP(ctx context.Context, header http.Header, ip string) ([]interface{}, error) {
	return t.getShieldConfig(ctx, header, ip)
}

func (t *tmpApi) getShieldConfig(ctx context.Context, header http.Header, ip string) ([]interface{}, error) {
	req := &CheckTMPReq{
		Method: "alarm.get_alarm_shield_config",
		Params: &CheckTMPParams{
			Ip: ip,
		},
	}

	subPath := "/tnm2_api/alarm.get_alarm_shield_config"
	ret := make([]interface{}, 0)
	err := t.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(&ret)

	if err != nil {
		return nil, err
	}

	if len(ret) != 3 {
		return nil, fmt.Errorf("check TMP alarm policy got invalid response format, resp: %+v", ret)
	}

	return ret[2].([]interface{}), nil
}

// AddShieldConfig add shield TMP alarm config
func (t *tmpApi) AddShieldConfig(ctx context.Context, header http.Header, req *AddShieldReq) (*AddShieldResp, error) {
	subPath := "/tnm2_api/alarmadapter.add_alarm_shield_config"
	resp := new(AddShieldResp)
	err := t.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(&resp)

	return resp, err
}
