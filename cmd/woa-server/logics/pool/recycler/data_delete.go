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

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/pkg/logs"
)

func (r *Recycler) dealDataDeleteTask(task *table.RecallDetail) error {
	if task.DataDeleteID == "" {
		return r.createDataDeleteTask(task)
	}

	return r.checkDataDeleteStatus(task)
}

func (r *Recycler) createDataDeleteTask(task *table.RecallDetail) error {
	// create job
	ip, ok := task.Labels[table.IPKey]
	if !ok || ip == "" {
		err := errors.New("get no ip from task label")
		logs.Errorf("failed to create data delete task, err: %v", err)

		errUpdate := r.updateTaskDataDeleteStatus(task, "", "", "", err.Error(), table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 根据IP获取主机信息
	hostInfo, err := r.esbCli.Cmdb().GetHostInfoByIP(r.kt.Ctx, r.kt.Header(), ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:data clear, get host info by host id failed, ip: %s, err: %v", ip, err)
		return err
	}

	// 根据bk_host_id，获取bk_biz_id
	bkBizIDs, err := r.esbCli.Cmdb().GetHostBizIds(r.kt.Ctx, r.kt.Header(), []int64{hostInfo.BkHostId})
	if err != nil {
		logs.Errorf("sops:process:check:data clear, get host biz id failed, ip: %s, bkHostId: %d, err: %v",
			ip, hostInfo.BkHostId, err)
		return err
	}
	bkBizID, ok := bkBizIDs[hostInfo.BkHostId]
	if !ok {
		logs.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
		return fmt.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
	}

	// 创建数据清理任务-只有Linux任务
	taskID, jobUrl, err := sops.CreateDataClearSopsTask(r.kt, r.sops, ip, bkBizID, hostInfo.BkOsType)
	if err != nil {
		logs.Errorf("sops:process:check:data clear, host %s failed to data delete, bkBizID: %d, err: %v, task: %+v",
			ip, bkBizID, err, task)

		errUpdate := r.updateTaskDataDeleteStatus(task, "", "", "", err.Error(), table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, ip: %s, bkBizID: %d, err: %v", ip, bkBizID, errUpdate)
		}

		return fmt.Errorf("host %s failed to data delete, err: %v", ip, err)
	}

	// update task status
	if err = r.updateTaskDataDeleteStatus(task, strconv.FormatInt(taskID, 10), strconv.FormatInt(bkBizID, 10),
		jobUrl, "", table.RecallStatusDataDeleting); err != nil {
		logs.Errorf("failed to update recall task status, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		return err
	}

	go func() {
		// query every 5 minutes
		time.Sleep(time.Minute * 5)
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) checkDataDeleteStatus(task *table.RecallDetail) error {
	ip, ok := task.Labels[table.IPKey]
	if !ok {
		err := errors.New("get no ip from task label")
		logs.Errorf("failed to create data delete task, err: %v", err)

		errUpdate := r.updateTaskDataDeleteStatus(task, "", "", "", err.Error(), table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	taskID, err := strconv.Atoi(task.DataDeleteID)
	if err != nil {
		logs.Errorf("failed to convert data delete id %s to int, err: %v", task.DataDeleteID, err)

		msg := fmt.Sprintf("failed to convert data delete id %s to int, err: %v", task.DataDeleteID, err)
		errUpdate := r.updateTaskDataDeleteStatus(task, "", "", "", msg, table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert data delete id %s to int, err: %v", task.DataDeleteID, err)
	}

	// 获取业务ID
	bkBizID, err := strconv.Atoi(task.DataDeleteBizID)
	if err != nil {
		logs.Errorf("sops:process:check:data delete status, failed to convert data delete biz id %s to int, err: %v",
			task.DataDeleteBizID, err)

		msg := fmt.Sprintf("failed to convert data delete biz id %s to int, err: %v", task.DataDeleteBizID, err)
		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", msg, table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert data delete biz id %s to int, err: %v", task.DataDeleteBizID, err)
	}

	if err = sops.CheckTaskStatus(r.kt, r.sops, int64(taskID), int64(bkBizID)); err != nil {
		logs.Infof("host %s failed to data delete, job id: %d, bkBizID: %d, err: %v", ip, taskID, bkBizID, err)

		errUpdate := r.updateTaskDataDeleteStatus(task, "", "", "", err.Error(), table.RecallStatusDataDeleteFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to data delete, job id: %d, bkBizID: %d, err: %v", ip, taskID, bkBizID, err)
	}

	// update task status
	if err = r.updateTaskDataDeleteStatus(task, "", "", "", "", table.RecallStatusConfChecking); err != nil {
		logs.Errorf("failed to update recall task status, ip: %s, taskID: %d, bkBizID: %d, err: %v",
			ip, taskID, bkBizID, err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) updateTaskDataDeleteStatus(task *table.RecallDetail, id, bkBizID, jobUrl, msg string,
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
		update["data_delete_id"] = id
		update["data_delete_biz_id"] = bkBizID
		update["data_delete_link"] = jobUrl
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
