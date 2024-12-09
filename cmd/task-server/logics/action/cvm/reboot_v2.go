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
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	actionflow "hcm/cmd/task-server/logics/flow"
	typecvm "hcm/pkg/adaptor/types/cvm"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/api/task-server/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// --------------------------[CVM开机]-----------------------------

var _ action.Action = new(RebootActionV2)
var _ action.ParameterAction = new(RebootActionV2)
var _ action.RollbackAction = new(RebootActionV2)

// RebootActionV2 cvm重启, 包含任务管理
type RebootActionV2 struct{}

// Name ...
func (c RebootActionV2) Name() enumor.ActionName {
	return enumor.ActionRebootCvmV2
}

// Run ...
func (c RebootActionV2) Run(kt run.ExecuteKit, params interface{}) (result interface{}, err error) {
	opt, ok := params.(*cvm.CvmOperationOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	if err = opt.Validate(); err != nil {
		return nil, err
	}

	asyncKit := kt.AsyncKit()

	// detail 状态检查
	detailList, err := actionflow.ListTaskDetail(asyncKit, opt.ManagementDetailIDs)
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
	err = actionflow.BatchUpdateTaskDetailState(asyncKit, opt.ManagementDetailIDs, enumor.TaskDetailRunning)
	if err != nil {
		return nil, fmt.Errorf("fail to update detail to running, err: %v", err)
	}

	err = c.rebootCvm(asyncKit, opt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c RebootActionV2) rebootCvm(kt *kit.Kit, opt *cvm.CvmOperationOption) error {

	switch opt.Vendor {
	case enumor.TCloud:
		err := c.rebootTCloudCvm(kt, opt)
		if err != nil {
			logs.Errorf("fail to start tcloud cvm, err: %v, req: %+v, rid: %s", err, opt, kt.Rid)
			return err
		}
	case enumor.TCloudZiyan:
		err := c.rebootTCloudZiyanCvm(kt, opt)
		if err != nil {
			logs.Errorf("fail to start tcloud ziyan cvm, err: %v, req: %+v, rid: %s", err, opt, kt.Rid)
			return err
		}
	default:
		return errf.New(errf.InvalidParameter, fmt.Sprintf("start cvm unsupported vendor: %s", opt.Vendor))
	}
	return nil
}

func (c RebootActionV2) rebootTCloudCvm(kt *kit.Kit, opt *cvm.CvmOperationOption) error {
	req := &hcprotocvm.TCloudBatchRebootReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
		StopType:  typecvm.SoftFirst,
	}
	executeErr := actcli.GetHCService().TCloud.Cvm.BatchRebootCvm(kt, req)
	if executeErr != nil {
		logs.Errorf("fail to call hc to start cvms, err: %v, req: %+v, rid: %s",
			executeErr, opt, kt.Rid)
		err := actionflow.BatchUpdateTaskDetailResultState(
			kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil, executeErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return err
	}

	// 更新任务状态为 success
	err := actionflow.BatchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailSuccess, nil, nil)
	if err != nil {
		logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
			err, kt.Rid)
		return err
	}
	return nil
}

// ParameterNew ...
func (c RebootActionV2) ParameterNew() (params interface{}) {
	return new(cvm.CvmOperationOption)
}

// Rollback 无需回滚
func (c RebootActionV2) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- RebootActionV2 Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
