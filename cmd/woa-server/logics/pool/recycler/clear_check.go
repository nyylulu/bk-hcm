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

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/logics/task/sops"
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

		errUpdate := r.updateTaskClearCheckStatus(task, "", "", "", err.Error(), table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 根据HostID获取主机信息
	hostInfo, err := r.esbCli.Cmdb().GetHostInfoByHostID(r.kt.Ctx, r.kt.Header(), task.HostID)
	if err != nil {
		logs.Errorf("sops:process:check:idle check, get host info by host id failed, bkHostID: %d, err: %v",
			task.HostID, err)
		return err
	}

	// 根据bk_host_id，获取bk_biz_id
	bkBizID, err := r.esbCli.Cmdb().GetHostBizId(r.kt.Ctx, r.kt.Header(), hostInfo.BkHostId)
	if err != nil {
		logs.Errorf("sops:process:check:idle check process, get host biz id failed, ip: %s, bkHostId: %d, "+
			"err: %v", ip, hostInfo.BkHostId, err)
		return err
	}

	// 创建空闲检查任务
	taskID, jobUrl, err := sops.CreateIdleCheckSopsTask(r.kt, r.sops, ip, bkBizID, hostInfo.BkOsType)
	if err != nil {
		logs.Errorf("sops:process:check:idle check, host %s failed to create clear check task, err: %v", ip, err)

		errUpdate := r.updateTaskClearCheckStatus(task, "", "", "", err.Error(), table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to create clear check task, err: %v", err)
	}

	// update task status
	if err = r.updateTaskClearCheckStatus(task, strconv.FormatInt(taskID, 10), strconv.FormatInt(bkBizID, 10),
		jobUrl, "", table.RecallStatusClearChecking); err != nil {
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
		errUpdate := r.updateTaskClearCheckStatus(task, "", "", "", msg, table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert clear check id %s to int, err: %v", task.ClearCheckID, err)
	}

	// 获取业务ID
	bkBizID, err := strconv.Atoi(task.ClearCheckBizID)
	if err != nil {
		logs.Errorf("sops:process:check:clear check status, failed to convert clear check biz id %s to int, err: %v",
			task.ClearCheckBizID, err)

		msg := fmt.Sprintf("failed to convert clear check biz id %s to int, err: %v", task.ClearCheckBizID, err)
		errUpdate := r.updateTaskClearCheckStatus(task, "", "", "", msg, table.RecallStatusClearCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert clear check biz id %s to int, err: %v", task.ClearCheckBizID, err)
	}

	// 检查标准运维的任务状态
	if err = sops.CheckTaskStatus(r.kt, r.sops, int64(taskID), int64(bkBizID)); err != nil {
		// if host ping death, go ahead to recycle
		if strings.Contains(err.Error(), "ping death") {
			logs.Infof("task %s ping death, bkBizID: %d, skip clear check step", taskID, bkBizID)
		} else {
			logs.Infof("sops:process:check:clear check status, failed to clear check, job id: %d, bkBizID: %d, "+
				"err: %v", taskID, bkBizID, err)

			errUpdate := r.updateTaskClearCheckStatus(task, "", "", "", err.Error(), table.RecallStatusClearCheckFailed)
			if errUpdate != nil {
				logs.Warnf("failed to update recall task status, taskID: %d, bkBizID: %d, err: %v",
					taskID, bkBizID, errUpdate)
			}

			return fmt.Errorf("failed to clear check, job id: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		}
	}

	// update task status
	if err = r.updateTaskClearCheckStatus(task, "", "", "", "", table.RecallStatusReinstalling); err != nil {
		logs.Errorf("failed to update recall task status, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

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

func (r *Recycler) updateTaskClearCheckStatus(task *table.RecallDetail, id, bkBizID, jobUrl, msg string,
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
		update["clear_check_biz_id"] = bkBizID
		update["clear_check_link"] = jobUrl
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
