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

package cmdb

import (
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// NewClient initialize a new cmdbApiGateWay client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer, esbClient esb.Client) (cmdb.Client, error) {
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
			Name:    "cmdbApiGateWay",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/v3")

	agw := &cmdbApiGateWay{
		config: cfg,
		client: restCli,
	}
	if esbClient != nil {
		agw.Client = esbClient.Cmdb()
	}
	return agw, nil
}

// cmdbApiGateWay is an esb client to request cmdbApiGateWay.
type cmdbApiGateWay struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
	// fall back to esbCall
	cmdb.Client
}

// ListBizHost ...
func (c *cmdbApiGateWay) ListBizHost(kt *kit.Kit, req *cmdb.ListBizHostParams) (
	*cmdb.ListBizHostResult, error) {

	return apigateway.ApiGatewayCall[cmdb.ListBizHostParams, cmdb.ListBizHostResult](c.client, c.config, rest.POST,
		kt, req, "/hosts/app/%d/list_hosts", req.BizID)
}

// UpdateCvmOS ...
func (c *cmdbApiGateWay) UpdateCvmOSAndSvrStatus(kt *kit.Kit, req *cmdb.UpdateCvmOSReq) error {

	err := req.Validate()
	if err != nil {
		return err
	}

	_, err = apigateway.ApiGatewayCall[cmdb.UpdateCvmOSReq, interface{}](c.client, c.config, rest.PUT,
		kt, req, "/shipper/update/reinstall/cmdb/cvm")
	if err != nil {
		logs.Errorf("call cmdb api gateway to update cvm os failed, err: %v", err)
		return err
	}
	return nil
}

// 其他请求使用esb 接口
