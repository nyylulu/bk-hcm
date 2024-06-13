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

// Package gcsapi implements GCS api
package gcsapi

import (
	"context"
	"net/http"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/golang-jwt/jwt"
	"github.com/prometheus/client_golang/prometheus"
)

// GcsClientInterface GCS api interface
type GcsClientInterface interface {
	// CheckGCS check if a host has GCS records or not
	CheckGCS(ctx context.Context, header http.Header, ip string) (*GCSResp, error)
}

// NewGcsClientInterface creates a GCS api instance
func NewGcsClientInterface(opts cc.GCSCli, reg prometheus.Registerer) (GcsClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "gcs api",
			servers: []string{opts.GcsApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &gcsApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}

	return client, nil
}

// gcsApi GCS api interface implementation
type gcsApi struct {
	client rest.ClientInterface
	opts   *cc.GCSCli
}

func (g *gcsApi) sign(rtx string) (tokenString string, err error) {
	// The token content.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  g.opts.SecretID,
		"user": rtx,
		"iat":  time.Now().Add(-1 * time.Hour).Unix(),
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(g.opts.SecretKey))
	return
}

// CheckGCS check if a host has GCS records or not
func (g *gcsApi) CheckGCS(ctx context.Context, header http.Header, ip string) (*GCSResp, error) {
	req := &GCSReq{
		Ip:      []string{ip},
		Columns: []string{"ip"},
	}

	subPath := "/gcscmdb/scene/gcsips"

	token, signErr := g.sign(g.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := new(GCSResp)
	err := g.client.Get().
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
