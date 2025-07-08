/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
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

	"hcm/pkg"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
)

// CheckOwnerMaxBatchSize 检查是否包含虚拟子机
const CheckOwnerMaxBatchSize = pkg.BKMaxInstanceLimit / 2

// CheckOwnerWorkGroup 查询bk Owner的主机池，如果该主机在Owner主机池里面，则不允许回收
type CheckOwnerWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	cc            CmdbOperator
	resultHandler StepResultHandler
}

// MaxBatchSize ...
func (t *CheckOwnerWorkGroup) MaxBatchSize() int {
	return CheckOwnerMaxBatchSize
}

// NewCheckOwnerWorkGroup ...
func NewCheckOwnerWorkGroup(cc cmdb.Client, resultHandler StepResultHandler,
	workerNum int) *CheckOwnerWorkGroup {

	return &CheckOwnerWorkGroup{
		cc:            NewCmdbOperator(cc),
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
	}
}

// HandleResult ...
func (t *CheckOwnerWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string,
	needRetry bool) {

	t.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

// Submit 提交检查
func (t *CheckOwnerWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	t.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// Start 启动worker
func (t *CheckOwnerWorkGroup) Start(kt *kit.Kit) {
	if !t.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < t.currency; i++ {
		subKit := kt.NewSubKit()
		go t.queryWorker(subKit, i)
	}
}

func (t *CheckOwnerWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("check owner query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check owner query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-t.stepBatchChan:
			logs.V(4).Infof("check owner worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			t.check(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

func (t *CheckOwnerWorkGroup) check(kt *kit.Kit, steps []*StepMeta) {
	assetIDs := slice.Map(steps, func(step *StepMeta) string { return step.Step.AssetID })

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_svr_owner_asset_id",
						Operator: querybuilder.OperatorIn,
						Value:    assetIDs,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_svr_owner_asset_id",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	hostVmMap := make(map[string][]*cmdb.Host)

	for {
		resp, err := t.cc.ListHost(kt, req)
		if err != nil {
			logs.Errorf("failed to get cc host info by bk_svr_owner_asset_id, err: %v", err)
			e := fmt.Errorf("failed to get cc host info by bk_svr_owner_asset_id, err: %v", err)
			t.HandleResult(kt, steps, e, err.Error(), true)
			return
		}
		for _, host := range resp.Info {
			hostVmMap[host.BKSvrOwnerAssetID] = append(hostVmMap[host.BKSvrOwnerAssetID], cvt.ValToPtr(host))
		}
		if len(resp.Info) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += req.Page.Limit
	}
	for _, step := range steps {
		if _, ok := hostVmMap[step.Step.AssetID]; ok {
			hostVmStr := structToStr(hostVmMap[step.Step.AssetID])
			err := fmt.Errorf("host has %d vm: %s", len(hostVmMap[step.Step.AssetID]), hostVmStr)
			t.HandleResult(kt, []*StepMeta{step}, err, hostVmStr, false)
			continue
		}
		log := fmt.Sprintf("no vm")
		t.HandleResult(kt, []*StepMeta{step}, nil, log, false)
	}
}
