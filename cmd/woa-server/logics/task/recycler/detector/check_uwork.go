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
	"strconv"
	"strings"
	"sync/atomic"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/xrayapi"
	"hcm/pkg/thirdparty/xshipapi"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"

	"golang.org/x/time/rate"
)

// GetUworkOpenTicketByAssetID 获取Uwork未完结的流程单
func (d *Detector) GetUworkOpenTicketByAssetID(kt *kit.Kit, assetID string) ([]string, error) {
	// 获取未结单的故障单
	_, tickets, err := d.checkXrayFaultTickets(kt, assetID)
	if err != nil {
		logs.Errorf("failed to check uwork-xray ticket, err: %v, assetID: %s, rid: %s", err, assetID, kt.Rid)
		return nil, err
	}

	// 获取未结单的x-ship流程单
	_, processes, err := d.checkXShipProcess(kt, assetID)
	if err != nil {
		logs.Errorf("failed to check uwork-xship process, err: %v, assetID: %s, rid: %s", err, assetID, kt.Rid)
		return nil, err
	}

	return append(tickets, processes...), nil
}

// checkXrayFaultTickets 检查尚未完结的故障单
func (d *Detector) checkXrayFaultTickets(kt *kit.Kit, assetID string) (string, []string, error) {
	var execInfo string

	respTicket, err := d.xray.CheckXrayFaultTickets(kt, []string{assetID}, enumor.XrayFaultTicketNotEnd)
	if err != nil {
		logs.Errorf("failed to check uwork-xray ticket, err: %v, assetID: %s, rid: %s", err, assetID, kt.Rid)
		return execInfo, nil, fmt.Errorf("failed to check uwork-xray, err: %v", err)
	}

	ticketRespStr := d.structToStr(respTicket)
	execInfo = fmt.Sprintf("uwork-xray ticket response: %s", ticketRespStr)

	ticketIds := make([]string, 0)
	for _, ticket := range respTicket.Data {
		if ticket.IsEnd == enumor.XrayFaultTicketNotEnd {
			ticketIds = append(ticketIds, strconv.Itoa(ticket.InstanceID))
		}
	}

	return execInfo, ticketIds, nil
}

// checkXShipProcess 检查尚未完结的x-ship流程单
func (d *Detector) checkXShipProcess(kt *kit.Kit, assetID string) (string, []string, error) {
	var execInfo string

	respProcess, err := d.xship.GetXServerProcess(kt, assetID)
	if err != nil {
		logs.Errorf("failed to check uwork-xship process, err: %v, assetID: %s, rid: %s", err, assetID, kt.Rid)
		return execInfo, nil, fmt.Errorf("failed to check uwork-xship, err: %v", err)
	}

	processRespStr := d.structToStr(respProcess)
	execInfo = fmt.Sprintf("uwork-xship process response: %s", processRespStr)

	processes := make([]string, 0)
	for _, process := range respProcess.Data.Processes {
		processes = append(processes, fmt.Sprintf("%d(%s)", process.ID, process.Name))
	}

	return execInfo, processes, nil
}

// CheckUworkMaxBatchSize ...
const CheckUworkMaxBatchSize = xrayapi.XRayCheckFaultTicketMaxLength

// CheckUworkWorkGroup 检查是否有Uwork故障或流程单据
type CheckUworkWorkGroup struct {
	stepBatchChan chan *stepBatch
	started       atomic.Bool
	currency      int
	xray          xrayapi.XrayClientInterface
	xship         xshipapi.XshipClientInterface
	resultHandler StepResultHandler
	// xship 只有单个查询接口，需要限流
	xshipRateLimiter *rate.Limiter
}

// MaxBatchSize ...
func (p *CheckUworkWorkGroup) MaxBatchSize() int {
	return CheckUworkMaxBatchSize
}

