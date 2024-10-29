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
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

func (r *Recycler) createCvmReinstallTask(task *table.RecallDetail) error {
	// 1. get cvm info
	cvmInfo, err := r.getCvmInfo(task)
	if err != nil {
		logs.Errorf("failed to get cvm info, err: %v", err)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 2. get password
	pwd, err := r.getPwd(task.HostID)
	if err != nil {
		logs.Errorf("failed to get host %d pwd", task.HostID)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return fmt.Errorf("failed to get host %d pwd", task.HostID)
	}

	// 3. get image id
	imageID := r.getImageID(task)

	// 4. create reinstall task
	taskID, err := r.createCvmReinstallOrder(cvmInfo, pwd, imageID)
	if err != nil {
		logs.Errorf("failed to create cvm reinstall order, err: %v", err)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// 5. update task status
	if err := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, taskID, "",
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

func (r *Recycler) checkCvmReinstallStatus(task *table.RecallDetail) error {
	// get cvm info
	cvmInfo, err := r.getCvmInfo(task)
	if err != nil {
		logs.Errorf("failed to get cvm info, err: %v", err)

		errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, "", err.Error(),
			table.RecallStatusReinstallFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}
	}

	id, key := r.tcOpt.Credential.ID, r.tcOpt.Credential.Key
	credential := common.NewCredential(id, key)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = r.tcOpt.Endpoints.Cvm
	cpf.HttpProfile.ReqTimeout = 30
	client, _ := cvm.NewClient(credential, cvmInfo.CloudRegion, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewDescribeInstancesRequest()
	request.InstanceIds = common.StringPtrs([]string{cvmInfo.InstanceId})

	// 返回的resp是一个DescribeInstancesResponse的实例，与请求对象对应
	resp, err := client.DescribeInstances(request)
	if err != nil {
		logs.Errorf("failed to describe cvm instance, err: %v", err)
		return err
	}

	status, err := r.parseCvmReinstallRst(resp, task)
	switch status {
	case ReinstallStatusSuccess:
		{
			err := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm,
				"", "", table.RecallStatusInitializing)
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
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			errUpdate := r.updateTaskReinstallStatus(task, types.ResourceTypeCvm, "", msg,
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

func (r *Recycler) getCvmInfo(task *table.RecallDetail) (*cvmapi.InstanceItem, error) {
	// create job
	ip, ok := task.Labels[table.IPKey]
	if !ok || ip == "" {
		return nil, errors.New("get no ip from task label")
	}

	req := &cvmapi.InstanceQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmInstanceStatusMethod,
		},
		Params: &cvmapi.InstanceQueryParam{
			LanIp: []string{ip},
		},
	}

	resp, err := r.cvm.QueryCvmInstances(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to query cvm instance, err: %v", err)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to query cvm instance, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		return nil, fmt.Errorf("failed to query cvm instance, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to query cvm instance, for result is nil")
		return nil, errors.New("failed to query cvm instance, for result is nil")
	}

	num := len(resp.Result.Data)
	if num != 1 {
		logs.Errorf("failed to query cvm instance, for data num %d != 1", num)
		return nil, fmt.Errorf("failed to query cvm instance, for data num %d != 1", num)
	}

	inst := resp.Result.Data[0]

	return inst, nil
}

func (r *Recycler) createCvmReinstallOrder(inst *cvmapi.InstanceItem, pwd, imageID string) (string, error) {
	id, key := r.tcOpt.Credential.ID, r.tcOpt.Credential.Key
	credential := common.NewCredential(id, key)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = r.tcOpt.Endpoints.Cvm
	cpf.HttpProfile.ReqTimeout = 30
	client, _ := cvm.NewClient(credential, inst.CloudRegion, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewResetInstanceRequest()
	request.InstanceId = common.StringPtr(inst.InstanceId)
	request.ImageId = common.StringPtr(imageID)
	request.LoginSettings = &cvm.LoginSettings{
		Password: common.StringPtr(pwd),
	}

	// 返回的resp是一个ResetInstanceResponse的实例，与请求对象对应
	resp, err := client.ResetInstance(request)
	if err != nil {
		logs.Errorf("failed to reset cvm, err: %v", err)
		return "", err
	}

	if resp.Response == nil {
		err := errors.New("failed to reset cvm, for response is nil")
		logs.Errorf("failed to reset cvm, for response is nil")

		return "", err
	}

	return *resp.Response.RequestId, nil
}

func (r *Recycler) parseCvmReinstallRst(resp *cvm.DescribeInstancesResponse, task *table.RecallDetail) (ReinstallStatus,
	error) {

	if resp.Response == nil {
		err := errors.New("failed to describe cvm instance, for response is nil")
		logs.Errorf("failed to describe cvm instance, for response is nil")

		return ReinstallStatusFailed, err
	}

	num := len(resp.Response.InstanceSet)
	if num != 1 {
		err := fmt.Errorf("failed to describe cvm instance, for response instance num %d != 1", num)
		logs.Errorf("failed to describe cvm instance, for response instance num %d != 1", num)

		return ReinstallStatusFailed, err
	}

	instance := resp.Response.InstanceSet[0]
	if instance == nil {
		err := errors.New("failed to describe cvm instance, for response instance is nil")
		logs.Errorf("failed to describe cvm instance, for response instance is nil")

		return ReinstallStatusFailed, err
	}

	if *instance.LatestOperation != "ResetInstance" {
		err := errors.New("failed to check reinstall status, for latest operation is not reset instance")
		logs.Errorf("failed to check reinstall status, for latest operation is not reset instance")

		return ReinstallStatusFailed, err
	}

	if *instance.LatestOperationRequestId != task.ReinstallID {
		err := fmt.Errorf("failed to check reinstall status, for latest operation request id %s != %s",
			*instance.LatestOperationRequestId, task.ReinstallID)
		logs.Errorf("failed to check reinstall status, for latest operation request id %s != %s",
			*instance.LatestOperationRequestId, task.ReinstallID)

		return ReinstallStatusFailed, err
	}

	switch *instance.LatestOperationState {
	case "SUCCESS":
		{
			logs.Infof("reinstall order %s is done", task.ReinstallID)
			return ReinstallStatusSuccess, nil
		}
	case "FAILED":
		{
			err := fmt.Errorf("reinstall order %s failed, status: %s", task.ReinstallID, *instance.LatestOperationState)
			logs.Errorf("reinstall order %s failed, status: %s", task.ReinstallID, *instance.LatestOperationState)

			return ReinstallStatusFailed, err
		}
	default:
		{
			logs.Infof("reinstall order %s handling, status: %s", task.ReinstallID, *instance.LatestOperationState)
			return ReinstallStatusRunning, nil
		}
	}
}

func (r *Recycler) getImageID(task *table.RecallDetail) string {
	imageID := cvmapi.DftImageID

	filter := &mapstr.MapStr{
		"id": task.RecallID,
	}

	recallOrder, err := dao.Set().RecallOrder().GetRecallOrder(context.Background(), filter)
	if err != nil {
		logs.Warnf("failed to get recall order by id: %d", task.RecallID)
		return imageID
	}

	if recallOrder == nil || recallOrder.RecyclePolicy == nil {
		logs.Warnf("get invalid nil recall order or recycle policy by id: %d", task.RecallID)
		return imageID
	}

	if recallOrder.RecyclePolicy.ImageID == "" {
		logs.Warnf("get invalid empty image id by id: %d", task.RecallID)
		return imageID
	}

	return recallOrder.RecyclePolicy.ImageID
}
