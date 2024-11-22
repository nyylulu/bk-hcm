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

package moa

import (
	"net/http"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	pkgmoa "hcm/pkg/thirdparty/moa"
	"hcm/pkg/tools/converter"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {
	etcdCfg, err := cc.WebServer().Service.Etcd.ToConfig()
	if err != nil {
		logs.Errorf("convert etcd config failed, err: %v", err)
		return
	}
	etcdCli, err := etcd3.New(etcdCfg)
	if err != nil {
		logs.Errorf("create etcd client failed, err: %v", err)
		return
	}
	svc := &service{
		client:  c.MoaCli,
		etcdCli: etcdCli,
	}

	h := rest.NewHandler()

	h.Add("MOARequest", http.MethodPost, "/moa/request", svc.Request)
	h.Add("MOAVerify", http.MethodPost, "/moa/verify", svc.Verify)

	h.Load(c.WebService)
}

type service struct {
	client  pkgmoa.Client
	etcdCli *etcd3.Client
}

// Request ...
func (s service) Request(cts *rest.Contexts) (interface{}, error) {
	req := new(pkgmoa.InitiateVerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	resp, err := s.client.Request(cts.Kit, req)
	if err != nil {
		logs.Errorf("request moa api failed, err: %v, req: %v, rid: %s", err, converter.PtrToVal(req), cts.Kit.Rid)
		return nil, err
	}
	return resp, nil
}

// Verify ...
func (s service) Verify(cts *rest.Contexts) (interface{}, error) {
	req := new(pkgmoa.VerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	resp, err := s.client.Verify(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	if resp.Status == enumor.VerificationStatusFinish && resp.ButtonType == enumor.VerificationResultConfirm {
		leaseResp, err := s.etcdCli.Grant(cts.Kit.Ctx, fiveMinutes)
		if err != nil {
			logs.Errorf("grant etcd lease failed, err: %v, leaseResp: %v, rid: %s", err, leaseResp, cts.Kit.Rid)
			return nil, err
		}
		putResp, err := s.etcdCli.Put(cts.Kit.Ctx, resp.SessionId, resp.ButtonType, etcd3.WithLease(leaseResp.ID))
		if err != nil {
			logs.Errorf("put etcd lease failed, err: %v, putResp: %v, rid: %s", err, putResp, cts.Kit.Rid)
			return nil, err
		}
	}
	return resp, nil
}

const (
	fiveMinutes = 5 * 60
)
