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
	"slices"
	"strconv"

	mtypes "hcm/cmd/woa-server/types/meta"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	rpdtablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

func convAdjustAbleQueryParam(req *ptypes.AdjustAbleDemandsReq) *cvmapi.CvmCbsAdjustAblePlanQueryParam {
	reqParams := new(cvmapi.CvmCbsAdjustAblePlanQueryParam)

	if len(req.RegionName) > 0 {
		reqParams.CityName = req.RegionName
	}

	if len(req.DeviceFamily) > 0 {
		reqParams.InstanceFamily = req.DeviceFamily
	}

	if len(req.DeviceType) > 0 {
		reqParams.InstanceModel = req.DeviceType
	}

	if len(req.ExpectTime) > 0 {
		reqParams.UseTime = req.ExpectTime
	}

	if len(req.PlanProductName) > 0 {
		reqParams.PlanProductName = req.PlanProductName
	}

	if len(req.ObsProject) > 0 {
		reqParams.ProjectName = string(req.ObsProject)
	}

	if len(req.DiskType) > 0 {
		// TODO 因CRP会在拆单时改变云盘的类型，且云盘不会参与模糊匹配，因此查询时暂时不考虑diskType参数.
		// reqParams.DiskTypeName = req.DiskType.Name()
	}

	if len(req.ResMode) > 0 {
		reqParams.ResourceMode = string(req.ResMode)
	}

	return reqParams
}

// QueryAdjustAbleDemands query demands that can be adjusted.
func (c *Controller) QueryAdjustAbleDemands(kt *kit.Kit, req *ptypes.AdjustAbleDemandsReq) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// init request parameter.
	queryReq := &cvmapi.CvmCbsAdjustAblePlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsAdjustAblePlanQueryMethod,
		},
		Params: convAdjustAbleQueryParam(req),
	}

	rst, err := c.crpCli.QueryAdjustAbleDemand(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		logs.Errorf("failed to query adjust able demands, err: %s, req: %+v, rid: %s", err, *queryReq.Params,
			kt.Rid)
		return nil, err
	}

	if rst.Error.Code != 0 {
		logs.Errorf("failed to query adjust able demands, err: %s, crp_trace: %s, rid: %s", rst.Error.Message,
			rst.TraceId, kt.Rid)
		return nil, errors.New(rst.Error.Message)
	}

	if rst.Result == nil || len(rst.Result.Data) == 0 {
		logs.Errorf("failed to query adjust able demands, return is empty, crp_trace: %s, rid: %s",
			rst.TraceId, kt.Rid)
		return nil, errors.New("no demands can be adjusted in CRP")
	}

	return rst.Result.Data, nil
}

