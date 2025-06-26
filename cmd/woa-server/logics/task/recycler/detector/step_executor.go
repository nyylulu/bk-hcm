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

package detector

import (
	"fmt"
	"sync"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"

	"golang.org/x/time/rate"
)

const emptyStepSleepInterval = time.Millisecond * 10

// Result 预检结果
type Result struct {
	StepName table.DetectStepName
	HostID   int64
	TaskID   string
	Error    error
}

// String ...
func (r *Result) String() string {
	return fmt.Sprintf("<R %s:%d,%s,E:%v>", r.StepName, r.HostID, r.TaskID, r.Error)
}

// ScoreFunc 打分函数
type ScoreFunc func(step *StepMeta) int64

// ScoreByCreateTime 根据入队时间打分, 越早分数越高
func ScoreByCreateTime(step *StepMeta) int64 {
	return -step.JoinedAt.UnixMilli()
}

// StepExecutor 预检步骤执行器
type StepExecutor interface {
	StepResultHandler
	// GetStepName 获取当前步骤名
	GetStepName() table.DetectStepName
	// SubmitSteps 提交子单内的全部主机的当前预检步骤类型的主机
	SubmitSteps(kt *kit.Kit, suborderID string, currentStepHosts []*table.DetectStep) (<-chan *Result, error)
	// Start 启动
	Start(kt *kit.Kit, workgroup DetectStepWorkGroup)
	// CancelSuborder 取消指定子单的执行
	CancelSuborder(kt *kit.Kit, suborderID string)
}

// DetectStepWorkGroup 任务执行组
type DetectStepWorkGroup interface {
	Start(kt *kit.Kit)
	// Submit 提交子单内的全部主机的当前预检步骤类型的主机，steps的数量不能超过 MaxBatchSize。单次Submit应该对应对第三方接口的单次请求。
	Submit(kt *kit.Kit, steps []*StepMeta)
	// MaxBatchSize Submit函数所能接受的最大数量，
	// 这里应该对应下游第三方系统的单次请求所能接受的参数数量，如果有多个步骤的请求应该以最小的限制为准。
	// 若WorkGroup有自行维护限流的可以不受这个限制，如：标准运维同时要遵守全局频率限制。
	MaxBatchSize() int
}

// NewDetectStepExecutor ...
func NewDetectStepExecutor(stepCfg *table.DetectStepCfg) *DetectStepExecutor {
	return &DetectStepExecutor{
		stepName:        stepCfg.Name,
		maxRetryTimes:   stepCfg.Retry,
		batchSize:       stepCfg.BatchSize,
		retryInterval:   time.Second * time.Duration(stepCfg.RetryIntervalSec),
		scoreFunc:       ScoreByCreateTime,
		suborderChanMap: sync.Map{},
		rateLimiter:     rate.NewLimiter(rate.Limit(stepCfg.RateLimitQps), stepCfg.RateLimitBurst),
		waitList:        &syncList{},
	}
}

// DetectStepExecutor 预检步骤执行器
type DetectStepExecutor struct {
	stepName      table.DetectStepName
	batchSize     int
	retryInterval time.Duration
	maxRetryTimes int
	scoreFunc     ScoreFunc
	// map of suborderID -> resultChan
	suborderChanMap sync.Map
	rateLimiter     *rate.Limiter
	waitList        *syncList
}

// String ...
func (d *DetectStepExecutor) String() string {
	return fmt.Sprintf("{StepName:%s,Batch:%d,RetryInterval:%s,Limiter:%f,%d}",
		d.stepName, d.batchSize, d.retryInterval, d.rateLimiter.Limit(), d.rateLimiter.Burst())
}

type suborderChan struct {
	rid       string
	ch        chan *Result
	remaining int
	locker    sync.Mutex
}

// SendResult ...
func (sc *suborderChan) SendResult(r *Result) (allFinished bool) {
	sc.locker.Lock()
	defer sc.locker.Unlock()
	if sc.ch == nil {
		return true
	}
	sc.remaining--
	sc.ch <- r
	allFinished = sc.remaining <= 0
	if allFinished {
		sc.closeWithoutLock()
	}
	return allFinished
}

// Rid ...
func (sc *suborderChan) Rid() string {
	return sc.rid
}

// String ...
func (sc *suborderChan) String() string {
	return fmt.Sprintf("suborderChan{remaining: %d, by rid: %s}", sc.remaining, sc.rid)
}

// Close ...
func (sc *suborderChan) Close() {
	sc.locker.Lock()
	defer sc.locker.Unlock()
	sc.closeWithoutLock()
}

