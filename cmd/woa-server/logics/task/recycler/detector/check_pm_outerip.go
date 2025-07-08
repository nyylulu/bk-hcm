/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

// Package detector ...
package detector

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/ngateapi"
	cvt "hcm/pkg/tools/converter"

	"k8s.io/client-go/util/workqueue"
)

const (
	checkPmOuterIPDelayTime = time.Second * 5
	checkPmOuterIPTimeout   = time.Minute * 10
)

// checkPmOuterIPWorkGroup ...
type checkPmOuterIPWorkGroup struct {
	resultHandler StepResultHandler
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	cliSet        *cliSet
	ccOp          CmdbOperator
	delayQueue    workqueue.DelayingInterface
}

// newCheckPmOuterIPWorkGroup ...
func newCheckPmOuterIPWorkGroup(resultHandler StepResultHandler, workerNum int,
	cliSet *cliSet) *checkPmOuterIPWorkGroup {

	return &checkPmOuterIPWorkGroup{
		resultHandler: resultHandler,
		stepBatchChan: make(chan *stepBatch, workerNum),
		currency:      workerNum,
		cliSet:        cliSet,
		ccOp:          NewCmdbOperator(cliSet.cc),
		delayQueue:    workqueue.NewDelayingQueue(),
	}
}

// Start ...
func (c *checkPmOuterIPWorkGroup) Start(kt *kit.Kit) {
	if !c.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < c.currency; i++ {
		subKit := kt.NewSubKit()
		go c.consume(subKit, i)
	}
	for i := 0; i < c.currency; i++ {
		subKit := kt.NewSubKit()
		go c.getConsumeResult(subKit, i)
	}
}

func (c *checkPmOuterIPWorkGroup) consume(kt *kit.Kit, idx int) {
	logs.Infof("consume worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("consume worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-c.stepBatchChan:
			logs.V(4).Infof("worker %d got steps: %d:%s, rid: %s", idx, len(batch.steps), batch.steps, kt.Rid)
			c.batchCheckPmOuterIP(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			c.delayQueue.ShutDownWithDrain()
			c.delayQueue.ShutDown()
			return
		}
	}
}

func (c *checkPmOuterIPWorkGroup) batchCheckPmOuterIP(kt *kit.Kit, steps []*StepMeta) {
	hostIDs := make([]int64, 0)
	for _, step := range steps {
		hostIDs = append(hostIDs, step.Step.HostID)
	}
	hosts, err := c.ccOp.GetHostBaseInfoByID(kt, hostIDs)
	if err != nil {
		logs.Errorf("failed to check pm outer ip, for get host from cc err: %v, host id: %v, rid: %s", err, hostIDs,
			kt.Rid)
		c.resultHandler.HandleResult(kt, steps, err, err.Error(), true)
		return
	}
	idHostMap := make(map[int64]cmdb.Host)
	for _, host := range hosts {
		idHostMap[host.BkHostID] = host
	}

	for _, step := range steps {
		host, ok := idHostMap[step.Step.HostID]
		if !ok {
			logs.Errorf("failed to check pm outer ip, can not find host, host id: %d, ip: %s, rid: %s",
				step.Step.HostID, step.Step.IP, kt.Rid)
			err = fmt.Errorf("can not find host, host id: %d, ip: %s", step.Step.HostID, step.Step.IP)
			c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, err.Error(), false)
			continue
		}

		c.checkPmOuterIP(kt, &host, step)
	}
}

// checkPmOuterIP 标准运维-物理机外网IP回收及清理检查
func (c *checkPmOuterIPWorkGroup) checkPmOuterIP(kt *kit.Kit, host *cmdb.Host, step *StepMeta) {
	// skip pm outer check if host is not physical machine
	if !host.IsPmAndOuterIPDevice() {
		logs.Infof("host is not pm, hostInfo: %+v, rid: %s", cvt.PtrToVal(host), kt.Rid)
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, nil, "跳过", false)
		return
	}

	exeInfos := make([]string, 0)
	// 调用公司sniper公网IP回收接口
	ngateExeInfos, retry, err := c.recycleNgateIP(kt, step.Step, host)
	if err != nil {
		logs.Errorf("failed to recycle ngate ip, err: %v, ip: %s, host id: %d, rid: %s", err, host.BkHostInnerIP,
			host.BkHostID, kt.Rid)
		exeInfos = append(exeInfos, err.Error())
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(exeInfos, "\n"), retry)
		return
	}
	// 记录ngate执行的信息
	exeInfos = append(exeInfos, ngateExeInfos...)

	// skip==0表示可以调用标准运维回收IP流程，否则跳过该流程(产品需求-需要支持跳过标准运维的流程，避免因该流程阻塞公司流程)
	if step.Step.Skip == 0 {
		c.recycleSopsOuterIP(kt, host, step, exeInfos)
	}
}

