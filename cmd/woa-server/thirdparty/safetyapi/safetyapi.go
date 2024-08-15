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

// Package safetyapi provides the safety api client
package safetyapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// SafetyClientInterface GCS api interface
type SafetyClientInterface interface {
	// CheckLog4jHost check if a host has log4j records or not
	CheckLog4jHost(ctx context.Context, header http.Header, ip string) (bool, error)
	// CheckLog4jContainer check if a container has log4j records or not
	CheckLog4jContainer(ctx context.Context, header http.Header, ip, parentIp string) (bool, error)
}

// NewSafetyClientInterface creates a safety api instance
func NewSafetyClientInterface(opts cc.SafetyCli, reg prometheus.Registerer) (SafetyClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "safety api",
			servers: []string{opts.SafetyApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &safetyApi{
		client: rest.NewClient(c, "/"),
	}

	return client, nil
}

// safetyApi safety api interface implementation
type safetyApi struct {
	client rest.ClientInterface
}

// CheckLog4jHost check if host has log4j records or not
func (s *safetyApi) CheckLog4jHost(ctx context.Context, header http.Header, ip string) (bool, error) {
	req := &BaseLineReq{
		TaskId:        156,
		Ip:            ip,
		BusinessGroup: "IEG",
		Department:    "all",
		Page:          1,
		PageSize:      10,
	}

	if header == nil {
		header = http.Header{}
	}
	headers := s.genHeaders()
	for k, v := range headers {
		header.Set(k, v)
	}

	subPath := "/api/baseline/get_task_data_new"
	resp := new(BaseLineRsp)
	err := s.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return false, err
	}

	if resp.Ret != 0 {
		logs.Errorf("%s check log4j failed, ret code: %d, msg: %s", ip, resp.Ret, resp.Msg)
		return false, fmt.Errorf("check log4j failed, ret code: %d, msg: %s", resp.Ret, resp.Msg)
	}

	pass := true
	if resp.Data.TotalCount > 0 {
		pass = false
		logs.Infof("%s has log4j records, resp: %+v", ip, resp)
	}

	return pass, nil
}

// CheckLog4jContainer check if container has log4j records or not
func (s *safetyApi) CheckLog4jContainer(ctx context.Context, header http.Header, ip, parentIp string) (bool, error) {
	req := &BaseLineReq{
		TaskId:        156,
		Ip:            parentIp,
		BusinessGroup: "IEG",
		Department:    "all",
		Page:          1,
		PageSize:      10000,
	}

	if header == nil {
		header = http.Header{}
	}
	headers := s.genHeaders()
	for k, v := range headers {
		header.Set(k, v)
	}

	subPath := "/api/baseline/get_task_data_new"
	resp := new(BaseLineRsp)
	err := s.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return false, err
	}

	if resp.Ret != 0 {
		logs.Errorf("%s check log4j failed, ret code: %d, msg: %s", ip, resp.Ret, resp.Msg)
		return false, fmt.Errorf("check log4j failed, ret code: %d, msg: %s", resp.Ret, resp.Msg)
	}

	pass := true
	for _, item := range resp.Data.Data {
		if item.ContainerIp == ip {
			pass = false
			logs.Infof("%s has log4j records, resp: %+v", ip, resp)
			break
		}
	}

	return pass, nil
}

func (s *safetyApi) genHeaders() map[string]string {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	min := 1000
	max := 9999
	random := strconv.Itoa(rand.Intn(max-min+1) + min)
	sign := now + appKey + apiName + random
	signHash := s.computeHmacSha256(sign, appSecret)

	headers := make(map[string]string)
	headers["XTimestamp"] = now
	headers["XRandom"] = random
	headers["XApiName"] = "baseline_get_task_data_new"
	headers["XSignature"] = signHash
	headers["XAppKey"] = appKey

	return headers
}

func (s *safetyApi) computeHmacSha256(msg, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(msg))
	//sha := hex.EncodeToString(h.Sum(nil))
	sha := h.Sum(nil)

	return base64.StdEncoding.EncodeToString([]byte(sha))
}
