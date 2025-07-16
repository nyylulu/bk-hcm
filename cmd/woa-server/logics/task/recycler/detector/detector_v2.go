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
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/metadata"
)

// StepResultHandler 预检步骤执行结果处理回调
type StepResultHandler interface {
	// HandleResult 处理结果接口，在错误非空的时候，needRetry表示是否需要重试
	HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string, needRetry bool)
}

func (d *Detector) initStepExecutor(backendKit *kit.Kit) error {
	d.StepExecutors = map[table.DetectStepName]StepExecutor{}

	// 获取预检步骤配置
	stepConfigs, err := d.getDetectStepConfigs()
	if err != nil {
		logs.Errorf("get detect step configs failed, err: %v, rid: %s", err, backendKit.Rid)
		return err
	}

	for _, stepCfg := range stepConfigs {
		executor := NewDetectStepExecutor(stepCfg)
		var workgroup DetectStepWorkGroup
		switch stepCfg.Name {
		case table.StepPreCheck:
			workgroup = NewPreCheckWorkGroup(d.cc, executor, stepCfg.Worker)
		case table.StepBasicCheck:
			workgroup = newCheckBasicWorkGroup(executor, stepCfg.Worker, &d.cliSet)
		case table.StepCvmCheck:
			workgroup = newCheckCvmWorkGroup(executor, stepCfg.Worker, &d.cliSet)
		case table.StepCheckProcess:
			workgroup = NewCheckProcessWorkGroup(d.sops, d.cc, executor, stepCfg.Worker)
		case table.StepCheckTcaplus:
			workgroup = NewCheckTcaplusWorkGroup(d.tcaplus, executor, stepCfg.Worker)
		case table.StepCheckDBM:
			workgroup = NewCheckDBMWorkGroup(d.bkDbm, executor, stepCfg.Worker)
		case table.StepCheckOwner:
			workgroup = NewCheckOwnerWorkGroup(d.cc, executor, stepCfg.Worker)
		case table.StepCheckUwork:
			workgroup = NewCheckUworkWorkGroup(d.xray, d.xship, executor, stepCfg.Worker,
				stepCfg.RateLimitQps, stepCfg.RateLimitBurst)
		case table.StepCheckPmOuterIP:
			workgroup = newCheckPmOuterIPWorkGroup(executor, stepCfg.Worker, &d.cliSet)
		case table.StepCheckReturn:
			workgroup = NewCheckReturnWorkGroup(d.cvm, executor, stepCfg.Worker)
		default:
			return errors.New(string("detect step not supported: " + stepCfg.Name))
		}
		if workgroup.MaxBatchSize() < stepCfg.BatchSize {
			logs.Errorf("detect step config batch size is too large, step: %s, batch size:%d, max: %d, rid: %s",
				stepCfg.Name, stepCfg.BatchSize, workgroup.MaxBatchSize(), backendKit.Rid)
			return fmt.Errorf("detect step config batch size is too large, step: %s, batch size:%d, max: %d, rid: %s",
				stepCfg.Name, stepCfg.BatchSize, workgroup.MaxBatchSize(), backendKit.Rid)
		}
		subKit := backendKit.NewSubKit()
		go executor.Start(subKit, workgroup)
		d.StepExecutors[stepCfg.Name] = executor
	}
	return nil
}

