/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package sops provides ...
package sops

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"hcm/cmd/woa-server/common/utils"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/thirdparty/sopsapi"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CreateInitSopsTask 创建标准运维-初始化任务
func CreateInitSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, ip, devnetIP string, bkBizID int64,
	bkOsType cmdb.OsType) (int64, string, error) {

	// 操作系统不是Linux、Windows的话，不处理
	if bkOsType != cmdb.LinuxOsType && bkOsType != cmdb.WindowsOsType {
		logs.Warnf("sops:process:check:ieod init, host:%s bkOsType is not Linux or Windows, bkBizID: %d, bkOsType: %s",
			ip, bkBizID, bkOsType)
		return 0, "", nil
	}

	// 操作系统类型(Linux:1 Windows:2)
	templateID := sopsapi.InitLinuxTemplateID
	taskName := sopsapi.InitLinuxTaskNamePrefix + "ieod_init"
	if bkOsType == cmdb.WindowsOsType {
		templateID = sopsapi.InitWindowsTemplateID
		taskName = sopsapi.InitWindowsTaskNamePrefix + "ieod_init"
	}

	params := map[string]interface{}{
		"${biz_cc_id}":   bkBizID,
		"${bk_biz_id}":   bkBizID,
		"${job_ip_list}": ip,
	}
	if bkOsType == cmdb.LinuxOsType {
		params["${devnetdls}"] = devnetIP
	}

	jobId, jobUrl, err := createSopsTask(kt, sopsCli, templateID, taskName, bkBizID, params)

	return jobId, jobUrl, err
}

// CreateIdleCheckSopsTask 创建空闲检查任务
func CreateIdleCheckSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, ip string,
	bkBizID int64, bkOsType cmdb.OsType) (int64, string, error) {

	// 操作系统不是Linux、Windows的话，不处理
	if bkOsType != cmdb.LinuxOsType && bkOsType != cmdb.WindowsOsType {
		logs.Warnf("sops:process:check:idle check process, host:%s bkOsType is not Linux or Windows, bkOsType: %s",
			ip, bkOsType)
		return 0, "", nil
	}

	// 操作系统类型(Linux:1 Windows:2)
	templateID := sopsapi.IdleCheckLinux
	taskName := sopsapi.IdleCheckLinuxTaskNamePrefix + "isclear"
	if bkOsType == cmdb.WindowsOsType {
		templateID = sopsapi.IdleCheckWindows
		taskName = sopsapi.IdleCheckWindowsTaskNamePrefix + "isclear"
	}

	params := map[string]interface{}{
		"${biz_cc_id}":   bkBizID,
		"${bk_biz_id}":   bkBizID,
		"${job_ip_list}": ip,
	}

	jobId, jobUrl, err := createSopsTask(kt, sopsCli, templateID, taskName, bkBizID, params)

	return jobId, jobUrl, err
}

// CreateConfigCheckSopsTask 创建配置检查任务-只有Linux任务
func CreateConfigCheckSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, ccCli cmdb.Client, ip string,
	bkBizID int64) (int64, string, error) {

	hostInfo, err := ccCli.GetHostInfoByIP(kt.Ctx, kt.Header(), ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:config check, get host info by host id failed, ip: %s, err: %v", ip, err)
		return 0, "", err
	}

	// 操作系统不是Linux的话，不处理
	if hostInfo.BkOsType != cmdb.LinuxOsType {
		logs.Warnf("sops:process:check:config check, host:%s bkOsType is not Linux, bkOsType: %s, "+
			"hostInfo: %+v", ip, hostInfo.BkOsType, hostInfo)
		return 0, "", nil
	}

	// 操作系统类型(Linux:1 Windows:2)
	templateID := sopsapi.ConfigCheckLinux
	taskName := sopsapi.ConfigCheckLinuxTaskNamePrefix + "confcheck"

	params := map[string]interface{}{
		"${biz_cc_id}":   bkBizID,
		"${bk_biz_id}":   bkBizID,
		"${job_ip_list}": ip,
	}

	jobId, jobUrl, err := createSopsTask(kt, sopsCli, templateID, taskName, bkBizID, params)

	return jobId, jobUrl, err
}

