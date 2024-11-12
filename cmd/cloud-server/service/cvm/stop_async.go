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
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchAsyncStopCvm batch stop cvm.
func (svc *cvmSvc) BatchAsyncStopCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchAsyncStopCvmSvc(cts, constant.UnassignedBiz, handler.ResOperateAuth)
}

// BatchAsyncStopBizCvm batch stop biz cvm.
func (svc *cvmSvc) BatchAsyncStopBizCvm(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchAsyncStopCvmSvc(cts, bizID, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchAsyncStopCvmSvc(cts *rest.Contexts, bkBizID int64, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.BatchStopCvmReqV2)
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

	taskManagementID, err := svc.buildFlowAndTaskManagement(cts.Kit, bkBizID, enumor.TaskStopCvm, cvmList)
	if err != nil {
		logs.Errorf("build flow and task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return proto.BatchOperateResp{
		TaskManagementID: taskManagementID,
	}, nil
}

func (svc *cvmSvc) buildFlowAndTaskManagement(kt *kit.Kit, bkBizID int64, taskOperation enumor.TaskOperation, cvmList []corecvm.BaseCvm) (string, error) {

	flowName, actionName, err := chooseFlowNameAndActionName(taskOperation)
	if err != nil {
		return "", err
	}

	vendorList, accountList, groupResult, detailList := groupCvmByVendorAndAccountAndRegion(cvmList)
	taskManagementID, err := svc.createTaskManagement(kt, bkBizID, vendorList, accountList,
		enumor.TaskManagementSourceAPI, taskOperation)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	err = svc.createTaskDetails(kt, bkBizID, taskManagementID, taskOperation, detailList)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flowID, err := svc.buildFlow(kt, actionName, flowName, groupResult)
	if err != nil {
		logs.Errorf("build flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if err = svc.updateTaskManagementAndDetails(kt, taskManagementID, flowID, detailList); err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return taskManagementID, nil
}

func chooseFlowNameAndActionName(taskOperation enumor.TaskOperation) (enumor.FlowName, enumor.ActionName, error) {
	var flowName enumor.FlowName
	var actionName enumor.ActionName
	switch taskOperation {
	case enumor.TaskStopCvm:
		flowName = enumor.FlowStopCvm
		actionName = enumor.ActionStopCvmV2
	case enumor.TaskStartCvm:
		flowName = enumor.FlowStartCvm
		actionName = enumor.ActionStartCvmV2
	case enumor.TaskRebootCvm:
		flowName = enumor.FlowRebootCvm
		actionName = enumor.ActionRebootCvmV2
	default:
		return "", "", fmt.Errorf("batch async operate cvm unsupported task operation: %s", taskOperation)
	}
	return flowName, actionName, nil
}
