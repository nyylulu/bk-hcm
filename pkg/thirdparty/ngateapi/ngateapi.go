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

// Package ngateapi ngate api
package ngateapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"git.woa.com/nops/ngate/ngate-sdk/ngate-go/ngatehmac"

	"github.com/prometheus/client_golang/prometheus"
)

// NgateClientInterface ngate api interface
type NgateClientInterface interface {
	// RecycleIP recycle ip
	RecycleIP(ctx context.Context, req *RecycleIPReq) (*RecycleIPResponse, error)
}

// NewNgateClientInterface creates a ngate api instance
func NewNgateClientInterface(opts cc.NgateCli, reg prometheus.Registerer) (NgateClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "ngate api",
			servers: []string{opts.Host},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	return &ngateApi{
		client: rest.NewClient(c, "/"),
		opts:   opts,
	}, nil
}

// ngateApi ngate api interface implementation
type ngateApi struct {
	client rest.ClientInterface
	opts   cc.NgateCli
}

// RecycleIP recycle ip
func (n *ngateApi) RecycleIP(ctx context.Context, req *RecycleIPReq) (*RecycleIPResponse, error) {
	reqByte, err := json.Marshal(req)
	if err != nil {
		logs.Errorf("json marshal extension failed, err: %v", err)
		return nil, err
	}

	headerMap := ngatehmac.GetAuthHeader(n.opts.AppCode, n.opts.AppSecret, reqByte)
	header := http.Header{}
	for k, v := range headerMap {
		header.Set(k, v)
	}

	subPath := "/config/iprms/ip/recycleIp"
	resp := new(RecycleIPResponse)
	err = n.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("ngate api call recycle ip failed, err: %v, req: %+v, subPath: %s", err, cvt.PtrToVal(req), subPath)
		return nil, err
	}

	if resp.ReturnCode != 0 || !resp.Success {
		return nil, fmt.Errorf("ngate recycle api response code failed, req: %+v, errCode: %d, errMsg: %s, traceID: %s",
			cvt.PtrToVal(req), resp.ReturnCode, resp.Message, resp.TraceID)
	}

	return resp, nil
}
