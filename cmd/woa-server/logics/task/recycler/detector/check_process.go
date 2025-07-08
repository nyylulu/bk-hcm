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

// Package detector ...
package detector

import (
	"fmt"
	"sync/atomic"
	"time"

	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"

	"k8s.io/client-go/util/workqueue"
)

// SopsCheckProcessTimeout 标准运维-空闲检查超时时间
const SopsCheckProcessTimeout = time.Minute * 60

// SopsCheckProcessFirstCheckDelay 标准运维-空闲检查首次检查结果延迟时间
const SopsCheckProcessFirstCheckDelay = time.Second * 20

// SopsCheckProcessRunningCheckInterval 标准运维-空闲检查运行中检查结果间隔
const SopsCheckProcessRunningCheckInterval = time.Second * 5

// SopsCheckProcessMaxBatchSize 标准运维-空闲检查最大并发数, create 和 start 只能逐个创建和启动，因此最大批次大小只能是1
const SopsCheckProcessMaxBatchSize = 1

// SopsRateLimiterWaitTimeout 标准运维-空闲检查限流等待超时时间
const SopsRateLimiterWaitTimeout = 3 * time.Minute

// CheckProcessWorkGroup 空闲检查，调用标准运维创建
type CheckProcessWorkGroup struct {
	createStepChan chan *checkProcessContext
	started        atomic.Bool
	currency       int
	resultHandler  StepResultHandler
	sopsCli        sopsapi.SopsClientInterface
	cc             CmdbOperator
	queryQueue     workqueue.DelayingInterface
}

type checkProcessContext struct {
	kt           *kit.Kit
	step         *StepMeta
	bizID        int64
	taskID       int64
	taskURL      string
	createdAt    time.Time
	queriedTimes int
}

func (c *checkProcessContext) String() string {
	return fmt.Sprintf("{step:%s, bizID:%d, taskURL:%s, createdAt:%s, queried:%d, rid:%s}",
		c.step.String(), c.bizID, c.taskURL, c.createdAt, c.queriedTimes, c.kt.Rid)
}

// Start ...
func (g *CheckProcessWorkGroup) Start(kt *kit.Kit) {
	if !g.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < g.currency; i++ {
		subKit := kt.NewSubKitWithSuffix(fmt.Sprintf("start%d", i))
		go g.createAndStartWorker(subKit, i)
	}
	for i := 0; i < g.currency; i++ {
		subKit := kt.NewSubKitWithSuffix(fmt.Sprintf("query%d", i))
		go g.queryResultWorker(subKit, i)
	}
}

// Submit ...
func (g *CheckProcessWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	for _, step := range steps {
		g.createStepChan <- &checkProcessContext{kt: kt, step: step}
	}
}

// MaxBatchSize ...
func (g *CheckProcessWorkGroup) MaxBatchSize() int {
	return SopsCheckProcessMaxBatchSize
}

// NewCheckProcessWorkGroup ...
func NewCheckProcessWorkGroup(sopsCli sopsapi.SopsClientInterface, cc cmdb.Client, resultHandler StepResultHandler,
	workerNum int) *CheckProcessWorkGroup {

	return &CheckProcessWorkGroup{
		createStepChan: make(chan *checkProcessContext, workerNum),
		currency:       workerNum,
		resultHandler:  resultHandler,
		sopsCli:        sopsCli,
		cc:             NewCmdbOperator(cc),
		queryQueue:     workqueue.NewDelayingQueue(),
	}
}

// handleCreateFailed 创建失败，需要重试
func (g *CheckProcessWorkGroup) handleCreateFailed(kt *kit.Kit, step *StepMeta, detectErr error, log string) {
	g.resultHandler.HandleResult(kt, []*StepMeta{step}, detectErr, log, true)
	return
}

// handleStartFailed 启动失败，需要重试
func (g *CheckProcessWorkGroup) handleStartFailed(kt *kit.Kit, step *StepMeta, detectErr error, log string) {
	g.resultHandler.HandleResult(kt, []*StepMeta{step}, detectErr, log, true)
	return
}

// handleQueryFailed 查询失败，需要重试
func (g *CheckProcessWorkGroup) handleQueryFailed(kt *kit.Kit, step *StepMeta, detectErr error, log string) {
	g.resultHandler.HandleResult(kt, []*StepMeta{step}, detectErr, log, true)
	return
}