// Detect 单据预检入口，等待全部主机的所有步骤结束后返回结果，可重入
func (d *Detector) Detect(kt *kit.Kit, order *table.RecycleOrder) error {
	// 获取预检步骤配置
	stepConfigs, err := d.getDetectStepConfigs()
	if err != nil {
		logs.Errorf("get detect step configs failed, order: %s, err: %v, rid: %s",
			order.SuborderID, err, kt.Rid)
		return err
	}

	// 初始化并获取需要处理的预检任务（主机）
	taskMap, err := d.initAndGetDetectTaskMap(kt, order.SuborderID, len(stepConfigs))
	if err != nil {
		logs.Errorf("fail to get detect task map, order: %s, err: %v, rid: %s",
			order.SuborderID, err, kt.Rid)
		return err
	}

	runner := newDetectStepRunner(order.SuborderID, order.BizID, len(stepConfigs), len(taskMap))
	for _, stepCfg := range stepConfigs {
		executor := d.StepExecutors[stepCfg.Name]
		if executor == nil {
			logs.Errorf("detecto failed, detect step executor not found, step name: %s, order: %s, rid: %s",
				stepCfg.Name, order.SuborderID, kt.Rid)
			return fmt.Errorf("detecto failed, detect step executor not found: %s", stepCfg.Name)

		}

		// 初始化当前预检步骤下需要执行的主机
		toSkipSteps, toExecuteSteps, err := d.prepareDetectStep(kt, order.SuborderID, stepCfg, taskMap)
		if err != nil {
			logs.Errorf("prepare detect step %s failed, order: %s, err: %v, rid: %s",
				stepCfg.Name, order.SuborderID, err, kt.Rid)
			return err
		}

		// 后台提交预检步骤
		go runner.Run(kt, executor, toExecuteSteps, toSkipSteps)
	}

	return d.handleDetectResult(kt, order.SuborderID, runner.ReadResults, taskMap)
}

// Cancel 单据取消入口, 会终止当前执行中单据，并丢弃现有结果。若单据未在执行中，不会报错
func (d *Detector) Cancel(kt *kit.Kit, suborderID string) error {
	// 获取预检步骤配置
	stepConfigs, err := d.getDetectStepConfigs()
	if err != nil {
		logs.Errorf("get detect step configs failed, order: %s, err: %v, rid: %s",
			suborderID, err, kt.Rid)
		return err
	}

	for _, stepCfg := range stepConfigs {
		executor := d.StepExecutors[stepCfg.Name]
		if executor == nil {
			logs.Errorf("cancel failed, detect step not found, step name: %s, order: %s, rid: %s",
				stepCfg.Name, suborderID, kt.Rid)
			return fmt.Errorf("cancel failed, detect step executor not found: %s", stepCfg.Name)
		}
		executor.CancelSuborder(kt, suborderID)
	}
	return nil
}

// handleDetectResult 收集预检步骤结果，更新预检进度
func (d *Detector) handleDetectResult(kt *kit.Kit, suborderID string, getResults func() (result []*Result, ok bool),
	taskMap map[int64]*table.DetectTask) error {

	defer logs.Infof("suborder all detect step done, order: %s, rid: %s", suborderID, kt.Rid)

	// 等待单内所有主机的所有预检步骤结果
	for {
		results, ok := getResults()
		if !ok {
			// 所有步骤处理完毕
			break
		}
		if len(results) == 0 {
			logs.Warnf("detect result is empty, order: %s, rid: %s", suborderID, kt.Rid)
			continue
		}
		// 合并同主机的更新结果，降低DB压力
		updatedTasks := make(map[int64]*table.DetectTask, len(taskMap))
		for _, result := range results {
			task := taskMap[result.HostID]
			if task == nil {
				logs.Warnf("detect task of host %d not found, order: %s, rid: %s", result.HostID, suborderID, kt.Rid)
				continue
			}
			logs.V(4).Infof("detector suborder: %s/%s got result: %s, rid: %s",
				suborderID, task.IP, result.String(), kt.Rid)

			if result.Error != nil {
				task.FailedNum++
				task.Status = table.DetectStatusFailed
				task.Message += result.Error.Error() + "\n"
			} else {
				task.SuccessNum++
				if task.Status != table.DetectStatusFailed {
					task.Status = table.DetectStatusSuccess
				}
			}
			task.PendingNum--
			updatedTasks[task.HostID] = task
		}
		for _, task := range updatedTasks {
			// 更新recycle task, recycle host 的状态
			err := d.updateRecycleTaskAndHostStatus(kt, task)
			if err != nil {
				logs.Errorf("skip update task %s failed, order: %s, err: %v, rid: %s",
					task.TaskID, suborderID, err, kt.Rid)
				// 忽略单次失败，防止阻塞Channel
			}
		}
	}

	return nil
}

