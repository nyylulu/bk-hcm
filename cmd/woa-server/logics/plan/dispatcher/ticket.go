/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

package dispatcher

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/times"
)

// listAndWatchTickets list and watch tickets
func (d *Dispatcher) listAndWatchTickets() error {
	logs.Infof("ready to list and watch tickets")
	if !d.sd.IsMaster() {
		// pop all pending orders
		d.ticketQueue.Clear()
		return nil
	}

	// list pending orders
	kt := core.NewBackendKit()
	pendingTkIDs, err := d.listAllPendingTickets(kt)
	if err != nil {
		logs.Errorf("failed to list pending resource plan tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// enqueue pending orders
	for _, tkID := range pendingTkIDs {
		d.ticketQueue.Enqueue(tkID)
	}

	return nil
}

// listAllPendingTickets list all pending tickets
func (d *Dispatcher) listAllPendingTickets(kt *kit.Kit) ([]string, error) {
	// list tickets of recent 7 days.
	dr := &times.DateRange{
		Start: time.Now().AddDate(0, 0, -enumor.PendingTicketTraceDay).Format(constant.DateLayout),
		End:   time.Now().Format(constant.DateLayout),
	}

	drOpt, err := tools.DateRangeExpression("submitted_at", dr)
	if err != nil {
		return nil, err
	}

	// TODO: 当单据数量超过500时，可能会漏单据。这里改为分页查询
	recentOpt := &types.ListOption{
		Fields: []string{"id"},
		Filter: drOpt,
		Page:   core.NewDefaultBasePage(),
	}

	tkRst, err := d.dao.ResPlanTicket().List(kt, recentOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	recentTkIDs := make([]string, 0)
	for _, ticket := range tkRst.Details {
		recentTkIDs = append(recentTkIDs, ticket.ID)
	}

	// list tickets with auditing status
	auditOpt := &types.ListOption{
		Fields: []string{"ticket_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("ticket_id", recentTkIDs),
			tools.RuleEqual("status", enumor.RPTicketStatusAuditing),
		),
		Page: core.NewDefaultBasePage(),
	}

	statusRst, err := d.dao.ResPlanTicketStatus().List(kt, auditOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pendingTkIDs := make([]string, 0)
	for _, ticket := range statusRst.Details {
		pendingTkIDs = append(pendingTkIDs, ticket.TicketID)
	}

	return pendingTkIDs, nil
}

// dealTicket deal a ticket
func (d *Dispatcher) dealTicket() error {
	logs.Infof("ready to deal ticket")

	// only master node handle plan tickets.
	if !d.sd.IsMaster() {
		return nil
	}

	// get one ticket from the work queue
	tkID, ok := d.ticketQueue.Pop()
	if !ok {
		return nil
	}

	logs.Infof("ready to handle ticket %s", tkID)

	// check the status of the ticket
	kt := core.NewBackendKit()
	tkInfo, err := d.resFetcher.GetTicketInfo(kt, tkID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
		return err
	}

	if tkInfo.Status != enumor.RPTicketStatusAuditing {
		logs.Warnf("need not handle ticket for its status %s != %s, id: %s, rid: %s", tkInfo.Status,
			enumor.RPTicketStatusAuditing, tkID, kt.Rid)
		return nil
	}

	if tkInfo.ItsmSN == "" {
		logs.Errorf("failed to handle ticket for itsm sn is empty, id: %s, rid: %s", tkID, kt.Rid)
		return errors.New("failed to handle ticket for itsm sn is empty")
	}

	if tkInfo.CrpSN != "" {
		return d.checkCrpTicket(kt, tkInfo)
	}

	return d.checkItsmTicket(kt, tkInfo)
}

func (d *Dispatcher) checkCrpTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	logs.Infof("ready to check crp flow, sn: %s, id: %s, rid: %s", ticket.CrpSN, ticket.ID, kt.Rid)

	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderQueryMethod,
		},
		Params: &cvmapi.QueryPlanOrderParam{
			OrderIds: []string{ticket.CrpSN},
		},
	}
	resp, err := d.crpCli.QueryPlanOrder(kt.Ctx, nil, req)
	if err != nil {
		logs.Errorf("failed to query crp plan order, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if resp.Error.Code != 0 {
		logs.Errorf("%s: failed to query crp plan order, code: %d, msg: %s, crp_sn: %s, rid: %s",
			constant.ResPlanTicketWatchFailed, resp.Error.Code, resp.Error.Message, ticket.CrpSN, kt.Rid)
		return fmt.Errorf("failed to query crp plan order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		logs.Errorf("%s: failed to query crp plan order, for result is empty, crp_sn: %s, rid: %s",
			constant.ResPlanTicketWatchFailed, ticket.CrpSN, kt.Rid)
		return errors.New("failed to query crp plan order, for result is empty")
	}
	planItem, ok := resp.Result[ticket.CrpSN]
	if !ok {
		logs.Errorf("%s: query crp plan order return no result by sn: %s, rid: %s",
			constant.ResPlanTicketWatchFailed, ticket.CrpSN, kt.Rid)
		return fmt.Errorf("query crp plan order return no result by sn: %s", ticket.CrpSN)
	}
	// CRP返回状态码为： 1 追加单， 2 调整单， 3 订单不存在， 4 其它错误（只有1 和 2 是正确的）
	if planItem.Code != 1 && planItem.Code != 2 {
		logs.Errorf("%s: failed to query crp plan order, order status is incorrect, code: %d, data: %+v, rid: %s",
			constant.ResPlanTicketWatchFailed, planItem.Code, planItem.Data, kt.Rid)
		return fmt.Errorf("crp plan order status is incorrect, code: %d, sn: %s", planItem.Code, ticket.CrpSN)
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
		CrpSN:    ticket.CrpSN,
		CrpURL:   ticket.CrpURL,
	}

	switch planItem.Data.BaseInfo.Status {
	case cvmapi.PlanOrderStatusRejected:
		update.Status = enumor.RPTicketStatusRejected
	case cvmapi.PlanOrderStatusApproved:
		return d.finishAuditFlow(kt, ticket)
	default:
		return d.checkTicketTimeout(kt, ticket)
	}
	if err := d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 单据被拒需要释放资源
	if update.Status != enumor.RPTicketStatusRejected {
		return nil
	}
	allDemandIDs := make([]string, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allDemandIDs = append(allDemandIDs, demand.Original.DemandID)
		}
	}
	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(allDemandIDs, 0)
	if err = d.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (d *Dispatcher) checkItsmTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	logs.Infof("ready to check itsm flow, sn: %s, id: %s", ticket.ItsmSN, ticket.ID)

	resp, err := d.itsmCli.GetTicketStatus(kt, ticket.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusRejected,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
	}

	switch resp.Data.CurrentStatus {
	case string(itsm.StatusFinished), string(itsm.StatusTerminated):
		// rejected
		update.Status = enumor.RPTicketStatusRejected
	case string(itsm.StatusRevoked):
		// revoked
		update.Status = enumor.RPTicketStatusRevoked
	case string(itsm.StatusRunning):
		// check if CRP audit state
		if len(resp.Data.CurrentSteps) == 0 {
			return d.checkTicketTimeout(kt, ticket)
		}

		if resp.Data.CurrentSteps[0].StateId != d.crpAuditNode.ID {
			return d.checkTicketTimeout(kt, ticket)
		}

		// CRP audit state, create CRP ticket
		return d.createCrpTicket(kt, ticket)
	default:
		return d.checkTicketTimeout(kt, ticket)
	}

	if err = d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if update.Status != enumor.RPTicketStatusRejected && update.Status != enumor.RPTicketStatusRevoked {
		return nil
	}
	// 单据被拒需要释放资源
	return d.unlockTicketOriginalDemands(kt, ticket)
}

// createCrpTicket create crp ticket.
func (d *Dispatcher) createCrpTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	if ticket == nil {
		logs.Errorf("failed to create crp ticket, ticket is nil, rid: %s", kt.Rid)
		return errors.New("ticket is nil")
	}

	// call crp api to create crp ticket.
	crpCreator := NewCrpTicketCreator(d.resFetcher, d.crpCli)
	sn, err := crpCreator.CreateCRPTicket(kt, ticket)
	if err != nil {
		// 因CRP单据修改冲突导致的提单失败，不返回报错，记录日志后返回队列继续等待
		if strings.Contains(err.Error(), constant.CRPResPlanDemandIsInProcessing) {
			logs.Warnf("failed to create crp ticket, as crp res plan demand is in processing, err: %v, "+
				"ticket_id: %s, rid: %s", err, ticket.ID, kt.Rid)
			return nil
		}

		// 这里主要返回的error是crp ticket创建失败，且ticket状态更新失败的日志在函数内已打印，这里可以忽略该错误
		_ = d.updateTicketStatusFailed(kt, ticket, err.Error())
		logs.Errorf("failed to create crp ticket with different ticket type, err: %v, ticket_id: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return err
	}

	// save crp sn and crp url to resource plan ticket status table.
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
		CrpSN:    sn,
		CrpURL:   cvmapi.CvmPlanLinkPrefix + sn,
	}

	if err = d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, ticket_id: %s, rid: %s", err, ticket.ID,
			kt.Rid)
		return err
	}

	return nil
}

