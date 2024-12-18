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

// Package detector implements rejected device detector
// which prevents serious recycle consequences.
package detector

import (
	"context"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/thirdparty/gcsapi"
	"hcm/pkg/thirdparty/l5api"
	"hcm/pkg/thirdparty/ngateapi"
	"hcm/pkg/thirdparty/safetyapi"
	"hcm/pkg/thirdparty/tcaplusapi"
	"hcm/pkg/thirdparty/tgwapi"
	"hcm/pkg/thirdparty/tmpapi"
	"hcm/pkg/thirdparty/uworkapi"
	"hcm/pkg/thirdparty/xshipapi"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/uuid"

	"golang.org/x/sync/errgroup"
)

// Detector detects rejected device for recycle
type Detector struct {
	cc      cmdb.Client
	uwork   uworkapi.UworkClientInterface
	xship   xshipapi.XshipClientInterface
	tmp     tmpapi.TMPClientInterface
	gcs     gcsapi.GcsClientInterface
	tcaplus tcaplusapi.TcaplusClientInterface
	tgw     tgwapi.TgwClientInterface
	l5      l5api.L5ClientInterface
	safety  safetyapi.SafetyClientInterface
	cvm     cvmapi.CVMClientInterface
	tcOpt   cc.TCloudCli
	sops    sopsapi.SopsClientInterface
	ngate   ngateapi.NgateClientInterface

	ctx context.Context
	kt  *kit.Kit
}

// New creates a detector
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client) (*Detector, error) {
	detector := &Detector{
		cc:      esbCli.Cmdb(),
		uwork:   thirdCli.Uwork,
		xship:   thirdCli.Xship,
		tmp:     thirdCli.Tmp,
		gcs:     thirdCli.GCS,
		tcaplus: thirdCli.Tcaplus,
		tgw:     thirdCli.TGW,
		l5:      thirdCli.L5,
		safety:  thirdCli.Safety,
		cvm:     thirdCli.CVM,
		tcOpt:   thirdCli.TencentCloudOpt,
		sops:    thirdCli.Sops,
		ngate:   thirdCli.Ngate,
		ctx:     ctx,
		kt:      &kit.Kit{Ctx: ctx, Rid: uuid.UUID()},
	}

	return detector, nil
}

// CheckDetectStatus checks if detection is finished
func (d *Detector) CheckDetectStatus(subOrderId string) error {
	filter := map[string]interface{}{
		"suborder_id": subOrderId,
		"status": mapstr.MapStr{
			pkg.BKDBNE: table.DetectStatusSuccess,
		},
	}
	cnt, err := dao.Set().DetectTask().CountDetectTask(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get detection task count, err: %v, subOrderId: %s", err, subOrderId)
		return err
	}

	if cnt == 0 {
		return nil
	}

	filterOrder := mapstr.MapStr{
		"suborder_id": subOrderId,
	}
	update := mapstr.MapStr{
		"failed_num": cnt,
		"update_at":  time.Now(),
	}
	// ignore and continue when update failed_num error
	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filterOrder, &update); err != nil {
		logs.Errorf("failed to update recycle order, ignore and continue when update failed_num error, "+
			"subOrderId: %s, err: %v", subOrderId, err)
	}

	logs.Errorf("recycle order detection failed, for detection tasks is not success, subOrderId: %s, failedStepNum: %d",
		subOrderId, cnt)
	return fmt.Errorf("recycle order detection failed, for detection tasks is not success, subOrderId: %s, failedStepNum"+
		": %d", subOrderId, cnt)
}

// DealRecycleOrder deals with recycle order by running detection tasks
func (d *Detector) DealRecycleOrder(orderId string) error {
	// get tasks by order id

	taskInfos, err := d.getRecycleTasks(orderId)
	if err != nil {
		logs.Errorf("failed to get recycle tasks by order id: %d, err: %v", orderId, err)
		return err
	}
	// run recycle tasks
	eg := errgroup.Group{}
	// 每个主机都会创建一个recycle task，这里防止无限制并发
	eg.SetLimit(5)
	for _, task := range taskInfos {
		taskInfo := task
		eg.Go(func() error {
			d.RunRecycleTask(taskInfo, 0)
			return nil
		})
	}
	return eg.Wait()
}

