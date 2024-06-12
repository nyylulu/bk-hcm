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

// Package thirdparty ...
package thirdparty

import (
	"hcm/cmd/woa-server/thirdparty/bkchatapi"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/thirdparty/dvmapi"
	"hcm/cmd/woa-server/thirdparty/erpapi"
	"hcm/cmd/woa-server/thirdparty/gcsapi"
	"hcm/cmd/woa-server/thirdparty/itsmapi"
	"hcm/cmd/woa-server/thirdparty/l5api"
	"hcm/cmd/woa-server/thirdparty/safetyapi"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
	"hcm/cmd/woa-server/thirdparty/sopsapi"
	"hcm/cmd/woa-server/thirdparty/tcaplusapi"
	"hcm/cmd/woa-server/thirdparty/tgwapi"
	"hcm/cmd/woa-server/thirdparty/tjjapi"
	"hcm/cmd/woa-server/thirdparty/tmpapi"
	"hcm/cmd/woa-server/thirdparty/uworkapi"
	"hcm/cmd/woa-server/thirdparty/xshipapi"
	"hcm/pkg/cc"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
)

// Client third party client
type Client struct {
	CVM             cvmapi.CVMClientInterface
	DVM             dvmapi.DVMClientInterface
	SoJob           sojobapi.SojobClientInterface
	Tjj             tjjapi.TjjClientInterface
	Xship           xshipapi.XshipClientInterface
	Erp             erpapi.ErpClientInterface
	Tmp             tmpapi.TMPClientInterface
	Uwork           uworkapi.UworkClientInterface
	GCS             gcsapi.GcsClientInterface
	Tcaplus         tcaplusapi.TcaplusClientInterface
	TGW             tgwapi.TgwClientInterface
	L5              l5api.L5ClientInterface
	Safety          safetyapi.SafetyClientInterface
	TencentCloudOpt cc.TCloudCli
	BkChat          bkchatapi.BkChatClientInterface
	Sops            sopsapi.SopsClientInterface
	ITSM            itsmapi.ITSMClientInterface
}

// NewClient new third party client
func NewClient(opts cc.ClientConfig, reg prometheus.Registerer) (*Client, error) {
	cvm, err := cvmapi.NewCVMClientInterface(opts.CvmOpt, reg)
	if err != nil {
		logs.Errorf("failed to new cvm api client, err: %v", err)
		return nil, err
	}

	dvm, err := dvmapi.NewDVMClientInterface(opts.DvmOpt, reg)
	if err != nil {
		logs.Errorf("failed to new dvm api client, err: %v", err)
		return nil, err
	}

	sojob, err := sojobapi.NewSojobClientInterface(opts.SojobOpt, reg)
	if err != nil {
		logs.Errorf("failed to new sojob api client, err: %v", err)
		return nil, err
	}

	tjj, err := tjjapi.NewTjjClientInterface(opts.TjjOpt, reg)
	if err != nil {
		logs.Errorf("failed to new tjj api client, err: %v", err)
		return nil, err
	}

	xship, err := xshipapi.NewXshipClientInterface(opts.XshipOpt, reg)
	if err != nil {
		logs.Errorf("failed to new xship api client, err: %v", err)
		return nil, err
	}

	client := &Client{
		CVM:   cvm,
		DVM:   dvm,
		SoJob: sojob,
		Tjj:   tjj,
		Xship: xship,
	}

	// 实例化API网关client
	client, err = newApiGWClient(opts, reg, client)
	if err != nil {
		return nil, err
	}

	// 实例化第三方服务client
	client, err = newThirdClient(opts, reg, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// newApiGWClient 实例化API网关client
func newApiGWClient(opts cc.ClientConfig, reg prometheus.Registerer, client *Client) (*Client, error) {
	sops, err := sopsapi.NewSopsClientInterface(opts.Sops, reg)
	if err != nil {
		logs.Errorf("failed to new sops api client, err: %v", err)
		return nil, err
	}
	client.Sops = sops

	return client, nil
}

// newThirdClient 实例化第三方服务client
func newThirdClient(opts cc.ClientConfig, reg prometheus.Registerer, client *Client) (*Client, error) {
	erp, err := erpapi.NewErpClientInterface(opts.ErpOpt, reg)
	if err != nil {
		logs.Errorf("failed to new erp api client, err: %v", err)
		return nil, err
	}
	client.Erp = erp

	tmp, err := tmpapi.NewTMPClientInterface(opts.TmpOpt, reg)
	if err != nil {
		logs.Errorf("failed to new tmp api client, err: %v", err)
		return nil, err
	}
	client.Tmp = tmp

	uwork, err := uworkapi.NewUworkClientInterface(opts.Uwork, reg)
	if err != nil {
		logs.Errorf("failed to new uwork api client, err: %v", err)
		return nil, err
	}
	client.Uwork = uwork

	gcs, err := gcsapi.NewGcsClientInterface(opts.GCS, reg)
	if err != nil {
		logs.Errorf("failed to new gcs api client, err: %v", err)
		return nil, err
	}
	client.GCS = gcs

	tcaplus, err := tcaplusapi.NewTcaplusClientInterface(opts.Tcaplus, reg)
	if err != nil {
		logs.Errorf("failed to new tcaplus api client, err: %v", err)
		return nil, err
	}
	client.Tcaplus = tcaplus

	tgw, err := tgwapi.NewTgwClientInterface(opts.TGW, reg)
	if err != nil {
		logs.Errorf("failed to new tgw api client, err: %v", err)
		return nil, err
	}
	client.TGW = tgw

	l5, err := l5api.NewL5ClientInterface(opts.L5, reg)
	if err != nil {
		logs.Errorf("failed to new l5 api client, err: %v", err)
		return nil, err
	}
	client.L5 = l5

	safety, err := safetyapi.NewSafetyClientInterface(opts.Safety, reg)
	if err != nil {
		logs.Errorf("failed to new safety api client, err: %v", err)
		return nil, err
	}
	client.Safety = safety

	bkchat, err := bkchatapi.NewBkChatClientInterface(opts.BkChat, reg)
	if err != nil {
		logs.Errorf("failed to new bkchat api client, err: %v", err)
		return nil, err
	}
	client.BkChat = bkchat

	itsm, err := itsmapi.NewITSMClientInterface(opts.ITSM, reg)
	if err != nil {
		logs.Errorf("failed to new itsm api client, err: %v", err)
		return nil, err
	}
	client.ITSM = itsm

	return client, nil
}
