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
	"math"
	"strconv"
	"sync"

	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// IsDeviceMatched return whether each device type in deviceTypeSlice can use deviceType's resource plan.
func (c *Controller) IsDeviceMatched(kt *kit.Kit, deviceTypeSlice []string, deviceType string) ([]bool, error) {
	// get device type map.
	deviceTypeMap, err := c.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]bool, len(deviceTypeSlice))
	for idx, ele := range deviceTypeSlice {
		// if ele and device type are equal, then they are matched.
		if ele == deviceType {
			result[idx] = true
		}

		if _, ok := deviceTypeMap[ele]; !ok {
			continue
		}

		if _, ok := deviceTypeMap[deviceType]; !ok {
			continue
		}

		// if device family and core type of ele and device type are equal, then they are matched.
		if deviceTypeMap[ele].DeviceFamily == deviceTypeMap[deviceType].DeviceFamily &&
			deviceTypeMap[ele].CoreType == deviceTypeMap[deviceType].CoreType {

			result[idx] = true
		}
	}

	return result, nil
}

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
		rst, err := c.crpCli.QueryCvmCbsPlans(kt.Ctx, kt.Header(), queryReq)
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

// GetProdResPlanPool get op product resource plan pool.
func (c *Controller) GetProdResPlanPool(kt *kit.Kit, prodID int64) (ResPlanPool, error) {
	// get op product all unlocked crp demand ids.
	opt, err := tools.And(
		tools.RuleEqual("locked", enumor.CrpDemandUnLocked),
		tools.RuleEqual("op_product_id", prodID),
	)
	if err != nil {
		logs.Errorf("failed to and filter expression, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get product all unlocked crp demand details.
	demands, err := c.listAllCrpDemandDetails(kt, opt)
	if err != nil {
		logs.Errorf("failed to list all crp demand details, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct resource plan pool.
	pool := make(ResPlanPool)
	strUnionFind := NewStrUnionFind()
	for _, demand := range demands {
		// merge matched device type.
		deviceType := demand.InstanceModel
		for _, ele := range strUnionFind.Elements() {
			matched, err := c.IsDeviceMatched(kt, []string{demand.InstanceModel}, ele)
			if err != nil {
				logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			if matched[0] {
				strUnionFind.Union(ele, demand.InstanceModel)
				deviceType = strUnionFind.Find(ele)
				break
			}
		}

		key := ResPlanPoolKey{
			PlanType:      enumor.PlanType(demand.InPlan).ToAnotherPlanType(),
			AvailableTime: NewAvailableTime(demand.Year, demand.Month),
			DeviceType:    deviceType,
			ObsProject:    enumor.ObsProject(demand.ProjectName),
			RegionName:    demand.CityName,
			ZoneName:      demand.ZoneName,
		}

		pool[key] += float64(demand.PlanCoreAmount)
	}

	return pool, nil
}

// listAllCrpDemandDetails list all crp resource plan demand details.
func (c *Controller) listAllCrpDemandDetails(kt *kit.Kit, opt *filter.Expression) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	// get op product all unlocked crp demand ids.
	listOpt := &types.ListOption{
		Fields: []string{"crp_demand_id"},
		Filter: opt,
		Page:   core.NewDefaultBasePage(),
	}

	// get all crp demand ids.
	crpDemandIDs := make([]int64, 0)
	for {
		rst, err := c.dao.ResPlanCrpDemand().List(kt, listOpt)
		if err != nil {
			logs.Errorf("failed to list resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for idx, detail := range rst.Details {
			crpDemandIDs[idx] = detail.CrpDemandID
		}

		if len(rst.Details) < int(listOpt.Page.Limit) {
			break
		}

		listOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	// query crp demand ids corresponding demand details.
	demands, err := c.QueryAllDemands(kt, &QueryAllDemandsReq{CrpDemandIDs: crpDemandIDs})
	if err != nil {
		logs.Errorf("failed to query all demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return demands, nil
}

// GetProdResConsumePool get op product resource consume pool.
func (c *Controller) GetProdResConsumePool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, error) {
	// get plan product all crp demand details.
	demands, err := c.listAllCrpDemandDetails(kt, tools.EqualExpression("plan_product_id", planProdID))
	if err != nil {
		logs.Errorf("failed to list all crp demand details, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get plan product all apply order consume map.
	orderConsumeMap, err := c.getApplyOrderConsumeMap(kt, demands)
	if err != nil {
		logs.Errorf("failed to get apply order consume map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get order ids of op product in order ids.
	planProdOrderIDs := converter.MapKeyToSlice(orderConsumeMap)
	prodOrderIDs, err := c.getProdOrders(kt, prodID, planProdOrderIDs)
	if err != nil {
		logs.Errorf("failed to get op product orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	prodConsumePool := make(ResPlanPool)
	strUnionFind := NewStrUnionFind()
	for _, prodOrderID := range prodOrderIDs {
		consume := orderConsumeMap[prodOrderID]
		deviceType := consume.DeviceType
		for _, ele := range strUnionFind.Elements() {
			matched, err := c.IsDeviceMatched(kt, []string{ele}, consume.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			if matched[0] {
				strUnionFind.Union(ele, consume.DeviceType)
				deviceType = strUnionFind.Find(ele)
				break
			}
		}

		key := ResPlanPoolKey{
			PlanType:      consume.PlanType,
			AvailableTime: consume.AvailableTime,
			DeviceType:    deviceType,
			ObsProject:    consume.ObsProject,
			RegionName:    consume.RegionName,
			ZoneName:      consume.ZoneName,
		}

		prodConsumePool[key] += consume.CpuCore
	}

	return prodConsumePool, nil
}

// getApplyOrderConsumeMap get crp demand ids corresponding apply order consume map.
func (c *Controller) getApplyOrderConsumeMap(kt *kit.Kit, demands []*cvmapi.CvmCbsPlanQueryItem) (
	map[string]ResPlanElem, error) {

	orderConsumeMap := make(map[string]ResPlanElem)
	mutex := sync.Mutex{}
	concurrent := make(chan struct{}, 10)
	wg := sync.WaitGroup{}
	var hitError error

	for _, demand := range demands {
		concurrent <- struct{}{}
		wg.Add(1)
		go func(demand *cvmapi.CvmCbsPlanQueryItem) {
			defer func() {
				<-concurrent
				wg.Done()
			}()

			crpDemandID, err := strconv.ParseInt(demand.DemandId, 10, 64)
			if err != nil {
				hitError = err
				logs.Errorf("failed to parse crp demand id, err: %v, rid: %s", err, kt.Rid)
				return
			}

			changelogs, err := c.getDemandAllChangelogs(kt, crpDemandID)
			if err != nil {
				hitError = err
				logs.Errorf("failed to get crp demand all changelogs, err: %v, rid: %s", err, kt.Rid)
				return
			}

			for _, changelog := range changelogs {
				// skip not apply changelog.
				// TODO：补充升降配
				if changelog.SourceType != enumor.CrpOrderSourceTypeApply {
					continue
				}

				elem := ResPlanElem{
					PlanType:      enumor.PlanType(demand.InPlan).ToAnotherPlanType(),
					AvailableTime: NewAvailableTime(demand.Year, demand.Month),
					DeviceType:    demand.InstanceModel,
					ObsProject:    enumor.ObsProject(demand.ProjectName),
					RegionName:    demand.CityName,
					ZoneName:      demand.ZoneName,
					CpuCore:       float64(changelog.ChangeCoreAmount),
				}
				mutex.Lock()
				orderConsumeMap[changelog.OrderId] = elem
				mutex.Unlock()
			}
		}(demand)

	}

	if hitError != nil {
		return nil, hitError
	}

	return orderConsumeMap, nil
}

// getDemandAllChangelogs get crp demand id corresponding all changelogs.
func (c *Controller) getDemandAllChangelogs(kt *kit.Kit, crpDemandID int64) (
	[]*cvmapi.DemandChangeLogQueryLogItem, error) {

	req := &cvmapi.DemandChangeLogQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsDemandChangeLogQueryMethod,
		},
		Params: &cvmapi.DemandChangeLogQueryParam{
			DemandIdList: []int64{crpDemandID},
			Page: &cvmapi.Page{
				Start: 0,
				Size:  int(core.DefaultMaxPageLimit),
			},
		},
	}

	result := make([]*cvmapi.DemandChangeLogQueryLogItem, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		req.Params.Page.Start = start
		resp, err := c.crpCli.QueryDemandChangeLog(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("failed to query crp demand change log, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if resp.Error.Code != 0 {
			logs.Errorf("failed to list crp demand change log, code: %d, msg: %s, rid: %s", resp.Error.Code,
				resp.Error.Message, kt.Rid)
			return nil, fmt.Errorf("failed to list crp demand change log, code: %d, msg: %s", resp.Error.Code,
				resp.Error.Message)
		}

		if resp.Result == nil {
			logs.Errorf("failed to list crp demand change log, for result is empty, rid: %s", kt.Rid)
			return nil, errors.New("failed to list crp demand change log, for result is empty")
		}

		if len(resp.Result.Data) == 0 {
			logs.Errorf("failed to list crp demand change log, for result data is empty, rid: %s", kt.Rid)
			return nil, errors.New("failed to list crp demand change log, for result data is empty")
		}

		result = append(result, resp.Result.Data[0].Info...)

		if len(resp.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
	}
	return result, nil
}

// getProdOrders get order ids of op product in order ids.
func (c *Controller) getProdOrders(kt *kit.Kit, prodID int64, orderIDs []string) ([]string, error) {
	req := &cvmapi.OrderQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmOrderStatusMethod,
		},
		Params: &cvmapi.OrderQueryParam{
			OrderId: orderIDs,
		},
	}

	resp, err := c.crpCli.QueryCvmOrders(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query cvm orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Result == nil {
		logs.Errorf("query cvm orders, but result is nil, rid: %s", kt.Rid)
		return nil, errors.New("query cvm orders, but result is nil")
	}

	prodOrderIDs := make([]string, 0)
	for _, order := range resp.Result.Data {
		if order.ProductId == prodID {
			prodOrderIDs = append(prodOrderIDs, order.OrderId)
		}
	}

	return prodOrderIDs, nil
}

// GetProdResRemainPool get op product resource remain pool.
// NOTE: remain = plan * 120% - consume.
func (c *Controller) GetProdResRemainPool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, error) {
	// get op product resource plan pool.
	prodPlanPool, err := c.GetProdResPlanPool(kt, prodID)
	if err != nil {
		logs.Errorf("failed to get op product resource plan pool, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// TODO: 确认下是否乘120，乘120的话写清楚原因
	for k, v := range prodPlanPool {
		prodPlanPool[k] = v * 1.2
	}

	// get op product resource consume pool.
	prodConsumePool, err := c.GetProdResConsumePool(kt, prodID, planProdID)
	if err != nil {
		logs.Errorf("failed to get op product resource consume pool, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// modify keys of prodPlanPool and prodConsumePool, set their matched device type to the same.
	if err = c.modifyResPlanPool(kt, prodPlanPool, prodConsumePool); err != nil {
		logs.Errorf("failed to modify resource plan pool, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// matching.
	for prodResPlanKey, consumeCpuCore := range prodConsumePool {
		// if zone name is empty, it must match region name.
		if prodResPlanKey.ZoneName == "" {
			plan, ok := prodPlanPool[prodResPlanKey]
			if !ok {
				logs.Errorf("record :%v is not found in op product resource plan pool, rid: %s", prodResPlanKey, kt.Rid)
				return nil, fmt.Errorf("record :%v is not found in op product resource plan pool", prodResPlanKey)
			}

			if consumeCpuCore > plan {
				logs.Errorf("record :%v is not find in op product resource plan pool, rid: %s", prodResPlanKey, kt.Rid)
				return nil, fmt.Errorf("record :%v is not find in op product resource plan pool", prodResPlanKey)
			}

			prodPlanPool[prodResPlanKey] -= consumeCpuCore
			continue
		}

		// zone name is not empty, it may use resource plan which zone name is equal or zone name is empty.
		consumeCpuCoreCopy := consumeCpuCore
		plan, ok := prodPlanPool[prodResPlanKey]
		if ok {
			prodPlanPool[prodResPlanKey] -= math.Min(plan, consumeCpuCoreCopy)
			consumeCpuCoreCopy -= math.Min(plan, consumeCpuCoreCopy)
		}

		if consumeCpuCoreCopy == 0 {
			continue
		}

		keyWithoutZone := ResPlanPoolKey{
			PlanType:      prodResPlanKey.PlanType,
			AvailableTime: prodResPlanKey.AvailableTime,
			DeviceType:    prodResPlanKey.DeviceType,
			ObsProject:    prodResPlanKey.ObsProject,
			RegionName:    prodResPlanKey.RegionName,
		}

		plan, ok = prodPlanPool[keyWithoutZone]
		if !ok {
			logs.Errorf("record :%v is not found in op product resource plan pool, rid: %s", keyWithoutZone, kt.Rid)
			return nil, fmt.Errorf("record :%v is not found in op product resource plan pool", keyWithoutZone)
		}

		if consumeCpuCoreCopy > plan {
			logs.Errorf("record :%v is not enough in op product resource plan pool, rid: %s", keyWithoutZone, kt.Rid)
			return nil, fmt.Errorf("record :%v is not enough in op product resource plan pool", keyWithoutZone)
		}

		prodPlanPool[keyWithoutZone] -= consumeCpuCoreCopy
	}

	return prodPlanPool, nil
}

// modifyResPlanPool modify resource plan pool1 and pool2, set their matched device type to the same.
func (c *Controller) modifyResPlanPool(kt *kit.Kit, pool1, pool2 ResPlanPool) error {
	strUnionFind := NewStrUnionFind()
	for k1 := range pool1 {
		for k2, v2 := range pool2 {
			if k1.DeviceType == k2.DeviceType {
				continue
			}

			matched, err := c.IsDeviceMatched(kt, []string{k1.DeviceType}, k2.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			if matched[0] {
				strUnionFind.Union(k1.DeviceType, k2.DeviceType)
				deviceType := strUnionFind.Find(k1.DeviceType)
				newK2 := ResPlanPoolKey{
					PlanType:      k2.PlanType,
					AvailableTime: k2.AvailableTime,
					DeviceType:    deviceType,
					ObsProject:    k2.ObsProject,
					RegionName:    k2.RegionName,
					ZoneName:      k2.ZoneName,
				}

				pool2[newK2] += v2
				delete(pool2, k2)
			}
		}
	}

	return nil
}

// VerifyProdDemands verify whether the needs of op product can be satisfied.
func (c *Controller) VerifyProdDemands(kt *kit.Kit, prodID, planProdID int64, needs []VerifyResPlanElem) (
	[]bool, error) {

	prodRemain, err := c.GetProdResRemainPool(kt, prodID, planProdID)
	if err != nil {
		logs.Errorf("failed to get prod resource remain pool, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]bool, len(needs))

	// match each need.
	for i, need := range needs {
		for k := range prodRemain {
			if needs[i].CpuCore == 0 {
				continue
			}

			matched, err := c.IsDeviceMatched(kt, []string{need.DeviceType}, k.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			if !matched[0] {
				continue
			}

			key := ResPlanPoolKey{
				PlanType:      need.PlanType,
				AvailableTime: need.AvailableTime,
				DeviceType:    k.DeviceType,
				ObsProject:    need.ObsProject,
				RegionName:    need.RegionName,
				ZoneName:      need.ZoneName,
			}

			if need.IsAnyPlanType {
				result[i] = verifyAnyPlanType(prodRemain, key, need.CpuCore)
			} else {
				result[i] = verifySpecPlanType(prodRemain, key, need.CpuCore)
			}
		}
	}

	return result, nil
}

func verifyAnyPlanType(prodRemain ResPlanPool, key ResPlanPoolKey, needCpuCore float64) bool {
	inPlanKey := ResPlanPoolKey{
		PlanType:      enumor.PlanTypeHcmInPlan,
		AvailableTime: key.AvailableTime,
		DeviceType:    key.DeviceType,
		ObsProject:    key.ObsProject,
		RegionName:    key.RegionName,
		ZoneName:      key.ZoneName,
	}

	remain, ok := prodRemain[inPlanKey]
	if ok {
		prodRemain[inPlanKey] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	outPlanKey := ResPlanPoolKey{
		PlanType:      enumor.PlanTypeHcmOutPlan,
		AvailableTime: key.AvailableTime,
		DeviceType:    key.DeviceType,
		ObsProject:    key.ObsProject,
		RegionName:    key.RegionName,
		ZoneName:      key.ZoneName,
	}

	remain, ok = prodRemain[outPlanKey]
	if ok {
		prodRemain[outPlanKey] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	inPlanKeyWithoutZone := ResPlanPoolKey{
		PlanType:      enumor.PlanTypeHcmInPlan,
		AvailableTime: key.AvailableTime,
		DeviceType:    key.DeviceType,
		ObsProject:    key.ObsProject,
		RegionName:    key.RegionName,
	}

	remain, ok = prodRemain[inPlanKeyWithoutZone]
	if ok {
		prodRemain[inPlanKeyWithoutZone] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	outPlanKeyWithoutZone := ResPlanPoolKey{
		PlanType:      enumor.PlanTypeHcmOutPlan,
		AvailableTime: key.AvailableTime,
		DeviceType:    key.DeviceType,
		ObsProject:    key.ObsProject,
		RegionName:    key.RegionName,
	}

	remain, ok = prodRemain[outPlanKeyWithoutZone]
	if ok {
		prodRemain[outPlanKeyWithoutZone] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	return needCpuCore == 0
}

func verifySpecPlanType(prodRemain ResPlanPool, key ResPlanPoolKey, needCpuCore float64) bool {
	remain, ok := prodRemain[key]
	if ok {
		prodRemain[key] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	keyWithoutZone := ResPlanPoolKey{
		PlanType:      key.PlanType,
		AvailableTime: key.AvailableTime,
		DeviceType:    key.DeviceType,
		ObsProject:    key.ObsProject,
		RegionName:    key.RegionName,
	}

	remain, ok = prodRemain[keyWithoutZone]
	if ok {
		prodRemain[keyWithoutZone] -= math.Min(remain, needCpuCore)
		needCpuCore -= math.Min(remain, needCpuCore)
	}

	return needCpuCore == 0
}
