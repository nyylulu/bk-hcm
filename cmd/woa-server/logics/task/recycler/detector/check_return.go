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
	"strings"
	"sync/atomic"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// CheckReturnMaxBatchSize Return 暂未限定长度，暂定200
const CheckReturnMaxBatchSize = 200

// crpProcessReturnAction 退回
const crpProcessReturnAction = "退回"

// CheckReturnWorkGroup check return work group
type CheckReturnWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	cvm           cvmapi.CVMClientInterface
	resultHandler StepResultHandler
}

// MaxBatchSize ...
func (t *CheckReturnWorkGroup) MaxBatchSize() int {
	return CheckReturnMaxBatchSize
}

// NewCheckReturnWorkGroup ...
func NewCheckReturnWorkGroup(cvm cvmapi.CVMClientInterface, resultHandler StepResultHandler,
	workerNum int) *CheckReturnWorkGroup {
	return &CheckReturnWorkGroup{
		cvm:           cvm,
		resultHandler: resultHandler,
		currency:      workerNum,
		stepBatchChan: make(chan *stepBatch, workerNum),
	}
}

// HandleResult ...
func (t *CheckReturnWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string, needRetry bool) {
	t.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

// Submit 提交检查
func (t *CheckReturnWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	t.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// Start 启动worker
func (t *CheckReturnWorkGroup) Start(kt *kit.Kit) {
	if !t.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < t.currency; i++ {
		subKit := kt.NewSubKit()
		go t.queryWorker(subKit, i)
	}
}

func (t *CheckReturnWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("check Return query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check Return query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-t.stepBatchChan:
			logs.V(4).Infof("check Return worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			t.check(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

func (t *CheckReturnWorkGroup) check(kt *kit.Kit, steps []*StepMeta) {
	if len(steps) == 0 {
		logs.Warnf("check return worker receive empty steps, rid: %s", kt.Rid)
		return
	}

	// hostExecInfoMap用于记录每个step对应的执行情况，key为suborder_id+"_"+assetID。发生错误的step会从hostExecInfoMap上移除
	hostExecInfoMap := make(map[string]*HostExecInfo, len(steps))
	assetIDs := make([]string, 0)
	for _, step := range steps {
		hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
		hostExecInfoMap[hostExecInfoKey] = &HostExecInfo{StepMeta: step}
		// 空assetID是非法的
		if len(step.Step.AssetID) == 0 {
			logs.Errorf("hostID %d failed to check cvm return process due to empty assetID, rid: %s",
				step.Step.HostID, kt.Rid)
			hostExecInfo := hostExecInfoMap[hostExecInfoKey]
			hostExecInfo.Error = fmt.Errorf("hostID %d failed to check cvm return process due to empty "+
				"assetID", step.Step.HostID)
			t.HandleResult(kt, []*StepMeta{step}, hostExecInfo.Error, hostExecInfo.ExecLog, false)
			delete(hostExecInfoMap, hostExecInfoKey)
			continue // 非法assetID不会被放入assetIDs
		}
		assetIDs = append(assetIDs, step.Step.AssetID)
	}

	// 因为同一个主机可能出现在多个回收子单，所以assetIDs需要去重以精简in语句
	assetIDs = slice.Unique(assetIDs)

	hostResType := make(map[string]table.ResourceType, len(assetIDs))
	// 首先分批取出全部待回收主机的具体信息
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKMaxInstanceLimit,
	}
	for _, parts := range slice.Split(assetIDs, t.MaxBatchSize()) {
		filter := mapstr.MapStr{
			"asset_id": &mapstr.MapStr{
				pkg.BKDBIN: parts,
			},
		}
		hosts, err := dao.Set().RecycleHost().FindManyRecycleHost(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get recycle host, err: %v, rid: %s", err, kt.Rid)
			for _, hostExecInfo := range hostExecInfoMap {
				hostExecInfo.Error = fmt.Errorf("failed to get recycle host, err: %v", err)
				t.HandleResult(kt, []*StepMeta{hostExecInfo.StepMeta}, hostExecInfo.Error, hostExecInfo.ExecLog, true)
			}
			return
		}
		for _, host := range hosts {
			hostResType[host.AssetID] = host.ResourceType
		}
	}

	// 然后根据类型将全部主机分类放进对应类型的切片，分别调用不同的接口查询回收状态
	cvmAssetIDs, pmAssetIDs := make([]string, 0), make([]string, 0)
	cvmSteps, pmSteps := make([]*StepMeta, 0), make([]*StepMeta, 0)
	for _, hostExecInfo := range hostExecInfoMap {
		step := hostExecInfo.StepMeta
		switch hostResType[step.Step.AssetID] {
		case table.ResourceTypeCvm:
			cvmAssetIDs = append(cvmAssetIDs, step.Step.AssetID)
			cvmSteps = append(cvmSteps, step)
		case table.ResourceTypePm:
			pmAssetIDs = append(pmAssetIDs, step.Step.AssetID)
			pmSteps = append(pmSteps, step)
		default:
			t.HandleResult(kt, []*StepMeta{step}, nil, hostExecInfo.ExecLog, false)
		}
	}
	t.checkCvmReturn(kt, cvmAssetIDs, cvmSteps, hostExecInfoMap)
	t.checkErpReturn(kt, pmAssetIDs, pmSteps, hostExecInfoMap)
}

// checkCvmReturn 检查cvm回收状态，assetIDs和steps的元素一一对应，assetIDs可能有重复
func (t *CheckReturnWorkGroup) checkCvmReturn(kt *kit.Kit, assetIDs []string, steps []*StepMeta,
	hostExecInfoMap map[string]*HostExecInfo) {
	assetIDs = slice.Unique(assetIDs)
	cvmProcess := make([]*cvmapi.CvmProcessItem, 0)

	for _, parts := range slice.Split(assetIDs, t.MaxBatchSize()) {
		req := &cvmapi.GetCvmProcessReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.CvmGetProcessMethod,
			},
			Params: &cvmapi.GetCvmProcessParam{
				AssetIds: parts,
			},
		}

		resp, err := t.cvm.GetCvmProcess(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("failed to check cvm return process, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req),
				kt.Rid)
			for _, step := range steps {
				hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
				hostExecInfoMap[hostExecInfoKey].Error = fmt.Errorf("failed to check cvm return process, "+
					"err: %v, req: %+v", err, cvt.PtrToVal(req))
				t.HandleResult(kt, []*StepMeta{step}, hostExecInfoMap[hostExecInfoKey].Error,
					hostExecInfoMap[hostExecInfoKey].ExecLog, true)
			}
			return
		}

		if resp.Error.Code != 0 {
			respStr := structToStr(resp)
			exeInfo := fmt.Sprintf("yunti response: %s", respStr)

			logs.Errorf("failed to check cvm return process, code: %d, msg: %s, rid: %s", resp.Error.Code,
				resp.Error.Message, kt.Rid)

			for _, step := range steps {
				hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
				hostExecInfoMap[hostExecInfoKey].ExecLog += exeInfo
				t.HandleResult(kt, []*StepMeta{step}, fmt.Errorf("check return process api return err: %s",
					resp.Error.Message), hostExecInfoMap[hostExecInfoKey].ExecLog, true)
			}
			return
		}
		cvmProcess = append(cvmProcess, resp.Result.Data...)
	}

	// key为assetID，value为该主机进行中的流程
	processes := make(map[string][]string)
	for _, item := range cvmProcess {
		if len(item.StatusDesc) > 0 {
			process := fmt.Sprintf("%s(%s)", item.OrderId, item.StatusDesc)
			processes[item.AssetId] = append(processes[item.AssetId], process)
		}
	}

	for _, step := range steps {
		respStr := structToStr(processes[step.Step.AssetID])
		exeInfo := fmt.Sprintf("yunti response: %s", respStr)
		hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
		hostExecInfoMap[hostExecInfoKey].ExecLog += exeInfo

		if len(processes[step.Step.AssetID]) > 0 {
			logs.Errorf("assetID: %s has cvm process: %s, rid: %s", step.Step.AssetID,
				strings.Join(processes[step.Step.AssetID], ";"), kt.Rid)
			hostExecInfoMap[hostExecInfoKey].Error = fmt.Errorf("assetID: %s has cvm process: %s",
				step.Step.AssetID, strings.Join(processes[step.Step.AssetID], ";"))
			t.HandleResult(kt, []*StepMeta{step}, hostExecInfoMap[hostExecInfoKey].Error,
				hostExecInfoMap[hostExecInfoKey].ExecLog, false)
			continue
		}
		t.HandleResult(kt, []*StepMeta{step}, nil, hostExecInfoMap[hostExecInfoKey].ExecLog, false)
	}
	return
}

// 检查erp回收状态，assetIDs和steps的元素一一对应，assetIDs可能有重复
func (t *CheckReturnWorkGroup) checkErpReturn(kt *kit.Kit, assetIDs []string, steps []*StepMeta,
	hostExecInfoMap map[string]*HostExecInfo) {
	assetIDs = slice.Unique(assetIDs)
	erpProcess := make([]*cvmapi.ErpProcessItem, 0)

	for _, parts := range slice.Split(assetIDs, t.MaxBatchSize()) {
		req := &cvmapi.GetErpProcessReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.GetErpProcessMethod,
			},
			Params: &cvmapi.GetErpProcessParam{
				AssetIds: parts,
			},
		}

		resp, err := t.cvm.GetErpProcess(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("failed to check erp return process, rid: %s", kt.Rid)
			for _, step := range steps {
				hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
				hostExecInfoMap[hostExecInfoKey].Error = fmt.Errorf("failed to check erp return process, "+
					"err: %v, assetId: %s, req: %+v", err, step.Step.AssetID, cvt.PtrToVal(req))
				t.HandleResult(kt, []*StepMeta{step}, hostExecInfoMap[hostExecInfoKey].Error,
					hostExecInfoMap[hostExecInfoKey].ExecLog, true)
			}
			return
		}

		if resp.Error.Code != 0 {
			respStr := structToStr(resp)
			exeInfo := fmt.Sprintf("yunti response: %s", respStr)

			logs.Errorf("failed to check erp return process, code: %d, msg: %s, rid: %s", resp.Error.Code,
				resp.Error.Message, kt.Rid)

			for _, step := range steps {
				hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
				hostExecInfoMap[hostExecInfoKey].ExecLog += exeInfo
				hostExecInfoMap[hostExecInfoKey].Error = fmt.Errorf("check return process api return err: %s",
					resp.Error.Message)
				t.HandleResult(kt, []*StepMeta{step}, hostExecInfoMap[hostExecInfoKey].Error,
					hostExecInfoMap[hostExecInfoKey].ExecLog, true)
			}
			return
		}
		erpProcess = append(erpProcess, resp.Result.Data...)
	}

	// 记录assetID对应主机所有单号
	processes := make(map[string][]string)
	for _, item := range erpProcess {
		if item.ActionType == crpProcessReturnAction {
			processes[item.AssetId] = append(processes[item.AssetId], item.OrderId)
		}
	}

	for _, step := range steps {
		respStr := structToStr(processes[step.Step.AssetID])
		exeInfo := fmt.Sprintf("yunti response: %s", respStr)
		hostExecInfoKey := t.getHostExecInfoKey(step.Step.SuborderID, step.Step.AssetID)
		hostExecInfoMap[hostExecInfoKey].ExecLog += exeInfo

		if len(processes[step.Step.AssetID]) > 0 {
			logs.Errorf("assetID %s has erp return order: %s, rid: %s", step.Step.AssetID,
				strings.Join(processes[step.Step.AssetID], ";"), kt.Rid)
			t.HandleResult(kt, []*StepMeta{step}, fmt.Errorf("assetID %s has erp return order: %s",
				step.Step.AssetID, strings.Join(processes[step.Step.AssetID], ";")),
				hostExecInfoMap[hostExecInfoKey].ExecLog, false)
			continue
		}
		t.HandleResult(kt, []*StepMeta{step}, nil, hostExecInfoMap[hostExecInfoKey].ExecLog, false)
	}
	return
}

func (t *CheckReturnWorkGroup) getHostExecInfoKey(suborderID string, assetID string) string {
	return suborderID + "_" + assetID
}