// CreateDataClearSopsTask 创建数据清理任务-只有Linux任务
func CreateDataClearSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, ip string, bkBizID int64,
	bkOsType cmdb.OsType) (int64, string, error) {

	// 操作系统不是Linux的话，不处理
	if bkOsType != cmdb.LinuxOsType {
		logs.Warnf("sops:process:check:data clear, host:%s bkOsType is not Linux, bkOsType: %s, rid: %s",
			ip, bkOsType, kt.Rid)
		return 0, "", nil
	}

	// 操作系统类型(Linux:1 Windows:2)
	templateID := sopsapi.DataClearLinux
	taskName := sopsapi.DataClearLinuxTaskNamePrefix + "delete_data"

	params := map[string]interface{}{
		"${biz_cc_id}":   bkBizID,
		"${bk_biz_id}":   bkBizID,
		"${job_ip_list}": ip,
	}

	jobId, jobUrl, err := createSopsTask(kt, sopsCli, templateID, taskName, bkBizID, params)

	return jobId, jobUrl, err
}

// createSopsTask create a sops task
func createSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, templateID int64, taskName string,
	bkBizID int64, constants map[string]interface{}) (int64, string, error) {

	currentTime := time.Now().Format("20060102150405")
	createReq := &sopsapi.CreateTaskReq{
		TemplateSource: sopsapi.CommonTemplateSource,
		Name:           fmt.Sprintf("%s-%s", taskName, currentTime),
		Constants:      constants,
	}

	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("sops:task:create:failed, bkBizID: %d, templateID: %d, taskName: %s, err: %v",
				bkBizID, templateID, createReq.Name, err)
		}
		if obj == nil {
			return false, fmt.Errorf("sops:task:create:failed, bkBizID: %d, templateID: %d, taskName: %s, resp is nil",
				bkBizID, templateID, createReq.Name)
		}
		resp, ok := obj.(*sopsapi.CreateTaskResp)
		if !ok {
			return false, fmt.Errorf("sops:task:create:failed, object is not a create sops task response: %+v", resp)
		}

		if !resp.Result {
			return false, fmt.Errorf("sops:task:create:failed, create sops task failed, code: %d, err: %s",
				resp.Code, resp.Message)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// 增加限频检查
		limiter := getSopsLimiter(createTask, writeLimit)
		if !limiter.Allow() {
			return nil, errors.New("exceeded Create API frequency limit times")
		}
		return sopsCli.CreateTask(kt.Ctx, kt.Header(), templateID, bkBizID, createReq)
	}

	rand.Seed(time.Now().UnixNano())
	createResp, err := utils.Retry(doFunc, checkFunc, 3600, uint64(rand.Intn(4)+1))
	if err != nil {
		logs.Errorf("create bksops task retry failed, bkBizID: %d, createReq: %+v, err: %v", bkBizID, createReq, err)
		return 0, "", err
	}

	resp, ok := createResp.(*sopsapi.CreateTaskResp)
	if !ok {
		return 0, "", fmt.Errorf("object is not a create sops task failed, bkBizID: %d, response: %+v",
			bkBizID, createResp)
	}

	taskId := resp.Data.TaskId
	if taskId <= 0 {
		return 0, "", fmt.Errorf("create sops task failed, bkBizID: %d, for response data invalid: %+v", bkBizID, resp)
	}

	err = startSopsTask(kt, sopsCli, taskId, bkBizID)
	if err != nil {
		return 0, "", fmt.Errorf("create sops task success, but start task failed, taskId: %d, bkBizID: %d, "+
			"for response data invalid: %+v", taskId, bkBizID, err)
	}

	return taskId, resp.Data.TaskUrl, nil
}

