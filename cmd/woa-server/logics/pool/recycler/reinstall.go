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
	"time"

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/thirdparty/xshipapi"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/logs"
)

// ReinstallStatus reinstall task status
type ReinstallStatus int

// ReinstallStatus ...
const (
	ReinstallStatusSuccess ReinstallStatus = 0
	ReinstallStatusRunning ReinstallStatus = 1
	ReinstallStatusFailed  ReinstallStatus = 2
)

func (r *Recycler) dealReinstallTask(task *table.RecallDetail) error {
	if task.ReinstallID == "" {
		return r.createReinstallTask(task)
	}

	return r.checkReinstallStatus(task)
}

func (r *Recycler) createReinstallTask(task *table.RecallDetail) error {
	resType, ok := task.Labels[table.ResourceTypeKey]
	if !ok {
		err := errors.New("get no resource type from task label")
		logs.Errorf("failed to create reinstall task, err: %v", err)

		errUpdate := r.updateTaskClearCheckStatus(task, "", err.Error(), table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	switch resType {
	case string(table.ResourceTypePm):
		return r.createPmReinstallTask(task)
	case string(table.ResourceTypeCvm):
		return r.createCvmReinstallTask(task)
	default:
		return fmt.Errorf("unsupported resource type %s, cannot reinstall", resType)
	}

	return nil
}

func (r *Recycler) checkReinstallStatus(task *table.RecallDetail) error {
	resType := task.Labels[table.ResourceTypeKey]
	switch resType {
	case string(table.ResourceTypePm):
		return r.checkPmReinstallStatus(task)
	case string(table.ResourceTypeCvm):
		return r.checkCvmReinstallStatus(task)
	default:
		return fmt.Errorf("unsupported resource type %s, cannot reinstall", resType)
	}

	return nil
}

func (r *Recycler) getPwd(hostID int64) (string, error) {
	// 1. get ip
	ip, err := r.getIpByHostID(hostID)
	if err != nil {
		logs.Errorf("failed to get host ip by id %d, err: %v", hostID, err)
		return "", err
	}

	pwd, err := r.tjj.GetPwd(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to get tjj pwd by ip %s, err: %v", ip, err)
		return "", err
	}

	return pwd, nil
}

func (r *Recycler) updateTaskReinstallStatus(task *table.RecallDetail, resType types.ResourceType, id, msg string,
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
		update["reinstall_id"] = id
		if resType == types.ResourceTypePm {
			update["reinstall_link"] = xshipapi.ReinstallLinkPrefix
		}
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
