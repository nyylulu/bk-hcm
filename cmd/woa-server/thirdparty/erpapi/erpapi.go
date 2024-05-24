/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package erpapi provides ERP system APIs
package erpapi

import (
	"context"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// ErpClientInterface erp api interface
type ErpClientInterface interface {
	// CreateDeviceReturnOrder creates device return order
	CreateDeviceReturnOrder(ctx context.Context, header http.Header, req *ErpReq) (*ErpResp, error)
	// QueryDeviceReturnOrders query device return orders
	QueryDeviceReturnOrders(ctx context.Context, header http.Header, req *ErpReq) (*ErpResp, error)
}

// NewErpClientInterface creates an erp api instance
func NewErpClientInterface(opts cc.ErpCli, reg prometheus.Registerer) (ErpClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "erp api",
			servers: []string{opts.ErpApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	erp := &erpApi{
		client: rest.NewClient(c, ""),
	}

	return erp, nil
}

// erpApi erp api interface implementation
type erpApi struct {
	client rest.ClientInterface
}

// CreateDeviceReturnOrder creates device return order
func (e *erpApi) CreateDeviceReturnOrder(ctx context.Context, header http.Header, req *ErpReq) (*ErpResp, error) {

	subPath := "/open/cloud/quota_api"
	resp := new(ErpResp)
	err := e.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryDeviceReturnOrders query device return orders
func (e *erpApi) QueryDeviceReturnOrders(ctx context.Context, header http.Header, req *ErpReq) (*ErpResp, error) {

	subPath := "/open/cloud/quota_api"
	resp := new(ErpResp)
	err := e.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}
