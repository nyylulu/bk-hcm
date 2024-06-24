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

// Package recycler ...
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

func (r *Recycler) dealConfCheckTask(task *table.RecallDetail) error {
	if task.ConfCheckID == "" {
		return r.createConfCheckTask(task)
	}

	return r.checkConfCheckStatus(task)
}

func (r *Recycler) createConfCheckTask(task *table.RecallDetail) error {
	// create job
	ip, ok := task.Labels[table.IPKey]
	if !ok || ip == "" {
		err := errors.New("get no ip from task label")
		logs.Errorf("failed to create conf check task, err: %v", err)

		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", err.Error(), table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 根据IP获取主机信息
	hostInfo, err := r.esbCli.Cmdb().GetHostInfoByIP(r.kt.Ctx, r.kt.Header(), ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:recycler:config check, get host info by ip failed, ip: %s, hostID: %d, "+
			"err: %v", ip, task.HostID, err)
		return err
	}

	// 根据bkHostID去cmdb获取bkBizID
	bkBizIDs, err := r.esbCli.Cmdb().GetHostBizIds(r.kt.Ctx, r.kt.Header(), []int64{hostInfo.BkHostId})
	if err != nil {
		logs.Errorf("sops:process:check:matcher:ieod init, get host info by host id failed, ip: %s, taskHostID: %d, "+
			"bkHostID: %d, err: %v", ip, task.HostID, hostInfo.BkHostId, err)
		return err
	}
	bkBizID, ok := bkBizIDs[hostInfo.BkHostId]
	if !ok {
		logs.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
		return fmt.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
	}

	// 创建配置检查任务-只有Linux任务
	taskID, jobUrl, err := sops.CreateConfigCheckSopsTask(r.kt, r.sops, r.esbCli.Cmdb(), ip, bkBizID)
	if err != nil {
		logs.Errorf("sops:process:check:config check, host %s failed to conf check, bkBizID: %d, err: %v",
			ip, bkBizID, err)

		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", err.Error(), table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, bkBizID: %d, err: %v", bkBizID, errUpdate)
		}

		return fmt.Errorf("host %s failed to conf check, err: %v", ip, err)
	}

	// update task status
	if err = r.updateTaskConfCheckStatus(task, strconv.FormatInt(taskID, 10), strconv.FormatInt(bkBizID, 10),
		jobUrl, "", table.RecallStatusConfChecking); err != nil {
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

func (r *Recycler) checkConfCheckStatus(task *table.RecallDetail) error {
	ip, ok := task.Labels[table.IPKey]
	if !ok {
		err := errors.New("get no ip from task label")
		logs.Errorf("failed to create conf check task, err: %v", err)

		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", err.Error(), table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	taskID, err := strconv.Atoi(task.ConfCheckID)
	if err != nil {
		logs.Errorf("sops:process:check:config check status, failed to convert conf check id %s to int, err: %v",
			task.ConfCheckID, err)

		msg := fmt.Sprintf("failed to convert conf check id %s to int, err: %v", task.ConfCheckID, err)
		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", msg, table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert conf check id %s to int, err: %v", task.ConfCheckID, err)
	}

	// 获取业务ID
	bkBizID, err := strconv.Atoi(task.ConfCheckBizID)
	if err != nil {
		logs.Infof("sops:process:check:config check status, failed to convert conf check biz id %s to int, err: %v",
			task.ConfCheckBizID, err)

		msg := fmt.Sprintf("failed to convert conf check biz id %s to int, err: %v", task.ConfCheckBizID, err)
		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", msg, table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert conf check biz id %s to int, err: %v", task.ConfCheckBizID, err)
	}

	if err = sops.CheckTaskStatus(r.kt, r.sops, int64(taskID), int64(bkBizID)); err != nil {
		logs.Errorf("sops:process:check:config check status, host %s failed to conf check, job id: %d, bkBizID: %d, "+
			"err: %v", ip, taskID, bkBizID, err)

		errUpdate := r.updateTaskConfCheckStatus(task, "", "", "", err.Error(), table.RecallStatusConfCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to conf check, job id: %d, bkBizID: %d, err: %v", ip, taskID, bkBizID, err)
	}

	// update task status
	if err = r.updateTaskConfCheckStatus(task, "", "", "", "", table.RecallStatusTransiting); err != nil {
		logs.Errorf("failed to update recall task status, err: %v", err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) updateTaskConfCheckStatus(task *table.RecallDetail, id, bkBizID, jobUrl, msg string,
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
		update["conf_check_id"] = id
		update["conf_check_biz_id"] = bkBizID
		update["conf_check_link"] = jobUrl
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
