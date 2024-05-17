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

package sojobapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/golang-jwt/jwt"
	"github.com/prometheus/client_golang/prometheus"
)

// SojobClientInterface sojob api interface
type SojobClientInterface interface {
	// CreateJob creates so job
	CreateJob(ctx context.Context, header http.Header, req *CreateJobReq) (*CreateJobResp, error)
	// GetJobStatus gets so job status simple info
	GetJobStatus(ctx context.Context, header http.Header, jobId int) (*GetJobStatusResp, error)
	// GetJobStatusDetail gets so job status detail info
	GetJobStatusDetail(ctx context.Context, header http.Header, jobId int) (*GetJobStatusDetailResp, error)
}

// NewSojobClientInterface creates a sojob api instance
func NewSojobClientInterface(opts cc.SojobCli, reg prometheus.Registerer) (SojobClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "sojob api",
			servers: []string{opts.SojobApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	sojob := &sojobApi{
		client: rest.NewClient(c, ""),
		opts:   &opts,
	}

	return sojob, nil
}

// sojobApi sojob api interface implementation
type sojobApi struct {
	client rest.ClientInterface
	opts   *cc.SojobCli
}

func (s *sojobApi) sign(rtx string) (tokenString string, err error) {
	// The token content.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  s.opts.SecretID,
		"user": rtx,
		"iat":  time.Now().Add(-1 * time.Hour).Unix(),
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(s.opts.SecretKey))
	return
}

// CreateJob creates so job
func (s *sojobApi) CreateJob(ctx context.Context, header http.Header, req *CreateJobReq) (*CreateJobResp, error) {
	subPath := "/sojob/v1/job/create"

	token, signErr := s.sign(s.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := new(CreateJobResp)
	err := s.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetJobStatus gets so job status simple info
func (s *sojobApi) GetJobStatus(ctx context.Context, header http.Header, jobId int) (*GetJobStatusResp, error) {
	subPath := "/sojob/v1/job/simple_status"

	token, signErr := s.sign(s.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := new(GetJobStatusResp)
	err := s.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithParam("job_id", strconv.Itoa(jobId)).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetJobStatusDetail gets so job status detail info
func (s *sojobApi) GetJobStatusDetail(ctx context.Context, header http.Header, jobId int) (*GetJobStatusDetailResp,
	error) {

	subPath := "/sojob/v1/job/status"

	token, signErr := s.sign(s.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := new(GetJobStatusDetailResp)
	err := s.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithParam("jobid", strconv.Itoa(jobId)).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}
