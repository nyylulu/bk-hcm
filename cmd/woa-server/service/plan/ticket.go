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

package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	mtypes "hcm/pkg/dal/dao/types/meta"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	dtypes "hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// ListResPlanTicket list resource plan ticket.
func (s *service) ListResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListResPlanTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// default request bkBizIDs is authorized bizs.
	bkBizIDs, err := s.logics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// if len request bkBizIDs > 0, request bkBizIDs = the intersection of authorized bkBizIDs and request bkBizIDs.
	if len(req.BkBizIDs) > 0 {
		bkBizIDs = slice.Intersect(bkBizIDs, req.BkBizIDs)
	}

	// if len bkBizIDs == 0, return empty response.
	if len(bkBizIDs) == 0 {
		return core.ListResultT[any]{Details: make([]any, 0)}, nil
	}

	// convert request to filter expression.
	opt, err := convToListTicketOption(bkBizIDs, req)
	if err != nil {
		logs.Errorf("failed to convert to list option, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	rst, err := s.dao.ResPlanTicket().ListWithStatusAndRes(cts.Kit, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket with status, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	return rst, nil
}

// convToListTicketOption convert request parameters to ListOption.
func convToListTicketOption(bkBizIDs []int64, req *ptypes.ListResPlanTicketReq) (*types.ListOption, error) {
	rules := []filter.RuleFactory{
		tools.ContainersExpression("bk_biz_id", bkBizIDs),
	}
	if len(req.OpProductIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("op_product_id", req.OpProductIDs))
	}
	if len(req.PlanProductIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("plan_product_id", req.PlanProductIDs))
	}
	if len(req.TicketIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("id", req.TicketIDs))
	}
	if len(req.Statuses) > 0 {
		rules = append(rules, tools.ContainersExpression("status", req.Statuses))
	}
	if len(req.TicketTypes) > 0 {
		rules = append(rules, tools.ContainersExpression("type", req.TicketTypes))
	}
	if len(req.Applicants) > 0 {
		rules = append(rules, tools.ContainersExpression("applicant", req.Applicants))
	}
	if req.SubmitTimeRange != nil {
		drOpt, err := tools.DateRangeExpression("submitted_at", req.SubmitTimeRange)
		if err != nil {
			return nil, err
		}

		rules = append(rules, drOpt)
	}

	// copy page for modifying.
	pageCopy := &core.BasePage{
		Count: req.Page.Count,
		Start: req.Page.Start,
		Limit: req.Page.Limit,
		Sort:  req.Page.Sort,
		Order: req.Page.Order,
	}

	// if count == false, default sort by submitted_at desc.
	if !pageCopy.Count {
		if pageCopy.Sort == "" {
			pageCopy.Sort = "submitted_at"
		}
		if pageCopy.Order == "" {
			pageCopy.Order = core.Descending
		}
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: pageCopy,
	}

	return opt, nil
}

