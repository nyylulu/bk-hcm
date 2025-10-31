/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package bkbotapproval bk审批助手
package bkbotapproval

import (
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client is an api-gateway client to request bkbotapproval.
type Client interface {
	SendMessage(kt *kit.Kit, params *SendMessageTplReq) error
}

// NewClient initialize a new client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &apigateway.Discovery{
			Name:    "bkBotApprovalApiGateWay",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/")

	agw := &bkBotApprovalApiGateWay{
		config: cfg,
		client: restCli,
	}
	return agw, nil
}

var _ Client = (*bkBotApprovalApiGateWay)(nil)

// bkBotApprovalApiGateWay is an apigw client to request bkBotApprovalApiGateWay.
type bkBotApprovalApiGateWay struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
}

// SendMessage 发送[蓝鲸审批助手]消息
func (c *bkBotApprovalApiGateWay) SendMessage(kt *kit.Kit, params *SendMessageTplReq) error {
	err := params.Validate()
	if err != nil {
		return err
	}

	_, err = apigateway.ApiGatewayCall[SendMessageTplReq, interface{}](c.client, c.config,
		rest.POST, kt, params, "/bkhcm_send_ticket/")
	if err != nil {
		return err
	}
	return nil
}