// NewCheckUworkWorkGroup ...
func NewCheckUworkWorkGroup(xray xrayapi.XrayClientInterface, xship xshipapi.XshipClientInterface,
	resultHandler StepResultHandler, workerNum int, limit int, burst int) *CheckUworkWorkGroup {

	return &CheckUworkWorkGroup{
		stepBatchChan:    make(chan *stepBatch, workerNum),
		currency:         workerNum,
		xray:             xray,
		xship:            xship,
		resultHandler:    resultHandler,
		xshipRateLimiter: rate.NewLimiter(rate.Limit(limit), burst),
	}
}

// Start 启动worker
func (p *CheckUworkWorkGroup) Start(kt *kit.Kit) {
	if !p.started.CompareAndSwap(false, true) {
		// already started
		return
	}

	for i := 0; i < p.currency; i++ {
		subKit := kt.NewSubKit()
		go p.queryWorker(subKit, i)
	}
}

// HandleResult ...
func (p *CheckUworkWorkGroup) HandleResult(kt *kit.Kit, steps []*StepMeta, detectErr error, log string,
	needRetry bool) {

	p.resultHandler.HandleResult(kt, steps, detectErr, log, needRetry)
}

func (p *CheckUworkWorkGroup) queryWorker(kt *kit.Kit, idx int) {
	logs.Infof("check uwork query worker %d start, rid: %s", idx, kt.Rid)
	defer logs.Infof("check uwork query worker %d exit, rid: %s", idx, kt.Rid)

	for {
		select {
		case batch := <-p.stepBatchChan:
			logs.V(4).Infof("check uwork worker %d got steps: %d:%s, rid: %s",
				idx, len(batch.steps), batch.steps, kt.Rid)
			p.Run(batch.kt, batch.steps)
		case <-kt.Ctx.Done():
			return
		}
	}
}

// Run 执行检查
func (p *CheckUworkWorkGroup) Run(kt *kit.Kit, steps []*StepMeta) {
	if len(steps) == 0 {
		logs.Warnf("check uwork worker receive empty steps, rid: %s", kt.Rid)
		return
	}

	hostMap := make(map[string][]*HostExecInfo, len(steps))
	for _, step := range steps {
		hostMap[step.Step.AssetID] = append(hostMap[step.Step.AssetID], &HostExecInfo{StepMeta: step})
	}
	assetIDs := maps.Keys(hostMap)

	worker := checkUworkWorker{
		xray:             p.xray,
		xship:            p.xship,
		hostMap:          hostMap,
		resultHandler:    p,
		xshipRateLimiter: p.xshipRateLimiter,
	}
	worker.checkAll(kt, assetIDs)
}

// Submit 提交检查
func (p *CheckUworkWorkGroup) Submit(kt *kit.Kit, steps []*StepMeta) {
	p.stepBatchChan <- &stepBatch{kt: kt, steps: steps}
}

// checkUworkWorker worker
type checkUworkWorker struct {
	xray             xrayapi.XrayClientInterface
	xship            xshipapi.XshipClientInterface
	hostMap          map[string][]*HostExecInfo
	resultHandler    StepResultHandler
	xshipRateLimiter *rate.Limiter
}

func (w *checkUworkWorker) checkAll(kt *kit.Kit, assetIDs []string) {
	// 1. 检查尚未完结的故障单
	opSuccessIDs := w.checkXrayFaultTickets(kt, assetIDs)

	if len(opSuccessIDs) == 0 {
		return
	}

	// 2. 检查尚未完结的x-ship流程单
	w.checkXShipProcess(kt, opSuccessIDs)
}

