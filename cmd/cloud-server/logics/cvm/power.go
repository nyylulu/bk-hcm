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
	"strings"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	cvmproto "hcm/pkg/api/task-server/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

/**
该文件包含CVM 开关机和重启的操作逻辑
*/

// CvmPowerOperation 开关机或重启CVM
func (c *cvm) CvmPowerOperation(kt *kit.Kit, bkBizID int64, uniqueID string, source enumor.TaskManagementSource,
	taskOperation enumor.TaskOperation, cvmList []corecvm.BaseCvm) (string, error) {

	flowName, actionName, taskType, err := chooseFlowNameAndActionName(taskOperation)
	if err != nil {
		return "", err
	}

	vendorList, accountList, groupResult, detailList := groupCvmByVendorAndAccountAndRegion(cvmList)
	taskManagementID, err := c.createTaskManagement(kt, bkBizID, vendorList, accountList,
		source, taskOperation, enumor.TaskManagementResCVM)
	if err != nil {
		logs.Errorf("create task management failed,bizID: %d, accountList: %v, err: %v, rid: %s",
			bkBizID, accountList, err, kt.Rid)
		return "", err
	}
	err = c.createTaskDetailsForPower(kt, bkBizID, taskManagementID, taskOperation, detailList)
	if err != nil {
		logs.Errorf("create task details failed, taskManagementID: %s, err: %v, rid: %s", taskManagementID, err, kt.Rid)
		return "", err
	}

	flowID, err := c.buildFlowForPower(kt, uniqueID, actionName, flowName, taskType, groupResult)
	if err != nil {
		logs.Errorf("build flow failed, err: %v, rid: %s", err, kt.Rid)
		// update task management state to failed
		detailIDs := make([]string, 0)
		for _, detail := range detailList {
			detailIDs = append(detailIDs, detail.taskDetailID)
		}
		if updateErr := c.updateTaskDetailsState(kt, enumor.TaskDetailFailed, detailIDs, err.Error()); updateErr != nil {
			logs.Errorf("update task management state failed, err: %v, rid: %s", updateErr, kt.Rid)
			return "", updateErr
		}
		return "", err
	}

	if err = c.updateTaskManagementAndDetailsForPower(kt, taskManagementID, flowID, detailList); err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return taskManagementID, nil
}

func (c *cvm) buildFlowForPower(kt *kit.Kit, uniqueID string, actionName enumor.ActionName, flowName enumor.FlowName,
	taskType enumor.TaskType, groupResult map[enumor.Vendor]map[string][]*cvmTaskDetail) (string, error) {

	lockRel, err := checkResFlowRel(kt, c.client.DataService(), uniqueID, enumor.CvmCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, uniqueID: %s, lockRel: %+v, rid: %s",
			err, uniqueID, converter.PtrToVal(lockRel), kt.Rid)
		return "", err
	}

	tasks := make([]ts.CustomFlowTask, 0)
	actionIDGenerator := counter.NewNumStringCounter(1, 10)
	for vendor, detailMap := range groupResult {
		for key, details := range detailMap {
			// 同一个account、region 共用一个flowTask, 所有flowTask共用一个flow
			split := strings.Split(key, "_")
			accountID := split[0]
			region := split[1]

			flowTasks, err := buildFlowTasks(actionName, vendor, accountID, region, details, actionIDGenerator)
			if err != nil {
				logs.Errorf("build flow task failed, err: %v, rid: %s", err, kt.Rid)
				return "", err
			}
			tasks = append(tasks, flowTasks...)
		}
	}
	flowID, err := c.createFlowTask(kt, tasks, taskType, flowName, uniqueID)
	if err != nil {
		logs.Errorf("create flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	err = lockResFlowStatus(kt, c.client.DataService(), c.client.TaskServer(), uniqueID,
		enumor.CvmCloudResType, flowID, taskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, uniqueID: %s, rid: %s", err, uniqueID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (c *cvm) createTaskDetailsForPower(kt *kit.Kit, bkBizID int64, taskManagementID string,
	taskOperation enumor.TaskOperation, details []*cvmTaskDetail) error {

	if len(details) == 0 {
		return nil
	}

	paramMap, err := c.getCvmWithExtMap(kt, details)
	if err != nil {
		logs.Errorf("get cvm with ext map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          bkBizID,
			TaskManagementID: taskManagementID,
			Operation:        taskOperation,
			Param:            paramMap[detail.cvm.ID],
			State:            enumor.TaskDetailInit,
		})
	}

	result, err := c.client.DataService().Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if len(result.IDs) != len(details) {
		return fmt.Errorf("create task details failed, expect created %d task details, but got %d",
			len(details), len(result.IDs))
	}

	for i := range result.IDs {
		details[i].taskDetailID = result.IDs[i]
	}
	return nil
}

