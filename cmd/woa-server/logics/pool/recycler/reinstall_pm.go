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
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/xshipapi"
)

func (r *Recycler) createPmReinstallTask(task *table.RecallDetail) error {
	// 1. get password
	pwd, err := r.getPwd(task.HostID)
	if err != nil {
		logs.Errorf("failed to get host %d pwd", task.HostID)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to get host %d pwd", task.HostID)
	}

	// 2. get os version
	osVersion := r.getOsType(task)

	// 3. create install order
	assetID, ok := task.Labels[table.AssetIDKey]
	if !ok || assetID == "" {
		logs.Errorf("get no asset id by host id %d", task.HostID)

		msg := fmt.Sprintf("get no asset id by host id %d", task.HostID)
		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", msg,
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("get no asset id by host id %d", task.HostID)
	}

	taskID, err := r.createXshipReinstallOrder(assetID, pwd, osVersion)
	if err != nil {
		logs.Errorf("failed to create xship reinstall order, err: %v", err)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 4. update task status
	if err := r.updateTaskReinstallStatus(task, types.ResourceTypePm, taskID, "",
		table.RecallStatusReinstalling); err != nil {
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

func (r *Recycler) createXshipReinstallOrder(assetID, pwd, osVersion string) (string, error) {
	req := &xshipapi.ReinstallReq{
		Assets: []*xshipapi.Asset{
			{
				AssetID: assetID,
				Variables: &xshipapi.Variables{
					OsVersion: osVersion,
					Password:  pwd,
				},
			},
		},
		Starter: xshipapi.DftStarter,
	}

	resp, err := r.xship.CreateReinstallTask(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to create xship reinstall task, err: %v", err)
		return "", err
	}

	if resp.Code != xshipapi.CodeSuccess {
		logs.Errorf("failed to create xship reinstall task, code: %s, msg: %s", resp.Code, resp.Message)
		return "", err
	}

	if resp.Data == nil {
		logs.Errorf("failed to create xship reinstall task, for return data is nil")
		return "", errors.New("failed to create xship reinstall task, for return data is nil")
	}

	orderNum := len(resp.Data.AcceptOrders)
	if orderNum != 1 {
		logs.Errorf("failed to create xship reinstall task, for return order num %d != 1", orderNum)
		return "", fmt.Errorf("failed to create xship reinstall task, for return order num %d != 1", orderNum)
	}

	order := resp.Data.AcceptOrders[0]

	if order == nil {
		logs.Errorf("failed to create xship reinstall task, for return order is nil")
		return "", errors.New("failed to create xship reinstall task, for return order is nil")
	}

	if order.AcceptStatus != 0 {
		logs.Errorf("failed to create xship reinstall task, accept status %d, msg: %s", order.AcceptStatus,
			order.AcceptMsg)
		return "", fmt.Errorf("failed to create xship reinstall task, accept status %d, msg: %s", order.AcceptStatus,
			order.AcceptMsg)
	}

	return order.OrderID, nil
}

func (r *Recycler) checkPmReinstallStatus(task *table.RecallDetail) error {
	resp, err := r.xship.GetReinstallTaskStatus(nil, nil, task.ReinstallID)
	if err != nil {
		logs.Errorf("failed to get xship reinstall task status, err: %v", err)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", err)
		}

		return err
	}

	status, err := r.parsePmReinstallRst(resp)
	switch status {
	case ReinstallStatusSuccess:
		{
			err := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", "", table.RecallStatusInitializing)
			if err != nil {
				logs.Warnf("failed to update recall task status, err: %v", err)
				return err
			}

			go func() {
				r.Add(task.ID)
			}()
		}
	case ReinstallStatusFailed:
		{
			errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypePm, "", err.Error(),
				table.RecallStatusReinstallFailed)
			if err != nil {
				logs.Warnf("failed to update recall task status, err: %v", errUpdate)
			}

			return err
		}
	case ReinstallStatusRunning:
		{
			go func() {
				// query every 5 minutes
				time.Sleep(time.Minute * 5)
				r.Add(task.ID)
			}()
		}
	default:
		{
			logs.Warnf("unknown reinstall status %d", status)
		}
	}

	return nil
}

func (r *Recycler) parsePmReinstallRst(resp *xshipapi.ReinstallStatusResp) (ReinstallStatus, error) {
	if resp.Code != xshipapi.CodeSuccess {
		err := fmt.Errorf("failed to get xship reinstall task status, code: %s, msg: %s", resp.Code, resp.Message)
		logs.Errorf("failed to get xship reinstall task status, code: %s, msg: %s", resp.Code, resp.Message)
		return ReinstallStatusFailed, err
	}

	if resp.Data == nil {
		err := errors.New("xship reinstall task status return data is nil")
		logs.Errorf("xship reinstall task status return data is nil")

		return ReinstallStatusFailed, err
	}

	orderNum := len(resp.Data.ReinstallInfos)
	if orderNum != 1 {
		err := fmt.Errorf("xship reinstall task status return order num %d != 1", orderNum)
		logs.Errorf("xship reinstall task status return order num %d != 1", orderNum)

		return ReinstallStatusFailed, err
	}

	order := resp.Data.ReinstallInfos[0]
	if order == nil {
		err := errors.New("xship reinstall task status return order is nil")
		logs.Errorf("xship reinstall task status return order is nil")

		return ReinstallStatusFailed, err
	}

	switch order.Status {
	case xshipapi.AcceptStatusDone:
		{
			logs.Infof("reinstall order %s is done", order.OrderID)
			return ReinstallStatusSuccess, nil
		}
	case xshipapi.AcceptStatusRejected, xshipapi.AcceptStatusExpired:
		{
			err := fmt.Errorf("reinstall order %s failed, status: %d, err: %s", order.OrderID, order.Status,
				order.ErrMsg)
			logs.Errorf("reinstall order %s failed, status: %s, err: %s", order.OrderID, order.Status, order.ErrMsg)

			return ReinstallStatusFailed, err
		}
	default:
		{
			logs.Infof("reinstall order %s handling, status: %s", order.OrderID, order.Status)
			return ReinstallStatusRunning, nil
		}
	}
}

func (r *Recycler) getOsType(task *table.RecallDetail) string {
	osVersion := "Tencent tlinux release 2.6 (tkernel4)"

	filter := &mapstr.MapStr{
		"id": task.RecallID,
	}

	recallOrder, err := dao.Set().RecallOrder().GetRecallOrder(context.Background(), filter)
	if err != nil {
		logs.Warnf("failed to get recall order by id: %d", task.RecallID)
		return osVersion
	}

	if recallOrder == nil || recallOrder.RecyclePolicy == nil {
		logs.Warnf("get invalid nil recall order or recycle policy by id: %d", task.RecallID)
		return osVersion
	}

	if recallOrder.RecyclePolicy.OsType == "" {
		logs.Warnf("get invalid empty os type by id: %d", task.RecallID)
		return osVersion
	}

	return recallOrder.RecyclePolicy.OsType
}
