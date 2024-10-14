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

	"hcm/cmd/woa-server/logics/plan"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// AdjustBizResPlanDemand adjust biz resource plan demand.
func (s *service) AdjustBizResPlanDemand(cts *rest.Contexts) (rst interface{}, err error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.AdjustRPDemandReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode adjust biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate adjust biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Update}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	crpDemandIDs := slice.Map(req.Adjusts, func(adjust ptypes.AdjustRPDemandReqElem) int64 { return adjust.CrpDemandID })

	// check whether all crp demand belong to the biz.
	allBelong, err := s.areAllCrpDemandBelongToBiz(cts.Kit, crpDemandIDs, bkBizID)
	if err != nil {
		logs.Errorf("failed to check whether all crp demand belong to biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if !allBelong {
		logs.Errorf("not all adjust crp demand belong to biz: %d, rid: %s", bkBizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, fmt.Errorf("not all adjust crp demand belong to biz: %d", bkBizID))
	}

	// examine whether all resource plan demand classes are the same, and get the demand class.
	demandClass, err := s.examineDemandClass(cts.Kit, crpDemandIDs)
	if err != nil {
		logs.Errorf("failed to examine demand class, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// examine and lock all resource plan demand.
	if err = s.dao.ResPlanCrpDemand().ExamineAndLockAllRPDemand(cts.Kit, crpDemandIDs); err != nil {
		logs.Errorf("failed to examine and lock all resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// defer is used to unlock all resource plan demand when some errors occur.
	defer func() {
		if err != nil {
			if tmpErr := s.dao.ResPlanCrpDemand().UnlockAllResPlanDemand(cts.Kit, crpDemandIDs); tmpErr != nil {
				logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", tmpErr, cts.Kit.Rid)
			}
		}
	}()

	// construct adjust biz resource plan demand request.
	adjustReq, err := s.constructAdjustReq(cts.Kit, bkBizID, demandClass, req)
	if err != nil {
		logs.Errorf("failed to construct adjust resource plan ticket request, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// create cancel resource plan ticket.
	ticketID, err := s.planController.CreateResPlanTicket(cts.Kit, adjustReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// create adjust resource plan ticket itsm audit flow.
	if err = s.planController.CreateAuditFlow(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil
}

// areAllCrpDemandBelongToBiz return whether all input crp demand ids belong to input biz.
func (s *service) areAllCrpDemandBelongToBiz(kt *kit.Kit, crpDemandIDs []int64, bkBizID int64) (bool, error) {
	listOpt := &types.ListOption{
		Fields: []string{"bk_biz_id"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	notAllBelong := slices.ContainsFunc(rst.Details, func(ele rpcd.ResPlanCrpDemandTable) bool {
		return ele.BkBizID != bkBizID
	})

	return !notAllBelong, nil
}

// examineDemandClass examine whether all demands are the same demand class, and return the demand class.
func (s *service) examineDemandClass(kt *kit.Kit, crpDemandIDs []int64) (enumor.DemandClass, error) {
	if len(crpDemandIDs) == 0 {
		return "", errors.New("crp demand ids is empty")
	}

	listOpt := &types.ListOption{
		Fields: []string{"demand_class"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := s.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(rst.Details) == 0 {
		logs.Errorf("list resource plan demand, but len detail is 0, rid: %s", kt.Rid)
		return "", errors.New("list resource plan demand, but len detail is 0")
	}

	demandClass := rst.Details[0].DemandClass
	for _, detail := range rst.Details {
		if detail.DemandClass != demandClass {
			logs.Errorf("not all demand classes are the same, rid: %s", kt.Rid)
			return "", errors.New("not all demand classes are the same")
		}
	}

	return demandClass, nil
}

// constructAdjustReq construct create resource plan ticket request of adjust.
func (s *service) constructAdjustReq(kt *kit.Kit, bkBizID int64, demandClass enumor.DemandClass,
	req *ptypes.AdjustRPDemandReq) (*plan.CreateResPlanTicketReq, error) {

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
	updates, err := s.constructUpdateDemands(kt, updateDemands, demandClass)
	if err != nil {
		logs.Errorf("failed to construct update demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct delay demands.
	delays, err := s.constructDelayDemands(kt, delayDemands, demandClass)
	if err != nil {
		logs.Errorf("failed to construct delay demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get biz org relation.
	bizOrgRel, err := s.logics.GetBizOrgRel(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	demands := append(updates, delays...)
	adjustReq := &plan.CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeAdjust,
		DemandClass: demandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     demands,
	}

	return adjustReq, nil
}

// constructUpdateDemands construct update demand.
func (s *service) constructUpdateDemands(kt *kit.Kit, updates []ptypes.AdjustRPDemandReqElem,
	demandClass enumor.DemandClass) ([]rpt.ResPlanDemand, error) {

	if len(updates) == 0 {
		return nil, nil
	}

	// get create resource plan ticket needed zoneMap, regionAreaMap and deviceTypeMap.
	zoneMap, regionAreaMap, deviceTypeMap, err := s.getMetaMaps(kt)
	if err != nil {
		logs.Errorf("failed to get meta maps, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	crpDemandIDs := slice.Map(updates, func(update ptypes.AdjustRPDemandReqElem) int64 {
		return update.CrpDemandID
	})

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	demandOriginMap, demandRemainMap, err := s.constructOriginalDemandMap(kt, crpDemandIDs)
	if err != nil {
		logs.Errorf("failed to construct original demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]rpt.ResPlanDemand, len(updates))
	for idx, update := range updates {
		result[idx] = rpt.ResPlanDemand{
			DemandClass: demandClass,
			Original:    demandOriginMap[update.CrpDemandID],
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

		if slices.Contains(update.UpdatedInfo.DemandResTypes, enumor.DemandResTypeCVM) {
			result[idx].Updated.Cvm.ResMode = update.UpdatedInfo.Cvm.ResMode
			result[idx].Updated.Cvm.DeviceType = update.UpdatedInfo.Cvm.DeviceType
			result[idx].Updated.Cvm.DeviceClass = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].DeviceClass
			result[idx].Updated.Cvm.DeviceFamily = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].DeviceFamily
			result[idx].Updated.Cvm.CoreType = deviceTypeMap[update.UpdatedInfo.Cvm.DeviceType].CoreType
			result[idx].Updated.Cvm.Os = *update.UpdatedInfo.Cvm.Os
			result[idx].Updated.Cvm.CpuCore = *update.UpdatedInfo.Cvm.CpuCore
			result[idx].Updated.Cvm.Memory = *update.UpdatedInfo.Cvm.Memory
		}

		if slices.Contains(update.UpdatedInfo.DemandResTypes, enumor.DemandResTypeCBS) {
			result[idx].Updated.Cbs.DiskType = update.UpdatedInfo.Cbs.DiskType
			result[idx].Updated.Cbs.DiskTypeName = update.UpdatedInfo.Cbs.DiskType.Name()
			result[idx].Updated.Cbs.DiskIo = *update.UpdatedInfo.Cbs.DiskIo
			result[idx].Updated.Cbs.DiskSize = *update.UpdatedInfo.Cbs.DiskSize
		}

		if result[idx].Updated.Cvm.CpuCore < result[idx].Original.Cvm.CpuCore-demandRemainMap[update.CrpDemandID] {
			logs.Errorf("update cpu core can not be less than applied, rid: %s", kt.Rid)
			return nil, errors.New("update cpu core can not be less than applied")
		}
	}

	return result, nil
}

// constructOriginalDemandMap construct original demand map.
// return crp demand id and demand class map, crp demand id and remain cpu core map.
func (s *service) constructOriginalDemandMap(kt *kit.Kit, crpDemandIDs []int64) (map[int64]*rpt.OriginalRPDemandItem,
	map[int64]float64, error) {

	if len(crpDemandIDs) == 0 {
		return make(map[int64]*rpt.OriginalRPDemandItem), make(map[int64]float64), nil
	}

	// call crp interface to get demand details.
	crpDemands, err := s.planController.QueryIEGDemands(kt, &plan.QueryIEGDemandsReq{CrpDemandIDs: crpDemandIDs})
	if err != nil {
		logs.Errorf("failed to query all demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// get meta name maps.
	zoneNameMap, regionAreaNameMap, err := s.getMetaNameMaps(kt)
	if err != nil {
		logs.Errorf("failed to get meta name maps, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// get device type mapping.
	deviceTypeMap, err := s.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	demandOriginMap := make(map[int64]*rpt.OriginalRPDemandItem)
	demandRemainMap := make(map[int64]float64)
	for _, crpDemand := range crpDemands {
		demandID, err := strconv.ParseInt(crpDemand.DemandId, 10, 64)
		if err != nil {
			logs.Errorf("failed to parse crp demand id, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		diskType, err := enumor.GetDiskTypeFromCrpName(crpDemand.DiskTypeName)
		if err != nil {
			logs.Errorf("failed to get disk type from crp name, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		deviceType := crpDemand.InstanceModel
		demandOriginMap[demandID] = &rpt.OriginalRPDemandItem{
			CrpDemandID: demandID,
			ObsProject:  enumor.ObsProject(crpDemand.ProjectName),
			ExpectTime:  crpDemand.UseTime,
			ZoneID:      zoneNameMap[crpDemand.ZoneName],
			ZoneName:    crpDemand.ZoneName,
			RegionID:    regionAreaNameMap[crpDemand.CityName].RegionID,
			RegionName:  crpDemand.CityName,
			AreaID:      regionAreaNameMap[crpDemand.CityName].AreaID,
			AreaName:    regionAreaNameMap[crpDemand.CityName].AreaName,
			Cvm: rpt.Cvm{
				ResMode:      crpDemand.ResourceMode,
				DeviceType:   deviceType,
				DeviceClass:  deviceTypeMap[deviceType].DeviceClass,
				DeviceFamily: deviceTypeMap[deviceType].DeviceFamily,
				CoreType:     deviceTypeMap[deviceType].CoreType,
				Os:           float64(crpDemand.PlanCvmAmount),
				CpuCore:      float64(crpDemand.PlanCoreAmount),
				Memory:       float64(crpDemand.PlanRamAmount),
			},
			Cbs: rpt.Cbs{
				DiskType:     diskType,
				DiskTypeName: diskType.Name(),
				DiskIo:       int64(crpDemand.InstanceIO),
				DiskSize:     float64(crpDemand.PlanDiskAmount),
			},
		}

		// crpDemand.CoreAmount is resource plan remained cpu core.
		demandRemainMap[demandID] = float64(crpDemand.CoreAmount)
	}

	return demandOriginMap, demandRemainMap, nil
}

// constructDelayDemands construct delay demand.
func (s *service) constructDelayDemands(kt *kit.Kit, delays []ptypes.AdjustRPDemandReqElem,
	demandClass enumor.DemandClass) ([]rpt.ResPlanDemand, error) {

	if len(delays) == 0 {
		return nil, nil
	}

	crpDemandIDs := slice.Map(delays, func(delay ptypes.AdjustRPDemandReqElem) int64 {
		return delay.CrpDemandID
	})

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	demandOriginMap, _, err := s.constructOriginalDemandMap(kt, crpDemandIDs)
	if err != nil {
		logs.Errorf("failed to construct original demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]rpt.ResPlanDemand, len(delays))
	for idx, delay := range delays {
		result[idx] = rpt.ResPlanDemand{
			DemandClass: demandClass,
			Original:    demandOriginMap[delay.CrpDemandID],
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

// CancelBizResPlanDemand cancel biz resource plan demand.
func (s *service) CancelBizResPlanDemand(cts *rest.Contexts) (rst interface{}, err error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.CancelRPDemandReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode cancel biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate cancel biz resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize biz resource plan operation.
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ResPlan, Action: meta.Delete}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// check whether all crp demand belong to the biz.
	allBelong, err := s.areAllCrpDemandBelongToBiz(cts.Kit, req.CrpDemandIDs, bkBizID)
	if err != nil {
		logs.Errorf("failed to check whether all crp demand belong to biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if !allBelong {
		logs.Errorf("not all adjust crp demand belong to biz: %d, rid: %s", bkBizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, fmt.Errorf("not all adjust crp demand belong to biz: %d", bkBizID))
	}

	// examine whether all resource plan demand classes are the same, and get the demand class.
	demandClass, err := s.examineDemandClass(cts.Kit, req.CrpDemandIDs)
	if err != nil {
		logs.Errorf("failed to examine demand class, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// examine and lock all resource plan demand.
	if err = s.dao.ResPlanCrpDemand().ExamineAndLockAllRPDemand(cts.Kit, req.CrpDemandIDs); err != nil {
		logs.Errorf("failed to examine and lock all resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// defer is used to unlock all resource plan demand when some errors occur.
	defer func() {
		if err != nil {
			if tmpErr := s.dao.ResPlanCrpDemand().UnlockAllResPlanDemand(cts.Kit, req.CrpDemandIDs); tmpErr != nil {
				logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", tmpErr, cts.Kit.Rid)
			}
		}
	}()

	// construct cancel biz resource plan demand request.
	cancelReq, err := s.constructCancelReq(cts.Kit, bkBizID, demandClass, req.CrpDemandIDs)
	if err != nil {
		logs.Errorf("failed to construct cancel resource plan ticket request, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// create cancel resource plan ticket.
	ticketID, err := s.planController.CreateResPlanTicket(cts.Kit, cancelReq)
	if err != nil {
		logs.Errorf("failed to create resource plan ticket, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", errf.NewFromErr(errf.Aborted, err)
	}

	// create cancel resource plan ticket itsm audit flow.
	if err = s.planController.CreateAuditFlow(cts.Kit, ticketID); err != nil {
		logs.Errorf("failed to create resource plan ticket audit flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": ticketID}, nil

}

// constructCancelReq construct create resource plan ticket request of cancel.
func (s *service) constructCancelReq(kt *kit.Kit, bkBizID int64, demandClass enumor.DemandClass, crpDemandIDs []int64) (
	*plan.CreateResPlanTicketReq, error) {

	// construct crp demand id and origin demand map, crp demand id and remain cpu core map.
	demandOriginMap, _, err := s.constructOriginalDemandMap(kt, crpDemandIDs)
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

	// get biz org relation.
	bizOrgRel, err := s.logics.GetBizOrgRel(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	req := &plan.CreateResPlanTicketReq{
		TicketType:  enumor.RPTicketTypeDelete,
		DemandClass: demandClass,
		BizOrgRel:   *bizOrgRel,
		Demands:     demands,
	}

	return req, nil
}
