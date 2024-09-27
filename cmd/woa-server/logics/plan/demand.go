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

	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// QueryAllDemands query all demands.
func (c *Controller) QueryAllDemands(kt *kit.Kit, req *QueryAllDemandsReq) ([]*cvmapi.CvmCbsPlanQueryItem, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	// init request parameter.
	queryReq := &cvmapi.CvmCbsPlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanQueryMethod,
		},
		Params: &cvmapi.CvmCbsPlanQueryParam{
			Page: &cvmapi.Page{
				Start: 0,
				Size:  int(core.DefaultMaxPageLimit),
			},
		},
	}

	// append filter parameters.
	if req.ExpectTimeRange != nil {
		queryReq.Params.UseTime = &cvmapi.UseTime{
			Start: req.ExpectTimeRange.Start,
			End:   req.ExpectTimeRange.End,
		}
	}

	if len(req.CrpDemandIDs) > 0 {
		queryReq.Params.DemandIdList = req.CrpDemandIDs
	}

	if len(req.CrpSns) > 0 {
		queryReq.Params.OrderIdList = req.CrpSns
	}

	if len(req.DeviceClasses) > 0 {
		queryReq.Params.InstanceType = req.DeviceClasses
	}

	if len(req.PlanProdNames) > 0 {
		queryReq.Params.PlanProductName = req.PlanProdNames
	}

	if len(req.ObsProjects) > 0 {
		queryReq.Params.ProjectName = req.ObsProjects
	}

	if len(req.RegionNames) > 0 {
		queryReq.Params.CityName = req.RegionNames
	}

	if len(req.ZoneNames) > 0 {
		queryReq.Params.ZoneName = req.ZoneNames
	}

	// query all demands.
	result := make([]*cvmapi.CvmCbsPlanQueryItem, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		queryReq.Params.Page.Start = start
		rst, err := c.crpCli.QueryCvmCbsPlans(kt.Ctx, nil, queryReq)
		if err != nil {
			return nil, err
		}

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}

		result = append(result, rst.Result.Data...)
	}

	return result, nil
}

// ExamineDemandClass examine whether all demands are the same demand class, and return the demand class.
func (c *Controller) ExamineDemandClass(kt *kit.Kit, crpDemandIDs []int64) (enumor.DemandClass, error) {
	listOpt := &types.ListOption{
		Fields: []string{"demand_class"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := c.dao.ResPlanCrpDemand().List(kt, listOpt)
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

// ExamineAndLockAllRPDemand examine all resource plan demand lock status and lock all resource plan demand.
// TODO：目前此函数不在一个事务中，会有并发问题
func (c *Controller) ExamineAndLockAllRPDemand(kt *kit.Kit, crpDemandIDs []int64) error {
	if len(crpDemandIDs) == 0 {
		return errors.New("crp demand ids is empty")
	}

	// examine whether all resource plan demand is unlocked.
	pass, err := c.examineAllRPDemandLockStatus(kt, crpDemandIDs)
	if err != nil {
		logs.Errorf("failed to examine all resource plan demand lock status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if !pass {
		logs.Errorf("some demands are locked, rid: %s", kt.Rid)
		return errors.New("some demands are locked")
	}

	// lock all resource plan demand.
	if err = c.lockAllResPlanDemand(kt, crpDemandIDs); err != nil {
		logs.Errorf("failed to lock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// examineAllRPDemandLockStatus examine all resource plan demand lock status, return whether pass or not.
func (c *Controller) examineAllRPDemandLockStatus(kt *kit.Kit, crpDemandIDs []int64) (bool, error) {
	listOpt := &types.ListOption{
		Fields: []string{"crp_demand_id", "locked"},
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := c.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	demandLockStatusMap := make(map[int64]enumor.CrpDemandLockStatus)
	for _, detail := range rst.Details {
		if detail.Locked == nil {
			logs.Errorf("locked of crp demand id: %d is empty, rid: %s", detail.CrpDemandID, kt.Rid)
			return false, fmt.Errorf("locked of crp demand id: %d is empty", detail.CrpDemandID)
		}

		demandLockStatusMap[detail.CrpDemandID] = *detail.Locked
	}

	for _, status := range demandLockStatusMap {
		if status == enumor.CrpDemandLocked {
			return false, nil
		}
	}

	return true, nil
}

// lockAllResPlanDemand lock all resource plan demand.
func (c *Controller) lockAllResPlanDemand(kt *kit.Kit, crpDemandIDs []int64) error {
	expr := tools.ContainersExpression("crp_demand_id", crpDemandIDs)
	lockedDemand := &rpcd.ResPlanCrpDemandTable{
		Locked: converter.ValToPtr(enumor.CrpDemandLocked),
	}

	if err := c.dao.ResPlanCrpDemand().Update(kt, expr, lockedDemand); err != nil {
		logs.Errorf("failed to lock resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UnlockAllResPlanDemand unlock all resource plan demand.
func (c *Controller) UnlockAllResPlanDemand(kt *kit.Kit, crpDemandIDs []int64) error {
	expr := tools.ContainersExpression("crp_demand_id", crpDemandIDs)
	unlockedDemand := &rpcd.ResPlanCrpDemandTable{
		Locked: converter.ValToPtr(enumor.CrpDemandUnLocked),
	}

	if err := c.dao.ResPlanCrpDemand().Update(kt, expr, unlockedDemand); err != nil {
		logs.Errorf("failed to unlock resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
