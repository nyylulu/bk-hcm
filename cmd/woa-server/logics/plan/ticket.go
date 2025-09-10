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
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpdaotypes "hcm/pkg/dal/dao/types/resource-plan"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// CreateResPlanTicket create resource plan ticket.
func (c *Controller) CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket request, err: %s, rid: %s", err, kt.Rid)
		return "", err
	}

	// construct resource plan ticket.
	ticket, err := c.constructResPlanTicket(kt, req, kt.User)
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
func (c *Controller) constructResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq, applicant string) (
	*rpt.ResPlanTicketTable, error) {

	var originalOs, updatedOs decimal.Decimal
	var originalCpuCore, originalMemory, originalDiskSize int64
	var updatedCpuCore, updatedMemory, updatedDiskSize int64
	for _, demand := range req.Demands {
		if demand.Original != nil {
			originalOs = originalOs.Add((*demand.Original).Cvm.Os.Decimal)
			originalCpuCore += (*demand.Original).Cvm.CpuCore
			originalMemory += (*demand.Original).Cvm.Memory
			originalDiskSize += (*demand.Original).Cbs.DiskSize
		}

		if demand.Updated != nil {
			// 期望交付时间的预测需求月和其自然月必须一致，否则需要选择该周的其他时间
			et, err := times.ParseDay(demand.Updated.ExpectTime)
			if err != nil {
				logs.Errorf("failed to parse expect time, err: %v, expect_time: %s, rid: %s", err,
					demand.Updated.ExpectTime, kt.Rid)
				return nil, err
			}
			isCross, err := c.demandTime.IsDayCrossMonth(kt, et)
			if err != nil {
				logs.Errorf("failed to check if expect time is cross month, err: %v, expect_time: %s, rid: %s",
					err, et.String(), kt.Rid)
				return nil, err
			}
			if isCross {
				return nil, fmt.Errorf("expect_time should not be cross month, expect_time: %s",
					demand.Updated.ExpectTime)
			}
			updatedOs = updatedOs.Add((*demand.Updated).Cvm.Os.Decimal)
			updatedCpuCore += (*demand.Updated).Cvm.CpuCore
			updatedMemory += (*demand.Updated).Cvm.Memory
			updatedDiskSize += (*demand.Updated).Cbs.DiskSize
		}
	}

	demandsJson, err := tabletypes.NewJsonField(req.Demands)
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
		OriginalOS:       originalOs.InexactFloat64(),
		OriginalCpuCore:  originalCpuCore,
		OriginalMemory:   originalMemory,
		OriginalDiskSize: originalDiskSize,
		UpdatedOS:        updatedOs.InexactFloat64(),
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

// GetResPlanTicketAudit get resource plan ticket audit.
func (c *Controller) GetResPlanTicketAudit(kt *kit.Kit, ticketID string, bkBizID int64) (
	*ptypes.GetResPlanTicketAuditResp, error) {

	return c.resFetcher.GetResPlanTicketAudit(kt, ticketID, bkBizID)
}

// GetResPlanTicketStatusInfo get resource plan ticket status information.
func (c *Controller) GetResPlanTicketStatusInfo(kt *kit.Kit, ticketID string) (
	*ptypes.GetRPTicketStatusInfo, error) {

	return c.resFetcher.GetResPlanTicketStatusInfo(kt, ticketID)
}

// ListResPlanTicketWithRes list res plan ticket with res.
func (c *Controller) ListResPlanTicketWithRes(kt *kit.Kit, req *core.ListReq) (*ptypes.RPTicketWithStatusAndResListRst,
	error) {

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rst, err := c.dao.ResPlanTicket().ListWithStatus(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list biz resource plan ticket with status, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if len(rst.Details) == 0 {
		return &ptypes.RPTicketWithStatusAndResListRst{Count: rst.Count}, nil
	}

	details, err := c.appendFieldToListResPlanTickets(kt, rst.Details)
	if err != nil {
		logs.Errorf("failed to append field to list res plan tickets, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &ptypes.RPTicketWithStatusAndResListRst{Count: 0, Details: details}, nil
}

func (c *Controller) appendFieldToListResPlanTickets(kt *kit.Kit, details []rpdaotypes.RPTicketWithStatus) (
	[]ptypes.RPTicketWithStatusAndRes, error) {

	ticketWithRes := make([]ptypes.RPTicketWithStatusAndRes, 0, len(details))
	for _, detail := range details {
		item := ptypes.RPTicketWithStatusAndRes{
			RPTicketWithStatus: detail,
			TicketTypeName:     detail.Type.Name(),
		}

		// set status name.
		item.StatusName = detail.Status.Name()
		// 资源需求报备数量
		switch detail.Type {
		case enumor.RPTicketTypeAdd:
			item.OriginalInfo = ptypes.NewNullResourceInfo()
			item.UpdatedInfo = ptypes.NewResourceInfo(detail.UpdatedCpuCore, detail.UpdatedMemory,
				detail.UpdatedDiskSize)
		case enumor.RPTicketTypeAdjust:
			item.OriginalInfo = ptypes.NewResourceInfo(detail.OriginalCpuCore, detail.OriginalMemory,
				detail.OriginalDiskSize)
			item.UpdatedInfo = ptypes.NewResourceInfo(detail.UpdatedCpuCore, detail.UpdatedMemory,
				detail.UpdatedDiskSize)
		case enumor.RPTicketTypeDelete:
			item.OriginalInfo = ptypes.NewResourceInfo(detail.OriginalCpuCore, detail.OriginalMemory,
				detail.OriginalDiskSize)
			item.UpdatedInfo = ptypes.NewNullResourceInfo()
		default:
			logs.Warnf("failed to append field to list res plan tickets: unsupported ticket type: %s, "+
				"ticket id: %s, rid: %s", detail.Type, detail.ID, kt.Rid)
		}

		ticketWithRes = append(ticketWithRes, item)
	}
	return ticketWithRes, nil
}

// GetResPlanTicketStatusByBiz 检查ticket是否存在，并校验业务id是否正确， bizID 为-1 表示不限制业务条件
func (c *Controller) GetResPlanTicketStatusByBiz(kt *kit.Kit, ticketID string, bizID int64) (
	*ptypes.GetRPTicketStatusInfo, error) {

	// 1. 检查ticket是否存在以及业务是否匹配
	rules := []*filter.AtomRule{tools.RuleEqual("id", ticketID)}
	if bizID != constant.AttachedAllBiz {
		rules = append(rules, tools.RuleEqual("bk_biz_id", bizID))
	}
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewCountPage(),
	}

	ticketRst, err := c.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket(%s,%d), err: %v, rid: %s", ticketID, bizID, err, kt.Rid)
		return nil, err
	}

	if ticketRst.Count < 1 {
		logs.Errorf("list resource plan ticket got %d != 1, rid: %s", ticketRst.Count, kt.Rid)
		return nil, fmt.Errorf("list resource plan ticket %s by biz %d failed", ticketID, bizID)
	}

	// 2. 查询对应状态单号
	statusOpt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	statusRst, err := c.dao.ResPlanTicketStatus().List(kt, statusOpt)
	if err != nil {
		logs.Errorf("failed to list status of resource plan ticket(%s), err: %v, rid: %s", ticketID, err, kt.Rid)
		return nil, err
	}

	if len(statusRst.Details) != 1 {
		logs.Errorf("list status of resource plan ticket got %d != 1, ticket_id: %s, rid: %s",
			len(statusRst.Details), ticketID, kt.Rid)
		return nil, errors.New("list status of resource plan ticket, but len != 1")
	}

	detail := statusRst.Details[0]
	result := &ptypes.GetRPTicketStatusInfo{
		Status:     detail.Status,
		StatusName: detail.Status.Name(),
		ItsmSn:     detail.ItsmSn,
		ItsmUrl:    detail.ItsmUrl,
		CrpSn:      detail.CrpSn,
		CrpUrl:     detail.CrpUrl,
		Message:    detail.Message,
	}

	return result, nil
}