// DealRecycleTask deals with recycle task
func (d *Detector) DealRecycleTask(taskId string) error {
	// get task by task id
	task, err := d.getRecycleTaskById(taskId)
	if err != nil {
		logs.Errorf("failed to get recycle task by task id: %d, err: %v", taskId, err)
		return err
	}

	go d.RunRecycleTask(task, 0)

	return nil
}

// getRecycleTasks gets recycle tasks by recycle order id
func (d *Detector) getRecycleTasks(orderId string) ([]*table.DetectTask, error) {
	filter := map[string]interface{}{
		"suborder_id": orderId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}
	tasks, err := dao.Set().DetectTask().FindManyDetectTask(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle tasks by order id: %s", orderId)
		return nil, err
	}

	return tasks, nil
}

// getRecycleTaskById gets recycle task by recycle task id
func (d *Detector) getRecycleTaskById(taskId string) (*table.DetectTask, error) {
	filter := map[string]interface{}{
		"task_id": taskId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}
	tasks, err := dao.Set().DetectTask().FindManyDetectTask(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle tasks by task id: %s", taskId)
		return nil, err
	}

	cnt := len(tasks)
	if cnt != 1 {
		logs.Errorf("get invalid recycle task count %d != 1 by task id %s", cnt, taskId)
		return nil, fmt.Errorf("get invalid recycle task count %d != 1 by task id %s", cnt, taskId)
	}

	return tasks[0], nil
}

// RunRecycleTask runs recycle task
func (d *Detector) RunRecycleTask(task *table.DetectTask, startStep uint) {
	// check task status
	if task.Status == table.DetectStatusSuccess || task.Status == table.DetectStatusRunning {
		logs.Infof("recycle task need not dispatch, taskId: %s, status: %s", task.TaskID, task.Status)
		return
	}

	// get recycle steps
	steps, err := d.getRecycleSteps()
	if err != nil {
		logs.Errorf("failed to run recycle task, taskId: %s, err: %v", task.TaskID, err)
		return
	}

	// init task status
	if err := d.initTaskStatus(task); err != nil {
		logs.Errorf("failed to init recycle task status, task id: %s, err: %v", task.TaskID, err)
		return
	}

	// run recycle steps in serial
	total := uint(len(steps))
	success, failed := task.SuccessNum, task.FailedNum
	if startStep == 0 {
		success, failed = 0, 0
	}

	var lastErr error = nil
	if failed != 0 {
		lastErr = fmt.Errorf("recycle some step failed")
	}

	for i := startStep; i < total; i++ {
		step := steps[i]
		errRun := d.runRecycleStep(task, step)
		if errRun != nil {
			logs.Errorf("failed to run recycle step, step name: %s, taskId: %s, err: %v", step.Name, task.TaskID,
				errRun)
			lastErr = errRun
			failed++
		} else {
			success++
		}
		if err = d.updateTaskProgress(task, total, success, failed); err != nil {
			logs.Errorf("recycler:logics:cvm:runRecycleTask:failed, failed to update recycle task status, "+
				"taskId: %s, subOrderID: %s, err: %v", task.TaskID, task.SuborderID, err)
		}
	}

	// update task status
	if err = d.updateRecycleTask(task, lastErr); err != nil {
		logs.Errorf("recycler:logics:cvm:runRecycleTask:failed, failed to update recycle task status, "+
			"taskId: %s, err: %v", task.TaskID, err)
	}

	// update recycle task
	if err = d.updateRecycleHost(task.SuborderID, task.IP, lastErr); err != nil {
		logs.Errorf("recycler:logics:cvm:runRecycleTask:failed, failed to update recycle host: %s, subOrderID: %s, "+
			"err: %v", task.IP, task.SuborderID, err)
	}

	logs.Infof("finish recycle order detect step, subOrderId: %s", task.SuborderID)
	return
}