func (d *Detector) initAndGetDetectTaskMap(kt *kit.Kit, suborderID string, stepCount int) (
	map[int64]*table.DetectTask, error) {

	// (重新)初始化task状态为running
	err := d.batchUpdateTaskStatus(kt, suborderID, table.DetectStatusRunning, stepCount)
	if err != nil {
		logs.Errorf("failed to update task status to running, order: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return nil, err
	}

	taskInfos, err := d.getDetectTasks(kt, suborderID)
	if err != nil {
		logs.Errorf("failed to get recycle tasks of order: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return nil, err
	}

	// 兼容存量数据没有bk host id、asset id的情况
	taskMap, err := d.fillTaskHostIDMap(kt, taskInfos, suborderID)
	if err != nil {
		logs.Errorf("failed to fill task host id map, order: %s, err: %v, rid: %s", suborderID, err, kt.Rid)
		return nil, err
	}

	for _, task := range taskInfos {
		task.Status = table.DetectStatusRunning
		task.FailedNum = 0
		task.SuccessNum = 0
		task.PendingNum = uint(stepCount)
		task.TotalNum = uint(stepCount)
		task.Message = ""
		taskMap[task.HostID] = task
	}
	return taskMap, nil
}

func newDetectStepRunner(suborderID string, bizID int64, stepTypeCount, hostCount int) stepRunner {
	allResult := make(chan *Result, stepTypeCount*hostCount)
	step := stepRunner{
		suborderID:     suborderID,
		bizID:          bizID,
		remainingSteps: &atomic.Int64{},
		allResult:      allResult,
	}
	step.remainingSteps.Add(int64(stepTypeCount))
	return step
}

// stepRunner 预检步骤执行器 封装对预检步骤的执行l流程
type stepRunner struct {
	suborderID     string
	bizID          int64
	remainingSteps *atomic.Int64
	allResult      chan *Result
}

// Run 执行预检步骤，并汇总其结果
func (sr *stepRunner) Run(kt *kit.Kit, executor StepExecutor, execSteps, skipSteps []*table.DetectStep) {
	defer sr.Done()

	// 将跳过的直接投到结果队列
	for _, step := range skipSteps {
		result := &Result{
			StepName: step.StepName,
			HostID:   step.HostID,
			TaskID:   step.TaskID,
		}
		sr.SendResult(result)
	}
	if len(execSteps) <= 0 {
		// 全部跳过
		return
	}
	// 初始化预检步骤
	resultCh, submitErr := executor.SubmitSteps(kt, sr.suborderID, sr.bizID, execSteps)
	if submitErr != nil {
		logs.Errorf("submit step %s failed, order: %s, err: %v, rid: %s",
			executor.GetStepName(), sr.suborderID, submitErr, kt.Rid)
		return
	}
	// 接收预检步骤结果并汇总
	for result := range resultCh {
		sr.SendResult(result)
	}
	logs.Infof("detect step runner finished, order: %s, stepName: %s, rid: %s",
		sr.suborderID, executor.GetStepName(), kt.Rid)
}

// Done 当前步骤执行结束
func (sr *stepRunner) Done() {
	if sr.remainingSteps.Add(-1) == 0 {
		// 全部流程结束，关闭总结果channel
		close(sr.allResult)
	}
}

// SendResult 发送结果
func (sr *stepRunner) SendResult(result *Result) {
	sr.allResult <- result
}

const stepRunnerBatchReadDefaultLength = 1000
const stepRunnerBatchReadDefaultTimeout = time.Millisecond * 200

// ReadResults 读取结果
func (sr *stepRunner) ReadResults() (result []*Result, ok bool) {
	return concurrence.BatchReadChannel(sr.allResult, stepRunnerBatchReadDefaultLength,
		stepRunnerBatchReadDefaultTimeout)
}

// prepareDetectStep 为给定task准备预检步骤：无则创建，若已失败则更新。最后返回所有预检步骤，包括已成功的
func (d *Detector) prepareDetectStep(kt *kit.Kit, suborderID string, cfg *table.DetectStepCfg,
	tasks map[int64]*table.DetectTask) (needToSkip, needToExecute []*table.DetectStep, err error) {

	currentStepMap, err := d.getCurrentStepMap(kt, suborderID, cfg.Name, tasks)
	if err != nil {
		logs.Errorf("failed to get detect step host id, order: %s, step: %s, err: %v, rid: %s",
			suborderID, cfg.Name, err, kt.Rid)
		return nil, nil, err
	}

	needToCreate := make([]*table.DetectStep, 0, len(tasks))
	needToRetry := make([]string, 0, len(tasks))

	for _, task := range tasks {
		curStep, ok := currentStepMap[task.HostID]
		if !ok {
			newStep := prepareStepForTask(task, cfg)
			needToCreate = append(needToCreate, newStep)
			continue
		}

		// 未成功、配置要求重试的 加入重试列表
		if curStep.Status != table.DetectStatusSuccess || cfg.RetryOnSuccess {
			needToRetry = append(needToRetry, curStep.ID)
			continue
		}

		// 已存在且成功的，跳过即可
		needToSkip = append(needToSkip, currentStepMap[task.HostID])
	}

	if len(needToCreate) > 0 {
		if err := dao.Set().DetectStep().BatchCreateDetectSteps(kt.Ctx, needToCreate); err != nil {
			logs.Errorf("failed to create detect step %s of order: %s, err: %v, rid: %s",
				cfg.Name, suborderID, err, kt.Rid)
			return nil, nil, err
		}
		needToExecute = append(needToExecute, needToCreate...)
	}

	if len(needToRetry) > 0 {
		// 重设状态到init
		resetSteps, err := d.resetStepInit(kt, suborderID, needToRetry)
		if err != nil {
			logs.Errorf("failed to reset detect step to init, order: %s, step: %s, err: %v, rid: %s",
				suborderID, cfg.Name, err, kt.Rid)
			return nil, nil, err
		}
		needToExecute = append(needToExecute, resetSteps...)
	}

	return needToSkip, needToExecute, nil
}

// getCurrentStepMap 获取当前单据已有的回收步骤map，格式为主机ID->步骤
func (d *Detector) getCurrentStepMap(kt *kit.Kit, suborderID string, stepName table.DetectStepName,
	tasks map[int64]*table.DetectTask) (map[int64]*table.DetectStep, error) {

	// 获取已创建的所有步骤信息
	currentSteps, err := d.listDetectStepByName(kt, suborderID, stepName)
	if err != nil {
		logs.Errorf("failed to list detect step %s of order: %s, err: %v, rid: %s",
			stepName, suborderID, err, kt.Rid)
		return nil, err
	}

	taskIDMap := make(map[string]*table.DetectTask, len(tasks))
	for _, task := range tasks {
		taskIDMap[task.TaskID] = task
	}

	// 主机ID->步骤
	currentStepMap := make(map[int64]*table.DetectStep, len(tasks))
	for _, step := range currentSteps {
		if len(step.AssetID) == 0 || step.HostID <= 0 {
			// 存量数据没有主机id和固资号，从task中补充
			task := taskIDMap[step.TaskID]
			if task == nil {
				logs.Errorf("failed to get task info by task id %s, rid: %s", step.TaskID, kt.Rid)
				return nil, fmt.Errorf("failed to get task info by task id %s", step.TaskID)
			}
			step.HostID = task.HostID
			step.AssetID = task.AssetID
		}
		currentStepMap[step.HostID] = step
	}

	return currentStepMap, nil
}

func (d *Detector) resetStepInit(kt *kit.Kit, suborder string, stepIDs []string) (
	updated []*table.DetectStep, err error) {

	filter := &mapstr.MapStr{
		"id": stepIDs,
	}
	doc := &mapstr.MapStr{
		"status":    table.DetectStatusInit,
		"message":   "init",
		"update_at": time.Now(),
	}
	if err := dao.Set().DetectStep().UpdateDetectStep(kt.Ctx, filter, doc); err != nil {
		logs.Errorf("failed to reset detect step to init, order: %s, err: %v, rid: %s",
			suborder, err, kt.Rid)
		return nil, err
	}
	resetSteps, err := d.listDetectTaskStepByID(kt, suborder, stepIDs)
	if err != nil {
		return nil, err
	}
	return resetSteps, nil
}

func (d *Detector) listDetectStepByName(kt *kit.Kit, suborder string, stepName table.DetectStepName) (
	[]*table.DetectStep, error) {

	filter := map[string]interface{}{
		"suborder_id": suborder,
		"step_name":   stepName,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	steps, err := dao.Set().DetectStep().FindManyDetectStep(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to list recycle step %s of order: %s, err: %v, rid: %s", stepName, suborder, err, kt.Rid)
		return nil, err
	}

	return steps, nil
}
func (d *Detector) listDetectTaskStepByID(kt *kit.Kit, suborder string, ids []string) (
	[]*table.DetectStep, error) {

	filter := map[string]interface{}{
		"suborder_id": suborder,
		"id": map[string]any{
			pkg.BKDBIN: ids,
		},
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	steps, err := dao.Set().DetectStep().FindManyDetectStep(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to list recycle step of order: %s, err: %v, id: %v, rid: %s", suborder, err, ids, kt.Rid)
		return nil, err
	}

	return steps, nil
}

func (d *Detector) batchUpdateTaskStatus(kt *kit.Kit, suborderID string, status table.DetectStatus,
	totalSteps int) error {

	filter := &mapstr.MapStr{
		"suborder_id": suborderID,
	}

	doc := mapstr.MapStr{
		"total_num":   totalSteps,
		"success_num": 0,
		"failed_num":  0,
		"pending_num": totalSteps,
		"message":     "",
		"update_at":   time.Now(),
		"status":      status,
	}

	if err := dao.Set().DetectTask().UpdateDetectTask(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update recycle task status, suborder: %s, status: %s, steps count: %d, err: %v, rid: %s",
			suborderID, status, totalSteps, err, kt.Rid)
		return err
	}
	return nil
}

// 更新recycle task, recycle host 的状态
func (d *Detector) updateRecycleTaskAndHostStatus(kt *kit.Kit, task *table.DetectTask) error {
	filter := &mapstr.MapStr{
		"task_id": task.TaskID,
	}

	doc := mapstr.MapStr{
		"total_num":   task.TotalNum,
		"success_num": task.SuccessNum,
		"failed_num":  task.FailedNum,
		"pending_num": task.PendingNum,
		"message":     task.Message,
		"update_at":   time.Now(),
	}
	if task.PendingNum == 0 {
		var targetStatus = table.DetectStatusFailed
		if task.FailedNum == 0 {
			targetStatus = table.DetectStatusSuccess
		}
		doc["status"] = targetStatus
	}

	if err := dao.Set().DetectTask().UpdateDetectTask(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update recycle task, task id: %s, update: %+v, err: %v, rid: %s",
			task.TaskID, doc, err, kt.Rid)
		return err
	}

	// 如果结束，更新recycle host 的状态
	if task.PendingNum == 0 && task.FailedNum != 0 {
		if err := d.updateRecycleHostStatus(kt, task.SuborderID, task.AssetID, task.IP,
			table.RecycleStatusDetectFailed); err != nil {
			logs.Errorf("failed to update recycle host, task id: %s, host: %s, assetID: %s, err: %v, rid: %s",
				task.TaskID, task.IP, task.AssetID, err, kt.Rid)
			return err
		}
	}

	return nil
}

func (d *Detector) updateRecycleHostStatus(kt *kit.Kit, suborder string, assetID string, ip string,
	status table.RecycleStatus) error {

	filter := &mapstr.MapStr{
		"suborder_id": suborder,
		"asset_id":    assetID,
		"ip":          ip,
	}

	doc := mapstr.MapStr{
		"status":    status,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update recycle host, suborder: %s, ip: %s, asset_id: %s, status: %s, err: %v, rid: %s",
			suborder, ip, assetID, status, err, kt.Rid)
		return err
	}
	return nil
}
