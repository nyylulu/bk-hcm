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
	actcli "hcm/cmd/task-server/logics/action/cli"
	typecvm "hcm/pkg/adaptor/types/cvm"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (c StartActionV2) startTCloudZiyanCvm(kt *kit.Kit, opt *CvmOperationOptionV2) error {

	req := &hcprotocvm.TCloudBatchStartReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchStartCvm(kt, req)
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

func (c StopActionV2) stopTCloudZiyanCvm(kt *kit.Kit, opt *CvmOperationOptionV2) error {

	req := &hcprotocvm.TCloudBatchStopReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchStopCvm(kt, req)
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

func (c RebootActionV2) rebootTCloudZiyanCvm(kt *kit.Kit, opt *CvmOperationOptionV2) error {

	req := &hcprotocvm.TCloudBatchRebootReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
		StopType:  typecvm.SoftFirst,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchRebootCvm(kt, req)
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
