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

// Package detector ...
package detector

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/common/utils"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
	"hcm/pkg/logs"
)

func (d *Detector) checkProcess(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		err = d.checkIsClear(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		exeInfo = err.Error()
	}

	return attempt, exeInfo, err
}

func (d *Detector) checkIsClear(ip string) error {
	// 1. create job
	ips := []string{ip}
	jobId, err := d.createSoJob("isclear", ips)
	if err != nil {
		logs.Errorf("host %s failed to check process, err: %v", ip, err)
		return fmt.Errorf("failed to check process, err: %v", err)
	}

	// 2. get job status
	if err := d.checkJobStatus(jobId); err != nil {
		// if host ping death, go ahead to recycle
		if strings.Contains(err.Error(), "ping death") {
			logs.Infof("host %s ping death, skip check process step", ip)
			return nil
		}
		logs.Infof("host %s failed to check process, job id: %d, err: %v", ip, err)
		return fmt.Errorf("host %s failed to check process, job id: %d, err: %v", ip, jobId, err)
	}
	return nil
}

// createSoJob starts a so job
func (d *Detector) createSoJob(jobName string, ips []string) (int, error) {
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
			return false, fmt.Errorf("create sojob failed, code: %d, msg: %s", resp.Code, resp.Message)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return d.sojob.CreateJob(nil, nil, createReq)
	}

	createResp, err := utils.Retry(doFunc, checkFunc, 30, 5)
	if err != nil {
		return 0, err
	}

	resp, ok := createResp.(*sojobapi.CreateJobResp)
	if !ok {
		return 0, fmt.Errorf("object is not a create job response: %+v", createResp)
	}

	jobId, ok := resp.Data.(json.Number)
	if !ok {
		return 0, fmt.Errorf("create sojob failed, for response data invalid: %+v", resp.Data)
	}

	id, err := jobId.Int64()
	if err != nil {
		return 0, fmt.Errorf("create sojob failed, for response data invalid: %+v", resp.Data)
	}

	return int(id), nil
}

// checkJobStatus check so job status
func (d *Detector) checkJobStatus(jobId int) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to get so job status by id %d, err: %v", jobId, err)
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

		if resp.Data.SubJobNum == 0 || len(resp.Data.SubJob) == 0 || resp.Data.SubJob[0] == nil {
			return false, fmt.Errorf("subjob %d count is 0, retry it", jobId)
		}

		if resp.Data.Status == sojobapi.JobDetailStatusTodo || resp.Data.Status == sojobapi.JobDetailStatusStart ||
			resp.Data.Status == sojobapi.JobDetailStatusDoing {
			return false, fmt.Errorf("so job %d handling", jobId)
		}

		if resp.Data.SubJob[0].Status == sojobapi.JobDetailStatusTodo ||
			resp.Data.SubJob[0].Status == sojobapi.JobDetailStatusStart ||
			resp.Data.SubJob[0].Status == sojobapi.JobDetailStatusDoing {
			return false, fmt.Errorf("so job %d handling", jobId)
		}

		// not pass process check
		if resp.Data.SubJobSucc == 0 {
			msg := resp.Data.SubJob[0].Comment
			if strings.HasPrefix(msg, "dity") {
				msg = d.removeUnusedComment(resp.Data.SubJob[0].CommentMore)
			}
			return true, fmt.Errorf("check process not pass: %s", msg)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return d.sojob.GetJobStatusDetail(nil, nil, jobId)
	}

	// timeout 20 minutes
	_, err := utils.Retry(doFunc, checkFunc, 1200, 10)
	if err != nil {
		return err
	}

	return nil
}

func (d *Detector) removeUnusedComment(comment string) string {
	var msg []string
	for _, line := range strings.Split(comment, "\n") {
		if strings.Contains(line, "STATUS") && !strings.Contains(line, "_num:0") {
			index := strings.Index(line, "STATUS")
			msg = append(msg, line[index:])
		}
	}

	return strings.Join(msg, "\n")
}
