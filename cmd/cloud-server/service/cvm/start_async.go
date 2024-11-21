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

package cvm

import (
	"fmt"
	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchAsyncStartCvm batch start cvm.
func (svc *cvmSvc) BatchAsyncStartCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchAsyncStartCvmSvc(cts, constant.UnassignedBiz, handler.ResOperateAuth)
}

// BatchAsyncStartBizCvm batch start biz cvm.
func (svc *cvmSvc) BatchAsyncStartBizCvm(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchAsyncStartCvmSvc(cts, bizID, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchAsyncStartCvmSvc(cts *rest.Contexts, bkBizID int64, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.BatchCvmPowerOperateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.validateAuthorize(cts, req.IDs, validHandler); err != nil {
		logs.Errorf("validate authorize and create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := svc.createAudit(cts, req.IDs); err != nil {
		logs.Errorf("create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmList, err := svc.listCvmByIDs(cts.Kit, req.IDs)
	if err != nil {
		logs.Errorf("list cvm by ids failed, ids: %v, err: %v, rid: %s", req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	uniqueID, err := calCvmResetUniqueID(cts.Kit, bkBizID, req.IDs)
	if err != nil {
		logs.Errorf("cal cvm reset unique key failed, err: %v, cvmIDs: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return "", err
	}

	taskManagementID, err := svc.cvmLgc.CvmPowerOperation(cts.Kit, bkBizID, uniqueID, enumor.TaskStartCvm, cvmList)
	if err != nil {
		logs.Errorf("build flow and task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return proto.BatchCvmOperateResp{
		TaskManagementID: taskManagementID,
	}, nil
}

func (svc *cvmSvc) listCvmByIDs(kt *kit.Kit, ids []string) ([]corecvm.BaseCvm, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is empty")
	}
	result := make([]corecvm.BaseCvm, 0, len(ids))
	for _, idList := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", idList),
			),
			Page: core.NewDefaultBasePage(),
		}
		cvm, err := svc.client.DataService().Global.Cvm.ListCvm(kt, listReq)
		if err != nil {
			logs.Errorf("list cvm failed, ids: %v, err: %v, rid: %s", idList, err, kt.Rid)
			return nil, err
		}
		if len(cvm.Details) == 0 {
			return nil, fmt.Errorf("no cvm found by ids: %v", ids)
		}
		result = append(result, cvm.Details...)
	}

	return result, nil
}

func (svc *cvmSvc) validateAuthorize(cts *rest.Contexts, ids []string,
	validHandler handler.ValidWithAuthHandler) error {

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list resource basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Start, BasicInfos: basicInfoMap})
	if err != nil {
		logs.Errorf("validate authorize failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

func (svc *cvmSvc) createAudit(cts *rest.Contexts, ids []string) error {
	if err := svc.audit.ResBaseOperationAudit(cts.Kit, enumor.CvmAuditResType, protoaudit.Start, ids); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}
	return nil
}
