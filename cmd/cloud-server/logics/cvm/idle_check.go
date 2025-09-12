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

	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	taskcvm "hcm/pkg/api/task-server/cvm"
	woaserver "hcm/pkg/api/woa-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// CvmIdleCheck 空闲检查CVM
func (c *cvm) CvmIdleCheck(kt *kit.Kit, bkBizID int64, bkHostIDs []int64, source enumor.TaskManagementSource,
	cvmList []corecvm.Cvm[corecvm.TCloudZiyanHostExtension]) (string, string, error) {

	baseCvms := slice.Map[corecvm.Cvm[corecvm.TCloudZiyanHostExtension], corecvm.BaseCvm](
		cvmList, func(c corecvm.Cvm[corecvm.TCloudZiyanHostExtension]) corecvm.BaseCvm {
			return c.BaseCvm
		})
	vendorList, accountList, _, detailList := groupCvmByVendorAndAccountAndRegion(baseCvms)
	taskManagementID, err := c.createTaskManagement(kt, bkBizID, vendorList, accountList,
		source, enumor.TaskIdleCheckCvm, enumor.TaskManagementResCVM)
	if err != nil {
		logs.Errorf("create task management failed, bizID: %d, accountList: %v, err: %v, rid: %s", bkBizID,
			accountList, err, kt.Rid)
		return "", "", err
	}

	hostIDToTaskDetailID, err := c.createTaskDetailsForIdleCheck(kt, bkBizID, taskManagementID, enumor.TaskIdleCheckCvm,
		detailList)
	if err != nil {
		logs.Errorf("create task details failed, taskManagementID: %s, err: %v, rid: %s", taskManagementID, err,
			kt.Rid)
		return "", "", err
	}

	defer func() {
		// 只有在发生错误时才更新任务状态为失败
		if err != nil {
			detailIDs := make([]string, 0)
			for _, detail := range detailList {
				detailIDs = append(detailIDs, detail.taskDetailID)
			}
			if len(detailIDs) == 0 {
				return
			}
			updateErr := c.updateTaskDetailsState(kt, enumor.TaskDetailFailed, detailIDs, err.Error())
			if updateErr != nil {
				logs.Errorf("update task management state failed, err: %v, rid: %s", updateErr, kt.Rid)
			}
		}
	}()

	assetIDs := make([]string, 0, len(cvmList))
	ips := make([]string, 0, len(cvmList))
	for _, one := range cvmList {
		// 检查BkAssetID是否为空
		if len(one.Extension.BkAssetID) == 0 {
			logs.Errorf("BkAssetID is empty, cvm: %v, rid: %s", one, kt.Rid)
			return "", "", fmt.Errorf("BkAssetID is empty")
		}
		// 检查PrivateIPv4Addresses是否为空
		if len(one.BaseCvm.PrivateIPv4Addresses) == 0 {
			logs.Errorf("PrivateIPv4Addresses is empty, cvm: %v, rid: %s", one, kt.Rid)
			return "", "", fmt.Errorf("PrivateIPv4Addresses is empty")
		}
		assetIDs = append(assetIDs, one.Extension.BkAssetID)
		ips = append(ips, one.BaseCvm.PrivateIPv4Addresses[0])
	}

	req := &woaserver.StartIdleCheckReq{
		HostIDs:  bkHostIDs,
		AssetIDs: assetIDs,
		IPs:      ips,
		BkBizID:  bkBizID,
	}
	result, err := c.client.WoaServer().Task.StartIdleCheck(kt, req)
	if err != nil {
		logs.Errorf("start idle check failed, err: %v, rid: %s", err, kt.Rid)
		return "", "", err
	}

	flowID, err := c.buildFlowForIdleCheck(kt, result, hostIDToTaskDetailID)
	if err != nil {
		logs.Errorf("build flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", "", err
	}

	if err = c.updateTaskManagementAndDetailsForCvm(kt, taskManagementID, flowID, detailList); err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", "", err
	}
	return taskManagementID, result.SuborderID, nil
}

// createTaskDetailsForIdleCheck 创建任务详情
func (c *cvm) createTaskDetailsForIdleCheck(kt *kit.Kit, bkBizID int64, taskManagementID string,
	taskOperation enumor.TaskOperation, details []*cvmTaskDetail) (map[int64]string, error) {

	if len(details) == 0 {
		return nil, nil
	}

	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          bkBizID,
			TaskManagementID: taskManagementID,
			Operation:        taskOperation,
			State:            enumor.TaskDetailInit,
			Param:            detail,
		})
	}

	result, err := c.client.DataService().Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(result.IDs) != len(details) {
		return nil, fmt.Errorf("create task details failed, expect created %d task details, but got %d",
			len(details), len(result.IDs))
	}

	hostIDToManagementDetailID := make(map[int64]string)
	for i := range result.IDs {
		details[i].taskDetailID = result.IDs[i]
		hostIDToManagementDetailID[details[i].cvm.BkHostID] = details[i].taskDetailID
	}
	return hostIDToManagementDetailID, nil
}

// buildFlowForIdleCheck build flow for monitor idle check
func (c *cvm) buildFlowForIdleCheck(kt *kit.Kit, rsp *woaserver.StartIdleCheckRsp,
	hostIDToManagementDetailID map[int64]string) (string, error) {

	taskOpt := &taskcvm.MonitorIdleCheckCvmOption{
		SuborderID:           rsp.SuborderID,
		HostIDToTaskDetailID: hostIDToManagementDetailID,
	}
	addFlowReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowIdleCheckCvm,
		Tasks: []ts.CustomFlowTask{
			{
				ActionID:   "1",
				ActionName: enumor.ActionMonitorIdleCheckCvm,
				Params:     taskOpt,
				DependOn:   nil,
			},
		},
	}
	result, err := c.client.TaskServer().CreateCustomFlow(kt, addFlowReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return result.ID, nil
}