// handleNotPass 不通过，不需要重试
func (g *CheckProcessWorkGroup) handleNotPass(kt *kit.Kit, step *StepMeta, detectErr error, log string) {
	g.resultHandler.HandleResult(kt, []*StepMeta{step}, detectErr, log, false)
	return
}

// handlePass 通过校验，不需要重试
func (g *CheckProcessWorkGroup) handlePass(kt *kit.Kit, step *StepMeta, log string) {
	g.resultHandler.HandleResult(kt, []*StepMeta{step}, nil, log, false)
	return
}

// createAndStartWorker 对应标准运维的 创建流程和启动流程两个操作
func (g *CheckProcessWorkGroup) createAndStartWorker(kt *kit.Kit, idx int) {
	logs.Infof("check process create worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check process create worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case stepCtx := <-g.createStepChan:
			logs.V(4).Infof("check process create %d got step: %s, rid: %s",
				idx, stepCtx.step.String(), kt.Rid)
			g.createAndStart(stepCtx.kt, stepCtx.step)
		case <-kt.Ctx.Done():
			g.queryQueue.ShutDownWithDrain()
			g.queryQueue.ShutDown()
			return
		}
	}
}

func (g *CheckProcessWorkGroup) queryResultWorker(kt *kit.Kit, idx int) {
	logs.Infof("check process query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check process query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		raw, shutdown := g.queryQueue.Get()
		if shutdown {
			return
		}
		stepCtx, ok := raw.(*checkProcessContext)
		if !ok {
			g.queryQueue.Done(raw)
			logs.Errorf("check process query worker %d got wrong type: %T, raw: %v, rid: %s", idx, raw, raw, kt.Rid)
			continue
		}
		g.queryResult(stepCtx.kt, stepCtx)
	}
}

// 对应标准运维的 创建流程、启动流程 两个操作
func (g *CheckProcessWorkGroup) createAndStart(kt *kit.Kit, step *StepMeta) {
	logs.Infof("check process create start, step: %s, rid: %s", step.String(), kt.Rid)

	// 1. 查询机型
	hostInfo, err := g.cc.GetHostInfoByHostID(kt, step.Step.HostID)
	if err != nil {
		logs.Errorf("fail to get host info by host id, err: %v, step: %s, rid: %s", err, step.String(), kt.Rid)
		g.handleCreateFailed(kt, step, err, fmt.Sprintf("fail to get host info by host id, err: %v", err))
		return
	}
	// 2. 构造参数
	idleCheckParams, supported, err := sops.GetIdleCheckParams(kt, hostInfo.BkOsType, step.Step.IP, step.BizID)
	if err != nil {
		logs.Errorf("fail to get idle check opt, err: %v, step: %s, rid: %s", err, step.String(), kt.Rid)
		g.handleCreateFailed(kt, step, err, fmt.Sprintf("fail to get idle check opt, err: %v", err))
		return
	}
	if !supported {
		g.handlePass(kt, step, fmt.Sprintf("os type unsupported: %s, skip", hostInfo.BkOsType))
		return
	}
	// 3. 创建标准运维任务
	if err := sops.WaitSopsCreateTaskLimiter(kt.Ctx, SopsRateLimiterWaitTimeout); err != nil {
		logs.Errorf("fail to wait create limiter, err: %v, rid: %s", err, kt.Rid)
		g.handleCreateFailed(kt, step, err, fmt.Sprintf("fail to wait create limiter, err: %v", err))
		return
	}
	taskResp, err := g.sopsCli.CreateTask(kt.Ctx, kt.Header(), idleCheckParams.TemplateID, step.BizID,
		idleCheckParams.CreateReq)
	if err != nil {
		logs.Errorf("fail to create sops idle check task, err: %v, step: %s, rid: %s", err, step.String(), kt.Rid)
		g.handleCreateFailed(kt, step, err, fmt.Sprintf("fail to create sops idle check task, err: %v", err))
		return
	}
	// 4. 启动标准运维任务
	if err := sops.WaitSopsStartTaskLimiter(kt.Ctx, SopsRateLimiterWaitTimeout); err != nil {
		logs.Errorf("fail to wait start limiter, err: %v, rid: %s", err, kt.Rid)
		g.handleStartFailed(kt, step, err, fmt.Sprintf("fail to wait start limiter, err: %v", err))
		return
	}
	_, err = g.sopsCli.StartTask(kt.Ctx, kt.Header(), taskResp.Data.TaskId, step.BizID)
	if err != nil {
		logs.Errorf("fail to start idle check task, err: %v, step: %s,  taskURL: %s, rid: %s",
			err, step.String(), taskResp.Data.TaskUrl, kt.Rid)
		g.handleStartFailed(kt, step, err,
			fmt.Sprintf("fail to start idle check task, sops url:%s, err: %v", taskResp.Data.TaskUrl, err))
		return
	}

	// 6. 加入查询列表
	stepCtx := &checkProcessContext{
		kt:           kt,
		step:         step,
		bizID:        step.BizID,
		taskID:       taskResp.Data.TaskId,
		taskURL:      taskResp.Data.TaskUrl,
		createdAt:    time.Now(),
		queriedTimes: 0,
	}
	go g.delayQuery(SopsCheckProcessFirstCheckDelay, stepCtx)
}

