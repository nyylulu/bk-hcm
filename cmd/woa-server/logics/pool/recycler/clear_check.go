/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package recycler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/common/utils"
	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
	"hcm/pkg/logs"
)

func (r *Recycler) dealClearCheckTask(task *table.RecallDetail) error {
	if task.ClearCheckID == "" {
		return r.createClearCheckTask(task)
	}

	return r.checkClearCheckStatus(task)
}

func (r *Recycler) createClearCheckTask(task *table.RecallDetail) error {
	// create job
	ip, ok := task.Labels[table.IPKey]
	if !ok || ip == "" {
		err := errors.New("get no ip from task label")
		logs.Errorf("failed to create clear check task, err: %v", err)

		errUpdate := r.updateTaskClearCheckStatus(task, "", err.Error(), table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	ips := []string{ip}
	taskID, err := r.createSoJob("isclear", ips)
	if err != nil {
		logs.Errorf("host %s failed to create clear check task, err: %v", ip, err)

		errUpdate := r.updateTaskClearCheckStatus(task, "", err.Error(), table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to create clear check task, err: %v", err)
	}

	// update task status
	if err := r.updateTaskClearCheckStatus(task, strconv.Itoa(taskID), "",
		table.RecallStatusClearChecking); err != nil {
		logs.Errorf("failed to update recall task status, err: %v", err)
		return err
	}

	go func() {
		// query every 5 minutes
		time.Sleep(time.Minute * 5)
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) checkClearCheckStatus(task *table.RecallDetail) error {
	taskID, err := strconv.Atoi(task.ClearCheckID)
	if err != nil {
		logs.Errorf("failed to convert clear check id %s to int, err: %v", task.ClearCheckID, err)

		msg := fmt.Sprintf("failed to convert clear check id %s to int, err: %v", task.ClearCheckID, err)
		errUpdate := r.updateTaskClearCheckStatus(task, "", msg, table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert clear check id %s to int, err: %v", task.ClearCheckID, err)
	}

	if err := r.checkClearJobStatus(taskID); err != nil {
		// if host ping death, go ahead to recycle
		if strings.Contains(err.Error(), "ping death") {
			logs.Infof("task %s ping death, skip clear check step", taskID)
		} else {
			logs.Infof("failed to clear check, job id: %d, err: %v", taskID, err)

			errUpdate := r.updateTaskClearCheckStatus(task, "", err.Error(), table.RecallStatusClearCheckFailed)
			if errUpdate != nil {
				logs.Warnf("failed to update recall task status, err: %v", errUpdate)
			}

			return fmt.Errorf("failed to clear check, job id: %d, err: %v", taskID, err)
		}
	}

	// update task status
	if err := r.updateTaskClearCheckStatus(task, "", "", table.RecallStatusReinstalling); err != nil {
		logs.Errorf("failed to update recall task status, err: %v", err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

// checkClearJobStatus check so clear check job status
func (r *Recycler) checkClearJobStatus(jobId int) error {
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

		if resp.Data.SubJobNum == 0 || len(resp.Data.SubJob) == 0 {
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
				msg = r.removeUnusedComment(resp.Data.SubJob[0].CommentMore)
			}
			return true, fmt.Errorf("check process not pass: %s", msg)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		return r.sojob.GetJobStatusDetail(nil, nil, jobId)
	}

	// timeout 20 minutes
	_, err := utils.Retry(doFunc, checkFunc, 1200, 30)
	if err != nil {
		return err
	}

	return nil
}

func (r *Recycler) removeUnusedComment(comment string) string {
	var msg []string
	for _, line := range strings.Split(comment, "\n") {
		if strings.Contains(line, "STATUS") && !strings.Contains(line, "_num:0") {
			index := strings.Index(line, "STATUS")
			msg = append(msg, line[index:])
		}
	}

	return strings.Join(msg, "\n")
}

func (r *Recycler) updateTaskClearCheckStatus(task *table.RecallDetail, id, msg string,
	status table.RecallStatus) error {

	filter := map[string]interface{}{
		"id": task.ID,
	}

	now := time.Now()
	update := map[string]interface{}{
		"status":    status,
		"update_at": now,
	}

	if id != "" {
		update["clear_check_id"] = id
		update["clear_check_link"] = sojobapi.TaskLinkPrefix + id
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
