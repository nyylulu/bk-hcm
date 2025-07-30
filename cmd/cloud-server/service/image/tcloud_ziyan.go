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

// Package image ...
package image

import (
	"hcm/cmd/cloud-server/service/common"
	"hcm/pkg/api/hc-service/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// TCloudZiyanQueryImage ...
func (svc *imageSvc) TCloudZiyanQueryImage(cts *rest.Contexts) (interface{}, error) {
	req, err := svc.decodeAndValidateTCloudImageListOpt(cts)
	if err != nil {
		logs.Errorf("decode and validate tcloud ziyan image list option failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.tcloudZiyanQueryImage(cts, constant.UnassignedBiz, req, handler.ResOperateAuth)
}

// TCLoudZiyanBizQueryImage ...
func (svc *imageSvc) TCLoudZiyanBizQueryImage(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req, err := svc.decodeAndValidateTCloudImageListOpt(cts)
	if err != nil {
		logs.Errorf("decode and validate tcloud ziyan image list option failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.tcloudZiyanQueryImage(cts, bizID, req, handler.BizOperateAuth)
}

func (svc *imageSvc) tcloudZiyanQueryImage(cts *rest.Contexts, bizID int64, req *image.TCloudImageListOption,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	// validate biz and authorize
	err := validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Image,
		Action: meta.Find, BasicInfo: common.GetCloudResourceBasicInfo(req.AccountID, bizID)})
	if err != nil {
		return nil, err
	}

	return svc.client.HCService().TCloudZiyan.Image.ListImage(cts.Kit, req)
}
