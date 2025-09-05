/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package actioncvm

import (
	"context"
	"errors"
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/cmd/woa-server/dal/task/table"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/task-server/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
)

// --------------------------[CVM空闲检查]-----------------------------

var _ action.Action = new(MonitorIdleCheckAction)
var _ action.ParameterAction = new(MonitorIdleCheckAction)
var _ action.RollbackAction = new(MonitorIdleCheckAction)

// MonitorIdleCheckAction 监听一批CVM空闲检查任务的执行情况
type MonitorIdleCheckAction struct{}

// Name ...
func (c MonitorIdleCheckAction) Name() enumor.ActionName {
	return enumor.ActionMonitorIdleCheckCvm
}

// Run ...
func (c MonitorIdleCheckAction) Run(kt run.ExecuteKit, params interface{}) (result interface{}, taskErr error) {
	opt, ok := params.(*cvm.MonitorIdleCheckCvmOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	asyncKit := kt.AsyncKit()
	taskDetailIDs := maps.Values(opt.HostIDToTaskDetailID)

	// detail 状态检查
	detailList, err := actionflow.ListTaskDetail(asyncKit, taskDetailIDs)
	if err != nil {
		return fmt.Sprintf("task detail query failed"), err
	}

	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			// 任务被取消，跳过该批次
			return fmt.Sprintf("task detail %s canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}

	// 更新任务状态为 running
	err = actionflow.BatchUpdateTaskDetailState(asyncKit, taskDetailIDs, enumor.TaskDetailRunning)
	if err != nil {
		return nil, fmt.Errorf("fail to update detail to running, err: %v", err)
	}

	defer func() {
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if taskErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err = actionflow.BatchUpdateTaskDetailResultState(asyncKit, taskDetailIDs, targetState, nil, taskErr)
		if err != nil {
			logs.Errorf("fail to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, asyncKit.Rid)
		}
	}()

	err = c.monitorIdleCheckCvm(asyncKit, opt, len(detailList))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c MonitorIdleCheckAction) monitorIdleCheckCvm(asyncKit *kit.Kit, opt *cvm.MonitorIdleCheckCvmOption,
	idleCheckCvmNum int) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "opt parameter cannot be nil")
	}

	req := &types.GetRecycleDetectReq{
		SuborderID: []string{opt.SuborderID},
		Status:     []table.DetectStatus{table.DetectStatusSuccess, table.DetectStatusFailed},
		Page:       metadata.BasePage{Limit: pkg.BKMaxInstanceLimit},
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// 使用watchdog的taskTimeoutSec作为超时时间
	taskTimeoutSec := time.Duration(cc.TaskServer().Async.WatchDog.TaskTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(asyncKit.Ctx, taskTimeoutSec)
	defer cancel()

	// 记录已经更新过的任务，避免重复更新
	updatedDetailIDs := make(map[string]bool)

	for {
		select {
		case <-timeoutCtx.Done():
			logs.Errorf("monitor idle check timeout after %d seconds, rid: %s",
				cc.TaskServer().Async.WatchDog.TaskTimeoutSec, asyncKit.Rid)
			return errors.New(fmt.Sprintf("monitor idle check timeout after %d seconds",
				cc.TaskServer().Async.WatchDog.TaskTimeoutSec))
		case <-ticker.C:
			rst, err := actcli.GetWoaServer().Task.ListDetectTask(asyncKit, req)
			if err != nil {
				logs.Errorf("fail to list detect task, err: %v, suborder: %v, rid: %s",
					err, req.SuborderID, asyncKit.Rid)
				return fmt.Errorf("fail to list detect task, err: %v, suborder: %v, rid: %s",
					err, req.SuborderID, asyncKit.Rid)
			}
			// 没有处于终态的空闲检查主机
			if len(rst.Info) == 0 {
				continue
			}

			// 只更新新完成的任务
			detailIDToState := make(map[string]enumor.TaskDetailState)
			for _, task := range rst.Info {
				detailID, ok := opt.HostIDToTaskDetailID[task.HostID]
				if !ok {
					logs.Errorf("host id %s not found in host id to task detail id map", task.HostID)
					return fmt.Errorf("host id %s not found in host id to task detail id map", task.HostID)
				}

				// 跳过已经更新过的任务
				if updatedDetailIDs[detailID] {
					continue
				}

				targetState := enumor.TaskDetailSuccess
				if task.Status == table.DetectStatusFailed {
					targetState = enumor.TaskDetailFailed
				}
				detailIDToState[detailID] = targetState
				updatedDetailIDs[detailID] = true
			}

			// 只有新完成的任务才需要更新
			if len(detailIDToState) > 0 {
				err = actionflow.BatchUpdateTaskDetailStatesIndividually(asyncKit, detailIDToState)
				if err != nil {
					logs.Errorf("fail to update task detail status, err: %v, rid: %s", err, asyncKit.Rid)
					return err
				}
			}

			// 全部主机空闲检查结束（全都处于终态即SUCCESS和FAILED）
			if len(rst.Info) == idleCheckCvmNum {
				return nil
			}
		}
	}
}

// ParameterNew ...
func (c MonitorIdleCheckAction) ParameterNew() (params interface{}) {
	return new(cvm.MonitorIdleCheckCvmOption)
}

// Rollback 无需回滚
func (c MonitorIdleCheckAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- MonitorIdleCheckAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