func (c *checkPmOuterIPWorkGroup) recycleNgateIP(kt *kit.Kit, step *table.DetectStep, hostInfo *cmdb.Host) ([]string,
	bool, error) {

	exeInfos := make([]string, 0)
	if step.User == "" {
		logs.Errorf("failed to recycle ngate outer ip, for invalid user is empty, rid: %s", kt.Rid)
		return exeInfos, false, fmt.Errorf("failed to check pm outerip, for invalid user is empty")
	}

	// 如果外网IP为空，无需处理，直接返回
	if !hostInfo.IsPmAndOuterIPDevice() {
		logs.Errorf("failed to recycle ngate outer ip, for invalid host outer ipv4 or ipv6 is empty, ip: %s, "+
			"hostInfo: %+v, rid: %s", step.IP, cvt.PtrToVal(hostInfo), kt.Rid)
		return exeInfos, false, fmt.Errorf("failed to check pm outerip, for invalid host outer ipv4 or ipv6 is empty")
	}

	for _, ipVersion := range []string{ngateapi.IPv4Version, ngateapi.IPv6Version} {
		addressList := make([]string, 0)
		switch ipVersion {
		case ngateapi.IPv4Version:
			if len(hostInfo.BkHostOuterIP) == 0 {
				continue
			}

			addressList = []string{hostInfo.BkHostOuterIP}
		case ngateapi.IPv6Version:
			if len(hostInfo.BkHostOuterIPv6) == 0 {
				continue
			}

			addressList = []string{hostInfo.BkHostOuterIPv6}
		}

		recycleIPReq := &ngateapi.RecycleIPReq{
			AssertIDList:  []string{hostInfo.BkAssetID},
			DeviceType:    ngateapi.ServerDeviceType,
			AddressList:   addressList,
			IPTypeEnum:    ngateapi.OuterIPType,
			IPVersionEnum: ipVersion,
			User:          step.User,
		}
		recycleIPResp, err := c.cliSet.ngate.RecycleIP(kt.Ctx, recycleIPReq)
		recycleIPRespStr := structToStr(recycleIPResp)
		ngateReqLogMsg := fmt.Sprintf("ngate recycle outer ip, ipVersion: %s, innerIP: %s, request: %s, response: %s",
			ipVersion, step.IP, structToStr(recycleIPReq), recycleIPRespStr)
		exeInfos = append(exeInfos, ngateReqLogMsg)
		logs.Infof("check pm outer ip, ngate response, %s, rid: %s", ngateReqLogMsg, kt.Rid)
		if err != nil {
			logs.Errorf("failed to use ngate recycle ip, err: %v, ipVersion: %s, step: %+v, recycleIPReq: %+v, rid: %s",
				err, ipVersion, cvt.PtrToVal(step), cvt.PtrToVal(recycleIPReq), kt.Rid)
			return exeInfos, true, fmt.Errorf("failed to check pm outer ip: %s, hostOuterIPv4: %s, hostOuterIPv6: %s, "+
				"stepName: %s, err: %v", step.IP, hostInfo.BkHostOuterIP, hostInfo.BkHostOuterIPv6, step.StepName, err)
		}

		if recycleIPResp.ReturnCode != 0 || !recycleIPResp.Success {
			return exeInfos, true, fmt.Errorf("recycle ngate outer ip: %s, ipVersion: %s, hostOuterIPv4: %s, "+
				"hostOuterIPv6: %s, api return err: %s", step.IP, ipVersion, hostInfo.BkHostOuterIP,
				hostInfo.BkHostOuterIPv6, recycleIPRespStr)
		}
	}

	return exeInfos, false, nil
}