// closeWithoutLock close and set to nil, use it after lock acquired
func (sc *suborderChan) closeWithoutLock() {
	if sc.ch != nil {
		close(sc.ch)
		sc.ch = nil
	}
}

// StepMeta 待执行任务信息
type StepMeta struct {
	Step       *table.DetectStep
	JoinedAt   time.Time
	Urgent     bool
	Score      int64
	RetryTimes int
	Rid        string
}

// String ...
func (m *StepMeta) String() string {
	if m == nil {
		return "nil"
	}
	// N: Normal
	urgent := 'N'
	if m.Urgent {
		// U: urgent
		urgent = 'U'
	}

	wait := time.Now().Sub(m.JoinedAt).Round(time.Millisecond).String()
	return fmt.Sprintf("{%s%c%d#%dw%s}", m.Step.Describe(), urgent, m.Score, m.RetryTimes, wait)
}

// GetStepName ...
func (d *DetectStepExecutor) GetStepName() table.DetectStepName {
	return d.stepName
}

// SubmitSteps 提交任务
func (d *DetectStepExecutor) SubmitSteps(kt *kit.Kit, suborderID string, currentStepHosts []*table.DetectStep) (
	<-chan *Result, error) {

	if len(currentStepHosts) == 0 {
		return nil, fmt.Errorf("can not submit empty step hosts, suborder: %s", suborderID)
	}

	for i, step := range currentStepHosts {
		if step.StepName != d.GetStepName() {
			logs.Errorf("StepExecutor got mismatch step name %s at idx: %d, suborder: %s, support step name: %s, "+
				"rid: %s", step.StepName, i, suborderID, d.GetStepName(), kt.Rid)
			return nil, fmt.Errorf("got mismatch step name %s at idx: %d, suborder: %s, support step name: %s",
				step.StepName, i, suborderID, d.GetStepName())
		}
	}
	sc := &suborderChan{
		ch:        make(chan *Result, len(currentStepHosts)),
		remaining: len(currentStepHosts),
		rid:       kt.Rid,
	}
	if actual, loaded := d.suborderChanMap.LoadOrStore(suborderID, sc); loaded {
		currentSc, ok := actual.(*suborderChan)
		if !ok {
			logs.Errorf("%s: suborder of %s already exist, but type is  %T instead of suborderChan",
				constant.CvmRecycleFailed, suborderID, actual)
			return nil, fmt.Errorf("suborder %s already exist, type assert error", suborderID)
		}
		return nil, fmt.Errorf("suborder %s already exist: %s", suborderID, currentSc.String())
	}

	d.addStepToWaitList(kt, currentStepHosts)
	return sc.ch, nil
}

func (d *DetectStepExecutor) addStepToWaitList(kt *kit.Kit, steps []*table.DetectStep) {
	stepMetas := make([]*StepMeta, 0, len(steps))
	for i := range steps {
		stepMetas = append(stepMetas, NewStepMeta(kt, steps[i]))
	}
	d.waitList.Add(stepMetas)
}

// NewStepMeta 创建一个stepMeta
func NewStepMeta(kt *kit.Kit, step *table.DetectStep) *StepMeta {
	return &StepMeta{
		Step:       step,
		JoinedAt:   time.Now(),
		Urgent:     false,
		Score:      0,
		RetryTimes: 0,
		Rid:        kt.Rid,
	}
}

// Start 启动调度循环
func (d *DetectStepExecutor) Start(kt *kit.Kit, workgroup DetectStepWorkGroup) {
	logs.Infof("detect step executor schedule loop start: %s, rid: %s", d.String(), kt.Rid)
	defer logs.Infof("detect step executor schedule loop stop: %s, rid: %s", d.stepName, kt.Rid)

	workgroup.Start(kt)
	for {
		select {
		case <-kt.Ctx.Done():
			return
		default:
			// 1. 限流
			err := d.rateLimiter.Wait(kt.Ctx)
			if err != nil {
				logs.Errorf("detect scheduled wait rate limiter failed, err: %s, step: %s, rid: %s",
					err, d.stepName, kt.Rid)
				return
			}
			subKit := kt.NewSubKit()
			// 2. 优先级计算, 3. 并取出任务
			d.scheduledOnce(subKit, workgroup)
		}
	}
}

