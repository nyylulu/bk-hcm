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
	"hcm/pkg/thirdparty/tcaplusapi"
	cvt "hcm/pkg/tools/converter"
)

// CheckTcaplusBatchSize iplist参数建议限制200
const CheckTcaplusBatchSize = tcaplusapi.TcapulsCheckIPExistsMaxLength

// CheckTcaplusWorkGroup 检查主机是否存在tcaplus
type CheckTcaplusWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	tcaplus       tcaplusapi.TcaplusClientInterface
	resultHandler StepResultHandler
}

// MaxBatchSize iplist参数建议限制200
func (t *CheckTcaplusWorkGroup) MaxBatchSize() int {
	return CheckTcaplusBatchSize
}

// NewCheckTcaplusWorkGroup ...
func NewCheckTcaplusWorkGroup(tcaplus tcaplusapi.TcaplusClientInterface, resultHandler StepResultHandler,
	workerNum int) *CheckTcaplusWorkGroup {

	return &CheckTcaplusWorkGroup{
		tcaplus:       tcaplus,
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
	}
}

// HandleResult ...
func (t *CheckTcaplusWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string,
	needRetry bool) {
	t.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

// Submit 提交检查
func (t *CheckTcaplusWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	t.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// Start 启动worker
func (t *CheckTcaplusWorkGroup) Start(kt *kit.Kit) {
	if !t.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < t.currency; i++ {
		subKit := kt.NewSubKit()
		go t.queryWorker(subKit, i)
	}
}

func (t *CheckTcaplusWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("check tcaplus query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check tcaplus query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-t.stepBatchChan:
			logs.V(4).Infof("check tcaplus worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			t.check(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

func (t *CheckTcaplusWorkGroup) check(kt *kit.Kit, steps []*StepMeta) {
	var ips []string
	var newSteps []*StepMeta
	for _, step := range steps {
		if step.Step == nil {
			logs.Errorf("IdleCheck:%s:failed to check tcaplus, step.Step is nil, rid: %s",
				table.StepCheckTcaplus, kt.Rid)
			err := fmt.Errorf("IdleCheck:%s, step.Step is nil", table.StepCheckTcaplus)
			t.HandleResult(kt, []*StepMeta{step}, err, err.Error(), false)
			continue
		}
		// 该主机对应的步骤已被设置为跳过
		if step.Step.Skip == enumor.DetectStepSkipYes {
			logs.Infof("IdleCheck:%s:SKIP ONE, subOrderID: %s, IP: %s, stepMeta: %+v, rid: %s",
				table.StepCheckTcaplus, step.Step.SuborderID, step.Step.IP, cvt.PtrToVal(step), kt.Rid)
			t.HandleResult(kt, []*StepMeta{step}, nil, "跳过", false)
			continue
		}
		ips = append(ips, step.Step.IP)
		newSteps = append(newSteps, step)
	}

	// 所有步骤都跳过了该步骤，则直接返回
	if len(ips) == 0 {
		logs.Warnf("IdleCheck:%s:SKIP ALL, steps: %+v, rid: %s", table.StepCheckTcaplus, cvt.PtrToSlice(steps), kt.Rid)
		return
	}

	resp, err := t.tcaplus.CheckTcaplus(kt, ips)
	if err != nil {
		log := fmt.Sprintf("check tcaplus failed, err: %s", err)
		t.HandleResult(kt, newSteps, err, log, true)
		return
	}
	existsMap := make(map[string][]*tcaplusapi.TcaplusItem, len(resp.Data))
	for _, item := range resp.Data {
		existsMap[item.IP] = append(existsMap[item.IP], item)
	}
	for _, step := range newSteps {
		if item, ok := existsMap[step.Step.IP]; ok {
			str := structToStr(item)
			terr := fmt.Errorf("%s found in tcaplus: %s", step.Step.IP, str)
			t.HandleResult(kt, []*StepMeta{step}, terr, terr.Error(), false)
			continue
		}
		log := fmt.Sprintf("%s not found in tcaplus, msg: %s", step.Step.IP, resp.Msg)
		t.HandleResult(kt, []*StepMeta{step}, nil, log, false)
	}
}
