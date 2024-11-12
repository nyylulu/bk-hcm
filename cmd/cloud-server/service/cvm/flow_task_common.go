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
	"strconv"

	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

func (svc *cvmSvc) createFlowTask(kt *kit.Kit, flowName enumor.FlowName, flowTasks []ts.CustomFlowTask) (string, error) {
	addReq := &ts.AddCustomFlowReq{
		Name:  flowName,
		Tasks: flowTasks,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return result.ID, nil
}

func buildFlowTasks(actionName enumor.ActionName, vendor enumor.Vendor, accountID, region string,
	details []*cvmTaskDetail) ([]ts.CustomFlowTask, error) {

	switch vendor {
	case enumor.TCloud, enumor.TCloudZiyan:
	default:
		return nil, fmt.Errorf("build flow task for vendor: %s not support", vendor)
	}

	paramMaps := make(map[string]*actioncvm.CvmOperationOptionV2)
	tasks := make([]ts.CustomFlowTask, 0, len(paramMaps))
	count := 1

	splitDetails := slice.Split(details, constant.BatchOperationMaxLimit)
	for _, list := range splitDetails {
		if len(list) == 0 {
			continue
		}

		opt := &actioncvm.CvmOperationOptionV2{
			Vendor:              vendor,
			AccountID:           accountID,
			Region:              region,
			IDs:                 make([]string, 0),
			ManagementDetailIDs: make([]string, 0),
		}

		for _, detail := range list {
			opt.IDs = append(opt.IDs, detail.cvm.ID)
			opt.ManagementDetailIDs = append(opt.ManagementDetailIDs, detail.taskDetailID)
			detail.actionID = strconv.Itoa(count)
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(strconv.Itoa(count)),
			ActionName: actionName,
			Params:     opt,
		})
		count++
	}

	return tasks, nil
}
