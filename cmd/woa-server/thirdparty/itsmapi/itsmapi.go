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

package itsmapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// ITSMClientInterface itsm api interface
type ITSMClientInterface interface {
	// CreateTicket create itsm ticket
	CreateTicket(ctx context.Context, header http.Header, user string, orderId uint64) (*CreateTicketResp, error)
	// OperateNode operate itsm ticket node
	OperateNode(ctx context.Context, header http.Header, req *OperateNodeReq) (*OperateNodeResp, error)
	// GetTicketStatus get itsm ticket status
	GetTicketStatus(ctx context.Context, header http.Header, id string) (*GetTicketStatusResp, error)
	// GetTicketLog get itsm ticket logs
	GetTicketLog(ctx context.Context, header http.Header, id string) (*GetTicketLogResp, error)
}

// NewIAMClientInterface creates iam api instance
func NewIAMClientInterface(opts cc.ApiGateway, reg prometheus.Registerer) (ITSMClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "itsm api",
			servers: opts.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	itsmClient := &itsmCli{
		client: rest.NewClient(c, ""),
	}

	return itsmClient, nil
}

// itsmApi itsm api interface implementation
type itsmCli struct {
	client rest.ClientInterface
	opts   *cc.ApiGateway
}

func (c *itsmCli) getAuthHeader() (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.opts.AppCode,
		c.opts.AppSecret, c.opts.User)

	return key, val
}

// CreateTicket create itsm ticket
func (c *itsmCli) CreateTicket(ctx context.Context, header http.Header, user string, orderId uint64) (
	*CreateTicketResp, error) {

	subPath := "/v2/itsm/create_ticket"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	req := &CreateTicketReq{
		ServiceId: int(c.opts.ServiceID),
		Creator:   user,
		Fields: []*TicketField{
			{
				Key:   TicketKeyTitle,
				Value: fmt.Sprintf(TicketValTitleFormat, orderId),
			},
			{
				Key:   TicketKeyApplyId,
				Value: strconv.Itoa(int(orderId)),
			},
			{
				Key:   TicketKeyApplyLink,
				Value: fmt.Sprintf(c.opts.ApplyLinkFormat, orderId),
			},
			{
				Key:   TicketKeyNeedSysAudit,
				Value: TicketValNeedSysAuditNo,
			},
		},
	}
	resp := new(CreateTicketResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// OperateNode operate itsm ticket node
func (c *itsmCli) OperateNode(ctx context.Context, header http.Header, req *OperateNodeReq) (*OperateNodeResp, error) {
	subPath := "/v2/itsm/operate_node"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(OperateNodeResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetTicketStatus get itsm ticket status
func (c *itsmCli) GetTicketStatus(ctx context.Context, header http.Header, id string) (*GetTicketStatusResp, error) {
	subPath := "/v2/itsm/get_ticket_status"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTicketStatusResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithParam("sn", id).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetTicketLog get itsm ticket logs
func (c *itsmCli) GetTicketLog(ctx context.Context, header http.Header, id string) (*GetTicketLogResp, error) {
	subPath := "/v2/itsm/get_ticket_logs"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTicketLogResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithParam("sn", id).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}
