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

package xshipapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// XshipClientInterface Xship api interface
type XshipClientInterface interface {
	// CreateReinstallTask create host reinstall task
	CreateReinstallTask(ctx context.Context, header http.Header, req *ReinstallReq) (*ReinstallResp, error)
	// GetReinstallTaskStatus get host reinstall task status
	GetReinstallTaskStatus(ctx context.Context, header http.Header, orderID string) (*ReinstallStatusResp, error)
}

// NewXshipClientInterface creates a Xship api instance
func NewXshipClientInterface(opts cc.XshipCli, reg prometheus.Registerer) (XshipClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "xship api",
			servers: []string{opts.XshipApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	xship := &xshipApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}

	return xship, nil
}

// xshipApi Xship api interface implementation
type xshipApi struct {
	client rest.ClientInterface
	opts   *cc.XshipCli
}

func (x *xshipApi) getAuthHeader() (string, string) {
	timeObj := time.Now()
	timestamp := strconv.FormatInt(timeObj.Unix(), 10)
	s := timestamp + x.opts.SecretKey

	h := sha256.New()
	h.Write([]byte(s))
	signature := hex.EncodeToString(h.Sum(nil))
	auth := "timestamp=" + timestamp + ",signature=" + signature

	return "Authorization", auth
}

// CreateReinstallTask create host reinstall task
func (x *xshipApi) CreateReinstallTask(ctx context.Context, header http.Header, req *ReinstallReq) (*ReinstallResp,
	error) {

	subPath := "xship-service-apiaccept/reinstallProcessAccept"

	key, val := x.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	header.Set("Content-Type", "application/json")
	header.Set("x-client-id", x.opts.ClientID)

	resp := new(ReinstallResp)
	err := x.client.Post().
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

// GetReinstallTaskStatus get host reinstall task status
func (x *xshipApi) GetReinstallTaskStatus(ctx context.Context, header http.Header, orderID string) (
	*ReinstallStatusResp, error) {

	subPath := "xship-service-process-reception/qureyReinstallProcessInfoWithoutGroup"

	key, val := x.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	header.Set("Content-Type", "application/json")
	header.Set("x-client-id", x.opts.ClientID)

	req := &GetReinstallStatusReq{
		OrderIDs: []string{orderID},
		Page:     1,
		PageSize: 1,
	}

	resp := new(ReinstallStatusResp)
	err := x.client.Post().
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
