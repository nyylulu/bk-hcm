/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

	moalogic "hcm/cmd/cloud-server/logics/moa"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/cloud-server/moa"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {

	svc := &service{
		authorizer: c.Authorizer,
		moaLogic:   c.Logics.Moa,
		apiClient:  c.ApiClient,
	}

	h := rest.NewHandler()

	h.Add("MoaBizRequest", http.MethodPost, "/bizs/{bk_biz_id}/moa/request", svc.MoaBizRequest)
	h.Add("MoaRequest", http.MethodPost, "/moa/request", svc.MoaRequest)

	h.Add("MoaBizVerify", http.MethodPost, "/bizs/{bk_biz_id}/moa/verify", svc.MoaBizVerify)
	h.Add("MoaVerify", http.MethodPost, "/moa/verify", svc.MoaVerify)

	h.Load(c.WebService)
}

type service struct {
	authorizer auth.Authorizer
	moaLogic   moalogic.Interface
	apiClient  *client.ClientSet
}

// MoaRequest moa biz request.
func (s service) MoaRequest(cts *rest.Contexts) (any, error) {

	return s.moaRequest(cts, constant.UnassignedBiz)
}

// MoaBizRequest moa biz request.
func (s service) MoaBizRequest(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.moaRequest(cts, bizID)
}

// MoaBizRequest ...
func (s service) moaRequest(cts *rest.Contexts, bizID int64) (any, error) {

	req := new(moa.InitiateVerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resType := req.Scene.GetResType()
	if err := s.checkBizWithPerm(cts.Kit, bizID, resType, req.ResIDs); err != nil {
		logs.Errorf("moa request fail to check biz , err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	sessionID, err := s.moaLogic.RequestMoa(cts.Kit, req.Scene, len(req.ResIDs), req.Language)
	if err != nil {
		return nil, err
	}
	result := moa.InitiateVerificationResp{SessionId: sessionID}
	return result, nil
}

func (s service) checkBizFindPermission(kt *kit.Kit, bizID int64, resType meta.ResourceType, resIDs []string) error {

	attrs := make([]meta.ResourceAttribute, 0, len(resIDs))
	for _, resID := range resIDs {
		attr := meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       resType,
				Action:     meta.Find,
				ResourceID: resID,
			},
			BizID: bizID,
		}
		attrs = append(attrs, attr)
	}
	// 检查对应资源的查看权限
	_, authorized, err := s.authorizer.Authorize(kt, attrs...)
	if err != nil {
		return nil
	}
	if !authorized {
		return errf.New(errf.PermissionDenied, "permission denied for request moa")
	}
	return err
}

func (s service) checkBizWithPerm(kt *kit.Kit, bizID int64, resType meta.ResourceType, resIDs []string) error {

	// 	判断资源是否在业务下
	basicRes := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.CloudResourceType(resType),
		IDs:          resIDs,
		Fields:       []string{"id", "bk_biz_id", "account_id"},
	}
	basicInfos, err := s.apiClient.DataService().Global.Cloud.ListResBasicInfo(kt, basicRes)
	if err != nil {
		logs.Errorf("fail to list resource basic info, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	resIDMap := make(map[string]struct{})
	for _, resID := range resIDs {
		resIDMap[resID] = struct{}{}
	}
	for i := range basicInfos {
		basicInfo := basicInfos[i]
		delete(resIDMap, basicInfo.ID)
		if basicInfo.BkBizID == bizID {
			continue
		}
		return errf.Newf(errf.InvalidParameter, "resource %s not belong to biz %d", basicInfo.ID, bizID)
	}
	if len(resIDMap) != 0 {
		return errf.Newf(errf.InvalidParameter, "resource %v not found", cvt.MapKeyToSlice(resIDMap))
	}

	// 检查权限
	if bizID < 0 {
		accountIDs := make([]string, 0, len(resIDs))
		for _, resID := range resIDs {
			accountIDs = append(accountIDs, basicInfos[resID].AccountID)
		}
		// 资源下检查账号权限
		err := s.checkBizFindPermission(kt, bizID, resType, accountIDs)
		if err != nil {
			logs.Errorf("fail to check permission , err: %v, rid: %s", err, kt.Rid)
			return err
		}
	} else {
		err := s.checkBizFindPermission(kt, bizID, resType, resIDs)
		if err != nil {
			logs.Errorf("fail to check permission , err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// MoaVerify moa biz verify.
func (s service) MoaVerify(cts *rest.Contexts) (any, error) {
	return s.moaVerify(cts, constant.UnassignedBiz)
}

// MoaBizVerify moa biz verify.
func (s service) MoaBizVerify(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	return s.moaVerify(cts, bizID)
}

// moaVerify ...
func (s service) moaVerify(cts *rest.Contexts, bizID int64) (any, error) {
	req := new(moa.VerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resType := req.Scene.GetResType()
	if err := s.checkBizWithPerm(cts.Kit, bizID, resType, req.ResIDs); err != nil {
		logs.Errorf("moa verify fail to check biz , err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	status, err := s.moaLogic.VerifyMoa(cts.Kit, req.Scene, req.SessionId)
	if err != nil {
		return nil, err
	}
	if status == enumor.MoaVerifyNotFound {
		return nil, errf.New(errf.MOAValidationTimeoutError, "session id expired or not found")
	}
	result := moa.VerificationResp{
		Status:    enumor.MoaStatusPending,
		SessionId: req.SessionId,
	}
	if status != enumor.MoaVerifyPending {
		result.Status = enumor.MoaStatusFinish
		if status == enumor.MoaVerifyConfirmed {
			result.ButtonType = enumor.MoaButtonTypeConfirm
		} else {
			result.ButtonType = enumor.MoaButtonTypeCancel
		}
	}

	return result, nil
}