// 获取尚未结单的故障单
func (w *checkUworkWorker) checkXrayFaultTickets(kt *kit.Kit, assetIDs []string) (succeed []string) {
	ticketResp, err := w.xray.CheckXrayFaultTickets(kt, assetIDs, enumor.XrayFaultTicketNotEnd)
	if err != nil {
		logs.Errorf("failed to check uwork-xray ticket, err: %v, assetIDs: %s, rid: %s", err, assetIDs, kt.Rid)
		// fail all
		w.handleHostBatchError(kt, assetIDs, err, err.Error())
		return nil
	}

	succeed = make([]string, 0, len(assetIDs))
	hostTicketMap := classifier.ClassifySlice(ticketResp.Data,
		func(ticket *xrayapi.Ticket) string { return ticket.ServerAssetId })
	for assetID, hostExecInfos := range w.hostMap {
		tickets := hostTicketMap[assetID]
		notEndedTicketIDs := make([]string, 0, len(tickets))
		for _, ticket := range tickets {
			if ticket.IsEnd == enumor.XrayFaultTicketNotEnd {
				notEndedTicketIDs = append(notEndedTicketIDs, strconv.Itoa(ticket.InstanceID))
			}
		}
		if len(notEndedTicketIDs) > 0 {
			hostExecLog := fmt.Sprintf("%d xray ticket(s): %s", len(notEndedTicketIDs),
				strings.Join(notEndedTicketIDs, ","))
			hostExecError := fmt.Errorf("has uwork-xray tickets: %s", notEndedTicketIDs)
			w.handleHostBatchResult(kt, []string{assetID}, hostExecError, hostExecLog)
			continue
		}
		for _, hostExecInfo := range hostExecInfos {
			hostExecInfo.ExecLog += fmt.Sprintf("xray ticket: %s", "no xray ticket")
		}
		// 检查成功
		succeed = append(succeed, assetID)
	}
	return succeed
}

func (w *checkUworkWorker) checkXShipProcess(kt *kit.Kit, assetIDs []string) {
	for _, assetID := range assetIDs {
		hosts := w.hostMap[assetID]
		err := w.xshipRateLimiter.Wait(kt.Ctx)
		if err != nil {
			hostExecErr := fmt.Errorf("failed to check uwork-xship wait rate limiter failed, err: %w", err)
			w.handleHostBatchError(kt, []string{assetID}, hostExecErr, hostExecErr.Error())
			continue
		}

		resp, err := w.xship.GetXServerProcess(kt, assetID)
		if err != nil {
			ips := make([]string, 0, len(hosts))
			for _, host := range hosts {
				ips = append(ips, host.Step.IP)
			}
			ips = slice.Unique(ips)
			logs.Errorf("failed to check uwork-xship process, err: %v, assetID: %s, ip: %v, rid: %s",
				err, assetID, ips, kt.Rid)
			hostExecErr := fmt.Errorf("failed to check uwork-xship, err: %w", err)
			w.handleHostBatchError(kt, []string{assetID}, hostExecErr, hostExecErr.Error())
			continue
		}
		processRespStr := structToStr(resp)
		execLog := fmt.Sprintf("\nuwork-xship process response: %s", processRespStr)
		var execErr error
		processes := make([]string, 0)
		for _, process := range resp.Data.Processes {
			processes = append(processes, fmt.Sprintf("%d(%s)", process.ID, process.Name))
		}
		if len(processes) > 0 {
			// 标记失败
			execErr = fmt.Errorf("has uwork-xship process: %s", strings.Join(processes, ";"))
		}
		w.handleHostBatchResult(kt, []string{assetID}, execErr, execLog)
	}
}

// handleHostBatchError 获取失败，需要重试
func (w *checkUworkWorker) handleHostBatchError(kt *kit.Kit, assetIDs []string, err error, execLog string) {
	for _, assetID := range assetIDs {
		resultHosts := w.hostMap[assetID]
		for _, host := range resultHosts {
			host.Error = err
			host.ExecLog += execLog
			w.resultHandler.HandleResult(kt, []*StepMeta{host.StepMeta}, host.Error, host.ExecLog, true)
		}
	}

}

// handleResult 处理结果, 不需要重试
func (w *checkUworkWorker) handleHostBatchResult(kt *kit.Kit, assetIDs []string, err error, execLog string) {
	for _, assetID := range assetIDs {
		resultHosts := w.hostMap[assetID]
		for _, host := range resultHosts {
			host.Error = err
			host.ExecLog += execLog
			w.resultHandler.HandleResult(kt, []*StepMeta{host.StepMeta}, host.Error, host.ExecLog, false)
		}
	}
}
