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
	"strconv"
	"sync"

	demandtime "hcm/cmd/woa-server/service/plan/demand-time"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	mtypes "hcm/pkg/dal/dao/types/meta"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

// ListResPlanDemand list resource plan demand.
func (s *service) ListResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListResPlanDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource demand parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanDemand(cts, req)
}

// ListBizResPlanDemand list biz res plan demand.
func (s *service) ListBizResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListResPlanDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list resource demand parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	bkBizIDs, err := s.logics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// if bizID not in authorized bkBizIDs, return empty response.
	if !slice.IsItemInSlice(bkBizIDs, bizID) {
		return core.ListResultT[any]{Details: make([]any, 0)}, nil
	}
	req.BkBizIDs = []int64{bizID}

	return s.listResPlanDemand(cts, req)
}

// ListResPlanDemand general logic for list demand
func (s *service) listResPlanDemand(cts *rest.Contexts, req *ptypes.ListResPlanDemandReq) (interface{}, error) {
	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, _, err := s.getMetaMaps(cts.Kit)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	// 拼装zone和region请求参数
	regionNames, err := convReqRegionIDsToNames(req.RegionIDs, regionAreaMap)
	if err != nil {
		logs.Errorf("failed to convert demand list request to crp request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound,
			fmt.Errorf("failed to convert demand list request to crp request: %s", err.Error()))
	}
	zoneNames, err := convReqZoneIDsToNames(req.ZoneIDs, zoneMap)
	if err != nil {
		logs.Errorf("failed to convert demand list request to crp request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound,
			fmt.Errorf("failed to convert demand list request to crp request: %s", err.Error()))
	}

	// 从CRP接口查询完整数据
	req.RegionNames = regionNames
	req.ZoneNames = zoneNames
	entireDetails, err := s.planController.ListCrpDemands(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// 按照请求参数中的其他约束过滤数据
	details, demandIDs, err := filterResPlanDemandResp(cts.Kit, req, entireDetails)
	if err != nil {
		logs.Errorf("failed to filter resource plan demand list result, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if len(details) == 0 {
		return core.ListResultT[any]{Details: make([]any, 0)}, nil
	}

	// 结合 res_plan_crp_demand 表追加剩余字段，并按照 bk_biz_id 二次过滤数据
	details, err = s.appendListResPlanDemandRespFieldWithTable(cts.Kit, details, demandIDs)
	if err != nil {
		logs.Errorf("failed to append field to resource plan demand list result, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	details = filterListResPlanDemandRespWithReqParams(req, details)
	details = filterListResPlanDemandRespWithRegion(cts.Kit, details, zoneNameMap, regionNameMap)

	rst, err := sortPageListResPlanCrpDemandResp(details, req.Page)
	if err != nil {
		logs.Errorf("failed to sort and page resource plan demand list result, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return rst, nil
}

// listResPlanCrpDemands list res plan crp demands, demandIDs length is unknown, page query
func (s *service) listResPlanCrpDemands(kt *kit.Kit, demandIDs []int64) (map[int64][]*rpcd.ResPlanCrpDemandTable,
	error) {

	var mapLock sync.Mutex
	result := make(map[int64][]*rpcd.ResPlanCrpDemandTable)

	// 查询参数最多500条，超过的需要分组查询
	batch := int(filter.DefaultMaxInLimit)
	concurrentParams := slice.Split(demandIDs, batch)

	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, concurrentParams,
		func(subDemandIDs []int64) error {
			opt := &types.ListOption{
				Filter: tools.ContainersExpression("crp_demand_id", subDemandIDs),
				Page:   core.NewDefaultBasePage(),
			}

			list, err := s.dao.ResPlanCrpDemand().List(kt, opt)
			if err != nil {
				logs.Errorf("list res plan crp demands failed, err: %v, demand_ids: %v, rid: %s", err, demandIDs,
					kt.Rid)
				return err
			}

			mapLock.Lock()
			for id, one := range list.Details {
				result[one.CrpDemandID] = append(result[one.CrpDemandID], &list.Details[id])
			}
			mapLock.Unlock()

			return nil
		})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func filterListResPlanDemandRespWithReqParams(req *ptypes.ListResPlanDemandReq,
	details []*ptypes.ListResPlanDemandItem) []*ptypes.ListResPlanDemandItem {

	rstDetails := make([]*ptypes.ListResPlanDemandItem, 0, len(details))
	for _, item := range details {
		if len(req.BkBizIDs) > 0 {
			if !slice.IsItemInSlice(req.BkBizIDs, item.BkBizID) {
				continue
			}
		}
		if len(req.OpProductIDs) > 0 {
			if !slice.IsItemInSlice(req.OpProductIDs, item.OpProductID) {
				continue
			}
		}
		if len(req.PlanProductIDs) > 0 {
			if !slice.IsItemInSlice(req.PlanProductIDs, item.PlanProductID) {
				continue
			}
		}
		if len(req.DemandClasses) > 0 {
			if !slice.IsItemInSlice(req.DemandClasses, item.DemandClass) {
				continue
			}
		}

		rstDetails = append(rstDetails, item)
	}
	return rstDetails
}

func filterListResPlanDemandRespWithRegion(kt *kit.Kit, details []*ptypes.ListResPlanDemandItem,
	zoneNameMap map[string]string, regionNameMap map[string]mtypes.RegionArea) []*ptypes.ListResPlanDemandItem {

	rstDetails := make([]*ptypes.ListResPlanDemandItem, 0, len(details))
	for _, item := range details {
		if err := item.SetRegionAndZoneID(zoneNameMap, regionNameMap); err != nil {
			logs.Warnf("failed to set region and zone id, err: %v, demand_id: %d, rid: %s", err, item.CrpDemandID,
				kt.Rid)
			continue
		}
		rstDetails = append(rstDetails, item)
	}
	return rstDetails
}

// appendListResPlanDemandRespFieldWithTable 根据本地数据库 res_plan_crp_demand 中的字段过滤数据
func (s *service) appendListResPlanDemandRespFieldWithTable(kt *kit.Kit, details []*ptypes.ListResPlanDemandItem,
	demandIDs []int64) ([]*ptypes.ListResPlanDemandItem, error) {

	resPlanCrpDemands, err := s.listResPlanCrpDemands(kt, demandIDs)
	if err != nil {
		return nil, err
	}

	rstDetails := make([]*ptypes.ListResPlanDemandItem, 0)
	for _, item := range details {
		if _, exists := resPlanCrpDemands[item.CrpDemandID]; !exists {
			// 本地没有的数据直接干掉，可能有一定风险
			logs.Warnf("cannot found demand_id: %d from res_plan_crp_demand, rid: %s", item.CrpDemandID, kt.Rid)
			continue
		}

		if err = item.ParseExpectTime(); err != nil {
			logs.Warnf("failed to parse expect_time, err: %v, demand_id: %d, rid: %s", err, item.CrpDemandID, kt.Rid)
			continue
		}

		// 因CRP合并，一个demand可能被多个业务共享
		for _, localDemand := range resPlanCrpDemands[item.CrpDemandID] {
			bizItem := item.Copy()

			if *localDemand.Locked == enumor.CrpDemandLocked {
				bizItem.Status = enumor.DemandStatusLocked
			}

			bizItem.StatusName = bizItem.Status.Name()
			bizItem.BkBizID = localDemand.BkBizID
			bizItem.BkBizName = localDemand.BkBizName
			bizItem.OpProductID = localDemand.OpProductID
			bizItem.OpProductName = localDemand.OpProductName
			bizItem.PlanProductID = localDemand.PlanProductID
			bizItem.PlanProductName = localDemand.PlanProductName
			bizItem.DemandClass = localDemand.DemandClass

			rstDetails = append(rstDetails, bizItem)
		}
	}
	return rstDetails, nil
}

// sortPageListResPlanCrpDemandResp sort and page list resource plan demands resp
func sortPageListResPlanCrpDemandResp(details []*ptypes.ListResPlanDemandItem, page *core.BasePage) (
	*ptypes.ListResPlanDemandResp, error) {

	overview := &ptypes.ListResPlanDemandOverview{}
	for _, item := range details {
		// 计算overview
		overview.TotalCpuCore += item.TotalCpuCore
		overview.TotalAppliedCore += item.AppliedCpuCore
		overview.ExpiringCpuCore += item.ExpiringCpuCore
		if item.PlanType.InPlan() {
			overview.InPlanCpuCore += item.TotalCpuCore
			overview.InPlanAppliedCpuCore += item.AppliedCpuCore
		} else {
			overview.OutPlanCpuCore += item.TotalCpuCore
			overview.OutPlanAppliedCpuCore += item.AppliedCpuCore
		}
	}

	// count查询
	if page.Count {
		return &ptypes.ListResPlanDemandResp{
			Overview: overview,
			Count:    len(details),
		}, nil
	}

	resp := &ptypes.ListResPlanDemandResp{
		Overview: overview,
		Details:  details,
	}
	// 根据page分页、排序
	err := resp.SortAndPage(page)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// filterResPlanDemandResp filter response of resource plan demand list
func filterResPlanDemandResp(kt *kit.Kit, req *ptypes.ListResPlanDemandReq, details []*ptypes.PlanDemandDetail) (
	[]*ptypes.ListResPlanDemandItem, []int64, error) {

	rstDetails := make([]*ptypes.ListResPlanDemandItem, 0)
	crpDemandIDs := make([]int64, 0)
	for _, item := range details {
		// device_type 不符合需求
		if len(req.DeviceTypes) > 0 {
			if !slice.IsItemInSlice(req.DeviceTypes, item.DeviceType) {
				continue
			}
		}

		// 计算预测内、预测外
		if err := item.PlanType.Validate(); err != nil {
			logs.Warnf("invalid plan type: %s, demand id: %s, err: %v, rid: %s", item.PlanType, item.CrpDemandID,
				err, kt.Rid)
			continue
		}
		if len(req.PlanTypes) > 0 {
			if !slice.IsItemInSlice(req.PlanTypes, item.PlanType) {
				continue
			}
		}

		// 筛选本月即将到期的
		expectTimeFmt, err := util.TimeStrToTimePtr(item.ExpectTime)
		if err != nil {
			logs.Warnf("failed to convert expectTime to time ptr, err: %v, demand id: %s, rid: %s", err,
				item.CrpDemandID,
				kt.Rid)
			continue
		}
		isAbortToExpire := demandtime.IsAboutToExpire(expectTimeFmt)
		if isAbortToExpire {
			item.ExpiringCpuCore = item.RemainedCpuCore
		}
		if req.ExpiringOnly {
			if !isAbortToExpire {
				continue
			}
		}

		// 将结果转换为 list detail
		rstItem, err := getListResPlanDemandItem(item)
		if err != nil {
			logs.Warnf("failed to convert crp demand item, err: %v, demand id: %s, rid: %s", err, item.CrpDemandID,
				kt.Rid)
			continue
		}
		// 计算demand状态，can_apply（可申领）、not_ready（未到申领时间）、expired（已过期）
		expectStartFmt, err := util.TimeStrToTimePtr(item.ExpectStartDate)
		if err != nil {
			logs.Warnf("failed to convert expectStartDate to time ptr, err: %v, demand id: %s, rid: %s", err,
				item.CrpDemandID, kt.Rid)
			continue
		}
		expectEndFmt, err := util.TimeStrToTimePtr(item.ExpectEndDate)
		if err != nil {
			logs.Warnf("failed to convert expectEndDate to time ptr, err: %v, demand id: %s, rid: %s", err,
				item.CrpDemandID, kt.Rid)
			continue
		}
		rstItem.SetStatus(demandtime.GetDemandStatus(expectStartFmt, expectEndFmt))

		crpDemandIDs = append(crpDemandIDs, rstItem.CrpDemandID)
		rstDetails = append(rstDetails, rstItem)
	}

	return rstDetails, crpDemandIDs, nil
}

func getListResPlanDemandItem(demandDetail *ptypes.PlanDemandDetail) (
	*ptypes.ListResPlanDemandItem, error) {

	demandIdInt, err := strconv.ParseInt(demandDetail.CrpDemandID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert crp demand id to int, err: %v", err)
	}

	diskType, err := enumor.GetDiskTypeFromCrpName(demandDetail.DiskTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk type from crp name, err: %v", err)
	}

	return &ptypes.ListResPlanDemandItem{
		CrpDemandID:        demandIdInt,
		AvailableYearMonth: fmt.Sprintf("%04d-%02d", demandDetail.Year, demandDetail.Month),
		ExpectTime:         demandDetail.ExpectTime,
		DeviceClass:        demandDetail.DeviceClass,
		DeviceType:         demandDetail.DeviceType,
		TotalOS:            demandDetail.TotalOS,
		AppliedOS:          demandDetail.AppliedOS,
		RemainedOS:         demandDetail.RemainedOS,
		TotalCpuCore:       demandDetail.TotalCpuCore,
		AppliedCpuCore:     demandDetail.AppliedCpuCore,
		RemainedCpuCore:    demandDetail.RemainedCpuCore,
		ExpiringCpuCore:    demandDetail.ExpiringCpuCore,
		TotalMemory:        demandDetail.TotalMemory,
		AppliedMemory:      demandDetail.AppliedMemory,
		RemainedMemory:     demandDetail.RemainedMemory,
		TotalDiskSize:      demandDetail.TotalDiskSize,
		AppliedDiskSize:    demandDetail.AppliedDiskSize,
		RemainedDiskSize:   demandDetail.RemainedDiskSize,
		RegionName:         demandDetail.RegionName,
		ZoneName:           demandDetail.ZoneName,
		PlanType:           demandDetail.PlanType,
		ObsProject:         demandDetail.ObsProject,
		GenerationType:     demandDetail.GenerationType,
		DeviceFamily:       demandDetail.DeviceFamily,
		DiskType:           diskType,
		DiskTypeName:       diskType.Name(),
		DiskIO:             demandDetail.DiskIO,
	}, nil
}

func convReqRegionIDsToNames(regionIDs []string, regionMap map[string]mtypes.RegionArea) ([]string, error) {
	reqRegionNames := make([]string, 0)
	for _, reqRegionID := range regionIDs {
		regionArea, exists := regionMap[reqRegionID]
		// 查询参数中的regionId如果数据库中查不到，查询直接失败
		if !exists {
			return nil, fmt.Errorf("region id: %s not found in woa_zone", reqRegionID)
		}
		reqRegionNames = append(reqRegionNames, regionArea.RegionName)
	}
	return reqRegionNames, nil
}

func convReqZoneIDsToNames(zoneIDs []string, zoneMap map[string]string) ([]string, error) {
	reqZoneNames := make([]string, 0)
	for _, reqZoneID := range zoneIDs {
		zoneName, exists := zoneMap[reqZoneID]
		// 查询参数中的zoneId如果数据库中查不到，查询直接失败
		if !exists {
			return nil, fmt.Errorf("zone id: %s not found in woa_zone", reqZoneID)
		}
		reqZoneNames = append(reqZoneNames, zoneName)
	}
	return reqZoneNames, nil
}

// GetPlanDemandDetail get plan demand detail.
func (s *service) GetPlanDemandDetail(cts *rest.Contexts) (interface{}, error) {
	demandID, err := cts.PathParameter("id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.getPlanDemandDetail(cts, demandID, []int64{})
}

// GetBizPlanDemandDetail get biz plan demand detail.
func (s *service) GetBizPlanDemandDetail(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	demandID, err := cts.PathParameter("id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	bkBizIDs, err := s.logics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// if bizID not in authorized bkBizIDs, return empty response.
	if !slice.IsItemInSlice(bkBizIDs, bizID) {
		return core.ListResultT[any]{Details: make([]any, 0)}, nil
	}
	bkBizIDs = []int64{bizID}

	return s.getPlanDemandDetail(cts, demandID, bkBizIDs)
}

// getPlanDemandDetail general logic for get demand details
func (s *service) getPlanDemandDetail(cts *rest.Contexts, demandID int64, bkBizIDs []int64) (interface{}, error) {
	// 从CRP接口查询完整数据
	req := &ptypes.ListResPlanDemandReq{
		CrpDemandIDs: []int64{demandID},
	}

	// 从 woa_zone 获取城市/地区的中英文对照
	zoneMap, regionAreaMap, _, err := s.getMetaMaps(cts.Kit)
	if err != nil {
		logs.Errorf("get meta maps failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	zoneNameMap, regionNameMap := getMetaNameMapsFromIDMap(zoneMap, regionAreaMap)

	// 拼装zone和region请求参数
	regionNames, err := convReqRegionIDsToNames(req.RegionIDs, regionAreaMap)
	if err != nil {
		logs.Errorf("failed to convert demand list request to crp request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound,
			fmt.Errorf("failed to convert demand list request to crp request: %s", err.Error()))
	}
	zoneNames, err := convReqZoneIDsToNames(req.ZoneIDs, zoneMap)
	if err != nil {
		logs.Errorf("failed to convert demand list request to crp request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound,
			fmt.Errorf("failed to convert demand list request to crp request: %s", err.Error()))
	}
	// get demand detail 和 list demands 调用的同一个接口
	req.RegionNames = regionNames
	req.ZoneNames = zoneNames
	listDetails, err := s.planController.ListCrpDemands(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to get plan demand detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	rst, err := convPlanDemandDetailResp(listDetails)
	if err != nil {
		logs.Errorf("failed to convert plan demand detail resp, err: %v, demand id: %d, rid: %s", err, demandID,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	rst, errCode, err := s.filterPlanDemandDetailRespByBkBizIDs(cts.Kit, bkBizIDs, rst, demandID, zoneNameMap,
		regionNameMap)
	if err != nil {
		logs.Errorf("failed to filter plan demand detail resp, err: %v, demand id: %d, rid: %s", err, demandID,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errCode, err)
	}

	return rst, nil
}

func (s *service) filterPlanDemandDetailRespByBkBizIDs(kt *kit.Kit, bkBizIDs []int64,
	src *ptypes.GetPlanDemandDetailResp, demandID int64, zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) (*ptypes.GetPlanDemandDetailResp, int32, error) {

	resPlanCrpDemands, err := s.listResPlanCrpDemands(kt, []int64{demandID})
	if err != nil {
		return nil, errf.Aborted, err
	}

	if _, exists := resPlanCrpDemands[demandID]; !exists {
		// 本地没有的数据直接干掉
		return nil, errf.Aborted, fmt.Errorf("cannot found demand_id: %d from res_plan_crp_demand", demandID)
	}

	// 因CRP合并，一个demand可能被多个业务共享，detail挑选其中一个有权限的业务返回
	authorized := false
	for _, localDemand := range resPlanCrpDemands[demandID] {
		// 业务无权限
		bkBizID := localDemand.BkBizID
		if len(bkBizIDs) > 0 {
			if !slice.IsItemInSlice(bkBizIDs, bkBizID) {
				continue
			}
		}
		authorized = true

		src.BkBizID = localDemand.BkBizID
		src.BkBizName = localDemand.BkBizName
		src.OpProductID = localDemand.OpProductID
		src.OpProductName = localDemand.OpProductName

		// 匹配到一个就break
		break
	}

	if !authorized {
		return nil, errf.PermissionDenied, fmt.Errorf("bk_biz_id: %v is not authorized", bkBizIDs)
	}

	if err = src.SetRegionAreaAndZoneID(zoneNameMap, regionNameMap); err != nil {
		return nil, errf.Aborted, err
	}
	if err = src.SetDiskType(); err != nil {
		return nil, errf.Aborted, err
	}

	return src, errf.OK, nil
}

// convPlanDemandDetailResp convert plan demand detail to crp resp.
func convPlanDemandDetailResp(listDetails []*ptypes.PlanDemandDetail) (*ptypes.GetPlanDemandDetailResp, error) {
	if len(listDetails) == 0 {
		return nil, errors.New("list demand detail return an empty result")
	}

	detail := listDetails[0]
	resp := &detail.GetPlanDemandDetailResp
	if err := resp.SetDiskType(); err != nil {
		return nil, err
	}

	return resp, nil
}

// ListBizPlanDemandChangeLog list biz plan demand change log.
func (s *service) ListBizPlanDemandChangeLog(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListDemandChangeLogReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list demand change log, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list demand change log parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	bkBizIDs, err := s.logics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list authorized biz, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// if bizID not in authorized bkBizIDs, return empty response.
	if !slice.IsItemInSlice(bkBizIDs, bizID) {
		return core.ListResultT[any]{Details: make([]any, 0)}, nil
	}

	return s.listPlanDemandChangeLog(cts, req, []int64{bizID})
}

// ListPlanDemandChangeLog list demand change log.
func (s *service) ListPlanDemandChangeLog(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListDemandChangeLogReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list demand change log, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list demand change log parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listPlanDemandChangeLog(cts, req, []int64{})
}

// listPlanDemandChangeLog list demand change log primary logic.
func (s *service) listPlanDemandChangeLog(cts *rest.Contexts, req *ptypes.ListDemandChangeLogReq, bkBizIDs []int64) (
	interface{}, error) {

	rst, err := s.planController.ListCrpDemandChangeLog(cts.Kit, req.CrpDemandId)
	if err != nil {
		logs.Errorf("failed to list demand change log by demand id: %d, err: %v, rid: %s,", req.CrpDemandId, err,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	resp, err := s.convCrpDemandChangeLogResp(cts.Kit, rst, bkBizIDs)
	if err != nil {
		logs.Errorf("failed to convert crp demand change log to hcm resp, demand id: %d, err: %v, rid: %s",
			req.CrpDemandId, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if req.Page.Count {
		return &ptypes.ListDemandChangeLogResp{Count: len(resp.Details)}, nil
	}
	resp.Page(req.Page)

	return resp, nil
}

func (s *service) convCrpDemandChangeLogResp(kt *kit.Kit, clogItems []*ptypes.ListDemandChangeLogItem,
	bkBizIDs []int64) (*ptypes.ListDemandChangeLogResp, error) {

	crpDemandIDs := make([]int64, 0, len(clogItems))
	for _, item := range clogItems {
		crpDemandIDs = append(crpDemandIDs, item.CrpDemandId)
	}
	rst := make([]*ptypes.ListDemandChangeLogItem, 0, len(clogItems))

	// 根据结果中的demandId，结合 res_plan_crp_demand 表追加运营产品字段
	resPlanCrpDemands, err := s.listResPlanCrpDemands(kt, crpDemandIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range clogItems {
		if _, exists := resPlanCrpDemands[item.CrpDemandId]; !exists {
			logs.Warnf("cannot found demand_id: %d from res_plan_crp_demand, rid: %s", item.CrpDemandId, kt.Rid)
			continue
		}

		// 因CRP合并，一个demand可能被多个业务共享
		for _, localDemand := range resPlanCrpDemands[item.CrpDemandId] {
			itemBkBizID := localDemand.BkBizID
			if len(bkBizIDs) > 0 {
				if !slice.IsItemInSlice(bkBizIDs, itemBkBizID) {
					continue
				}
			}

			item.OpProductName = localDemand.OpProductName

			// 匹配到一个就break
			break
		}

		rst = append(rst, item)
	}

	return &ptypes.ListDemandChangeLogResp{
		Details: rst,
	}, nil
}