// scheduledOnce 调度任务
func (d *DetectStepExecutor) scheduledOnce(kt *kit.Kit, workgroup DetectStepWorkGroup) {
	steps := d.PopTopK(d.batchSize)
	if len(steps) == 0 {
		time.Sleep(emptyStepSleepInterval)
		return
	}
	// 1. 更新任务状态
	stepIDs := slice.Map(steps, func(step *StepMeta) string { return step.Step.ID })
	// 2. 更新任务状态
	err := d.batchUpdateRecycleStep(kt, stepIDs, table.DetectStatusRunning)
	if err != nil {
		logs.Errorf("failed to update detect steps %v status to %s, err: %s, rid: %s",
			stepIDs, table.DetectStatusRunning, err, kt.Rid)
		// 状态更新失败时不影响继续执行
	}
	// 3. 调用workgroup执行
	workgroup.Submit(kt, steps)
}

// PopTopK 更新优先级, 并取出最高优先级的k个任务
func (d *DetectStepExecutor) PopTopK(k int) []*StepMeta {
	scoreFunc := d.scoreFunc
	if scoreFunc == nil {
		scoreFunc = ScoreByCreateTime
	}
	return d.waitList.PopTopK(scoreFunc, k)
}

// HandleResult 处理结果：1.持久化单个预检步骤结果到db 2. 通知Detector对应子单执行结果 3. 维护子单Channel状态
func (d *DetectStepExecutor) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string, needRetry bool) {
	targetStatus := table.DetectStatusSuccess
	errStr := "success"

	if detectErr != nil {
		targetStatus = table.DetectStatusFailed
		errStr = detectErr.Error()
	}

	for _, step := range steps {
		if step == nil {
			continue
		}
		// 1. 检查任务所属结果channel是否存在，且当前任务的rid和结果channel所属rid相符，否则跳过
		suborderChan := d.getChannelBySuborderID(kt, step.Step.SuborderID)
		if suborderChan == nil || suborderChan.ch == nil {
			logs.Errorf("failed to get channel of suborder %s, got nil data, detectErr: %v, suborderChan: %+v, rid: %s",
				step.Step.SuborderID, detectErr, suborderChan, kt.Rid)
			continue
		}
		// rid不相符，说明已经被重新入队
		if suborderChan.rid != step.Rid {
			logs.Warnf("discard detect result of suborder %s, stepRid: %s, err: %s, log: %s, rid: %s",
				step.Step.SuborderID, step.Rid, errStr, log, kt.Rid)
			continue
		}

		d.recordMetric(step, detectErr)

		// 重试次数小于配置
		if needRetry && detectErr != nil && step.RetryTimes+1 < d.maxRetryTimes {
			// 延迟入队，暂不写入结果
			step.RetryTimes++
			// 写入中间结果到db
			err := d.updateRecycleStep(kt, step.Step.ID, table.DetectStatusRunning, step.RetryTimes, errStr, log)
			if err != nil {
				logs.Errorf("%s: failed to update internal recycle step status, err: %v, step: %s, rid: %s",
					constant.CvmRecycleFailed, err, step.Step.Describe(), kt.Rid)
				return
			}
			go d.delayRetry(kt, step)
			continue
		}

		// 持久化结果到db
		err := d.updateRecycleStep(kt, step.Step.ID, targetStatus, step.RetryTimes, errStr, log)
		if err != nil {
			logs.Errorf("%s: failed to update recycle step status, err: %v, step: %s, rid: %s",
				constant.CvmRecycleFailed, err, step.Step.Describe(), kt.Rid)
			return
		}

		result := &Result{
			StepName: step.Step.StepName,
			HostID:   step.Step.HostID,
			TaskID:   step.Step.TaskID,
			Error:    detectErr,
		}
		allFinished := suborderChan.SendResult(result)
		if allFinished {
			// 发送完毕，清理信息
			d.suborderChanMap.Delete(step.Step.SuborderID)
			logs.Infof("detect step finished, step: %s, suborder: %s, suborderChan: %s, rid: %s",
				d.GetStepName(), step.Step.SuborderID, suborderChan.String(), kt.Rid)
		}
	}
}

// CancelSuborder 取消子单，关闭结果channel
func (d *DetectStepExecutor) CancelSuborder(kt *kit.Kit, suborderID string) {
	logs.Infof("try to cancel suborder %s, rid: %s", suborderID, kt.Rid)
	resultScRaw, loaded := d.suborderChanMap.LoadAndDelete(suborderID)
	if !loaded || resultScRaw == nil {
		return
	}
	resultSc, ok := resultScRaw.(*suborderChan)
	if !ok {
		return
	}

	// 通知上层
	resultSc.Close()
	logs.Infof("canceled suborder %s, resultSc: %s, rid: %s", suborderID, resultSc.String(), kt.Rid)
}

