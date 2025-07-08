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
	"strings"
	"sync/atomic"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/classifier"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
)

// PrecheckMaxBatchSize 检查CC模块和负责人最大批次大小
const PrecheckMaxBatchSize = 500

// PreCheckWorkGroup 检查CC模块和负责人以及业务是否发生变化
type PreCheckWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	cc            CmdbOperator
	resultHandler StepResultHandler
}

// MaxBatchSize ...
func (p *PreCheckWorkGroup) MaxBatchSize() int {
	return PrecheckMaxBatchSize
}

// NewPreCheckWorkGroup ...
func NewPreCheckWorkGroup(cc cmdb.Client, resultHandler StepResultHandler, workerNum int) *PreCheckWorkGroup {
	return &PreCheckWorkGroup{
		cc:            NewCmdbOperator(cc),
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
	}
}

// HandleResult ...
func (p *PreCheckWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string, needRetry bool) {
	p.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

// Start 启动worker
func (p *PreCheckWorkGroup) Start(kt *kit.Kit) {
	if !p.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < p.currency; i++ {
		subKit := kt.NewSubKit()
		go p.queryWorker(subKit, i)
	}
}

func (p *PreCheckWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("pre-check query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("pre-check query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-p.stepBatchChan:
			logs.V(4).Infof("pre-check worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			p.Run(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

// Run 执行检查
func (p *PreCheckWorkGroup) Run(kt *kit.Kit, steps []*StepMeta) {
	if len(steps) == 0 {
		logs.Warnf("pre-check worker receive empty steps, rid: %s", kt.Rid)
		return
	}

	hostMap := make(map[int64][]*HostExecInfo, len(steps))
	for _, step := range steps {
		hostMap[step.Step.HostID] = append(hostMap[step.Step.HostID], &HostExecInfo{StepMeta: step})
	}
	hostIDs := maps.Keys(hostMap)

	worker := preCheckWorker{
		hostMap:       hostMap,
		cc:            p.cc,
		resultHandler: p,
	}
	worker.checkAll(kt, hostIDs)
}

// Submit 提交检查
func (p *PreCheckWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	p.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// preCheckWorker worker
type preCheckWorker struct {
	cc            CmdbOperator
	hostMap       map[int64][]*HostExecInfo
	resultHandler StepResultHandler
}

func (w *preCheckWorker) checkAll(kt *kit.Kit, hostIDs []int64) {
	// 1. 检查操作人是否为主机负责人或备份负责人
	opSuccessIDs := w.checkOperator(kt, hostIDs)

	if len(opSuccessIDs) == 0 {
		return
	}

	// 2. 检查是否在待回收模块
	w.checkHostModule(kt, opSuccessIDs)
}

// 检查操作人是否为主机业务主备负责人
func (w *preCheckWorker) checkOperator(kt *kit.Kit, hostIDs []int64) (succeed []int64) {
	leftHostIDSet := make(map[int64]struct{}, len(hostIDs))
	for _, hostID := range hostIDs {
		leftHostIDSet[hostID] = struct{}{}
	}
	cmdbHosts, err := w.cc.GetHostBaseInfoByID(kt, hostIDs)
	if err != nil {
		logs.Errorf("get host base info by id failed, err: %s, rid: %s", err, kt.Rid)
		// fail all
		w.handleHostBatchError(kt, hostIDs, err)
		return nil
	}

	succeed = make([]int64, 0, len(cmdbHosts))
	for i := range cmdbHosts {
		host := cmdbHosts[i]
		delete(leftHostIDSet, host.BkHostID)

		hostExecInfos := w.hostMap[host.BkHostID]
		if len(hostExecInfos) == 0 {
			continue
		}
		for _, hostExecInfo := range hostExecInfos {
			// 记录当时主机信息
			hostExecInfo.ExecLog += fmt.Sprintf("operator: %s, bak operator: %s", host.Operator, host.BkBakOperator)

			if !isValidOperator(&host, hostExecInfo.Step) {
				logs.Errorf("checkRecyclability failed, %s is not operator or bak operator of host %s(%d), rid: %s",
					hostExecInfo.Step.User, hostExecInfo.Step.IP, hostExecInfo.Step.HostID, kt.Rid)
				hostExecInfo.Error = fmt.Errorf("%s is not operator or bak operator of host %s",
					hostExecInfo.Step.User, hostExecInfo.Step.IP)

				w.handleResult(kt, []*HostExecInfo{hostExecInfo}, false)
				continue
			}
		}

		// 检查成功
		succeed = append(succeed, host.BkHostID)
	}

	for hostID := range leftHostIDSet {
		for _, host := range w.hostMap[hostID] {
			e := fmt.Errorf("host %s(%d) not found on cmdb", host.Step.IP, host.Step.HostID)
			w.handleHostError(kt, host, e)
		}
	}
	return succeed
}

// isValidOperator 检查是否是主备负责人之一
func isValidOperator(host *cmdb.Host, step *table.DetectStep) bool {
	return strings.Contains(host.Operator, step.User) || strings.Contains(host.BkBakOperator, step.User)
}

func (w *preCheckWorker) checkHostModule(kt *kit.Kit, hostIDs []int64) {
	// 转为set
	leftHostIDSet := cvt.SliceToMap(hostIDs, func(hostID int64) (int64, struct{}) { return hostID, struct{}{} })

	req := &cmdb.HostModuleRelationParams{
		HostID: hostIDs,
	}
	relations, err := w.cc.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("get host topo info failed, err: %s, rid: %s", err, kt.Rid)
		// fail all
		e := fmt.Errorf("fail to get host topo info err: %v", err)
		w.handleHostBatchError(kt, hostIDs, e)
		return
	}

	rels := cvt.PtrToVal(relations)
	// 按业务分组检查
	bizRelList := classifier.ClassifySlice(rels, func(rel cmdb.HostTopoRelation) int64 { return rel.BizID })

	for bizId, rels := range bizRelList {

		// 主机可能属于多个模块
		hostRelMap := make(map[int64][]cmdb.HostTopoRelation, len(rels))
		moduleIDs := make([]int64, 0, len(rels))
		bizHostIDs := make([]int64, 0, len(rels))
		for _, rel := range rels {
			hostRelMap[rel.HostID] = append(hostRelMap[rel.HostID], rel)
			moduleIDs = append(moduleIDs, rel.BkModuleID)
			bizHostIDs = append(bizHostIDs, rel.HostID)
		}

		// 获取模块信息
		moduleMap, err := w.cc.GetBizModuleMap(kt, bizId, moduleIDs)
		if err != nil {
			logs.Errorf("failed to get module info, err: %v, module: %v, rid: %s", err, moduleIDs, kt.Rid)
			e := fmt.Errorf("fail to get module info err: %v", err)
			w.handleHostBatchError(kt, bizHostIDs, e)
			continue
		}

		// 检查是否在待回收模块
		for hostID, relList := range hostRelMap {
			delete(leftHostIDSet, hostID)
			for _, host := range w.hostMap[hostID] {
				w.checkRecycleModule(kt, moduleMap, bizId, host, relList)
			}
		}
	}

	for hostID := range leftHostIDSet {
		for _, host := range w.hostMap[hostID] {
			e := fmt.Errorf("host %s(%d) topo relation not found on cmdb", host.Step.IP, host.Step.HostID)
			w.handleHostError(kt, host, e)
		}
	}
	return
}

func (w *preCheckWorker) checkRecycleModule(kt *kit.Kit, moduleMap map[int64]*cmdb.ModuleInfo, bizID int64,
	host *HostExecInfo, relList []cmdb.HostTopoRelation) {

	if host.BizID != bizID {
		host.Error = fmt.Errorf("主机业务发生变化，新业务: %d, 原业务: %d, IP: %s, 固资号：%s",
			bizID, host.BizID, host.Step.IP, host.Step.AssetID)
		logs.Errorf("host: %s(%d,%s) topo relation biz changed, new biz: %d, old biz: %d, rid: %s",
			host.Step.IP, host.Step.HostID, host.Step.AssetID, bizID, host.BizID, kt.Rid)
		w.handleResult(kt, []*HostExecInfo{host}, false)
		return
	}

	for _, rel := range relList {
		module, ok := moduleMap[rel.BkModuleID]
		if !ok {
			// 找不到模块
			e := fmt.Errorf("module: %d of host: %s not found", rel.BkModuleID, host.Step.IP)
			logs.Errorf("failed to get module info, biz: %d, module: %d, rid: %s",
				bizID, rel.BkModuleID, kt.Rid)
			w.handleHostError(kt, host, e)
			return
		}
		host.ExecLog += fmt.Sprintf("\nmodule: %d, name: %s, dft: %d",
			module.BkModuleId, module.BkModuleName, module.Default)
		if module.Default != cmdb.DftModuleRecycle {
			// 任何一个模块不是待回收模块则失败
			host.Error = fmt.Errorf("主机(%s)不在空闲机池下的待回收模块, 当前模块为: %s(%d,dft:%d)",
				host.Step.IP, module.BkModuleName, rel.BkModuleID, module.Default)
			logs.Errorf("module: %d of host: %s is not 待回收, biz: %d, rid: %s",
				rel.BkModuleID, host.Step.IP, bizID, kt.Rid)
			w.handleResult(kt, []*HostExecInfo{host}, false)
			return
		}
	}
	// 成功
	w.handleResult(kt, []*HostExecInfo{host}, false)
}

// handleHostBatchError 批量失败，需要重试
func (w *preCheckWorker) handleHostBatchError(kt *kit.Kit, hostIDs []int64, err error) {
	if len(hostIDs) == 0 {
		return
	}
	hostList := make([]*HostExecInfo, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		for _, host := range w.hostMap[hostID] {
			host.Error = err
			hostList = append(hostList, host)
		}
	}
	w.handleResult(kt, hostList, true)
}

// handleHostError 单个主机失败，需要重试
func (w *preCheckWorker) handleHostError(kt *kit.Kit, host *HostExecInfo, err error) {
	if err != nil {
		host.Error = err
	}
	w.handleResult(kt, []*HostExecInfo{host}, true)
}

func (w *preCheckWorker) handleResult(kt *kit.Kit, resultHosts []*HostExecInfo, needRetry bool) {
	for _, host := range resultHosts {
		w.resultHandler.HandleResult(kt, []*StepMeta{host.StepMeta}, host.Error, host.ExecLog, needRetry)
	}
}