// AdjustBizResPlanDemand adjust biz res plan demand.
func (c *Controller) AdjustBizResPlanDemand(kt *kit.Kit, req *ptypes.AdjustRPDemandReq, bkBizID int64,
	bizOrgRel *mtypes.BizOrgRel) (ticketID string, retErr error) {

	demandIDs := slice.Map(req.Adjusts,
		func(adjust ptypes.AdjustRPDemandReqElem) string { return adjust.DemandID })

	// check whether all crp demand belong to the biz.
	allBelong, err := c.AreAllDemandBelongToBiz(kt, demandIDs, bkBizID)
	if err != nil {
		logs.Errorf("failed to check whether all demand belong to biz, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if !allBelong {
		logs.Errorf("not all adjust demand belong to biz: %d, rid: %s", bkBizID, kt.Rid)
		return "", fmt.Errorf("not all adjust crp demand belong to biz: %d", bkBizID)
	}

	// examine whether all resource plan demand classes are the same, and get the demand class.
	demandClass, err := c.ExamineDemandClass(kt, demandIDs)
	if err != nil {
		logs.Errorf("failed to examine demand class, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// lock all resource plan demand.
	lockReq := &rpproto.ResPlanDemandLockOpReq{
		IDs: demandIDs,
	}
	if err = c.client.DataService().Global.ResourcePlan.LockResPlanDemand(kt, lockReq); err != nil {
		logs.Errorf("failed to lock all resource plan demand, err: %v, demandIDs: %v, rid: %s", err, demandIDs,
			kt.Rid)
		return "", err
	}

	// defer is used to unlock all resource plan demand when some errors occur.
	defer func() {
		if retErr != nil {
			if tmpErr := c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, lockReq); tmpErr != nil {
				logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", tmpErr, kt.Rid)
			}
		}
	}()

	// construct adjust biz resource plan demand request.
	adjustReq, err := c.constructAdjustReq(kt, bizOrgRel, demandClass, req)
	if err != nil {
		logs.Errorf("failed to construct adjust resource plan ticket request, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// create cancel resource plan ticket.
	ticketID, err = c.CreateResPlanTicket(kt, adjustReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// create adjust resource plan ticket itsm audit flow.
	if err = c.CreateAuditFlow(kt, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return ticketID, nil
}

// CancelBizResPlanDemand cancel biz res plan demand.
func (c *Controller) CancelBizResPlanDemand(kt *kit.Kit, req *ptypes.CancelRPDemandReq, bkBizID int64,
	bizOrgRel *mtypes.BizOrgRel) (string, error) {

	demandIDs := slice.Map(req.CancelDemands,
		func(cancel ptypes.CancelRPDemandReqElem) string { return cancel.DemandID })

	// check whether all crp demand belong to the biz.
	allBelong, err := c.AreAllDemandBelongToBiz(kt, demandIDs, bkBizID)
	if err != nil {
		logs.Errorf("failed to check whether all demand belong to biz, err: %v, demand_ids: %v, bk_biz_id: %d, rid: %s",
			err, demandIDs, bkBizID, kt.Rid)
		return "", err
	}

	if !allBelong {
		logs.Errorf("not all adjust demand belong to biz: %d, rid: %s", bkBizID, kt.Rid)
		return "", fmt.Errorf("not all adjust crp demand belong to biz: %d", bkBizID)
	}

	// examine whether all resource plan demand classes are the same, and get the demand class.
	demandClass, err := c.ExamineDemandClass(kt, demandIDs)
	if err != nil {
		logs.Errorf("failed to examine demand class, err: %v, demand_ids: %v, rid: %s", err, demandIDs, kt.Rid)
		return "", err
	}

	// lock all resource plan demand.
	lockReq := &rpproto.ResPlanDemandLockOpReq{
		IDs: demandIDs,
	}
	if err = c.client.DataService().Global.ResourcePlan.LockResPlanDemand(kt, lockReq); err != nil {
		logs.Errorf("failed to lock all resource plan demand, err: %v, demandIDs: %v, rid: %s", err,
			demandIDs, kt.Rid)
		return "", err
	}

	// defer is used to unlock all resource plan demand when some errors occur.
	defer func() {
		if err != nil {
			if tmpErr := c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, lockReq); tmpErr != nil {
				logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", tmpErr, kt.Rid)
			}
		}
	}()

	// construct cancel biz resource plan demand request.
	cancelReq, err := c.constructCancelReq(kt, bizOrgRel, demandClass, req.CancelDemands)
	if err != nil {
		logs.Errorf("failed to construct adjust resource plan ticket request, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// create cancel resource plan ticket.
	ticketID, err := c.CreateResPlanTicket(kt, cancelReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// create adjust resource plan ticket itsm audit flow.
	if err = c.CreateAuditFlow(kt, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return ticketID, nil
}

// AreAllDemandBelongToBiz return whether all input demand ids belong to input biz.
func (c *Controller) AreAllDemandBelongToBiz(kt *kit.Kit, demandIDs []string, bkBizID int64) (bool, error) {
	if len(demandIDs) == 0 {
		return false, errors.New("demand ids is empty")
	}

	listReq := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", demandIDs),
				tools.RuleEqual("bk_biz_id", bkBizID),
			),
			Page: core.NewCountPage(),
		},
	}

	rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	return len(demandIDs) == int(rst.Count), nil
}

// ExamineDemandClass examine whether all demands are the same demand class, and return the demand class.
func (c *Controller) ExamineDemandClass(kt *kit.Kit, demandIDs []string) (enumor.DemandClass, error) {
	listReq := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Fields: []string{"demand_class"},
			Filter: tools.ContainersExpression("id", demandIDs),
			Page:   core.NewDefaultBasePage(),
		},
	}

	rstDetails := make([]rpdtablers.ResPlanDemandTable, 0)
	for {
		rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		rstDetails = append(rstDetails, rst.Details...)

		if len(rst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	if len(rstDetails) == 0 {
		logs.Errorf("list resource plan demand, but len detail is 0, rid: %s", kt.Rid)
		return "", errors.New("list resource plan demand, but len detail is 0")
	}

	demandClass := rstDetails[0].DemandClass
	for _, detail := range rstDetails {
		if detail.DemandClass != demandClass {
			logs.Errorf("not all demand classes are the same, rid: %s", kt.Rid)
			return "", errors.New("not all demand classes are the same")
		}
	}

	return demandClass, nil
}

// constructAdjustReq construct create resource plan ticket request of adjust.
func (c *Controller) constructAdjustReq(kt *kit.Kit, bizOrgRel *mtypes.BizOrgRel, demandClass enumor.DemandClass,
	req *ptypes.AdjustRPDemandReq) (*CreateResPlanTicketReq, error) {

	updateDemands := make([]ptypes.AdjustRPDemandReqElem, 0)
	delayDemands := make([]ptypes.AdjustRPDemandReqElem, 0)
	for _, adjust := range req.Adjusts {
		switch adjust.AdjustType {
		case enumor.RPDemandAdjustTypeUpdate:
			updateDemands = append(updateDemands, adjust)
		case enumor.RPDemandAdjustTypeDelay:
			delayDemands = append(delayDemands, adjust)
		default:
			return nil, fmt.Errorf("unsupported resource plan demand adjust type: %s", adjust.AdjustType)
		}
	}

	// construct update demands.
	updates, err := c.constructUpdateDemands(kt, updateDemands, demandClass)
	if err != nil {
		logs.Errorf("failed to construct update demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct delay demands.
	delays, err := c.constructDelayDemands(kt, delayDemands, demandClass)
	if err != nil {
		logs.Errorf("failed to construct delay demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	demands := append(updates, delays...)
	adjustReq := &CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeAdjust,
		DemandClass: demandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     demands,
	}

	return adjustReq, nil
}

// constructCancelReq construct create resource plan ticket request of cancel.
func (c *Controller) constructCancelReq(kt *kit.Kit, bizOrgRel *mtypes.BizOrgRel, demandClass enumor.DemandClass,
	cancelDemands []ptypes.CancelRPDemandReqElem) (*CreateResPlanTicketReq, error) {

	originDemandMap := make(map[string]ptypes.CreateResPlanDemandResource)
	for _, cancelD := range cancelDemands {
		originDemandMap[cancelD.DemandID] = ptypes.CreateResPlanDemandResource{
			Os:       decimal.NewFromInt(0),
			CpuCore:  cancelD.RemainedCpuCore,
			Memory:   0,
			DiskSize: 0,
		}
	}

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	demandOriginMap, err := c.constructOriginalDemandMap(kt, originDemandMap)
	if err != nil {
		logs.Errorf("failed to construct original demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct demands.
	demands := make(rpt.ResPlanDemands, 0, len(demandOriginMap))
	for _, origin := range demandOriginMap {
		demands = append(demands, rpt.ResPlanDemand{
			DemandClass: demandClass,
			Original:    origin,
		})
	}

	req := &CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeDelete,
		DemandClass: demandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     demands,
	}

	return req, nil
}

// constructUpdateDemands construct update demand.
func (c *Controller) constructUpdateDemands(kt *kit.Kit, updates []ptypes.AdjustRPDemandReqElem,
	demandClass enumor.DemandClass) ([]rpt.ResPlanDemand, error) {

	if len(updates) == 0 {
		return nil, nil
	}

	// get create resource plan ticket needed zoneMap, regionAreaMap and deviceTypeMap.
	zoneMap, regionAreaMap, deviceTypeMap, err := c.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("failed to get meta maps, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	originDemandMap := slice.FuncToMap(updates,
		func(update ptypes.AdjustRPDemandReqElem) (string, ptypes.CreateResPlanDemandResource) {
			return update.DemandID, update.OriginalInfo.GetResource()
		})
	demandOriginMap, err := c.constructOriginalDemandMap(kt, originDemandMap)
	if err != nil {
		logs.Errorf("failed to construct original demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]rpt.ResPlanDemand, len(updates))
	for idx, update := range updates {
		result[idx] = rpt.ResPlanDemand{
			DemandClass: demandClass,
			Original:    demandOriginMap[update.DemandID],
			Updated: &rpt.UpdatedRPDemandItem{
				ObsProject:   update.UpdatedInfo.ObsProject,
				ExpectTime:   update.UpdatedInfo.ExpectTime,
				ZoneID:       update.UpdatedInfo.ZoneID,
				ZoneName:     zoneMap[update.UpdatedInfo.ZoneID],
				RegionID:     update.UpdatedInfo.RegionID,
				RegionName:   regionAreaMap[update.UpdatedInfo.RegionID].RegionName,
				AreaID:       regionAreaMap[update.UpdatedInfo.RegionID].AreaID,
				AreaName:     regionAreaMap[update.UpdatedInfo.RegionID].AreaName,
				DemandSource: update.DemandSource,
			},
		}

		// TODO 目前CRP接口不支持修改CBS类型的预测，且CBS类型的预测不会产生罚金，因此暂时不允许单独修改CBS类型的预测
		if len(update.UpdatedInfo.DemandResTypes) == 1 &&
			slices.Contains(update.UpdatedInfo.DemandResTypes, enumor.DemandResTypeCBS) {
			return nil, errors.New("cannot adjust cbs plan demand")
		}

		result[idx].Updated.Cvm.ResMode = update.UpdatedInfo.Cvm.ResMode
		result[idx].Updated.Cvm.DeviceType = update.UpdatedInfo.Cvm.DeviceType
		result[idx].Updated.Cvm.DeviceClass = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].DeviceClass
		result[idx].Updated.Cvm.DeviceFamily = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].DeviceFamily
		result[idx].Updated.Cvm.CoreType = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].CoreType
		result[idx].Updated.Cvm.Os = types.Decimal{Decimal: cvt.PtrToVal(update.UpdatedInfo.Cvm.Os)}
		result[idx].Updated.Cvm.CpuCore = cvt.PtrToVal(update.UpdatedInfo.Cvm.CpuCore)
		result[idx].Updated.Cvm.Memory = cvt.PtrToVal(update.UpdatedInfo.Cvm.Memory)

		if slices.Contains(update.UpdatedInfo.DemandResTypes, enumor.DemandResTypeCBS) {
			result[idx].Updated.Cbs.DiskType = update.UpdatedInfo.Cbs.DiskType
			result[idx].Updated.Cbs.DiskTypeName = update.UpdatedInfo.Cbs.DiskType.Name()
			result[idx].Updated.Cbs.DiskIo = cvt.PtrToVal(update.UpdatedInfo.Cbs.DiskIo)
			result[idx].Updated.Cbs.DiskSize = cvt.PtrToVal(update.UpdatedInfo.Cbs.DiskSize)
		}
	}

	return result, nil
}

// constructOriginalDemandMap construct original demand map.
// return demand id and demand class map, demand id and remain cpu core map.
func (c *Controller) constructOriginalDemandMap(kt *kit.Kit,
	originDemandMap map[string]ptypes.CreateResPlanDemandResource) (map[string]*rpt.OriginalRPDemandItem, error) {

	if len(originDemandMap) == 0 {
		return make(map[string]*rpt.OriginalRPDemandItem), nil
	}

	demandIDs := maps.Keys(originDemandMap)

	// get demand details
	listReq := &ptypes.ListResPlanDemandReq{
		DemandIDs: demandIDs,
		Page:      core.NewDefaultBasePage(),
	}
	demands, _, err := c.listAllResPlanDemand(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	demandOriginMap := make(map[string]*rpt.OriginalRPDemandItem)
	for _, demand := range demands {
		// 变更前资源量以请求中的变更前数据为准
		originOS := originDemandMap[demand.ID].Os

		expectTimeStr, err := times.TransTimeStrWithLayout(strconv.Itoa(demand.ExpectTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Errorf("failed to convert expect time to string, err: %v, expect time: %d, rid: %s", err,
				demand.ExpectTime, kt.Rid)
			return nil, err
		}

		demandOriginMap[demand.ID] = &rpt.OriginalRPDemandItem{
			DemandID:   demand.ID,
			ObsProject: demand.ObsProject,
			ExpectTime: expectTimeStr,
			ZoneID:     demand.ZoneID,
			ZoneName:   demand.ZoneName,
			RegionID:   demand.RegionID,
			RegionName: demand.RegionName,
			AreaID:     demand.AreaID,
			AreaName:   demand.AreaName,
			Cvm: rpt.Cvm{
				ResMode:      demand.ResMode.Name(),
				DeviceType:   demand.DeviceType,
				DeviceClass:  demand.DeviceClass,
				DeviceFamily: demand.DeviceFamily,
				CoreType:     demand.CoreType,
				Os:           types.Decimal{Decimal: originOS},
				CpuCore:      originDemandMap[demand.ID].CpuCore,
				Memory:       originDemandMap[demand.ID].Memory,
			},
			Cbs: rpt.Cbs{
				DiskType:     demand.DiskType,
				DiskTypeName: demand.DiskTypeName,
				DiskIo:       demand.DiskIO,
				DiskSize:     originDemandMap[demand.ID].DiskSize,
			},
		}
	}

	return demandOriginMap, nil
}

// constructDelayDemands construct delay demand.
func (c *Controller) constructDelayDemands(kt *kit.Kit, delays []ptypes.AdjustRPDemandReqElem,
	demandClass enumor.DemandClass) ([]rpt.ResPlanDemand, error) {

	if len(delays) == 0 {
		return nil, nil
	}

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	originDemandMap := slice.FuncToMap(delays,
		func(update ptypes.AdjustRPDemandReqElem) (string, ptypes.CreateResPlanDemandResource) {
			return update.DemandID, update.OriginalInfo.GetResource()
		})
	demandOriginMap, err := c.constructOriginalDemandMap(kt, originDemandMap)
	if err != nil {
		logs.Errorf("failed to construct original demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]rpt.ResPlanDemand, len(delays))
	for idx, delay := range delays {
		result[idx] = rpt.ResPlanDemand{
			DemandClass: demandClass,
			Original:    demandOriginMap[delay.DemandID],
		}

		// delay updated equals to original, except expect time.
		result[idx].Updated = &rpt.UpdatedRPDemandItem{
			ObsProject: result[idx].Original.ObsProject,
			ExpectTime: delay.ExpectTime,
			ZoneID:     result[idx].Original.ZoneID,
			ZoneName:   result[idx].Original.ZoneName,
			RegionID:   result[idx].Original.RegionID,
			RegionName: result[idx].Original.RegionName,
			AreaID:     result[idx].Original.AreaID,
			AreaName:   result[idx].Original.AreaName,
			Cvm: rpt.Cvm{
				ResMode:      result[idx].Original.Cvm.ResMode,
				DeviceType:   result[idx].Original.Cvm.DeviceType,
				DeviceClass:  result[idx].Original.Cvm.DeviceClass,
				DeviceFamily: result[idx].Original.Cvm.DeviceFamily,
				CoreType:     result[idx].Original.Cvm.CoreType,
				Os:           result[idx].Original.Cvm.Os,
				CpuCore:      result[idx].Original.Cvm.CpuCore,
				Memory:       result[idx].Original.Cvm.Memory,
			},
			Cbs: rpt.Cbs{
				DiskType:     result[idx].Original.Cbs.DiskType,
				DiskTypeName: result[idx].Original.Cbs.DiskTypeName,
				DiskIo:       result[idx].Original.Cbs.DiskIo,
				DiskSize:     result[idx].Original.Cbs.DiskSize,
			},
		}
	}

	return result, nil
}