func chooseFlowNameAndActionName(taskOperation enumor.TaskOperation) (
	enumor.FlowName, enumor.ActionName, enumor.TaskType, error) {

	var flowName enumor.FlowName
	var actionName enumor.ActionName
	var taskType enumor.TaskType
	switch taskOperation {
	case enumor.TaskStopCvm:
		flowName = enumor.FlowStopCvm
		actionName = enumor.ActionStopCvmV2
		taskType = enumor.StopCvmTaskType
	case enumor.TaskStartCvm:
		flowName = enumor.FlowStartCvm
		actionName = enumor.ActionStartCvmV2
		taskType = enumor.StartCvmTaskType
	case enumor.TaskRebootCvm:
		flowName = enumor.FlowRebootCvm
		actionName = enumor.ActionRebootCvmV2
		taskType = enumor.RebootCvmTaskType
	default:
		return "", "", "", fmt.Errorf("batch async operate cvm unsupported task operation: %s", taskOperation)
	}
	return flowName, actionName, taskType, nil
}

// groupCvmByVendorAndAccountAndRegion group cvm by vendor, account and region
// result vendor -> account_region -> []*cvmTaskDetail
func groupCvmByVendorAndAccountAndRegion(cvmList []corecvm.BaseCvm) (vendorList []enumor.Vendor, accountList []string,
	result map[enumor.Vendor]map[string][]*cvmTaskDetail, detailList []*cvmTaskDetail) {

	vendorMap := make(map[enumor.Vendor]struct{})
	accountMap := make(map[string]struct{})
	groupResult := make(map[enumor.Vendor]map[string][]*cvmTaskDetail)
	detailList = make([]*cvmTaskDetail, 0)
	for _, cvm := range cvmList {
		vendorMap[cvm.Vendor] = struct{}{}
		accountMap[cvm.AccountID] = struct{}{}

		m, exist := groupResult[cvm.Vendor]
		if !exist {
			m = make(map[string][]*cvmTaskDetail)
		}
		key := cvm.AccountID + "_" + cvm.Region
		l, exist := m[key]
		if !exist {
			l = make([]*cvmTaskDetail, 0)
		}
		detail := &cvmTaskDetail{cvm: cvm}
		l = append(l, detail)
		detailList = append(detailList, detail)
		m[key] = l
		groupResult[cvm.Vendor] = m
	}

	return converter.MapKeyToSlice(vendorMap), converter.MapKeyToSlice(accountMap), groupResult, detailList
}

type cvmTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string

	cvm corecvm.BaseCvm
}

// getCvmWithExtMap map cvm id to cvm with ext
func (c *cvm) getCvmWithExtMap(kt *kit.Kit, details []*cvmTaskDetail) (map[string]interface{}, error) {
	cvmIDGroupByVendor := make(map[enumor.Vendor][]string)
	for _, detail := range details {
		cvmIDGroupByVendor[detail.cvm.Vendor] = append(cvmIDGroupByVendor[detail.cvm.Vendor], detail.cvm.ID)
	}

	result := make(map[string]interface{})
	for vendor, ids := range cvmIDGroupByVendor {
		switch vendor {
		case enumor.TCloud:
			cvmMap, err := c.listTCloudCvmWithExt(kt, ids)
			if err != nil {
				logs.Errorf("list tcloud cvm with ext failed, ids: %v, err: %v, rid: %s", ids, err, kt.Rid)
				return nil, err
			}
			for key, value := range cvmMap {
				result[key] = value
			}
		case enumor.TCloudZiyan:
			cvms, err := c.listTCloudZiyanCvmWithExt(kt, ids)
			if err != nil {
				logs.Errorf("list tcloud ziyan cvm with ext failed, ids: %v, err: %v, rid: %s", ids, err, kt.Rid)
				return nil, err
			}
			for key, value := range cvms {
				result[key] = value
			}
		default:
			return nil, fmt.Errorf("getCvmWithExtMap, unsupported vendor: %s", vendor)
		}
	}
	return result, nil
}

