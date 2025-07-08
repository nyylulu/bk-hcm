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

// Package xrayapi xray api
package xrayapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// XRayCheckFaultTicketMaxLength xray api check fault ticket max length
const XRayCheckFaultTicketMaxLength = 1000

// XrayClientInterface xray api interface
type XrayClientInterface interface {
	// CheckXrayFaultTickets check if a host xray fault tickets
	CheckXrayFaultTickets(kt *kit.Kit, assetIDs []string, isEnd enumor.XrayFaultTicketIsEnd) (
		*QueryFaultTicketResponse, error)
}

// NewXrayClientInterface create a xray api instance
func NewXrayClientInterface(opts cc.XrayCli, reg prometheus.Registerer) (XrayClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "xray api",
			servers: []string{opts.XrayApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	return &xrayApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}, nil
}

// xrayApi xray api interface implementation
type xrayApi struct {
	client rest.ClientInterface
	opts   *cc.XrayCli
}

func (x *xrayApi) getAuthHeader() (string, string) {
	timeObj := time.Now()
	timestamp := strconv.FormatInt(timeObj.Unix(), 10)
	s := timestamp + x.opts.SecretKey

	h := sha256.New()
	h.Write([]byte(s))
	signature := hex.EncodeToString(h.Sum(nil))
	auth := "timestamp=" + timestamp + ",signature=" + signature

	return "Authorization", auth
}

// CheckXrayFaultTickets check if a host xray fault tickets, 入参不超过1000，建议QPS 50
// @doc https://iwiki.woa.com/p/1479549975
func (x *xrayApi) CheckXrayFaultTickets(kt *kit.Kit, assetIDs []string, isEnd enumor.XrayFaultTicketIsEnd) (
	*QueryFaultTicketResponse, error) {

	if len(assetIDs) > XRayCheckFaultTicketMaxLength {
		return nil, fmt.Errorf("assetIDs length greater than limit, input length: %d, limit: %d",
			len(assetIDs), XRayCheckFaultTicketMaxLength)
	}
	req := &QueryFaultTicketReq{
		ServerAssetIdList: assetIDs,
		IsEnd:             isEnd,
	}
	subPath := "/xray-srv-process-reception/queryMergedFaultInfo"

	key, val := x.getAuthHeader()
	header := kt.Header()
	header.Set(key, val)
	header.Set("Content-Type", "application/json")
	header.Set("x-client-id", x.opts.ClientID)

	resp := new(QueryFaultTicketResponse)
	err := x.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("check:xray:fault:ticket:failed, err: %+v, subPath: %v, req: %+v, rid: %s",
			err, subPath, req, kt.Rid)
		return nil, err
	}

	if resp.Code != "0" {
		logs.Errorf("check:xray:fault:ticket:return code is not \"0\", message: %s, code: %s, xrayTraceID: %s, rid: %s",
			resp.Message, resp.Code, resp.TraceID, kt.Rid)
		return nil, fmt.Errorf("check:xray:fault:ticket return code is not \"0\": %s(%s), "+
			"xrayTraceID: %s", resp.Message, resp.Code, resp.TraceID)
	}

	return resp, nil
}
