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
 */

// Package cert ...
package cert

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/logics/ziyan"
	cloudserver "hcm/pkg/api/cloud-server"
	cscert "hcm/pkg/api/cloud-server/cert"
	hccert "hcm/pkg/api/hc-service/cert"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// CreateCert create resource cert.
func (svc *certSvc) CreateCert(cts *rest.Contexts) (interface{}, error) {
	return svc.createCert(cts, handler.ResOperateAuth, false)
}

// CreateBizCert create biz cert.
func (svc *certSvc) CreateBizCert(cts *rest.Contexts) (interface{}, error) {
	return svc.createCert(cts, handler.BizOperateAuth, true)
}

func (svc *certSvc) createCert(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler, bizRequired bool) (
	interface{}, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create cert request decode failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	var bkBizID int64 = constant.UnassignedBiz
	var err error
	if bizRequired {
		bkBizID, err = cts.PathParameter("bk_biz_id").Int64()
		if err != nil {
			return nil, err
		}
	}

	// create authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cert,
		Action: meta.Create, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("create cert auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, accID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.createTCloudCert(cts.Kit, req.Data, bkBizID)
	case enumor.TCloudZiyan:
		return svc.createTCloudZiyanCert(cts.Kit, req.Data, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *certSvc) createTCloudCert(kt *kit.Kit, body json.RawMessage, bkBizID int64) (interface{}, error) {
	req := new(hccert.TCloudCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	publicKey, err := base64.URLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		logs.Errorf("create tcloud cert decode publickey failed, pk: %s, err: %v, rid: %s", req.PublicKey, err, kt.Rid)
		return nil, err
	}

	privateKey, err := base64.URLEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		logs.Errorf("create tcloud cert decode privatekey failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	req.PublicKey = string(publicKey)
	req.PrivateKey = string(privateKey)
	req.BkBizID = bkBizID

	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cert.CreateCert(kt, req)
	if err != nil {
		logs.Errorf("create tcloud cert failed, req: %+v, result: %+v, err: %v, rid: %s", req, result, err, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *certSvc) createTCloudZiyanCert(kt *kit.Kit, body json.RawMessage, bkBizID int64) (interface{}, error) {
	req := new(cscert.ZiyanCreateCertReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tags, err := ziyan.GenTagsForBizsWithManager(kt, svc.cmdbCli, bkBizID, req.Manager, req.BakManager)
	if err != nil {
		logs.Errorf("create tcloud ziyan cert gen tags failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	publicKey, err := base64.URLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		logs.Errorf("create tcloud cert decode publickey failed, pk: %s, err: %v, rid: %s", req.PublicKey, err, kt.Rid)
		return nil, err
	}

	privateKey, err := base64.URLEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		logs.Errorf("create tcloud ziyan cert decode privatekey failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	createReq := &hccert.TCloudCreateReq{
		BkBizID:    bkBizID,
		AccountID:  req.AccountId,
		Vendor:     req.Vendor,
		Name:       req.Name,
		Memo:       req.Memo,
		CertType:   enumor.CertType(req.CertType),
		PublicKey:  string(publicKey),
		PrivateKey: string(privateKey),
		Tags:       tags,
	}
	if err = createReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloudZiyan.Cert.CreateCert(kt, createReq)
	if err != nil {
		logs.Errorf("create tcloud cert failed, req: %+v, result: %+v, err: %v, rid: %s",
			createReq, result, err, kt.Rid)
		return result, err
	}

	return result, nil
}
