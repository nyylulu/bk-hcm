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
	"hcm/pkg/tools/converter"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
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
	opt := req.GenListTicketsOption()
	ticketsInfo, err := s.getResPlanTickets(cts.Kit, opt)
	if err != nil {
		logs.Errorf("failed to get resource plan tickets, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// get ticket demands
	opt = req.GenListDemandsOption()
	ticketDemands, err := s.getResPlanDemandByTicketIDs(cts.Kit, opt)
	if err != nil {
		logs.Errorf("failed to get resource plan demand by ticket ids, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update res_plan_ticket
	err = s.updateResPlanTicketsWithDemands(cts.Kit, enumor.RPTicketTypeAdd, ticketsInfo, ticketDemands)
	if err != nil {
		logs.Errorf("failed to update resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	var successfulIDs []string
	// insert res_plan_crp_demand with ticket demands
	for ticketID, info := range ticketsInfo {
		if _, ok := ticketDemands[ticketID]; !ok {
			logs.Warnf("ticket %s has no demand, skip it, rid: %s", ticketID, cts.Kit.Rid)
			continue
		}
		err := s.createResPlanCrpDemand(cts.Kit, info, ticketDemands[ticketID])
		if err != nil {
			logs.Warnf("failed to create res plan crp demand, err: %v, ticket_id: %s, rid: %s", err, ticketID,
				cts.Kit.Rid)
			continue
		}
		successfulIDs = append(successfulIDs, ticketID)
	}

	return map[string]interface{}{"count": len(successfulIDs), "ids": successfulIDs}, nil
}

func (s *service) getResPlanTickets(kt *kit.Kit, opt *types.ListOption) (map[string]*rptypes.RPTicketWithStatus,
	error) {
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
	for _, ticket := range tickets {
		rst[ticket.ID] = &ticket
	}

	return rst, nil
}

func (s *service) getResPlanDemandByTicketIDs(kt *kit.Kit, opt *types.ListOption) (
	map[string][]*rpd.ResPlanDemandTable, error) {

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
	for _, demand := range demands {
		ticketID := demand.TicketID
		if _, ok := rst[ticketID]; !ok {
			rst[ticketID] = make([]*rpd.ResPlanDemandTable, 0)
		}
		rst[ticketID] = append(rst[ticketID], &demand)
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
	tickets map[string]*rptypes.RPTicketWithStatus, demands map[string][]*rpd.ResPlanDemandTable) error {

	for ticketID, ticketInfo := range tickets {
		ticketDemands := demands[ticketID]

		// merge source ticket info & ticket demands(json)
		demandsJson, err := convTicketDemandsInfoToJson(kt, ticketInfo.DemandClass, ticketDemands)
		if err != nil {
			logs.Errorf("failed to convert ticket demands to json, err: %v, ticket_id: %s, rid: %s", err, ticketID,
				kt.Rid)
			return err
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
			return err
		}
	}

	return nil
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

func (s *service) createResPlanCrpDemand(kt *kit.Kit, ticketInfo *rptypes.RPTicketWithStatus,
	demands []*rpd.ResPlanDemandTable) error {

	demandIDs, err := s.getCrpDemandIDByCrpOrderID(kt, ticketInfo.CrpSn)
	if err != nil {
		logs.Errorf("failed to get crp demand id by crp order id, err: %v, crp_sn: %s, rid: %s", err, ticketInfo.CrpSn,
			kt.Rid)
		return err
	}

	if len(demands) != len(demandIDs) {
		logs.Errorf("hcm demands count not equal to crp demands count, ticket_id: %s, crp_sn: %s, rid: %s",
			ticketInfo.ID, ticketInfo.CrpSn, kt.Rid)
		return errors.New("ticket demands count not equal to crp demands count")
	}

	inserts := make([]rpcd.ResPlanCrpDemandTable, len(demands))
	for idx, demand := range demands {
		dID, err := strconv.ParseInt(demandIDs[idx], 10, 64)
		if err != nil {
			logs.Errorf("failed to parse crp demand id to int64, err: %v, demand_id: %s, rid: %s", err, demandIDs[idx],
				kt.Rid)
			return err
		}

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
			Creator:         demand.Creator,
			Reviser:         demand.Reviser,
		}
	}
	_, err = s.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return s.dao.ResPlanCrpDemand().CreateWithTx(kt, txn, inserts)
	})
	if err != nil {
		logs.Errorf("failed to create resource plan crp demands, err: %v, rid: %s", err, kt.Rid)
	}
	return err
}
