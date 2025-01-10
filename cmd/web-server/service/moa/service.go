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
	"hcm/pkg/api/web-server/moa"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	pkgmoa "hcm/pkg/thirdparty/moa"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/util"

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
		client:     c.MoaCli,
		etcdCli:    etcdCli,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("MoaBizRequest", http.MethodPost, "/bizs/{bk_biz_id}/moa/request", svc.MoaBizRequest)
	h.Add("MoaBizVerify", http.MethodPost, "/bizs/{bk_biz_id}/moa/verify", svc.MoaBizVerify)

	h.Load(c.WebService)
}

type service struct {
	client     pkgmoa.Client
	etcdCli    *etcd3.Client
	authorizer auth.Authorizer
}

// MoaBizRequest moa biz request.
func (s service) MoaBizRequest(cts *rest.Contexts) (any, error) {
	return s.moaRequest(cts, handler.ListBizAuthRes, meta.Biz, meta.Access)
}

// MoaBizRequest ...
func (s service) moaRequest(cts *rest.Contexts, validHandler handler.ListAuthResHandler, resType meta.ResourceType,
	action meta.Action) (any, error) {

	req := new(moa.InitiateVerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate biz and authorize
	validReq := &handler.ListAuthResOption{Authorizer: s.authorizer, ResType: resType, Action: action}
	_, noPerm, err := validHandler(cts, validReq)
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for request moa")
	}

	opt := &pkgmoa.InitiateVerificationReq{
		Username:      req.Username,
		Channel:       req.Channel,
		Language:      req.Language,
		PromptPayload: req.PromptPayload,
	}
	resp, err := s.client.Request(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request moa api failed, err: %v, req: %v, rid: %s", err, converter.PtrToVal(req), cts.Kit.Rid)
		return nil, err
	}

	result := moa.InitiateVerificationResp{SessionId: resp.SessionId}
	return result, nil
}

// MoaBizVerify moa biz verify.
func (s service) MoaBizVerify(cts *rest.Contexts) (any, error) {
	return s.moaVerify(cts, handler.ListBizAuthRes)
}

// moaVerify ...
func (s service) moaVerify(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	req := new(moa.VerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验是否同一个用户
	if req.Username != cts.Kit.User {
		return nil, errf.Newf(errf.InvalidParameter, "username is not match")
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: s.authorizer, ResType: meta.Biz, Action: meta.Access})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for verify moa")
	}

	opt := &pkgmoa.VerificationReq{
		SessionId: req.SessionId,
		Username:  req.Username,
	}
	resp, err := s.client.Verify(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if resp.Status == enumor.VerificationStatusFinish && resp.ButtonType == enumor.VerificationResultConfirm {
		leaseResp, err := s.etcdCli.Grant(cts.Kit.Ctx, fiveMinutes)
		if err != nil {
			logs.Errorf("grant etcd lease failed, err: %v, leaseResp: %v, rid: %s", err, leaseResp, cts.Kit.Rid)
			return nil, err
		}
		moaValue := util.JoinStrings(cts.Kit.User, resp.ButtonType, "-")
		putResp, err := s.etcdCli.Put(cts.Kit.Ctx, resp.SessionId, moaValue, etcd3.WithLease(leaseResp.ID))
		if err != nil {
			logs.Errorf("put etcd lease failed, err: %v, putResp: %v, rid: %s", err, putResp, cts.Kit.Rid)
			return nil, err
		}
	}
	result := moa.VerificationResp{
		Status:     resp.Status,
		ButtonType: resp.ButtonType,
		SessionId:  resp.SessionId,
	}
	return result, nil
}

const (
	fiveMinutes = 5 * 60
)
