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

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkdbm"
	cvt "hcm/pkg/tools/converter"
)

// CheckDBMMaxBatchSize DBM 暂未限定长度，暂定200
const CheckDBMMaxBatchSize = 200

// CheckDBMWorkGroup 查询bk dbm的主机池，如果该主机在dbm主机池里面，则不允许回收
type CheckDBMWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	dbmCLi        bkdbm.Client
	resultHandler StepResultHandler
}

// MaxBatchSize ...
func (t *CheckDBMWorkGroup) MaxBatchSize() int {
	return CheckDBMMaxBatchSize
}

// NewCheckDBMWorkGroup ...
func NewCheckDBMWorkGroup(dbmCli bkdbm.Client, resultHandler StepResultHandler, workerNum int) *CheckDBMWorkGroup {

	return &CheckDBMWorkGroup{
		dbmCLi:        dbmCli,
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
	}
}

// HandleResult ...
func (t *CheckDBMWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string, needRetry bool) {
	t.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

// Submit 提交检查
func (t *CheckDBMWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	t.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// Start 启动worker
func (t *CheckDBMWorkGroup) Start(kt *kit.Kit) {
	if !t.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < t.currency; i++ {
		subKit := kt.NewSubKit()
		go t.queryWorker(subKit, i)
	}
}

func (t *CheckDBMWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("check DBM query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check DBM query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-t.stepBatchChan:
			logs.V(4).Infof("check DBM worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			t.check(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

func (t *CheckDBMWorkGroup) check(kt *kit.Kit, steps []*StepMeta) {
	var hostIDs []int64
	var newSteps []*StepMeta
	for _, step := range steps {
		if step.Step == nil {
			logs.Errorf("IdleCheck:%s:failed to check dbm, step.Step is nil, rid: %s", table.StepCheckDBM, kt.Rid)
			err := fmt.Errorf("IdleCheck:%s, step.Step is nil", table.StepCheckDBM)
			t.HandleResult(kt, []*StepMeta{step}, err, err.Error(), false)
			continue
		}
		// 该主机对应的步骤已被设置为跳过
		if step.Step.Skip == enumor.DetectStepSkipYes {
			logs.Infof("IdleCheck:%s:SKIP ONE, subOrderID: %s, IP: %s, stepMeta: %+v, rid: %s",
				table.StepCheckDBM, step.Step.SuborderID, step.Step.IP, cvt.PtrToVal(step), kt.Rid)
			t.HandleResult(kt, []*StepMeta{step}, nil, "跳过", false)
			continue
		}
		hostIDs = append(hostIDs, step.Step.HostID)
		newSteps = append(newSteps, step)
	}

	// 所有步骤都跳过了该步骤，则直接返回
	if len(hostIDs) == 0 {
		logs.Warnf("IdleCheck:%s:SKIP ALL, steps: %+v, rid: %s", table.StepCheckDBM, cvt.PtrToSlice(steps), kt.Rid)
		return
	}

	req := &bkdbm.ListMachinePool{HostIDs: hostIDs, Offset: 0, Limit: int64(len(newSteps))}
	resp, err := t.dbmCLi.QueryMachinePool(kt, req)
	if err != nil {
		t.HandleResult(kt, newSteps, err, fmt.Sprintf("check DBM failed, err: %s", err), true)
		return
	}
	existsMap := make(map[string][]*bkdbm.MachinePoolResult, len(resp.Results))
	for _, item := range resp.Results {
		existsMap[item.IP] = append(existsMap[item.IP], cvt.ValToPtr(item))
	}
	for _, step := range newSteps {
		if _, ok := existsMap[step.Step.IP]; ok {
			str := structToStr(existsMap[step.Step.IP])
			terr := fmt.Errorf("该主机在DBM中使用，不允许回收: %s", step.Step.IP)
			t.HandleResult(kt, []*StepMeta{step}, terr, str, false)
			continue
		}
		t.HandleResult(kt, []*StepMeta{step}, nil, fmt.Sprintf("dbm response: %+v", existsMap[step.Step.IP]), false)
	}
}
