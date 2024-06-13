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
	"errors"
	"fmt"

	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/common/utils"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
)

// createSoJob starts a so job
func (m *Matcher) createSoJob(jobName string, ips []string) (int, error) {
	createReq := &sojobapi.CreateJobReq{
		JobName: jobName,
	}
	for _, ip := range ips {
		createReq.List = append(createReq.List, &sojobapi.HostItem{Ip: ip})
	}

	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to create sojob, err: %v", err)
		}
		if obj == nil {
			return false, fmt.Errorf("create sojob resp not found")
		}
		resp, ok := obj.(*sojobapi.CreateJobResp)
		if !ok {
			return false, fmt.Errorf("object is not a create job response: %+v", resp)
		}

		if resp.Code != 0 {
			return false, fmt.Errorf("create sojob failed, code: %d, err: %s", resp.Code, resp.Message)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return m.sojob.CreateJob(nil, nil, createReq)
	}

	createResp, err := utils.Retry(doFunc, checkFunc, 30, 5)
	if err != nil {
		return 0, err
	}

	resp, ok := createResp.(*sojobapi.CreateJobResp)
	if !ok {
		return 0, fmt.Errorf("object is not a create job response: %+v", createResp)
	}

	id, err := util.GetInt64ByInterface(resp.Data)
	if err != nil {
		return 0, fmt.Errorf("create sojob failed, for response data invalid: %+v", resp.Data)
	}

	return int(id), nil
}

// checkJobStatus check so job status
func (m *Matcher) checkJobStatus(jobId int) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to get so job status by id %d, err: %v", jobId, err)
		}
		if obj == nil {
			return false, fmt.Errorf("so job %d not found", jobId)
		}
		resp, ok := obj.(*sojobapi.GetJobStatusResp)
		if !ok {
			return false, fmt.Errorf("object with job id %d is not a job response: %+v", jobId, resp)
		}

		if resp.Code != 0 {
			return false, fmt.Errorf("so job %d failed, err: %s", jobId, resp.Message)
		}

		if resp.Data == nil {
			return false, fmt.Errorf("object with job id %d is not a job response: %+v", jobId, resp)
		}

		if resp.Data.CheckResultCode == sojobapi.JobStatusHandling {
			return false, fmt.Errorf("so job %d handling", jobId)
		}

		if resp.Data.CheckResultCode != sojobapi.JobStatusSuccess {
			return true, fmt.Errorf("so job %d failed, code: %d", jobId, resp.Data.CheckResultCode)
		}

		if len(resp.Data.SubJob) == 0 || resp.Data.SubJob[0] == nil {
			return true, fmt.Errorf("so job %d failed, for sub job is empty", jobId)
		}

		if resp.Data.SubJob[0].Result != "success" {
			return true, fmt.Errorf("so job %d failed, err: %s", jobId, resp.Data.SubJob[0].Comment)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return m.sojob.GetJobStatus(nil, nil, jobId)
	}

	_, err := utils.Retry(doFunc, checkFunc, 1200, 10)
	if err != nil {
		return err
	}

	return nil
}

// checkJobDetailStatus check so job detail status
func (m *Matcher) checkJobStatusDetail(jobId int) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to get so job detail status by id %d, err: %v", jobId, err)
		}
		if obj == nil {
			return false, fmt.Errorf("so job %d not found", jobId)
		}
		resp, ok := obj.(*sojobapi.GetJobStatusDetailResp)
		if !ok {
			return false, fmt.Errorf("object with job id %d is not a job response: %+v", jobId, resp)
		}

		if resp.Code != 0 {
			return false, fmt.Errorf("so job %d failed, err: %s", jobId, resp.Message)
		}

		if resp.Data == nil {
			return false, fmt.Errorf("object with job id %d is not a job response: %+v", jobId, resp)
		}

		if resp.Data.SubJobNum == 0 || len(resp.Data.SubJob) == 0 {
			return false, fmt.Errorf("subjob %d count is 0, retry it", jobId)
		}

		if resp.Data.Status != sojobapi.JobDetailStatusDone {
			return false, fmt.Errorf("so job %d handling", jobId)
		}

		if len(resp.Data.SubJob) == 0 || resp.Data.SubJob[0] == nil {
			return false, fmt.Errorf("subjob %d is empty, retry it", jobId)
		}

		if resp.Data.SubJob[0].Status != sojobapi.JobDetailStatusDone {
			return false, fmt.Errorf("so job %d handling", jobId)
		}

		if resp.Data.SubJob[0].IsFailed {
			return true, errors.New("disk check not pass")
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return m.sojob.GetJobStatusDetail(nil, nil, jobId)
	}

	_, err := utils.Retry(doFunc, checkFunc, 1200, 10)
	if err != nil {
		return err
	}

	return nil
}