func (d *Detector) runRecycleStep(task *table.DetectTask, stepCfg *table.DetectStepCfg) error {
	// check step status
	stepId := fmt.Sprintf("%s-%d", task.TaskID, stepCfg.ID)
	steps, err := d.getRecycleTaskStep(stepId)
	if err != nil {
		logs.Errorf("failed to get recycle step, task id: %s, step name: %s, err: %v", task.TaskID, stepCfg.Name, err)
		return err
	}

	step := new(table.DetectStep)
	cnt := len(steps)
	if cnt > 1 {
		logs.Errorf("recycler:logics:cvm:runRecycleStep:failed, failed to get recycle step, for invalid count > 1, "+
			"step id: %s, taskID: %s, subOrderID: %s, IP: %s", stepId, task.TaskID, task.SuborderID, task.IP)
		return fmt.Errorf("failed to get recycle step, for invalid count > 1, step id: %s", stepId)
	} else if cnt == 1 {
		step = steps[0]
	} else {
		// init step status
		step, err = d.initRecycleStep(task, stepCfg)
		if err != nil {
			logs.Errorf("failed to init recycle step, task id: %s, step name: %s, err: %v", task.TaskID, stepCfg.Name,
				err)
			return err
		}
	}

	// can not skip pre-check step
	if step.Status == table.DetectStatusSuccess && step.StepName != table.StepPreCheck {
		logs.Infof("recycler:logics:cvm:runRecycleStep:has success, step %s already success, skip", stepId)
		return nil
	} else if step.Status == table.DetectStatusRunning {
		logs.Errorf("recycler:logics:cvm:runRecycleStep:running, step %s is running, can not execute again", stepId)
		return fmt.Errorf("step %s is running, can not execute again", stepId)
	}

	// update step status to running
	if err = d.updateRecycleStep(step, table.DetectStatusRunning, 0, "running", ""); err != nil {
		logs.Errorf("failed to update recycle step, step id: %s, err: %v", step.ID, err)
		return err
	}

	// execute step
	attempt, exeInfo, errExec := d.executeRecycleStep(step, stepCfg.Retry)
	if errExec != nil {
		logs.Errorf("recycler:logics:cvm:runRecycleStep:failed, failed to execute recycle step, step id: %s, "+
			"stepName: %s, err: %v, subOrderID: %s, IP: %s", step.ID, step.StepName, errExec, task.SuborderID, task.IP)
	} else {
		logs.Infof("recycler:logics:cvm:runRecycleStep:success, success to execute recycle step, step id: %s, "+
			"stepName: %s, subOrderID: %s, IP: %s", step.ID, step.StepName, task.SuborderID, task.IP)
	}

	// update step status
	if errExec != nil {
		if err = d.updateRecycleStep(step, table.DetectStatusFailed, attempt, errExec.Error(), exeInfo); err != nil {
			logs.Errorf("failed to update recycle step, step id: %s, err: %v", step.ID, err)
			return err
		}
	} else {
		if err = d.updateRecycleStep(step, table.DetectStatusSuccess, attempt, "success", exeInfo); err != nil {
			logs.Errorf("failed to update recycle step, step id: %s, err: %v", step.ID, err)
			return err
		}
	}

	// return execution result
	return errExec
}

func (d *Detector) initTaskStatus(task *table.DetectTask) error {
	task.Status = table.DetectStatusRunning

	filter := &mapstr.MapStr{
		"task_id": task.TaskID,
	}

	doc := &mapstr.MapStr{
		"status":     task.Status,
		"status_seq": table.DetectStatusSeqRunning,
		"message":    "running",
		"update_at":  time.Now(),
	}

	if err := dao.Set().DetectTask().UpdateDetectTask(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update recycle task, ip: %s, update: %+v, err: %v", task.IP, doc, err)
		return err
	}

	return nil
}

func (d *Detector) updateRecycleTask(task *table.DetectTask, lastErr error) error {
	task.Status = table.DetectStatusSuccess

	filter := &mapstr.MapStr{
		"task_id": task.TaskID,
	}

	if lastErr != nil {
		task.Status = table.DetectStatusFailed
	}

	seq, ok := table.DetectStatus2Seq[task.Status]
	if !ok {
		logs.Warnf("found no recycle status seq by %v", task.Status)
		seq = 0
	}
	doc := &mapstr.MapStr{
		"status":     task.Status,
		"status_seq": seq,
		"update_at":  time.Now(),
	}

	if err := dao.Set().DetectTask().UpdateDetectTask(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update detection task, task id: %s, update: %+v, err: %v", task.TaskID, doc, err)
		return err
	}

	return nil
}