func (d *DetectStepExecutor) getChannelBySuborderID(kt *kit.Kit, suborderID string) *suborderChan {
	resultScRaw, ok := d.suborderChanMap.Load(suborderID)
	if !ok || resultScRaw == nil {
		return nil
	}

	resultSc, ok := resultScRaw.(*suborderChan)
	if !ok {
		logs.Errorf("%s: type of result channel of suborder %s mismatch resultScRaw: %+v, rid: %s",
			constant.CvmRecycleFailed, suborderID, resultScRaw, kt.Rid)
		return nil
	}
	return resultSc
}
func (d *DetectStepExecutor) batchUpdateRecycleStep(kt *kit.Kit, stepIDList []string, status table.DetectStatus) error {
	filter := &mapstr.MapStr{
		"id": map[string]any{
			pkg.BKDBIN: stepIDList,
		},
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"update_at": now,
	}

	switch status {
	case table.DetectStatusSuccess, table.DetectStatusFailed:
		doc["end_at"] = now
	case table.DetectStatusRunning:
		doc["start_at"] = now
	default:
		// do nothing for other status
	}

	if err := dao.Set().DetectStep().UpdateDetectStep(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update recycle step: %s, update: %+v, err: %v", stepIDList, doc, err)
		return err
	}

	return nil
}

func (d *DetectStepExecutor) updateRecycleStep(kt *kit.Kit, stepID string, status table.DetectStatus,
	attempt int, msg, log string) error {

	filter := &mapstr.MapStr{
		"id": stepID,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"retry_time": attempt,
		"status":     status,
		"message":    msg,
		"log":        log,
		"update_at":  now,
	}

	switch status {
	case table.DetectStatusSuccess, table.DetectStatusFailed:
		doc["end_at"] = now
	case table.DetectStatusRunning:
		doc["start_at"] = now
	default:
		// do nothing for other status
	}

	if err := dao.Set().DetectStep().UpdateDetectStep(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update recycle step: %s, update: %+v, err: %v", stepID, doc, err)
		return err
	}

	return nil
}

func (d *DetectStepExecutor) delayRetry(kt *kit.Kit, step *StepMeta) {
	sleepTime := d.retryInterval
	// sleep at least 3 seconds
	if sleepTime < 3*time.Second {
		sleepTime = 3 * time.Second
	}
	time.Sleep(sleepTime)

	d.addStepMeta(step)
}

func (d *DetectStepExecutor) addStepMeta(step *StepMeta) {
	d.waitList.Add([]*StepMeta{step})
}

func (d *DetectStepExecutor) recordMetric(step *StepMeta, err error) {
	labels := map[string]string{
		"step_name": string(step.Step.StepName),
	}

	sinceJoined := time.Since(step.JoinedAt)
	detectorMetrics.DetectStepCostSec.With(labels).Observe(sinceJoined.Seconds())

	if err != nil {
		detectorMetrics.DetectStepErrCounter.With(labels).Inc()
	}
}

// syncList 用于存储待执行任务 并发安全
type syncList struct {
	list   []*StepMeta
	locker sync.RWMutex
}

// PopTopK 计算分数，并返回分数最高的k个
func (wl *syncList) PopTopK(scoreFunc ScoreFunc, k int) (steps []*StepMeta) {
	length := wl.Length()
	if length == 0 {
		return []*StepMeta{}
	}

	k = max(k, 1)

	// 锁住waitList防止计算过程中被修改
	wl.locker.Lock()
	defer wl.locker.Unlock()

	length = wl.unlockedLength()
	// 数量较少无需排序
	if length <= k {
		steps = wl.list
		// 取完清空waitList
		wl.list = []*StepMeta{}
		return steps
	}

	// 1. 计算优先级
	for _, step := range wl.list {
		step.Score = scoreFunc(step)
	}

	// 2.1 部分排序
	slice.TopKSort(k, wl.list, func(a, b *StepMeta) bool {
		return a.Score < b.Score
	})

	// 2.2 取出优先级最高的k个
	topK := wl.list[length-k:]
	wl.list = wl.list[:length-k]
	return topK
}

// Length 长度
func (wl *syncList) Length() int {
	wl.locker.RLock()
	defer wl.locker.RUnlock()

	return wl.unlockedLength()
}

// Add 添加任务
func (wl *syncList) Add(steps []*StepMeta) {
	wl.locker.Lock()
	defer wl.locker.Unlock()

	wl.list = append(wl.list, steps...)
}

func (wl *syncList) unlockedLength() int {
	return len(wl.list)
}
