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

package plan

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	rtypes "hcm/pkg/dal/dao/types/resource-plan"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	dtypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/jmoiron/sqlx"
)

// CreateResPlanTicket create resource plan ticket.
func (c *Controller) CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket request, err: %s, rid: %s", err, kt.Rid)
		return "", err
	}

	// construct resource plan ticket.
	ticket, err := constructResPlanTicket(req, kt.User)
	if err != nil {
		logs.Errorf("failed to construct resource plan ticket, err: %s, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketID, err := c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ticketIDs, err := c.dao.ResPlanTicket().CreateWithTx(kt, txn, []rpt.ResPlanTicketTable{*ticket})
		if err != nil {
			logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		if len(ticketIDs) != 1 {
			logs.Errorf("create resource plan ticket, but len ticketIDs != 1, rid: %s", kt.Rid)
			return "", errors.New("create resource plan ticket, but len ticketIDs != 1")
		}

		ticketID := ticketIDs[0]

		// create resource plan ticket status.
		statuses := []rpts.ResPlanTicketStatusTable{{
			TicketID: ticketID,
			Status:   enumor.RPTicketStatusInit,
		}}
		if err = c.dao.ResPlanTicketStatus().CreateWithTx(kt, txn, statuses); err != nil {
			logs.Errorf("create resource plan ticket status failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		return ticketID, nil
	})

	if err != nil {
		logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketIDStr, ok := ticketID.(string)
	if !ok {
		logs.Errorf("convert resource plan ticket id %v from interface to string failed, err: %v, rid: %s",
			ticketID, err, kt.Rid)
		return "", fmt.Errorf("convert resource plan ticket id %v from interface to string failed", ticketID)
	}

	return ticketIDStr, nil
}

// constructResPlanTicket construct resource plan ticket.
func constructResPlanTicket(req *CreateResPlanTicketReq, applicant string) (*rpt.ResPlanTicketTable, error) {
	var originalOs, originalCpuCore, originalMemory, originalDiskSize float64
	var updatedOs, updatedCpuCore, updatedMemory, updatedDiskSize float64
	for _, demand := range req.Demands {
		if demand.Original != nil {
			originalOs += (*demand.Original).Cvm.Os
			originalCpuCore += (*demand.Original).Cvm.CpuCore
			originalMemory += (*demand.Original).Cvm.Memory
			originalDiskSize += (*demand.Original).Cbs.DiskSize
		}

		if demand.Updated != nil {
			updatedOs += (*demand.Updated).Cvm.Os
			updatedCpuCore += (*demand.Updated).Cvm.CpuCore
			updatedMemory += (*demand.Updated).Cvm.Memory
			updatedDiskSize += (*demand.Updated).Cbs.DiskSize
		}
	}

	demandsJson, err := dtypes.NewJsonField(req.Demands)
	if err != nil {
		return nil, err
	}

	result := &rpt.ResPlanTicketTable{
		Type:             req.TicketType,
		Demands:          demandsJson,
		Applicant:        applicant,
		BkBizID:          req.BizOrgRel.BkBizID,
		BkBizName:        req.BizOrgRel.BkBizName,
		OpProductID:      req.BizOrgRel.OpProductID,
		OpProductName:    req.BizOrgRel.OpProductName,
		PlanProductID:    req.BizOrgRel.PlanProductID,
		PlanProductName:  req.BizOrgRel.PlanProductName,
		VirtualDeptID:    req.BizOrgRel.VirtualDeptID,
		VirtualDeptName:  req.BizOrgRel.VirtualDeptName,
		DemandClass:      req.DemandClass,
		OriginalOS:       originalOs,
		OriginalCpuCore:  originalCpuCore,
		OriginalMemory:   originalMemory,
		OriginalDiskSize: originalDiskSize,
		UpdatedOS:        updatedOs,
		UpdatedCpuCore:   updatedCpuCore,
		UpdatedMemory:    updatedMemory,
		UpdatedDiskSize:  updatedDiskSize,
		Remark:           req.Remark,
		Creator:          applicant,
		Reviser:          applicant,
		SubmittedAt:      time.Now().Format(constant.DateTimeLayout),
	}

	return result, nil
}

// GetItsmAndCrpAuditStatus get itsm and crp audit status.
func (c *Controller) GetItsmAndCrpAuditStatus(kt *kit.Kit, ticketStatus *ptypes.GetRPTicketStatusInfo) (
	*ptypes.GetRPTicketItsmAudit, *ptypes.GetRPTicketCrpAudit, error) {

	itsmAudit := &ptypes.GetRPTicketItsmAudit{
		ItsmSn:  ticketStatus.ItsmSn,
		ItsmUrl: ticketStatus.ItsmUrl,
	}
	crpAudit := &ptypes.GetRPTicketCrpAudit{
		CrpSn:  ticketStatus.CrpSn,
		CrpUrl: ticketStatus.CrpUrl,
	}

	// 审批未开始
	if ticketStatus.ItsmSn == "" {
		itsmAudit.Status = enumor.RPTicketStatusInit
		itsmAudit.StatusName = itsmAudit.Status.Name()
		crpAudit.Status = enumor.RPTicketStatusInit
		crpAudit.StatusName = crpAudit.Status.Name()
		return itsmAudit, crpAudit, nil
	}

	// 获取ITSM审批记录和当前审批节点
	itsmStatus, err := c.itsmCli.GetTicketStatus(kt, ticketStatus.ItsmSn)
	if err != nil {
		logs.Errorf("failed to get itsm audit status, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSn, kt.Rid)
		return nil, nil, err
	}
	itsmLogs, err := c.itsmCli.GetTicketLog(kt, ticketStatus.ItsmSn)
	if err != nil {
		logs.Errorf("failed to get itsm audit log, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSn, kt.Rid)
		return nil, nil, err
	}
	if itsmLogs.Data == nil {
		logs.Errorf("itsm audit log is empty, sn: %s, rid: %s", ticketStatus.ItsmSn, kt.Rid)
		return nil, nil, fmt.Errorf("itsm audit log is empty, sn: %s", ticketStatus.ItsmSn)
	}

	itsmAudit, err = c.setItsmAuditDetails(itsmAudit, itsmStatus, itsmLogs.Data)
	if err != nil {
		logs.Errorf("failed to set itsm audit details, err: %v, sn: %s, rid: %s", err, ticketStatus.ItsmSn, kt.Rid)
		return nil, nil, err
	}

	// ITSM审批中或审批终止在itsm阶段
	if ticketStatus.CrpSn == "" {
		// ITSM流程没有正常结束，将单据审批状态作为ITSM流程的当前状态
		if itsmAudit.Status != enumor.RPTicketStatusDone {
			itsmAudit.Status = ticketStatus.Status
			itsmAudit.StatusName = itsmAudit.Status.Name()
			itsmAudit.Message = ticketStatus.Message
			crpAudit.Status = enumor.RPTicketStatusInit
			crpAudit.StatusName = crpAudit.Status.Name()
			return itsmAudit, crpAudit, nil
		}
		// ITSM流程正常结束，CRP单据尚未创建
		crpAudit.Status = ticketStatus.Status
		crpAudit.StatusName = crpAudit.Status.Name()
		crpAudit.Message = ticketStatus.Message
		return itsmAudit, crpAudit, nil
	}
	// itsm审批流已结束
	itsmAudit.Status = enumor.RPTicketStatusDone
	itsmAudit.StatusName = itsmAudit.Status.Name()

	// 流程走到CRP步骤，获取CRP审批记录和当前审批节点
	crpCurrentSteps, err := c.GetCrpCurrentApprove(kt, ticketStatus.CrpSn)
	if err != nil {
		logs.Errorf("failed to get crp current approve, err: %v, sn: %s, rid: %s", err, ticketStatus.CrpSn, kt.Rid)
		return nil, nil, err
	}
	crpApproveLogs, err := c.GetCrpApproveLogs(kt, ticketStatus.CrpSn)
	if err != nil {
		logs.Errorf("failed to get crp approve logs, err: %v, sn: %s, rid: %s", err, ticketStatus.CrpSn, kt.Rid)
		return nil, nil, err
	}

	// CRP审批状态赋值
	crpAudit.Status = ticketStatus.Status
	crpAudit.StatusName = crpAudit.Status.Name()
	crpAudit.Message = ticketStatus.Message
	crpAudit.CurrentSteps = crpCurrentSteps
	crpAudit.Logs = crpApproveLogs

	return itsmAudit, crpAudit, nil
}

func (c *Controller) setItsmAuditDetails(itsmAudit *ptypes.GetRPTicketItsmAudit, current *itsm.GetTicketStatusResp,
	logData *itsm.GetTicketLogRst) (*ptypes.GetRPTicketItsmAudit, error) {

	// current steps
	itsmAudit.CurrentSteps = make([]*ptypes.ItsmAuditStep, len(current.Data.CurrentSteps))
	for i, step := range current.Data.CurrentSteps {
		itsmAudit.CurrentSteps[i] = &ptypes.ItsmAuditStep{
			StateID:    step.StateId,
			Name:       step.Name,
			Processors: strings.Split(step.Processors, ","),
		}
	}

	// logs
	itsmAudit.Logs = make([]*ptypes.ItsmAuditLog, 0, len(logData.Logs))
	for _, log := range logData.Logs {
		// 流程开始、结束、CRP审批 不展示
		if log.Message == itsm.AuditNodeStart || log.Message == itsm.AuditNodeEnd ||
			log.Operator == TicketOperatorNameCrpAudit {
			continue
		}

		itsmAudit.Logs = append(itsmAudit.Logs, &ptypes.ItsmAuditLog{
			Operator:  log.Operator,
			OperateAt: log.OperateAt,
			Message:   log.Message,
		})
	}

	// 如果itsm审批流已经到了CRP阶段，需要赋值为结束状态
	if len(current.Data.CurrentSteps) > 0 && current.Data.CurrentSteps[0].StateId == c.crpAuditNode.ID {
		itsmAudit.Status = enumor.RPTicketStatusDone
		itsmAudit.StatusName = itsmAudit.Status.Name()
		itsmAudit.CurrentSteps = itsmAudit.CurrentSteps[:0]
	}

	return itsmAudit, nil
}

// GetCrpCurrentApprove 查询当前审批节点
func (c *Controller) GetCrpCurrentApprove(kt *kit.Kit, orderID string) ([]*ptypes.CrpAuditStep, error) {
	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderQueryMethod,
		},
		Params: &cvmapi.QueryPlanOrderParam{
			OrderIds: []string{orderID},
		},
	}
	resp, err := c.crpCli.QueryPlanOrder(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query crp plan order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to query crp plan order, code: %d, msg: %s, order id: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, orderID, kt.Rid)
		return nil, fmt.Errorf("failed to query crp plan order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to query crp plan order, for result is empty, order id: %s, rid: %s", orderID, kt.Rid)
		return nil, errors.New("failed to query crp plan order, for result is empty")
	}

	orderItem, ok := resp.Result[orderID]
	if !ok {
		logs.Errorf("query crp plan order return no result by order id: %s, rid: %s", orderID, kt.Rid)
		return nil, fmt.Errorf("query crp plan order return no result by order id: %s", orderID)
	}

	// 如果processors为空，说明审批已经结束
	processors := orderItem.Data.BaseInfo.CurrentProcessor
	if processors == "" {
		return []*ptypes.CrpAuditStep{}, nil
	}

	currentStep := &ptypes.CrpAuditStep{
		StateID:    "", // CRP接口暂时没有节点的ID，后续实现审批操作功能时，必须补全这个ID
		Name:       orderItem.Data.BaseInfo.StatusDesc,
		Processors: strings.Split(processors, ";"),
	}

	return []*ptypes.CrpAuditStep{currentStep}, nil
}

// GetCrpApproveLogs 查询Crp审批记录
func (c *Controller) GetCrpApproveLogs(kt *kit.Kit, orderID string) ([]*ptypes.CrpAuditLog, error) {
	req := &cvmapi.GetApproveLogReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.GetApproveLogMethod,
		},
		Params: &cvmapi.GetApproveLogParams{
			OrderId: []string{orderID},
		},
	}

	resp, err := c.crpCli.GetApproveLog(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to get crp approve log, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get crp approve log, code: %d, msg: %s, order id: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, req.Params.OrderId, kt.Rid)
		return nil, fmt.Errorf("failed to get crp approve log, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to get crp approve log, for result is empty, order id: %s, rid: %s", req.Params.OrderId,
			kt.Rid)
		return nil, errors.New("failed to get crp approve log, for result is empty")
	}

	orderLogs, ok := resp.Result[orderID]
	if !ok {
		return []*ptypes.CrpAuditLog{}, nil
	}

	// crp返回的审批记录是倒序的，需要反转
	auditLogs := make([]*ptypes.CrpAuditLog, len(orderLogs))
	for i := len(orderLogs) - 1; i >= 0; i-- {
		auditLogs[len(orderLogs)-1-i] = &ptypes.CrpAuditLog{
			Operator:  orderLogs[i].Operator,
			OperateAt: orderLogs[i].OperateTime,
			Message:   orderLogs[i].OperateResult,
			Name:      orderLogs[i].Activity,
		}
	}

	return auditLogs, nil
}

// listAllResPlanTicket list all res plan ticket.
func (c *Controller) listAllResPlanTicket(kt *kit.Kit, listFilter *filter.Expression) ([]rtypes.RPTicketWithStatus,
	error) {

	listReq := &types.ListOption{
		Filter: listFilter,
		Page:   core.NewDefaultBasePage(),
	}

	rstDetails := make([]rtypes.RPTicketWithStatus, 0)
	for {
		rst, err := c.dao.ResPlanTicket().ListWithStatus(kt, listReq)
		if err != nil {
			return nil, err
		}

		rstDetails = append(rstDetails, rst.Details...)

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}

		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return rstDetails, nil
}
