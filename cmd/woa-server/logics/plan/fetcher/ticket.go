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

package fetcher

import (
	"encoding/json"
	"errors"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpdaotypes "hcm/pkg/dal/dao/types/resource-plan"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// ListAllResPlanTicket list all res plan ticket.
func (f *ResPlanFetcher) ListAllResPlanTicket(kt *kit.Kit, listFilter *filter.Expression) (
	[]rpdaotypes.RPTicketWithStatus, error) {

	listReq := &types.ListOption{
		Filter: listFilter,
		Page:   core.NewDefaultBasePage(),
	}

	rstDetails := make([]rpdaotypes.RPTicketWithStatus, 0)
	for {
		rst, err := f.dao.ResPlanTicket().ListWithStatus(kt, listReq)
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

// GetTicketInfo get ticket info
func (f *ResPlanFetcher) GetTicketInfo(kt *kit.Kit, ticketID string) (*ptypes.TicketInfo, error) {
	base, err := f.getTicketBaseInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v", err)
		return nil, err
	}

	var demands rpt.ResPlanDemands
	if err = json.Unmarshal([]byte(base.Demands), &demands); err != nil {
		logs.Errorf("failed to unmarshal demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	status, err := f.getTicketStatusInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket status info, err: %v", err)
		return nil, err
	}

	brief := &ptypes.TicketInfo{
		ID:               ticketID,
		Type:             base.Type,
		Applicant:        base.Applicant,
		BkBizID:          base.BkBizID,
		BkBizName:        base.BkBizName,
		OpProductID:      base.OpProductID,
		OpProductName:    base.OpProductName,
		PlanProductID:    base.PlanProductID,
		PlanProductName:  base.PlanProductName,
		VirtualDeptID:    base.VirtualDeptID,
		VirtualDeptName:  base.VirtualDeptName,
		DemandClass:      base.DemandClass,
		OriginalCpuCore:  base.OriginalCpuCore,
		OriginalMemory:   base.OriginalMemory,
		OriginalDiskSize: base.OriginalDiskSize,
		UpdatedCpuCore:   base.UpdatedCpuCore,
		UpdatedMemory:    base.UpdatedMemory,
		UpdatedDiskSize:  base.UpdatedDiskSize,
		Remark:           base.Remark,
		Demands:          demands,
		SubmittedAt:      base.SubmittedAt,
		Status:           status.Status,
		ItsmSn:           status.ItsmSn,
		ItsmUrl:          status.ItsmUrl,
		CrpSn:            status.CrpSn,
		CrpUrl:           status.CrpUrl,
	}

	return brief, nil
}

func (f *ResPlanFetcher) getTicketBaseInfo(kt *kit.Kit, ticketID string) (*rpt.ResPlanTicketTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := f.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket, but len details != 1")
	}

	return &rst.Details[0], nil
}

func (f *ResPlanFetcher) getTicketStatusInfo(kt *kit.Kit, ticketID string) (*rpts.ResPlanTicketStatusTable, error) {
	// search resource plan ticket table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := f.dao.ResPlanTicketStatus().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket status, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket status, but len details != 1")
	}

	return &rst.Details[0], nil
}

// GetResPlanTicketAudit get resource plan ticket audit.
func (f *ResPlanFetcher) GetResPlanTicketAudit(kt *kit.Kit, ticketID string, bkBizID int64) (
	*ptypes.GetResPlanTicketAuditResp, error) {

	resp := new(ptypes.GetResPlanTicketAuditResp)
	resp.TicketID = ticketID

	// 查询Itsm单号和Crp单号
	statusInfo, err := f.GetResPlanTicketStatusInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket status info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	itsmAudit, crpAudit, err := f.getItsmAndCrpAuditStatus(kt, bkBizID, statusInfo)
	if err != nil {
		logs.Errorf("get itsm and crp audit status failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	resp.ItsmAudit = itsmAudit
	resp.CrpAudit = crpAudit

	return resp, nil
}

// GetResPlanTicketStatusInfo get resource plan ticket status information.
func (f *ResPlanFetcher) GetResPlanTicketStatusInfo(kt *kit.Kit, ticketID string) (
	*ptypes.GetRPTicketStatusInfo, error) {

	// search resource plan ticket status table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := f.dao.ResPlanTicketStatus().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket status, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket status, but len details != 1")
	}

	detail := rst.Details[0]
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
