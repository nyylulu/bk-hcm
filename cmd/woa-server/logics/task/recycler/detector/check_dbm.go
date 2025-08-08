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

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkdbm"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
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
	hostIDs := slice.Map(steps, func(step *StepMeta) int64 { return step.Step.HostID })

	req := &bkdbm.ListMachinePool{HostIDs: hostIDs, Offset: 0, Limit: int64(len(steps))}
	resp, err := t.dbmCLi.QueryMachinePool(kt, req)
	if err != nil {
		t.HandleResult(kt, steps, err, fmt.Sprintf("check DBM failed, err: %s", err), true)
		return
	}
	existsMap := make(map[string][]*bkdbm.MachinePoolResult, len(resp.Results))
	for _, item := range resp.Results {
		existsMap[item.IP] = append(existsMap[item.IP], cvt.ValToPtr(item))
	}
	for _, step := range steps {
		if _, ok := existsMap[step.Step.IP]; ok {
			str := structToStr(existsMap[step.Step.IP])
			terr := fmt.Errorf("该主机在DBM中使用，不允许回收: %s", step.Step.IP)
			t.HandleResult(kt, []*StepMeta{step}, terr, str, false)
			continue
		}
		t.HandleResult(kt, []*StepMeta{step}, nil, fmt.Sprintf("dbm response: %+v", existsMap[step.Step.IP]), false)
	}
}
