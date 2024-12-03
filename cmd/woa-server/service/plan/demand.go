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
	"sync"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
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
	resp, err := s.planController.ListResPlanDemandAndOverview(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list res plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return resp, nil
}

// listResPlanCrpDemands list res plan crp demands, demandIDs length is unknown, page query
func (s *service) listResPlanCrpDemands(kt *kit.Kit, demandIDs []int64) (map[int64]*rpcd.ResPlanCrpDemandTable, error) {
	var mapLock sync.Mutex
	result := make(map[int64]*rpcd.ResPlanCrpDemandTable)

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
				result[one.CrpDemandID] = &list.Details[id]
			}
			mapLock.Unlock()

			return nil
		})
	if err != nil {
		return nil, err
	}

	return result, nil
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
	demandID := cts.PathParameter("id").String()

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
	demandID := cts.PathParameter("id").String()

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
func (s *service) getPlanDemandDetail(cts *rest.Contexts, demandID string, bkBizIDs []int64) (interface{}, error) {
	resp, err := s.planController.GetResPlanDemandDetail(cts.Kit, demandID, bkBizIDs)
	if err != nil {
		logs.Errorf("failed to get plan demand detail for demand id: %s, err: %v, rid: %s", demandID, err,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return resp, nil
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
	// 业务无权限
	bkBizID := resPlanCrpDemands[demandID].BkBizID
	if len(bkBizIDs) > 0 {
		if !slice.IsItemInSlice(bkBizIDs, bkBizID) {
			return nil, errf.PermissionDenied, fmt.Errorf("bk_biz_id: %d is not authorized", bkBizID)
		}
	}
	if err = src.SetRegionAreaAndZoneID(zoneNameMap, regionNameMap); err != nil {
		return nil, errf.Aborted, err
	}
	if err = src.SetDiskType(); err != nil {
		return nil, errf.Aborted, err
	}

	src.BkBizID = resPlanCrpDemands[demandID].BkBizID
	src.BkBizName = resPlanCrpDemands[demandID].BkBizName
	src.OpProductID = resPlanCrpDemands[demandID].OpProductID
	src.OpProductName = resPlanCrpDemands[demandID].OpProductName

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

	demandReq := &ptypes.ListResPlanDemandReq{
		BkBizIDs:  bkBizIDs,
		DemandIDs: []string{req.DemandID},
		Page:      core.NewCountPage(),
	}
	rstCount, err := s.planController.ListResPlanDemandAndOverview(cts.Kit, demandReq)
	if err != nil {
		logs.Errorf("failed to list res plan demand to authorize, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if rstCount.Count == 0 {
		return nil, errf.NewFromErr(errf.PermissionDenied, fmt.Errorf("bk_biz is not authorized"))
	}

	resp, err := s.planController.ListCrpDemandChangeLog(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list demand change log by demand id: %s, err: %v, rid: %s,", req.DemandID, err,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return resp, nil
}

// RepairResPlanDemand 按业务修复一段时间的历史预测数据.
// 接口非幂等，使用时需清理业务该段时间的历史预测，否则数据可能和预期不符。
func (s *service) RepairResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.RepairRPDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to repair res plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate repair res plan demand parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	err := s.planController.RepairResPlanDemandFromTicket(cts.Kit, req.BkBizIDs, req.RepairTicketRange)
	if err != nil {
		logs.Errorf("failed to repair res plan demand from ticket, err: %v, req: %+v, rid: %s", err, *req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}
