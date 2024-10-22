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
	"encoding/json"
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rptypes "hcm/pkg/dal/dao/types/resource-plan"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	types2 "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// TransferResPlanTicket transfer res plan ticket
func (s *service) TransferResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.TransferResPlanTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode transfer resource plan ticket request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate transfer resource plan ticket request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get tickets
	ticketsInfo, err := s.getResPlanTickets(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to get resource plan tickets, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// get ticket demands
	ticketDemands, err := s.getResPlanDemandByTicketIDs(cts.Kit, ticketsInfo)
	if err != nil {
		logs.Errorf("failed to get resource plan demand by ticket ids, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update res_plan_ticket
	successfulTickets := s.updateResPlanTicketsWithDemands(cts.Kit, enumor.RPTicketTypeAdd, ticketsInfo, ticketDemands)

	var successfulIDs []string
	// insert res_plan_crp_demand with ticket demands
	for ticketID, info := range successfulTickets {
		if _, ok := ticketDemands[ticketID]; !ok {
			logs.Warnf("ticket %s has no demand, skip it, rid: %s", ticketID, cts.Kit.Rid)
			continue
		}
		err := s.createResPlanCrpDemand(cts.Kit, info)
		if err != nil {
			logs.Warnf("failed to create res plan crp demand, err: %v, ticket_id: %s, rid: %s", err, ticketID,
				cts.Kit.Rid)
			continue
		}
		successfulIDs = append(successfulIDs, ticketID)
	}

	return map[string]interface{}{"count": len(successfulIDs), "ids": successfulIDs}, nil
}

func (s *service) getResPlanTickets(kt *kit.Kit, req *ptypes.TransferResPlanTicketReq) (
	map[string]*rptypes.RPTicketWithStatus, error) {

	basePage := &core.BasePage{
		Start: 0,
		Limit: core.DefaultMaxPageLimit,
	}
	opt := req.GenListTicketsOption(basePage)

	tickets := make([]rptypes.RPTicketWithStatus, 0)
	for {
		resp, err := s.dao.ResPlanTicket().ListWithStatus(kt, opt)
		if err != nil {
			logs.Errorf("failed to list res plan tickets, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		tickets = append(tickets, resp.Details...)

		if len(resp.Details) < int(opt.Page.Limit) {
			break
		}
		opt.Page.Start += uint32(opt.Page.Limit)
	}

	rst := make(map[string]*rptypes.RPTicketWithStatus)
	for idx, ticket := range tickets {
		rst[ticket.ID] = &tickets[idx]
	}

	return rst, nil
}

func (s *service) getResPlanDemandByTicketIDs(kt *kit.Kit, tickets map[string]*rptypes.RPTicketWithStatus) (
	map[string][]*rpd.ResPlanDemandTable, error) {

	ticketIDs := make([]string, 0, len(tickets))
	for id := range tickets {
		ticketIDs = append(ticketIDs, id)
	}
	rules := []filter.RuleFactory{
		tools.ContainersExpression("ticket_id", ticketIDs),
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}

	demands := make([]rpd.ResPlanDemandTable, 0)
	for {
		resp, err := s.dao.ResPlanDemand().List(kt, opt)
		if err != nil {
			logs.Errorf("failed to list res plan demands by ticket ids, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		demands = append(demands, resp.Details...)

		if len(resp.Details) < int(opt.Page.Limit) {
			break
		}
		opt.Page.Start += uint32(opt.Page.Limit)
	}

	rst := make(map[string][]*rpd.ResPlanDemandTable)
	for idx, demand := range demands {
		ticketID := demand.TicketID
		if _, ok := rst[ticketID]; !ok {
			rst[ticketID] = make([]*rpd.ResPlanDemandTable, 0)
		}
		rst[ticketID] = append(rst[ticketID], &demands[idx])
	}

	return rst, nil
}

func (s *service) getCrpDemandIDByCrpOrderID(kt *kit.Kit, orderID string) ([]string, error) {
	req := &ptypes.ListResPlanDemandReq{
		OrderIDs: []string{orderID},
	}

	rst, err := s.planController.ListCrpDemands(kt, req)
	if err != nil {
		return nil, err
	}

	rstDemandIDs := make([]string, 0)
	for _, item := range rst {
		rstDemandIDs = append(rstDemandIDs, item.CrpDemandID)
	}

	return rstDemandIDs, err
}

func (s *service) updateResPlanTicketsWithDemands(kt *kit.Kit, ticketType enumor.RPTicketType,
	tickets map[string]*rptypes.RPTicketWithStatus,
	demands map[string][]*rpd.ResPlanDemandTable) map[string]*rptypes.RPTicketWithStatus {

	successfulTickets := make(map[string]*rptypes.RPTicketWithStatus)
	for ticketID, ticketInfo := range tickets {
		ticketDemands, ok := demands[ticketID]
		if !ok {
			logs.Warnf("the ticket's demands from res_plan_demand is empty, ticket_id: %s, rid: %s", ticketID, kt.Rid)
			continue
		}

		// merge source ticket info & ticket demands(json)
		demandsJson, err := convTicketDemandsInfoToJson(kt, ticketInfo.DemandClass, ticketDemands)
		if err != nil {
			logs.Errorf("failed to convert ticket demands to json, err: %v, ticket_id: %s, rid: %s", err, ticketID,
				kt.Rid)
			continue
		}

		expr := tools.EqualExpression("id", ticketID)
		model := &rpt.ResPlanTicketTable{
			Type:    ticketType,
			Demands: demandsJson,
			Reviser: ticketInfo.Reviser,
		}
		err = s.dao.ResPlanTicket().Update(kt, expr, model)
		if err != nil {
			logs.Errorf("failed to update res plan ticket, err: %v, ticket_id: %s, rid: %s", err, ticketID, kt.Rid)
			continue
		}
		successfulTickets[ticketID] = ticketInfo
	}

	return successfulTickets
}

func convTicketDemandsInfoToJson(kt *kit.Kit, class enumor.DemandClass, oldDemands []*rpd.ResPlanDemandTable) (
	types2.JsonField, error) {

	demandIDs := make([]string, 0)
	newDemands := make(rpt.ResPlanDemands, 0)
	for _, od := range oldDemands {
		newDemand := rpt.ResPlanDemand{
			DemandClass: class,
			Original:    nil, // 历史预测单都是新增单，没有original
			Updated: &rpt.UpdatedRPDemandItem{
				ObsProject:   od.ObsProject,
				ExpectTime:   od.ExpectTime,
				ZoneID:       od.ZoneID,
				ZoneName:     od.ZoneName,
				RegionID:     od.RegionID,
				RegionName:   od.RegionName,
				AreaID:       od.AreaID,
				AreaName:     od.AreaName,
				DemandSource: od.DemandSource,
				Remark:       od.Remark,
			},
		}

		var oldCvm rpd.Cvm
		var oldCbs rpd.Cbs
		if err := json.Unmarshal([]byte(od.Cvm), &oldCvm); err != nil {
			logs.Errorf("failed to unmarshal old cvm, err: %v, demand_id: %s, rid: %s", err, od.ID, kt.Rid)
			return "", err
		}
		if err := json.Unmarshal([]byte(od.Cbs), &oldCbs); err != nil {
			logs.Errorf("failed to unmarshal old cbs, err: %v, demand_id: %s, rid: %s", err, od.ID, kt.Rid)
			return "", err
		}

		newDemand.Updated.Cvm = rpt.Cvm{
			ResMode:      oldCvm.ResMode,
			DeviceType:   oldCvm.DeviceType,
			DeviceClass:  oldCvm.DeviceClass,
			DeviceFamily: oldCvm.DeviceFamily,
			CoreType:     oldCvm.CoreType,
			Os:           oldCvm.Os,
			CpuCore:      oldCvm.CpuCore,
			Memory:       oldCvm.Memory,
		}

		newDemand.Updated.Cbs = rpt.Cbs{
			DiskType:     oldCbs.DiskType,
			DiskTypeName: oldCbs.DiskTypeName,
			DiskIo:       oldCbs.DiskIo,
			DiskSize:     oldCbs.DiskSize,
		}

		demandIDs = append(demandIDs, od.ID)
		newDemands = append(newDemands, newDemand)
	}

	demandsJson, err := types2.NewJsonField(newDemands)
	if err != nil {
		logs.Errorf("demands json marshal failed, err: %v, demand_ids: %v, rid: %s", err, demandIDs, kt.Rid)
		return "", err
	}

	return demandsJson, nil
}

func (s *service) createResPlanCrpDemand(kt *kit.Kit, ticketInfo *rptypes.RPTicketWithStatus) error {

	// 只有审批通过的单据才更新demand表
	if ticketInfo.Status != enumor.RPTicketStatusDone {
		return nil
	}

	demandIDs, err := s.getCrpDemandIDByCrpOrderID(kt, ticketInfo.CrpSn)
	if err != nil {
		logs.Errorf("failed to get crp demand id by crp order id, err: %v, crp_sn: %s, rid: %s", err, ticketInfo.CrpSn,
			kt.Rid)
		return err
	}

	insertDemandIDs, err := s.demandIDNotExistsInResPlanCrpDemand(kt, demandIDs)
	if err != nil {
		logs.Errorf("failed to get demand id not exists in res plan crp demand, err: %v, ticket_id: %s, rid: %s",
			err, ticketInfo.ID, kt.Rid)
		return err
	}

	if len(insertDemandIDs) == 0 {
		logs.Infof("demands for the ticket do not need to be created, ticket_id: %s, rid: %s", ticketInfo.ID, kt.Rid)
		return nil
	}

	inserts := make([]rpcd.ResPlanCrpDemandTable, len(insertDemandIDs))
	for idx, dID := range insertDemandIDs {
		inserts[idx] = rpcd.ResPlanCrpDemandTable{
			CrpDemandID:     dID,
			Locked:          converter.ValToPtr(enumor.CrpDemandUnLocked),
			DemandClass:     ticketInfo.DemandClass,
			BkBizID:         ticketInfo.BkBizID,
			BkBizName:       ticketInfo.BkBizName,
			OpProductID:     ticketInfo.OpProductID,
			OpProductName:   ticketInfo.OpProductName,
			PlanProductID:   ticketInfo.PlanProductID,
			PlanProductName: ticketInfo.PlanProductName,
			VirtualDeptID:   ticketInfo.VirtualDeptID,
			VirtualDeptName: ticketInfo.VirtualDeptName,
			Creator:         ticketInfo.Creator,
			Reviser:         ticketInfo.Reviser,
		}
	}
	_, err = s.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return s.dao.ResPlanCrpDemand().CreateWithTx(kt, txn, inserts)
	})
	if err != nil {
		logs.Errorf("failed to create resource plan crp demands, err: %v, ticket_id: %s, rid: %s", err,
			ticketInfo.ID, kt.Rid)
	}
	return err
}

// demandIDNotExistsInResPlanCrpDemand 保证幂等，只返回在res_plan_crp_demand表中不存在的demand_id
func (s *service) demandIDNotExistsInResPlanCrpDemand(kt *kit.Kit, demandIDs []string) ([]int64, error) {
	rstDemandIDs := make([]int64, 0, len(demandIDs))

	optDemandList := make([]int64, 0, len(demandIDs))
	for _, idString := range demandIDs {
		dID, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			logs.Warnf("failed to parse crp demand id to int64, err: %v, demand_id: %s, rid: %s", err, idString,
				kt.Rid)
			continue
		}

		optDemandList = append(optDemandList, dID)
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				tools.ContainersExpression("crp_demand_id", optDemandList),
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}

	existsDemandIDs := make(map[int64]interface{})
	for {
		resp, err := s.dao.ResPlanCrpDemand().List(kt, opt)
		if err != nil {
			logs.Errorf("failed to list res_plan_crp_demand by demand ids, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, demand := range resp.Details {
			existsDemandIDs[demand.CrpDemandID] = nil
		}

		if len(resp.Details) < int(opt.Page.Limit) {
			break
		}
		opt.Page.Start += uint32(opt.Page.Limit)
	}

	for _, id := range optDemandList {
		if _, ok := existsDemandIDs[id]; !ok {
			rstDemandIDs = append(rstDemandIDs, id)
		}
	}

	return rstDemandIDs, nil
}
