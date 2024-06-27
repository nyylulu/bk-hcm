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
		logs.Errorf("failed to create initialize task, err: %v", err)

		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 根据IP获取主机信息
	hostInfo, err := r.esbCli.Cmdb().GetHostInfoByIP(r.kt.Ctx, r.kt.Header(), ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:recycler:ieod init, get host info by host id failed, ip: %s, err: %v", ip, err)
		return err
	}

	// 根据bkHostID去cmdb获取bkBizID
	bkBizID, err := r.esbCli.Cmdb().GetHostBizId(r.kt.Ctx, r.kt.Header(), hostInfo.BkHostId)
	if err != nil {
		logs.Errorf("sops:process:check:recycler:ieod init, get host info by host id failed, ip: %s, bkHostID: %d, "+
			"err: %v", ip, hostInfo.BkHostId, err)
		return err
	}

	// 创建标准运维-初始化任务
	taskID, jobUrl, err := sops.CreateInitSopsTask(r.kt, r.sops, ip, r.sopsOpt.DevnetIP, bkBizID, hostInfo.BkOsType)
	if err != nil {
		logs.Errorf("sops:process:check:recycler:ieod init, host %s failed to initialize, hostID: %d, bkBizID: %d, "+
			"err: %v", ip, task.HostID, bkBizID, err)

		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to initialize, hostID: %d, err: %v", ip, task.HostID, err)
	}

	// update task status
	if err = r.updateTaskInitializeStatus(task, strconv.FormatInt(taskID, 10), strconv.FormatInt(bkBizID, 10),
		jobUrl, "", table.RecallStatusInitializing); err != nil {
		logs.Errorf("failed to update recall task status, ip: %s, bkBizID: %d, taskID: %d, jobUrl: %s, hostID: %d, "+
			"err: %v", ip, bkBizID, taskID, jobUrl, task.HostID, err)
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
		logs.Errorf("failed to create initialize task, err: %v", err)

		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	taskID, err := strconv.Atoi(task.InitializeID)
	if err != nil {
		logs.Errorf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)

		msg := fmt.Sprintf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)
		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", msg, table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert initialize id %s to int, err: %v", task.InitializeID, err)
	}

	bkBizID, err := strconv.Atoi(task.InitializeBizID)
	if err != nil {
		logs.Errorf("sops:process:check:ieod init status, failed to convert conf initial biz id %s to int, err: %v",
			task.InitializeBizID, err)

		msg := fmt.Sprintf("failed to convert initial biz id %s to int, err: %v", task.InitializeBizID, err)
		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", msg, table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to convert initial biz id %s to int, err: %v", task.InitializeBizID, err)
	}

	if err = sops.CheckTaskStatus(r.kt, r.sops, int64(taskID), int64(bkBizID)); err != nil {
		logs.Infof("sops:process:check:matcher:ieod init device, host %s failed to initialize, job id: %d, "+
			"bkBizID: %d, err: %v", ip, taskID, bkBizID, err)

		errUpdate := r.updateTaskInitializeStatus(task, "", "", "", err.Error(), table.RecallStatusInitializeFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("host %s failed to initialize, job id: %d, bkBizID: %d, err: %v", ip, taskID, bkBizID, err)
	}

	// update task status
	if err = r.updateTaskInitializeStatus(task, "", "", "", "", table.RecallStatusDataDeleting); err != nil {
		logs.Errorf("recycler:failed to update recall task status, err: %v", err)

		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

func (r *Recycler) updateTaskInitializeStatus(task *table.RecallDetail, id, bkBizID, jobUrl, msg string,
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
		update["initialize_biz_id"] = bkBizID
		update["initialize_link"] = jobUrl
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