// checkTicketTimeout check ticket timeout
func (d *Dispatcher) checkTicketTimeout(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	submitTime, err := time.Parse(constant.TimeStdFormat, ticket.SubmittedAt)
	if err != nil {
		logs.Errorf("failed to parse ticket submit time %s, err: %v, rid: %s", ticket.SubmittedAt, err, kt.Rid)
		return err
	}

	// set timeout as 5 days
	if time.Now().Before(submitTime.AddDate(0, 0, enumor.AuditFlowTimeoutDay)) {
		return nil
	}

	return d.updateTicketStatusFailed(kt, ticket, "audit flow timeout")
}

// finishAuditFlow 单据的所有子单均已结单，汇总成功的子单并应用到本地
func (d *Dispatcher) finishAuditFlow(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	itsmStatus, err := d.itsmCli.GetTicketStatus(kt, ticket.ItsmSN)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	// 将itsm单结单，如果itsm单已结单，继续走后续流程
	if len(itsmStatus.Data.CurrentSteps) > 0 && itsmStatus.Data.CurrentSteps[0].StateId == d.crpAuditNode.ID {
		approveReq := &itsm.ApproveReq{
			Sn:       ticket.ItsmSN,
			StateID:  int(d.crpAuditNode.ID),
			Approver: d.crpAuditNode.Approver,
			Action:   "true",
			Remark:   "",
		}
		if err := d.itsmCli.Approve(kt, approveReq); err != nil {
			logs.Errorf("request itsm ticket approve failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	// crp单据通过后更新本地数据表
	if err := d.applyResPlanDemandChange(kt, ticket); err != nil {
		logs.Errorf("%s: failed to upsert crp demand, err: %v, rid: %s", constant.DemandChangeAppliedFailed,
			err, kt.Rid)
		return err
	}
	return nil
}

// updateTicketStatus update ticket status.
func (d *Dispatcher) updateTicketStatus(kt *kit.Kit, ticket *rpts.ResPlanTicketStatusTable) error {
	expr := tools.EqualExpression("ticket_id", ticket.TicketID)
	if err := d.dao.ResPlanTicketStatus().Update(kt, expr, ticket); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// UpdateTicketStatusFailed update ticket status to failed.
func (d *Dispatcher) updateTicketStatusFailed(kt *kit.Kit, ticket *ptypes.TicketInfo, msg string) error {
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusFailed,
		ItsmSN:   ticket.ItsmSN,
		ItsmURL:  ticket.ItsmURL,
		CrpSN:    ticket.CrpSN,
		CrpURL:   ticket.CrpURL,
		Message:  msg,
	}

	if len(msg) > 255 {
		logs.Warnf("failure message is truncated to 255, origin message: %s, rid: %s", msg, kt.Rid)
		update.Message = msg[:255]
	}

	if err := d.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 失败需要释放资源
	allDemandIDs := make([]string, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allDemandIDs = append(allDemandIDs, (*demand.Original).DemandID)
		}
	}
	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(allDemandIDs, 0)
	if err := d.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
