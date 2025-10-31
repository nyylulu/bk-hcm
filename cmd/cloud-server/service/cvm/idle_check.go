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

	"hcm/cmd/woa-server/dal/task/table"
	types "hcm/cmd/woa-server/types/task"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
)

// BatchIdleCheckBizCvm batch precheck biz cvm.
func (svc *cvmSvc) BatchIdleCheckBizCvm(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	req := new(cscvm.BatchIdleCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmList, err := svc.listZiYanCvmByBkHostIDs(cts.Kit, req.BkHostIDs)
	if err != nil {
		logs.Errorf("list cvm by host ids failed, hostIDs: %v, err: %v, rid: %s", req.BkHostIDs, err, cts.Kit.Rid)
		return nil, err
	}
	// 提供的BkHostIDs要求是自研云的主机
	if len(cvmList) != len(req.BkHostIDs) {
		return nil, fmt.Errorf("cvm list length(%d) not equal to bkHostIDs length(%d)",
			len(cvmList), len(req.BkHostIDs))
	}

	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	})
	if err != nil {
		logs.Errorf("failed to check idle check permission, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	taskManagementID, suborderID, err := svc.cvmLgc.CvmIdleCheck(cts.Kit, bizID, req.BkHostIDs,
		enumor.TaskManagementSourceAPI, cvmList)
	if err != nil {
		logs.Errorf("build flow and task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return proto.BatchIdleCheckCvmRsp{
		TaskManagementID:    taskManagementID,
		IdleCheckSuborderID: suborderID,
	}, nil
}

// listZiYanCvmByBkHostIDs 根据BkHostIDs查询自研云Cvms
func (svc *cvmSvc) listZiYanCvmByBkHostIDs(kt *kit.Kit, bkHostIDs []int64) (
	[]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], error) {
	if len(bkHostIDs) == 0 {
		return nil, fmt.Errorf("BkHostIDs is empty")
	}
	result := make([]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], 0, len(bkHostIDs))
	listReq := &protocloud.CvmListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("bk_host_id", bkHostIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	cvm, err := svc.client.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list cvm failed, bkHostIDs: %v, err: %v, rid: %s", bkHostIDs, err, kt.Rid)
		return nil, err
	}
	if len(cvm.Details) == 0 {
		return nil, fmt.Errorf("no cvm found by ids: %v", bkHostIDs)
	}
	result = append(result, cvm.Details...)

	return result, nil
}

// GetIdleCheckCvmResult 获取空闲检查Cvm结果
func (svc *cvmSvc) GetIdleCheckCvmResult(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	suborderID := cts.PathParameter("suborder_id").String()
	if len(suborderID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("suborder_id is empty"))
	}
	req := new(proto.IdleCheckResultReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	page := req.Page
	err = svc.validateGetIdleCheckCvmResultReq(cts, bizID, suborderID, page)
	if err != nil {
		return nil, err
	}

	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	})
	if err != nil {
		logs.Errorf("failed to check permission in GetIdleCheckCvmResult, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.getIdleCheckCvmResult(cts.Kit, suborderID, page)
}

// validateGetIdleCheckCvmResultReq 校验GetIdleCheckCvmResult请求参数
func (svc *cvmSvc) validateGetIdleCheckCvmResultReq(cts *rest.Contexts, bizID int64, suborderID string,
	page *core.BasePage) error {

	orders, err := svc.client.WoaServer().Task.ListRecycleOrder(cts.Kit, &types.GetRecycleOrderReq{
		SuborderID: []string{suborderID},
		BizID:      []int64{bizID},
	})
	if err != nil {
		return err
	}
	if len(orders.Info) == 0 {
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no order found by "+
			"suborder_id: %s, bizID: %d", suborderID, bizID))
	}

	// 1台待空闲检查的主机->1个detectTask->10个detectStep，因为500/10=50，所以限制查询主机数为50
	validateOpt := &core.PageOption{
		MaxLimit: table.DetectTaskMaxPageLimit,
	}
	if err := page.Validate(validateOpt); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}

// getIdleCheckCvmResult 构建空闲检查结果
func (svc *cvmSvc) getIdleCheckCvmResult(kt *kit.Kit, suborderID string, page *core.BasePage) (
	*proto.IdleCheckResultRsp, error) {

	listDetectTaskReq := &types.GetRecycleDetectReq{
		SuborderID: []string{suborderID},
		Page: metadata.BasePage{
			Sort:        page.Sort,
			Limit:       int(page.Limit),
			Start:       int(page.Start),
			EnableCount: page.Count,
		},
	}
	tasks, err := svc.client.WoaServer().Task.ListDetectTask(kt, listDetectTaskReq)
	if err != nil {
		logs.Errorf("list detect task failed, suborderID: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return nil, err
	}

	if page.Count {
		return &proto.IdleCheckResultRsp{Count: uint64(tasks.Count)}, nil
	}

	details := make([]*proto.IdleCheckResultRspItem, 0, len(tasks.Info))
	// 根据taskID取得对应空闲检查任务执行结果的响应结构体
	taskIDToRspItem := make(map[string]*proto.IdleCheckResultRspItem)
	for _, task := range tasks.Info {
		if task == nil {
			logs.Errorf("detect task is nil, suborderID: %s, rid: %s", suborderID, kt.Rid)
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("detect task is nil"))
		}
		if task.TaskID == "" {
			logs.Errorf("detect task taskID is empty, suborderID: %s, rid: %s", suborderID, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("detect task taskID is empty"))
		}
		rspItem := &proto.IdleCheckResultRspItem{
			DetectTask: converter.PtrToVal(task),
		}
		taskIDToRspItem[task.TaskID] = rspItem
		details = append(details, rspItem)
	}
	listDetectStepReq := &types.GetDetectStepReq{
		SuborderID: []string{suborderID},
	}
	steps, err := svc.client.WoaServer().Task.ListDetectStep(kt, listDetectStepReq)
	if err != nil {
		logs.Errorf("list detect step failed, suborderID: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return nil, err
	}
	for _, step := range steps.Info {
		if step == nil {
			logs.Errorf("detect step is nil, suborderID: %s, rid: %s", suborderID, kt.Rid)
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("detect step is nil"))
		}
		if _, ok := taskIDToRspItem[step.TaskID]; !ok {
			logs.Errorf("detect step task id not found in task map, step taskID: %s, suborderID: %s, "+
				"stepIP: %s, rid: %s", step.TaskID, suborderID, step.IP, kt.Rid)
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("detect step task id %s not found "+
				"in task map", step.TaskID))
		}
		// 将当前空闲检查步骤添加到对应空闲检查任务的步骤列表中，通过taskID找到对应的响应项，将空闲检查步骤追加到其DetectSteps切片中
		// 这样可以将同一个空闲检查任务的多个空闲检查步骤聚合在一起返回给客户端
		taskIDToRspItem[step.TaskID].DetectSteps = append(taskIDToRspItem[step.TaskID].DetectSteps,
			converter.PtrToVal(step))
	}
	return &proto.IdleCheckResultRsp{Details: details}, nil
}
