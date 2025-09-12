/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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
	"sync/atomic"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

const (
	defaultMaxBatchSize = 500
	onlyOneBatchSize    = 1
)

type stepBatch struct {
	kt    *kit.Kit
	steps []*StepMeta
}

type handleFunc func(kt *kit.Kit, steps []*StepMeta, resultHandler StepResultHandler, cliSet *cliSet)

type baseWorkGroup struct {
	stepName      enumor.DetectStepName
	resultHandler StepResultHandler
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	handleFunc    handleFunc
	cliSet        *cliSet
}

// MaxBatchSize 最大批量数
func (b *baseWorkGroup) MaxBatchSize() int {
	return defaultMaxBatchSize
}

func newBaseWorkGroup(stepName enumor.DetectStepName, resultHandler StepResultHandler, workerNum int,
	handleFunc handleFunc, cliSet *cliSet) baseWorkGroup {

	return baseWorkGroup{
		stepName:      stepName,
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
		handleFunc:    handleFunc,
		cliSet:        cliSet,
	}
}

// Start 启动worker
func (b *baseWorkGroup) Start(kt *kit.Kit) {
	if !b.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < b.currency; i++ {
		subKit := kt.NewSubKit()
		go b.consume(subKit, i)
	}
}

func (b *baseWorkGroup) consume(kt *kit.Kit, idx int) {
	logs.Infof("%s consume worker %d start, rid: %s", b.stepName, idx, kt.Rid)
	defer logs.Infof("%s consume worker %d exit, rid: %s", b.stepName, idx, kt.Rid)

	for {
		select {
		case batch := <-b.stepBatchChan:
			logs.V(4).Infof("%s worker %d got steps: %d:%s, rid: %s", b.stepName, idx, len(batch.steps), batch.steps,
				kt.Rid)
			b.handleFunc(batch.kt, batch.steps, b.resultHandler, b.cliSet)
		case <-kt.Ctx.Done():
			return
		}
	}
}

// Submit 提交任务
func (b *baseWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	b.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

type querySopsResultTask struct {
	kt          *kit.Kit
	step        *StepMeta
	bizID       int64
	taskID      int64
	taskUrl     string
	createdTime time.Time
	exeInfos    []string
}
