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

// Package bkchatapi ...
package bkchatapi

import (
	"context"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// BkChatClientInterface bkchat api interface
type BkChatClientInterface interface {
	// SendApplyDoneMsg send apply done bkchat message
	SendApplyDoneMsg(ctx context.Context, header http.Header, user, content string) (*SendMsgResp, error)
	// GetNoticeFmt get apply done notice format
	GetNoticeFmt() string
}

// NewBkChatClientInterface creates an itsm api instance
func NewBkChatClientInterface(opts cc.BkChatCli, reg prometheus.Registerer) (BkChatClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "bkchat api",
			servers: []string{opts.BkChatApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &bkchatApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}

	return client, nil
}

// bkchatApi bkchat api interface implementation
type bkchatApi struct {
	client rest.ClientInterface
	opts   *cc.BkChatCli
}

// SendApplyDoneMsg create itsm ticket
func (c *bkchatApi) SendApplyDoneMsg(ctx context.Context, header http.Header, user, content string) (*SendMsgResp,
	error) {

	subPath := "/sendById"
	req := &SendMsgReq{
		Id:      user,
		Content: content,
	}
	resp := new(SendMsgResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetNoticeFmt get apply done notice format
func (c *bkchatApi) GetNoticeFmt() string {
	return c.opts.NoticeFmt
}
