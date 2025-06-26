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
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/bkdbm"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/l5api"
	"hcm/pkg/thirdparty/ngateapi"
	"hcm/pkg/thirdparty/safetyapi"
	"hcm/pkg/thirdparty/tcaplusapi"
	"hcm/pkg/thirdparty/tgwapi"
	"hcm/pkg/thirdparty/tmpapi"
	"hcm/pkg/thirdparty/xrayapi"
	"hcm/pkg/thirdparty/xshipapi"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// Detector detects rejected device for recycle
type Detector struct {
	cc      cmdb.Client
	xray    xrayapi.XrayClientInterface
	xship   xshipapi.XshipClientInterface
	tmp     tmpapi.TMPClientInterface
	tcaplus tcaplusapi.TcaplusClientInterface
	tgw     tgwapi.TgwClientInterface
	l5      l5api.L5ClientInterface
	safety  safetyapi.SafetyClientInterface
	cvm     cvmapi.CVMClientInterface
	tcOpt   cc.TCloudCli
	sops    sopsapi.SopsClientInterface
	ngate   ngateapi.NgateClientInterface
	bkDbm   bkdbm.Client

	cliSet *client.ClientSet

	// 仅能作为后台操作kit，不能用到单个单据的执行
	backendKit *kit.Kit

	StepExecutors map[table.DetectStepName]StepExecutor
}

// New creates a detector
func New(ctx context.Context, thirdCli *thirdparty.Client, cmdbCli cmdb.Client, cliSet *client.ClientSet) (
	*Detector, error) {

	kt := core.NewBackendKit()
	kt.Ctx = ctx
	detector := &Detector{
		backendKit: kt,
		cc:         cmdbCli,
		xray:       thirdCli.Xray,
		xship:      thirdCli.Xship,
		tmp:        thirdCli.Tmp,
		tcaplus:    thirdCli.Tcaplus,
		tgw:        thirdCli.TGW,
		l5:         thirdCli.L5,
		safety:     thirdCli.Safety,
		cvm:        thirdCli.CVM,
		tcOpt:      thirdCli.TencentCloudOpt,
		sops:       thirdCli.Sops,
		ngate:      thirdCli.Ngate,
		bkDbm:      thirdCli.BkDbm,
		cliSet:     cliSet,
	}
	err := detector.initStepExecutor(detector.backendKit)
	if err != nil {
		logs.Errorf("failed to init step executors, err: %v", err)
		return nil, err
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

// getDetectTasks 查询预检任务，每个主机会有一个DetectTask
func (d *Detector) getDetectTasks(kt *kit.Kit, orderId string) ([]*table.DetectTask, error) {
	filter := map[string]interface{}{
		"suborder_id": orderId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}
	tasks, err := dao.Set().DetectTask().FindManyDetectTask(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle tasks by order id: %s, rid: %s", orderId, kt.Rid)
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

func (d *Detector) getDetectStepConfigs() ([]*table.DetectStepCfg, error) {
	filter := map[string]interface{}{
		"enable": true,
	}
	page := metadata.BasePage{
		Sort:  "sequence",
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

func prepareStepForTask(task *table.DetectTask, cfg *table.DetectStepCfg) *table.DetectStep {
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
		HostID:     task.HostID,
		AssetID:    task.AssetID,
		User:       task.User,
		RetryTime:  0,
		Status:     table.DetectStatusInit,
		Message:    "",
		StartAt:    now,
		EndAt:      now,
		CreateAt:   now,
		UpdateAt:   now,
	}
	return step
}

// TODO 原逻辑参考 改造完删除
func (d *Detector) executeRecycleStep(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error

	switch step.StepName {
	// case table.StepPreCheck:
	// 	attempt, exeInfo, err = d.preCheck(step, retry)
	case table.StepCheckUwork:
		attempt, exeInfo, err = d.checkUwork(step, retry)
	case table.StepCheckTcaplus:
		attempt, exeInfo, err = d.checkTcaplus(step, retry)
	case table.StepCheckDBM:
		attempt, exeInfo, err = d.checkDbm(step, retry)
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
	// case table.StepCheckProcess:
	// 	attempt, exeInfo, err = d.checkProcess(step, retry)
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

func (d *Detector) fillTaskHostIDMap(kt *kit.Kit, taskList []*table.DetectTask,
	suborderID string) (map[int64]*table.DetectTask, error) {

	hostIDTaskMap := make(map[int64]*table.DetectTask, len(taskList))
	taskIPMap := make(map[string]*table.DetectTask)
	ipList := make([]string, 0)

	for _, task := range taskList {
		if task.HostID >= 0 {
			hostIDTaskMap[task.HostID] = task
			continue
		}
		taskIPMap[task.IP] = task
		ipList = append(ipList, task.IP)
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKMaxInstanceLimit,
	}

	for _, ipBatch := range slice.Split(ipList, pkg.BKMaxInstanceLimit) {
		filter := map[string]interface{}{
			"suborder_id": suborderID,
			"ip": mapstr.MapStr{
				pkg.BKDBIN: ipBatch,
			},
		}
		hostList, err := dao.Set().RecycleHost().FindManyRecycleHost(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get recycle hosts, err: %v", err)
			return nil, err
		}
		for _, inst := range hostList {
			task := taskIPMap[inst.IP]
			if task == nil {
				logs.Errorf("get host by ip got unknown ip: %s, rid: %s", inst.IP, kt.Rid)
				return nil, fmt.Errorf("get host by ip got unknown ip: %s, rid: %s", inst.IP, kt.Rid)
			}
			task.AssetID = inst.AssetID
			task.HostID = inst.HostID
			hostIDTaskMap[inst.HostID] = task
			delete(taskIPMap, inst.IP)
		}
	}

	if len(taskIPMap) > 0 {
		logs.Errorf("failed to get host id by ip, task ip map: %+v, rid: %s", taskIPMap, kt.Rid)
		return nil, fmt.Errorf("failed to get host id by ip, task ip map: %+v, rid: %s", taskIPMap, kt.Rid)
	}
	return hostIDTaskMap, nil
}