func (c *cvm) listTCloudCvmWithExt(kt *kit.Kit, ids []string) (
	map[string]corecvm.Cvm[corecvm.TCloudCvmExtension], error) {

	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is empty")
	}
	cvmList := make([]corecvm.Cvm[corecvm.TCloudCvmExtension], 0, len(ids))
	for _, idList := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.CvmListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", idList),
			),
			Page: core.NewDefaultBasePage(),
		}
		cvms, err := c.client.DataService().TCloud.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list cvm failed, ids: %v, err: %v, rid: %s", idList, err, kt.Rid)
			return nil, err
		}
		if len(cvms.Details) == 0 {
			return nil, fmt.Errorf("no cvm found by ids: %v", ids)
		}
		cvmList = append(cvmList, cvms.Details...)
	}
	result := make(map[string]corecvm.Cvm[corecvm.TCloudCvmExtension])
	for _, item := range cvmList {
		result[item.ID] = item
	}

	return result, nil
}

func (c *cvm) listTCloudZiyanCvmWithExt(kt *kit.Kit, ids []string) (
	map[string]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], error) {

	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is empty")
	}
	cvmList := make([]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], 0, len(ids))
	for _, idList := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.CvmListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", idList),
			),
			Page: core.NewDefaultBasePage(),
		}
		cvms, err := c.client.DataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list cvm failed, ids: %v, err: %v, rid: %s", idList, err, kt.Rid)
			return nil, err
		}
		if len(cvms.Details) == 0 {
			return nil, fmt.Errorf("no cvm found by ids: %v", ids)
		}
		cvmList = append(cvmList, cvms.Details...)
	}
	result := make(map[string]corecvm.Cvm[corecvm.TCloudZiyanHostExtension])
	for _, item := range cvmList {
		result[item.ID] = item
	}

	return result, nil
}

func (c *cvm) updateTaskManagementAndDetailsForPower(kt *kit.Kit, taskManagementID, flowID string, details []*cvmTaskDetail) error {

	for _, detail := range details {
		detail.flowID = flowID
	}
	if err := c.updateTaskDetailsForPower(kt, details); err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if err := c.updateTaskManagement(kt, taskManagementID, []string{flowID}); err != nil {
		logs.Errorf("update task management failed, taskManagementID: %s, flowID: %s, err: %v, rid: %s",
			taskManagementID, flowID, err, kt.Rid)
		return err
	}
	return nil
}

func (c *cvm) updateTaskDetailsForPower(kt *kit.Kit, details []*cvmTaskDetail) error {
	if len(details) == 0 {
		return nil
	}
	updateItems := make([]task.UpdateTaskDetailField, 0, len(details))
	for _, detail := range details {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:            detail.taskDetailID,
			FlowID:        detail.flowID,
			TaskActionIDs: []string{detail.actionID},
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.client.DataService().Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func buildFlowTasks(actionName enumor.ActionName, vendor enumor.Vendor, accountID, region string,
	details []*cvmTaskDetail, actionIDGenerator func() string) ([]ts.CustomFlowTask, error) {

	switch vendor {
	case enumor.TCloud, enumor.TCloudZiyan:
	default:
		return nil, fmt.Errorf("build flow task for vendor: %s not support", vendor)
	}

	paramMaps := make(map[string]*cvmproto.CvmOperationOption)
	tasks := make([]ts.CustomFlowTask, 0, len(paramMaps))

	splitDetails := slice.Split(details, constant.BatchOperationMaxLimit)
	for _, list := range splitDetails {
		if len(list) == 0 {
			continue
		}

		actionID := actionIDGenerator()
		opt := &cvmproto.CvmOperationOption{
			Vendor:              vendor,
			AccountID:           accountID,
			Region:              region,
			IDs:                 make([]string, 0),
			ManagementDetailIDs: make([]string, 0),
		}

		for _, detail := range list {
			opt.IDs = append(opt.IDs, detail.cvm.ID)
			opt.ManagementDetailIDs = append(opt.ManagementDetailIDs, detail.taskDetailID)
			detail.actionID = actionID
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(actionID),
			ActionName: actionName,
			Params:     opt,
		})
	}

	return tasks, nil
}
