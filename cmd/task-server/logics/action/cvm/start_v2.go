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
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// --------------------------[CVM开机]-----------------------------

var _ action.Action = new(StartActionV2)
var _ action.ParameterAction = new(StartActionV2)

// StartActionV2 cvm开机, 包含任务管理
type StartActionV2 struct{}

// Name ...
func (c StartActionV2) Name() enumor.ActionName {
	return enumor.ActionStartCvmV2
}

// Run ...
func (c StartActionV2) Run(kt run.ExecuteKit, params interface{}) (result interface{}, err error) {
	opt, ok := params.(*CvmOperationOptionV2)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	if err = opt.Validate(); err != nil {
		return nil, err
	}

	asyncKit := kt.AsyncKit()

	// detail 状态检查
	detailList, err := listTaskDetail(asyncKit, opt.ManagementDetailIDs)
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
	if err := batchUpdateTaskDetailState(asyncKit, opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("fail to update detail to running, err: %v", err)
	}

	err = c.startCvm(asyncKit, opt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c StartActionV2) startCvm(kt *kit.Kit, opt *CvmOperationOptionV2) error {

	switch opt.Vendor {
	case enumor.TCloud:
		err := c.startTCloudCvm(kt, opt)
		if err != nil {
			logs.Errorf("fail to start tcloud cvm, err: %v, req: %+v, rid: %s", err, opt, kt.Rid)
			return err
		}
	case enumor.TCloudZiyan:
		err := c.startTCloudZiyanCvm(kt, opt)
		if err != nil {
			logs.Errorf("fail to start tcloud ziyan cvm, err: %v, req: %+v, rid: %s", err, opt, kt.Rid)
			return err
		}
	default:
		return errf.New(errf.InvalidParameter, fmt.Sprintf("start cvm unsupported vendor: %s", opt.Vendor))
	}
	return nil
}

func (c StartActionV2) startTCloudCvm(kt *kit.Kit, opt *CvmOperationOptionV2) error {

	req := &hcprotocvm.TCloudBatchStartReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
	}
	executeErr := actcli.GetHCService().TCloud.Cvm.BatchStartCvm(kt, req)
	if executeErr != nil {
		logs.Errorf("fail to call hc to start cvms, err: %v, req: %+v, rid: %s",
			executeErr, opt, kt.Rid)
		err := batchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil, executeErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return err
	}

	// 更新任务状态为 success
	err := batchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailSuccess, nil, nil)
	if err != nil {
		logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
			err, kt.Rid)
		return err
	}
	return nil
}

// ParameterNew ...
func (c StartActionV2) ParameterNew() (params interface{}) {
	return new(CvmOperationOptionV2)
}

// CvmOperationOptionV2 operation cvm option.
type CvmOperationOptionV2 struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"omitempty"`
	// IDs TCloud/HuaWei/Aws 支持批量操作，Azure/Gcp 仅支持单个操作
	IDs                 []string `json:"ids" validate:"required,min=1,max=100"`
	ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1,max=100"`
}

// Validate operation cvm option.
func (opt CvmOperationOptionV2) Validate() error {

	switch opt.Vendor {
	case enumor.TCloud, enumor.TCloudZiyan:
		if len(opt.Region) == 0 {
			return fmt.Errorf("vendor: %s region is required", opt.Vendor)
		}
	default:
		return fmt.Errorf("cvm operation option unsupported vendor: %s", opt.Vendor)
	}
	if len(opt.ManagementDetailIDs) != len(opt.IDs) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and IDs length not match: %d! = %d",
			len(opt.ManagementDetailIDs), len(opt.IDs))
	}

	return validator.Validate.Struct(opt)
}