func (d *Detector) updateTaskProgress(task *table.DetectTask, total, success, failed uint) error {
	filter := &mapstr.MapStr{
		"task_id": task.TaskID,
	}

	doc := &mapstr.MapStr{
		"total_num":   total,
		"success_num": success,
		"failed_num":  failed,
		"update_at":   time.Now(),
	}

	if err := dao.Set().DetectTask().UpdateDetectTask(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update recycle task, task id: %s, update: %+v, err: %v", task.TaskID, doc, err)
		return err
	}

	return nil
}

func (d *Detector) updateRecycleHost(orderId, ip string, lastErr error) error {
	status := table.RecycleStatusDetecting

	filter := &mapstr.MapStr{
		"suborder_id": orderId,
		"ip":          ip,
	}

	if lastErr != nil {
		status = table.RecycleStatusDetectFailed
	}

	doc := &mapstr.MapStr{
		"status":    status,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), filter, doc); err != nil {
		return err
	}

	return nil
}

func (d *Detector) getRecycleSteps() ([]*table.DetectStepCfg, error) {
	filter := map[string]interface{}{
		"enable": true,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	steps, err := dao.Set().DetectStepCfg().GetDetectStepConfig(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle step config, err: %v", err)
		return nil, err
	}

	return steps, nil
}

func (d *Detector) getRecycleTaskStep(stepId string) ([]*table.DetectStep, error) {
	filter := map[string]interface{}{
		"id": stepId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}

	insts, err := dao.Set().DetectStep().FindManyDetectStep(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle task step, err: %v, step id: %s", err, stepId)
		return nil, err
	}

	return insts, nil
}

func (d *Detector) initRecycleStep(task *table.DetectTask, cfg *table.DetectStepCfg) (*table.DetectStep, error) {
	now := time.Now()
	step := &table.DetectStep{
		OrderID:    task.OrderID,
		SuborderID: task.SuborderID,
		TaskID:     task.TaskID,
		ID:         fmt.Sprintf("%s-%d", task.TaskID, cfg.ID),
		StepID:     cfg.Sequence,
		StepName:   cfg.Name,
		StepDesc:   cfg.Description,
		IP:         task.IP,
		User:       task.User,
		RetryTime:  0,
		Status:     table.DetectStatusInit,
		Message:    "",
		StartAt:    now,
		EndAt:      now,
		CreateAt:   now,
		UpdateAt:   now,
	}

	if err := dao.Set().DetectStep().CreateDetectStep(context.Background(), step); err != nil {
		logs.Errorf("failed to save step, step id: %s", step.ID)
		return nil, fmt.Errorf("failed to save step, step id: %s", step.ID)
	}

	return step, nil
}

func (d *Detector) executeRecycleStep(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error

	switch step.StepName {
	case table.StepPreCheck:
		attempt, exeInfo, err = d.preCheck(step, retry)
	case table.StepCheckUwork:
		attempt, exeInfo, err = d.checkUwork(step, retry)
	case table.StepCheckGCS:
		attempt, exeInfo, err = d.checkGCS(step, retry)
	case table.StepBasicCheck:
		attempt, exeInfo, err = d.basicCheck(step, retry)
	case table.StepCheckOwner:
		attempt, exeInfo, err = d.checkOwner(step, retry)
	case table.StepCvmCheck:
		attempt, exeInfo, err = d.cvmCheck(step, retry)
	case table.StepCheckSafety:
		attempt, exeInfo, err = d.checkSecurityBaseline(step, retry)
	case table.StepCheckReturn:
		attempt, exeInfo, err = d.checkReturn(step, retry)
	case table.StepCheckProcess:
		attempt, exeInfo, err = d.checkProcess(step, retry)
	case table.StepCheckPmOuterIP: // 物理机外网IP回收及清理检查
		attempt, exeInfo, err = d.checkPmOuterIP(step, retry)
	default:
		logs.Errorf("unknown recycle step %s", step.StepName)
		err = fmt.Errorf("unknown recycle step %s", step.StepName)
	}

	return attempt, exeInfo, err
}

func (d *Detector) updateRecycleStep(step *table.DetectStep, status table.DetectStatus, attempt int, msg,
	log string) error {

	filter := &mapstr.MapStr{
		"id": step.ID,
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
	}

	if err := dao.Set().DetectStep().UpdateDetectStep(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update recycle step, step id: %s, update: %+v, err: %v", step.ID, doc, err)
		return err
	}

	return nil
}
