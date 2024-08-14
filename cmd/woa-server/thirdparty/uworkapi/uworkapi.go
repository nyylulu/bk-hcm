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

// Package uworkapi uwork api
package uworkapi

import (
	"context"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// UworkClientInterface Uwork api interface
type UworkClientInterface interface {
	// CheckUworkTicket check if a host has Uwork tickets or not
	CheckUworkTicket(ctx context.Context, header http.Header, user, ip string) (*UworkTicketResponse, error)
	// CheckUworkProcess check if a host has Uwork process or not
	CheckUworkProcess(ctx context.Context, header http.Header, assetId string) (*UworkProcessResponse, error)
}

// NewUworkClientInterface creates a Uwork api instance
func NewUworkClientInterface(opts cc.UworkCli, reg prometheus.Registerer) (UworkClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "uwork api",
			servers: []string{opts.UworkApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &uworkApi{
		client: rest.NewClient(c, "/"),
	}

	return client, nil
}

// uworkApi Uwork api interface implementation
type uworkApi struct {
	client rest.ClientInterface
}

// CheckUwork check if a host has Uwork tickets or not
func (u *uworkApi) CheckUworkTicket(ctx context.Context, header http.Header, user, ip string) (*UworkTicketResponse,
	error) {

	systemId := "16"
	req := &QueryServerEventReq{
		Action:   "QueryServerEvent",
		FlowId:   "4",
		Starter:  user,
		SystemId: systemId,
		Data: &QueryServerEventParams{
			ResultColumns: &ResultColumns{
				TicketNo:    "",
				ProcessDesc: "",
				IsEnd:       "",
			},
			SearchCondition: &SearchCondition{
				ServerIP: ip,
			},
		},
	}
	subPath := "open_api/logic"

	resp := new(UworkTicketResponse)
	err := u.client.Post().
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

// CheckUwork check if a host has Uwork process or not
func (u *uworkApi) CheckUworkProcess(ctx context.Context, header http.Header, assetId string) (*UworkProcessResponse,
	error) {

	systemId := "16"
	req := &QueryServerProcessReq{
		Action:   "QueryData",
		Method:   "serverInUwork",
		SystemId: systemId,
		Data: &QueryServerProcessParams{
			AssetID: assetId,
		},
	}
	subPath := "open_api/logic"

	resp := new(UworkProcessResponse)
	err := u.client.Post().
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
