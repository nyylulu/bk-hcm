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

package sopsapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// SopsClientInterface sops api interface
type SopsClientInterface interface {
	// CreateTask create sops task
	CreateTask(ctx context.Context, header http.Header, templateId, bkBizId string, req *CreateTaskReq) (
		*CreateTaskResp, error)
	// GetTaskStatus get sops task status
	GetTaskStatus(ctx context.Context, header http.Header, taskId, bkBizId string) (*GetTaskStatusResp, error)
}

// NewSopsClientInterface creates a sops api instance
func NewSopsClientInterface(opts cc.SopsCli, reg prometheus.Registerer) (SopsClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "sops api",
			servers: []string{opts.SopsApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	sopsCli := &sopsApi{
		client: rest.NewClient(c, ""),
		opts:   opts,
	}

	return sopsCli, nil
}

// sopsApi sops api interface implementation
type sopsApi struct {
	client rest.ClientInterface
	opts   cc.SopsCli
}

func (c *sopsApi) getAuthHeader() (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.opts.AppCode,
		c.opts.AppSecret, c.opts.Operator)

	return key, val
}

// CreateTask create sops task
func (c *sopsApi) CreateTask(ctx context.Context, header http.Header, templateId, bkBizId string, req *CreateTaskReq) (
	*CreateTaskResp, error) {

	subPath := "/create_task/%s/%s"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(CreateTaskResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath, templateId, bkBizId).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetTaskStatus get sops task status
func (c *sopsApi) GetTaskStatus(ctx context.Context, header http.Header, taskId, bkBizId string) (*GetTaskStatusResp,
	error) {

	subPath := "/get_task_status/%s/%s"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskStatusResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath, taskId, bkBizId).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}
