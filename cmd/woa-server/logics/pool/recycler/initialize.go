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
	"time"

	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
)

func (r *Recycler) dealInitializeTask(task *table.RecallDetail) error {
	if task.InitializeID == "" {
		return r.createInitializeTask(task)
	}

	return r.checkInitializeStatus(task)
}

func (r *Recycler) createInitializeTask(task *table.RecallDetail) error {
	// create job
	ip, ok := task.Labels[table.IPKey]
	if !ok || ip == "" {
		err := errors.New("get no ip from task label")
		blog.Errorf("failed to create initialize task, err: %v", err)

		errUpdate := r.updateTaskInitializeStatus(task, "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			blog.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	taskID, err := r.createSoJob("ieod_init", []string{ip})
	if err != nil {
		blog.Errorf("host %s failed to initialize, err: %v", ip, err)

		errUpdate := r.updateTaskInitializeStatus(task, "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			blog.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to initialize, err: %v", ip, err)
	}

	// update task status
	if err := r.updateTaskInitializeStatus(task, strconv.Itoa(taskID), "", table.RecallStatusInitializing); err != nil {
		blog.Errorf("failed to update recall task status, err: %v", err)
		return err
	}

	go func() {
		// query every 5 minutes
		time.Sleep(time.Minute * 5)
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) checkInitializeStatus(task *table.RecallDetail) error {
	ip, ok := task.Labels[table.IPKey]
	if !ok {
		err := errors.New("get no ip from task label")
		blog.Errorf("failed to create initialize task, err: %v", err)

		errUpdate := r.updateTaskInitializeStatus(task, "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			blog.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	taskID, err := strconv.Atoi(task.InitializeID)
	if err != nil {
		blog.Errorf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)

		msg := fmt.Sprintf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)
		errUpdate := r.updateTaskInitializeStatus(task, "", msg, table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			blog.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)
	}

	if err := r.checkJobStatus(taskID); err != nil {
		blog.Infof("host %s failed to initialize, job id: %d, err: %v", ip, taskID, err)

		errUpdate := r.updateTaskInitializeStatus(task, "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			blog.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to initialize, job id: %d, err: %v", ip, taskID, err)
	}

	// update task status
	if err := r.updateTaskInitializeStatus(task, "", "", table.RecallStatusDataDeleting); err != nil {
		blog.Errorf("failed to update recall task status, err: %v", err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) updateTaskInitializeStatus(task *table.RecallDetail, id, msg string,
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
		update["initialize_id"] = id
		update["initialize_link"] = sojobapi.TaskLinkPrefix + id
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