// CreateResPlanTicket create resource plan ticket.
func (s *service) CreateResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.CreateResPlanTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Create}, BizID: req.BkBizID}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	ticketID, err := s.createResPlanTicket(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if err = s.planController.CreateAuditFlow(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil
}

// createResPlanTicket create resource plan ticket.
func (s *service) createResPlanTicket(kt *kit.Kit, req *ptypes.CreateResPlanTicketReq) (string, error) {
	// convert request to resource plan ticket table slice.
	tickets, err := s.convToRPTicketTableSlice(kt, req)
	if err != nil {
		logs.Errorf("convert to resource plan ticket table slice failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketID, err := s.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ticketIDs, err := s.dao.ResPlanTicket().CreateWithTx(kt, txn, tickets)
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
		statuses := []rpts.ResPlanTicketStatusTable{
			{
				TicketID: ticketID,
				Status:   enumor.RPTicketStatusInit,
			},
		}
		if err = s.dao.ResPlanTicketStatus().CreateWithTx(kt, txn, statuses); err != nil {
			logs.Errorf("create resource plan ticket status failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		// create resource plan demands.
		demands, err := s.convToRPDemandTableSlice(kt, ticketID, req.Demands)
		if err != nil {
			logs.Errorf("convert to resource plan demand table slice failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		if _, err = s.dao.ResPlanDemand().CreateWithTx(kt, txn, demands); err != nil {
			logs.Errorf("create resource plan demand failed, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("convert resource plan ticket id %v from interface to string failed, err: %v, rid: %s", ticketID,
			err, kt.Rid)
		return "", fmt.Errorf("convert resource plan ticket id %v from interface to string failed", ticketID)
	}

	return ticketIDStr, nil
}

// convert request to ResPlanTicketTable slice.
func (s *service) convToRPTicketTableSlice(kt *kit.Kit, req *ptypes.CreateResPlanTicketReq) (
	[]rpt.ResPlanTicketTable, error) {

	// get biz org relation.
	bizOrgRel, err := s.logics.GetBizOrgRel(kt, req.BkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// calculate total os, cpu_core, memory and disk_size.
	var os, cpuCore, memory, diskSize int64
	for _, demand := range req.Demands {
		if demand.Cvm != nil {
			os += *demand.Cvm.Os
			cpuCore += *demand.Cvm.CpuCore
			memory += *demand.Cvm.Memory
		}

		if demand.Cbs != nil {
			diskSize += *demand.Cbs.DiskSize
		}
	}

	tickets := []rpt.ResPlanTicketTable{
		{
			Applicant:       kt.User,
			BkBizID:         req.BkBizID,
			BkBizName:       bizOrgRel.BkBizName,
			OpProductID:     bizOrgRel.OpProductID,
			OpProductName:   bizOrgRel.OpProductName,
			PlanProductID:   bizOrgRel.PlanProductID,
			PlanProductName: bizOrgRel.PlanProductName,
			VirtualDeptID:   bizOrgRel.VirtualDeptID,
			VirtualDeptName: bizOrgRel.VirtualDeptName,
			DemandClass:     req.DemandClass,
			UpdatedOS:       os,
			UpdatedCpuCore:  cpuCore,
			UpdatedMemory:   memory,
			UpdatedDiskSize: diskSize,
			Remark:          req.Remark,
			Creator:         kt.User,
			Reviser:         kt.User,
			SubmittedAt:     time.Now().Format(constant.DateTimeLayout),
		},
	}

	return tickets, nil
}

// getMetaMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap.
func (s *service) getMetaMaps(kt *kit.Kit) (map[string]string, map[string]mtypes.RegionArea,
	map[string]wdt.WoaDeviceTypeTable, error) {

	// get zone id name mapping.
	zoneMap, err := s.dao.WoaZone().GetZoneMap(kt)
	if err != nil {
		logs.Errorf("get zone map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get region area mapping.
	regionAreaMap, err := s.dao.WoaZone().GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("get region area map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get device type mapping.
	deviceTypeMap, err := s.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	return zoneMap, regionAreaMap, deviceTypeMap, nil
}

// getMetaNameMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap. map key is name
func (s *service) getMetaNameMaps(kt *kit.Kit) (map[string]string, map[string]mtypes.RegionArea, error) {
	zoneMap, regionAreaMap, _, err := s.getMetaMaps(kt)
	if err != nil {
		return nil, nil, err
	}

	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)
	return zoneNameMap, regionNameMap, nil
}

func getMetaNameMapsFromIDMap(zoneMap map[string]string, regionAreaMap map[string]mtypes.RegionArea) (
	map[string]string, map[string]mtypes.RegionArea) {

	zoneNameMap := make(map[string]string)
	for id, name := range zoneMap {
		zoneNameMap[name] = id
	}
	regionNameMap := make(map[string]mtypes.RegionArea)
	for _, item := range regionAreaMap {
		regionNameMap[item.RegionName] = item
	}
	return zoneNameMap, regionNameMap
}

// convert CreateResPlanDemandReq slice to ResPlanTicketTable slice.
func (s *service) convToRPDemandTableSlice(kt *kit.Kit, ticketID string, requests []ptypes.CreateResPlanDemandReq) (
	[]rpd.ResPlanDemandTable, error) {

	// get create resource plan ticket needed zoneMap, regionAreaMap and deviceTypeMap.
	zoneMap, regionAreaMap, deviceTypeMap, err := s.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	demands := make([]rpd.ResPlanDemandTable, 0, len(requests))
	for _, req := range requests {
		var cvm, cbs dtypes.JsonField
		if slices.Contains(req.DemandResTypes, enumor.DemandResTypeCVM) {
			deviceType := req.Cvm.DeviceType
			cvm, err = dtypes.NewJsonField(rpd.Cvm{
				ResMode:      req.Cvm.ResMode,
				DeviceType:   deviceType,
				DeviceClass:  deviceTypeMap[deviceType].DeviceClass,
				DeviceFamily: deviceTypeMap[deviceType].DeviceFamily,
				CoreType:     deviceTypeMap[deviceType].CoreType,
				Os:           *req.Cvm.Os,
				CpuCore:      *req.Cvm.CpuCore,
				Memory:       *req.Cvm.Memory,
			})
			if err != nil {
				logs.Errorf("cvm new json field failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if slices.Contains(req.DemandResTypes, enumor.DemandResTypeCBS) {
			cbs, err = dtypes.NewJsonField(rpd.Cbs{
				DiskType:     req.Cbs.DiskType,
				DiskTypeName: req.Cbs.DiskType.Name(),
				DiskIo:       *req.Cbs.DiskIo,
				DiskSize:     *req.Cbs.DiskSize,
			})
			if err != nil {
				logs.Errorf("cbs new json field failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		demands = append(demands, rpd.ResPlanDemandTable{
			TicketID:     ticketID,
			ObsProject:   req.ObsProject,
			ExpectTime:   req.ExpectTime,
			ZoneID:       req.ZoneID,
			ZoneName:     zoneMap[req.ZoneID],
			RegionID:     req.RegionID,
			RegionName:   regionAreaMap[req.RegionID].RegionName,
			AreaID:       regionAreaMap[req.RegionID].AreaID,
			AreaName:     regionAreaMap[req.RegionID].AreaName,
			DemandSource: req.DemandSource,
			Remark:       *req.Remark,
			Cvm:          cvm,
			Cbs:          cbs,
			Creator:      kt.User,
			Reviser:      kt.User,
		})
	}

	return demands, nil
}

// GetResPlanTicket get resource plan ticket detail.
func (s *service) GetResPlanTicket(cts *rest.Contexts) (interface{}, error) {
	ticketID := cts.PathParameter("id").String()
	if len(ticketID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("ticket id can not be empty"))
	}

	resp := new(ptypes.GetResPlanTicketResp)
	resp.ID = ticketID

	// get base info.
	baseInfo, err := s.getRPTicketBaseInfo(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket base info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.BaseInfo = baseInfo

	// authorize biz access.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: baseInfo.BkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// get status info.
	statusInfo, err := s.getRPTicketStatusInfo(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket status info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.StatusInfo = statusInfo

	// get demands.
	demands, err := s.getRPTicketDemands(cts.Kit, ticketID)
	if err != nil {
		logs.Errorf("get resource plan ticket demands failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	resp.Demands = demands

	return resp, nil
}

// getRPTicketBaseInfo get resource plan ticket base information.
func (s *service) getRPTicketBaseInfo(kt *kit.Kit, ticketID string) (*ptypes.GetRPTicketBaseInfo, error) {
	// search resource plan ticket table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket, but len details != 1")
	}

	detail := rst.Details[0]
	result := &ptypes.GetRPTicketBaseInfo{
		Applicant:       detail.Applicant,
		BkBizID:         detail.BkBizID,
		BkBizName:       detail.BkBizName,
		OpProductID:     detail.OpProductID,
		OpProductName:   detail.OpProductName,
		PlanProductID:   detail.PlanProductID,
		PlanProductName: detail.PlanProductName,
		VirtualDeptID:   detail.VirtualDeptID,
		VirtualDeptName: detail.VirtualDeptName,
		DemandClass:     detail.DemandClass,
		Remark:          detail.Remark,
		SubmittedAt:     detail.SubmittedAt,
	}

	return result, nil
}

// getRPTicketStatusInfo get resource plan ticket status information.
func (s *service) getRPTicketStatusInfo(kt *kit.Kit, ticketID string) (*ptypes.GetRPTicketStatusInfo, error) {
	// search resource plan ticket status table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanTicketStatus().List(kt, opt)
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
	}

	return result, nil
}

// getRPTicketDemands get resource plan ticket demands.
func (s *service) getRPTicketDemands(kt *kit.Kit, ticketID string) ([]ptypes.GetRPTicketDemand, error) {
	// search resource plan demand table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanDemand().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]ptypes.GetRPTicketDemand, 0, len(rst.Details))
	for _, detail := range rst.Details {
		var cvm *rpd.Cvm
		var cbs *rpd.Cbs
		if err = json.Unmarshal([]byte(detail.Cvm), &cvm); err != nil {
			logs.Errorf("failed to unmarshal cvm, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if err = json.Unmarshal([]byte(detail.Cbs), &cbs); err != nil {
			logs.Errorf("failed to unmarshal cbs, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result = append(result, ptypes.GetRPTicketDemand{
			ObsProject:   detail.ObsProject,
			ExpectTime:   detail.ExpectTime,
			ZoneID:       detail.ZoneID,
			ZoneName:     detail.ZoneName,
			RegionID:     detail.RegionID,
			RegionName:   detail.RegionName,
			AreaID:       detail.AreaID,
			AreaName:     detail.AreaName,
			DemandSource: detail.DemandSource,
			Remark:       detail.Remark,
			Cvm:          cvm,
			Cbs:          cbs,
		})
	}

	return result, nil
}