func (g *CheckProcessWorkGroup) delayQuery(delay time.Duration, stepCtx *checkProcessContext) {
	g.queryQueue.AddAfter(stepCtx, delay)
}

func (g *CheckProcessWorkGroup) queryResult(kt *kit.Kit, stepCtx *checkProcessContext) {
	defer g.queryQueue.Done(stepCtx)
	logs.Infof("check process query start, step: %s, rid: %s", stepCtx.String(), kt.Rid)
	// 0. 限流
	err := sops.WaitSopsGetTaskStatusLimiter(kt.Ctx, SopsRateLimiterWaitTimeout)
	if err != nil {
		logs.Errorf("fail to wait query limiter, err: %v, rid: %s", err, kt.Rid)
		log := fmt.Sprintf("fail to wait query limiter, err: %v", err)
		g.handleQueryFailed(kt, stepCtx.step, err, log)
		return
	}

	// 1. 查询任务状态
	statusResp, err := g.sopsCli.GetTaskStatus(kt.Ctx, kt.Header(), stepCtx.taskID, stepCtx.bizID)
	if err != nil {
		logs.Errorf("fail to query sops idle check task statusResp, err: %v, step: %s, task: %s, rid: %s",
			err, stepCtx.step.String(), stepCtx.taskURL, kt.Rid)
		// 查询失败返回上层
		log := fmt.Sprintf("get task status failed, sops url: %s, err: %v", stepCtx.taskURL, err)
		g.handleQueryFailed(kt, stepCtx.step, err, log)
		return
	}
	stepCtx.queriedTimes++
	state := statusResp.Data.State

	// 2. 判断任务状态
	if state == sopsapi.TaskStateRunning || state == sopsapi.TaskStateCreated {
		queryCost := time.Since(stepCtx.createdAt)
		if queryCost > SopsCheckProcessTimeout {
			// 超时失败
			err := fmt.Errorf("task state query timeout, sops url: %s, cost: %s, current: %s",
				stepCtx.taskURL, queryCost, state)
			g.handleQueryFailed(kt, stepCtx.step, err, err.Error())
			return
		}
		// 2.2 任务还在执行中, 延迟重试
		go g.delayQuery(SopsCheckProcessRunningCheckInterval, stepCtx)
		return
	}

	if state != sopsapi.TaskStateFinished {
		// 2.3 任务状态异常，返回上层
		// 如果失败的是JOB节点，则获取JOB平台的链接
		jobInstURL, jobInstErr := sops.GetIdleCheckFailedJobUrl(kt, g.sopsCli, stepCtx.taskID, stepCtx.bizID,
			statusResp)
		if jobInstErr != nil {
			logs.Errorf("check job status failed, step: %s, sopsUrl: %s, jobInstURL: %s, jobInstErr: %v, rid: %s",
				stepCtx.step.String(), stepCtx.taskURL, jobInstURL, jobInstErr, kt.Rid)
		}
		jobErrMsg := ""
		if jobInstURL != "" {
			jobErrMsg = fmt.Sprintf("作业平台(JOB): %s,", jobInstURL)
		}
		err := fmt.Errorf("host %s failed to check process, sops url: %s, state: %s, %s",
			stepCtx.step.Step.IP, stepCtx.taskURL, state, jobErrMsg)
		g.handleNotPass(kt, stepCtx.step, err, err.Error())
		return
	}

	// 2.4 成功
	msg := fmt.Sprintf("sops url: %s, task state: %s", stepCtx.taskURL, state)
	g.handlePass(kt, stepCtx.step, msg)
}
