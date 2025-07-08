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
	"hcm/pkg/thirdparty/tcaplusapi"
	"hcm/pkg/tools/slice"
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
	ips := slice.Map(steps, func(step *StepMeta) string { return step.Step.IP })
	resp, err := t.tcaplus.CheckTcaplus(kt, ips)
	logs.Infof("DEBUG: check tcaplus resp: %+v, err: %v, rid: %s", resp, err, kt.Rid)
	if err != nil {
		log := fmt.Sprintf("check tcaplus failed, err: %s", err)
		t.HandleResult(kt, steps, err, log, true)
		return
	}
	existsMap := make(map[string][]*tcaplusapi.TcaplusItem, len(resp.Data))
	for _, item := range resp.Data {
		existsMap[item.IP] = append(existsMap[item.IP], item)
	}
	for _, step := range steps {
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
