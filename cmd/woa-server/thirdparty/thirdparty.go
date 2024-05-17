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

package thirdparty

import (
	"github.com/prometheus/client_golang/prometheus"
	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
	"hcm/cmd/woa-server/thirdparty/tjjapi"
	"hcm/cmd/woa-server/thirdparty/xshipapi"
	"hcm/pkg/cc"
)

// Client third party client
type Client struct {
	CVM   cvmapi.CVMClientInterface
	SoJob sojobapi.SojobClientInterface
	Tjj   tjjapi.TjjClientInterface
	Xship xshipapi.XshipClientInterface
}

// NewClient new third party client
func NewClient(opts cc.ClientConfig, reg prometheus.Registerer) (*Client, error) {
	cvm, err := cvmapi.NewCVMClientInterface(opts.CvmOpt, reg)
	if err != nil {
		blog.Errorf("failed to new cvm api client, err: %v", err)
		return nil, err
	}

	sojob, err := sojobapi.NewSojobClientInterface(opts.SojobOpt, reg)
	if err != nil {
		blog.Errorf("failed to new sojob api client, err: %v", err)
		return nil, err
	}

	tjj, err := tjjapi.NewTjjClientInterface(opts.TjjOpt, reg)
	if err != nil {
		blog.Errorf("failed to new tjj api client, err: %v", err)
		return nil, err
	}

	xship, err := xshipapi.NewXshipClientInterface(opts.XshipOpt, reg)
	if err != nil {
		blog.Errorf("failed to new xship api client, err: %v", err)
		return nil, err
	}

	return &Client{
		CVM:   cvm,
		SoJob: sojob,
		Tjj:   tjj,
		Xship: xship,
	}, nil
}