// startSopsTask start a sops task
func startSopsTask(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, taskID, bkBizID int64) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("sops:task:start:failed, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		}
		if obj == nil {
			return false, fmt.Errorf("sops:task:start:failed, taskID: %d, bkBizID: %d, resp is nil", taskID, bkBizID)
		}
		resp, ok := obj.(*sopsapi.StartTaskResp)
		if !ok {
			return true, fmt.Errorf("sops:task:start:failed, object is not a start sops task response: %+v", resp)
		}

		if !resp.Result {
			return true, fmt.Errorf("sops:task:start:failed, start sops task failed, code: %d, err: %s",
				resp.Code, resp.Message)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// 增加限频检查
		limiter := getSopsLimiter(startTask, writeLimit)
		if !limiter.Allow() {
			return nil, errors.New("exceeded Create API frequency limit times")
		}
		return sopsCli.StartTask(kt.Ctx, kt.Header(), taskID, bkBizID)
	}

	rand.Seed(time.Now().UnixNano())
	startResp, err := utils.Retry(doFunc, checkFunc, 3600, uint64(rand.Intn(4)+1))
	if err != nil {
		logs.Errorf("start bksops task retry failed, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		return err
	}

	_, ok := startResp.(*sopsapi.StartTaskResp)
	if !ok {
		return fmt.Errorf("object is not a start sops task response: %+v", startResp)
	}

	return nil
}

// CheckTaskStatus check sops task status
func CheckTaskStatus(kt *kit.Kit, sopsCli sopsapi.SopsClientInterface, taskID, bkBizID int64) error {
	// 如果该任务不是Linux或Windows任务，则不会创建标准运维任务，就不用去标准运维查询任务状态
	if taskID <= 0 && bkBizID <= 0 {
		logs.Infof("sops:process:check task status empty")
		return nil
	}

	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("sops:task:check:status:failed, taskID: %d, bkBizID: %d, err: %v",
				taskID, bkBizID, err)
		}
		if obj == nil {
			return false, fmt.Errorf("sops:task:check:status:failed, taskID: %d not found, bkBizID: %d, resp is nil",
				taskID, bkBizID)
		}

		resp, ok := obj.(*sopsapi.GetTaskStatusResp)
		if !ok {
			return false, fmt.Errorf("sops:task:check:status:failed, object with taskID: %d, bkBizID: %d, "+
				"is not right task response: %+v", taskID, bkBizID, resp)
		}

		if !resp.Result {
			return false, fmt.Errorf("sops:task:check:status:failed, sops taskID: %d failed, bkBizID: %d, "+
				"code: %d, err: %s", taskID, bkBizID, resp.Code, resp.Message)
		}

		if resp.Data == nil {
			return false, fmt.Errorf("sops:task:check:status:failed, object with taskID: %d, bkBizID: %d, "+
				"resp.Data is nil, response: %+v", taskID, bkBizID, resp)
		}

		if resp.Data.State == sopsapi.TaskStateRunning || resp.Data.State == sopsapi.TaskStateCreated {
			return false, fmt.Errorf("sops:task:check:status, sops taskID: %d, bkBizID: %d is handling, state: %s",
				taskID, bkBizID, resp.Data.State)
		}

		if resp.Data.State != sopsapi.TaskStateFinished {
			return true, fmt.Errorf("sops:task:check:status, sops taskID: %d failed, bkBizID: %d, currentState: %s",
				taskID, bkBizID, resp.Data.State)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// 增加限频检查
		limiter := getSopsLimiter(getTaskStatus, readLimit)
		if !limiter.Allow() {
			return nil, errors.New("exceeded Query API frequency limit times")
		}
		return sopsCli.GetTaskStatus(kt.Ctx, kt.Header(), taskID, bkBizID)
	}

	_, err := utils.Retry(doFunc, checkFunc, 3600, 10)
	if err != nil {
		return err
	}

	return nil
}