// recycleSopsOuterIP 回收外网IP
func (c *checkPmOuterIPWorkGroup) recycleSopsOuterIP(kt *kit.Kit, host *cmdb.Host, step *StepMeta, exeInfos []string) {

	// 1. 构造参数
	params, supported, err := sops.GetRecycleOuterIPParams(kt, host.BkHostInnerIP, step.BizID, host.BkOsType)
	if err != nil {
		logs.Errorf("failed to recycle outer ip, ip: %s, bizID: %d, err: %v, rid: %s", host.BkHostInnerIP, step.BizID,
			err, kt.Rid)
		exeInfos = append(exeInfos, err.Error())
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(exeInfos, "\n"), false)
		return
	}
	if !supported {
		exeInfos = append(exeInfos, "不支持该主机操作系统类型")
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, nil, strings.Join(exeInfos, "\n"), false)
		return
	}

	// 2. 创建标准运维任务
	if err = sops.WaitSopsCreateTaskLimiter(kt.Ctx, SopsRateLimiterWaitTimeout); err != nil {
		logs.Errorf("fail to wait create limiter, err: %v, rid: %s", err, kt.Rid)
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(exeInfos, "\n"), true)
		return
	}
	taskResp, err := c.cliSet.sops.CreateTask(kt.Ctx, kt.Header(), params.TemplateID, step.BizID, params.CreateReq)
	if err != nil {
		logs.Errorf("fail to create sops  task, err: %v, ip: %s, rid: %s", err, host.BkHostInnerIP, kt.Rid)
		exeInfos = append(exeInfos, err.Error())
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(exeInfos, "\n"), true)
		return
	}

	// 3. 启动标准运维任务
	if err := sops.WaitSopsStartTaskLimiter(kt.Ctx, SopsRateLimiterWaitTimeout); err != nil {
		logs.Errorf("fail to wait start limiter, err: %v, rid: %s", err, kt.Rid)
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err,
			fmt.Sprintf("fail to wait start limiter, err: %v", err), true)
		return
	}
	if _, err = c.cliSet.sops.StartTask(kt.Ctx, kt.Header(), taskResp.Data.TaskId, step.BizID); err != nil {
		logs.Errorf("fail to start sops task, err: %v, ip: %s,  taskURL: %s, rid: %s", err, host.BkHostInnerIP,
			taskResp.Data.TaskUrl, kt.Rid)
		exeInfos = append(exeInfos, err.Error())
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(exeInfos, "\n"), true)
		return
	}

	// 4. 加入查询结果延迟队列
	task := querySopsResultTask{
		kt:          kt,
		step:        step,
		bizID:       step.BizID,
		taskID:      taskResp.Data.TaskId,
		taskUrl:     taskResp.Data.TaskUrl,
		createdTime: time.Now(),
		exeInfos:    exeInfos,
	}
	go c.delayQueue.AddAfter(task, checkPmOuterIPDelayTime)
}

func (c *checkPmOuterIPWorkGroup) getConsumeResult(kt *kit.Kit, idx int) {
	logs.Infof("get consume result worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("get consume result worker %d exit, rid: %s", idx, kt.Rid)

	for {
		raw, shutdown := c.delayQueue.Get()
		if shutdown {
			return
		}
		task, ok := raw.(*querySopsResultTask)
		if !ok {
			c.delayQueue.Done(raw)
			logs.Errorf("query result worker %d got wrong type: %T, raw: %v, rid: %s", idx, raw, raw, kt.Rid)
			continue
		}
		c.getResult(task)
	}
}

func (c *checkPmOuterIPWorkGroup) getResult(task *querySopsResultTask) {
	kt := task.kt
	step := task.step
	taskID := task.taskID
	bizID := task.bizID
	taskUrl := task.taskUrl
	createdTime := task.createdTime

	defer c.delayQueue.Done(task)
	err := sops.WaitSopsGetTaskStatusLimiter(kt.Ctx, SopsRateLimiterWaitTimeout)
	if err != nil {
		logs.Errorf("fail to wait query limiter, err: %v, rid: %s", err, kt.Rid)
		task.exeInfos = append(task.exeInfos, fmt.Sprintf("fail to wait query limiter, err: %v", err))
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(task.exeInfos, "\n"), true)
		return
	}

	// 1.查询任务状态
	statusResp, err := c.cliSet.sops.GetTaskStatus(kt.Ctx, kt.Header(), taskID, bizID)
	if err != nil {
		logs.Errorf("fail to get sops check task status, err: %v, task: %s, rid: %s", err, taskUrl, kt.Rid)
		task.exeInfos = append(task.exeInfos,
			fmt.Sprintf("get task status failed, sops url: %s, err: %v", taskUrl, err))
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(task.exeInfos, "\n"), true)
		return
	}

	// 2.判断任务状态
	state := statusResp.Data.State
	if state == sopsapi.TaskStateRunning || state == sopsapi.TaskStateCreated {
		queryCost := time.Since(createdTime)
		if queryCost > checkPmOuterIPTimeout {
			err = fmt.Errorf("task state query timeout, sops url: %s, cost: %s, current: %s", taskUrl, queryCost,
				state)
			task.exeInfos = append(task.exeInfos, err.Error())
			c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(task.exeInfos, "\n"), true)
			return
		}
		// 任务还在执行中, 延迟重试
		go c.delayQueue.AddAfter(task, SopsCheckProcessRunningCheckInterval)
		return
	}

	if state != sopsapi.TaskStateFinished {
		err = fmt.Errorf("host failed to check process, ip: %s, sops url: %s, state: %s", task.step.Step.IP, taskUrl,
			state)
		task.exeInfos = append(task.exeInfos, err.Error())
		c.resultHandler.HandleResult(kt, []*StepMeta{step}, err, strings.Join(task.exeInfos, "\n"), true)
		return
	}

	task.exeInfos = append(task.exeInfos, fmt.Sprintf("sops url: %s, task state: %s", taskUrl, state))
	c.resultHandler.HandleResult(kt, []*StepMeta{step}, nil, strings.Join(task.exeInfos, "\n"), false)
}

// Submit 提交任务
func (c *checkPmOuterIPWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	c.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// MaxBatchSize 最大批量数
func (c *checkPmOuterIPWorkGroup) MaxBatchSize() int {
	return defaultMaxBatchSize
}
