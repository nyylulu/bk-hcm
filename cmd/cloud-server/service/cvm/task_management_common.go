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

	"hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (svc *cvmSvc) createTaskManagement(kt *kit.Kit, bkBizID int64, vendors []enumor.Vendor, accountIDs []string,
	source enumor.TaskManagementSource, operation enumor.TaskOperation) (string, error) {

	taskManagementCreateReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    bkBizID,
				Source:     source,
				Vendors:    vendors,
				AccountIDs: accountIDs,
				Resource:   enumor.TaskManagementResCVM,
				State:      enumor.TaskManagementRunning,
				Operations: []enumor.TaskOperation{operation},
			},
		},
	}

	result, err := svc.client.DataService().Global.TaskManagement.Create(kt, taskManagementCreateReq)
	if err != nil {
		logs.Errorf("create task management failed, req: %v, err: %v, rid: %s", taskManagementCreateReq, err, kt.Rid)
		return "", err
	}
	if len(result.IDs) == 0 {
		return "", fmt.Errorf("create task management failed")
	}
	return result.IDs[0], nil
}

func (svc *cvmSvc) createTaskDetails(kt *kit.Kit, bkBizID int64, taskManagementID string,
	taskOperation enumor.TaskOperation, details []*cvmTaskDetail) error {

	if len(details) == 0 {
		return nil
	}

	paramMap, err := svc.getCvmWithExtMap(kt, details)
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

	result, err := svc.client.DataService().Global.TaskDetail.Create(kt, taskDetailsCreateReq)
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

func (svc *cvmSvc) updateTaskManagement(kt *kit.Kit, taskID string, flowIDs ...string) error {

	if len(flowIDs) == 0 {
		return nil
	}
	updateItem := task.UpdateTaskManagementField{
		ID:      taskID,
		FlowIDs: flowIDs,
	}
	updateReq := &task.UpdateManagementReq{
		Items: []task.UpdateTaskManagementField{updateItem},
	}
	err := svc.client.DataService().Global.TaskManagement.Update(kt, updateReq)
	if err != nil {
		logs.Errorf("update task management failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (svc *cvmSvc) updateTaskDetails(kt *kit.Kit, details []*cvmTaskDetail) error {
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
	err := svc.client.DataService().Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
