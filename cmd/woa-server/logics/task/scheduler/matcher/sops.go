/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package matcher provides ...
package matcher

import (
	"fmt"
	"strconv"

	"hcm/cmd/woa-server/common/utils"
	"hcm/cmd/woa-server/thirdparty/sopsapi"
)

// createSopsTask starts a sops task
func (m *Matcher) createSopsTask(taskName string, ips []string) (int, error) {
	createReq := &sopsapi.CreateTaskReq{}

	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to create sops task, err: %v", err)
		}
		if obj == nil {
			return false, fmt.Errorf("create sops task resp not found")
		}
		resp, ok := obj.(*sopsapi.CreateTaskResp)
		if !ok {
			return false, fmt.Errorf("object is not a create sops task response: %+v", resp)
		}

		if resp.Result != true {
			return false, fmt.Errorf("create sops task failed, code: %d, err: %s", resp.Code, resp.Message)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return m.sops.CreateTask(nil, nil, "", "", createReq)
	}

	createResp, err := utils.Retry(doFunc, checkFunc, 30, 5)
	if err != nil {
		return 0, err
	}

	resp, ok := createResp.(*sopsapi.CreateTaskResp)
	if !ok {
		return 0, fmt.Errorf("object is not a create sops task response: %+v", createResp)
	}

	taskId := resp.Data.TaskId
	if !ok {
		return 0, fmt.Errorf("create sops task failed, for response data invalid: %+v", resp.Data)
	}

	return taskId, nil
}

// checkJobStatus check sops task status
func (m *Matcher) checkTaskStatus(taskId int) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to get sops task status by id %d, err: %v", taskId, err)
		}
		if obj == nil {
			return false, fmt.Errorf("sops task %d not found", taskId)
		}
		resp, ok := obj.(*sopsapi.GetTaskStatusResp)
		if !ok {
			return false, fmt.Errorf("object with task id %d is not a task response: %+v", taskId, resp)
		}

		if resp.Result != true {
			return false, fmt.Errorf("sops task %d failed, err: %s", taskId, resp.Message)
		}

		if resp.Data == nil {
			return false, fmt.Errorf("object with task id %d is not a task response: %+v", taskId, resp)
		}

		if resp.Data.State == sopsapi.TaskStateRunning || resp.Data.State == sopsapi.TaskStateCreated {
			return false, fmt.Errorf("sops task %d handling", taskId)
		}

		if resp.Data.State != sopsapi.TaskStateFinished {
			return true, fmt.Errorf("sops task %d failed, err: %s", taskId, resp.Data.State)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return m.sops.GetTaskStatus(nil, nil, strconv.Itoa(taskId), "")
	}

	_, err := utils.Retry(doFunc, checkFunc, 3600, 10)
	if err != nil {
		return err
	}

	return nil
}
