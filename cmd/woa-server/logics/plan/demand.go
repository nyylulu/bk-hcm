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
	"slices"
	"strconv"
	"sync"
	"time"

	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/concurrence"
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

// QueryIEGDemands query IEG crp demands.
func (c *Controller) QueryIEGDemands(kt *kit.Kit, req *QueryIEGDemandsReq) ([]*cvmapi.CvmCbsPlanQueryItem, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	if err := req.Validate(); err != nil {
		return nil, err
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
			BgName: []string{cvmapi.CvmCbsPlanQueryBgName},
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

		result = append(result, rst.Result.Data...)

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
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

// GetProdResPlanPool get op product resource plan pool.
func (c *Controller) GetProdResPlanPool(kt *kit.Kit, prodID int64) (ResPlanPool, error) {
	// get op product all unlocked crp demand ids.
	opt, err := tools.And(
		tools.RuleEqual("locked", int8(enumor.CrpDemandUnLocked)),
		tools.RuleEqual("op_product_id", prodID),
	)
	if err != nil {
		logs.Errorf("failed to and filter expression, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get product all unlocked crp demand details.
	demands, err := c.listCrpDemandDetails(kt, opt)
	if err != nil {
		logs.Errorf("failed to list all crp demand details, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct resource plan pool.
	pool := make(ResPlanPool)
	strUnionFind := NewStrUnionFind()
	for _, demand := range demands {
		strUnionFind.Add(demand.InstanceModel)
	}

	for _, demand := range demands {
		// merge match device type.
		deviceType := demand.InstanceModel
		matches, err := c.IsDeviceMatched(kt, strUnionFind.Elements(), demand.InstanceModel)
		if err != nil {
			logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for idx, match := range matches {
			if match {
				strUnionFind.Union(strUnionFind.Elements()[idx], demand.InstanceModel)
				deviceType = strUnionFind.Find(demand.InstanceModel)
				break
			}
		}

		key := ResPlanPoolKey{
			PlanType:      enumor.PlanType(demand.InPlan).ToAnotherPlanType(),
			AvailableTime: NewAvailableTime(demand.Year, time.Month(demand.Month)),
			DeviceType:    deviceType,
			ObsProject:    enumor.ObsProject(demand.ProjectName),
			RegionName:    demand.CityName,
			ZoneName:      demand.ZoneName,
		}

		pool[key] += float64(demand.PlanCoreAmount)
	}

	return pool, nil
}

// listCrpDemandDetails list crp resource plan demand details.
func (c *Controller) listCrpDemandDetails(kt *kit.Kit, opt *filter.Expression) ([]*cvmapi.CvmCbsPlanQueryItem, error) {
	if opt == nil {
		return nil, errors.New("expression is nil")
	}

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

		for _, detail := range rst.Details {
			crpDemandIDs = append(crpDemandIDs, detail.CrpDemandID)
		}

		if len(rst.Details) < int(listOpt.Page.Limit) {
			break
		}

		listOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	if len(crpDemandIDs) == 0 {
		return nil, nil
	}

	// query crp demand ids corresponding demand details.
	demands, err := c.QueryIEGDemands(kt, &QueryIEGDemandsReq{CrpDemandIDs: crpDemandIDs})
	if err != nil {
		logs.Errorf("failed to query ieg demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return demands, nil
}

// GetProdResConsumePool get op product resource consume pool.
func (c *Controller) GetProdResConsumePool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, error) {
	// get plan product all crp demand details.
	demands, err := c.listCrpDemandDetails(kt, tools.EqualExpression("plan_product_id", planProdID))
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
	for _, consume := range orderConsumeMap {
		strUnionFind.Add(consume.DeviceType)
	}

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
	limit := constant.SyncConcurrencyDefaultMaxLimit

	err := concurrence.BaseExec(limit, demands, func(demand *cvmapi.CvmCbsPlanQueryItem) error {
		crpDemandID, err := strconv.ParseInt(demand.DemandId, 10, 64)
		if err != nil {
			logs.Errorf("failed to parse crp demand id, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		changelogs, err := c.getDemandAllChangelogs(kt, crpDemandID)
		if err != nil {
			logs.Errorf("failed to get crp demand all changelogs, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		for _, changelog := range changelogs {
			// skip changelog which does not consume resource plan.
			if !slices.Contains(enumor.GetCrpConsumeResPlanSourceTypes(), changelog.SourceType) {
				continue
			}

			// 消耗预测CRP changelog中ChangeCoreAmount对应负值，因此需要乘-1取反
			consumeCpuCore := -float64(changelog.ChangeCoreAmount)
			elem := ResPlanElem{
				PlanType:      enumor.PlanType(demand.InPlan).ToAnotherPlanType(),
				AvailableTime: NewAvailableTime(demand.Year, time.Month(demand.Month)),
				DeviceType:    demand.InstanceModel,
				ObsProject:    enumor.ObsProject(demand.ProjectName),
				RegionName:    demand.CityName,
				ZoneName:      demand.ZoneName,
				CpuCore:       consumeCpuCore,
			}
			mutex.Lock()
			orderConsumeMap[changelog.OrderId] = elem
			mutex.Unlock()
		}

		return nil
	})

	if err != nil {
		return nil, err
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
// @param prodID is the op product id.
// @param planProdID is the corresponding plan product id of the op product id.
// @return prodRemainedPool is the op product in plan and out plan remained resource plan pool.
// @return prodMaxAvailablePool is the op product in plan and out plan remained max available resource plan pool.
// NOTE: maxAvailableInPlanPool = totalInPlan * 120% - consumeInPlan, because the special rules of the crp system.
func (c *Controller) GetProdResRemainPool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, ResPlanPool, error) {
	// get op product resource plan pool.
	prodPlanPool, err := c.GetProdResPlanPool(kt, prodID)
	if err != nil {
		logs.Errorf("failed to get op product resource plan pool, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// get op product resource consume pool.
	prodConsumePool, err := c.GetProdResConsumePool(kt, prodID, planProdID)
	if err != nil {
		logs.Errorf("failed to get op product resource consume pool, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// compact keys of prodPlanPool and prodConsumePool, set their matched device type to the same.
	if err = c.compactResPlanPool(kt, prodPlanPool, prodConsumePool); err != nil {
		logs.Errorf("failed to compact resource plan pool, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// construct product max available resource plan pool.
	prodMaxAvailablePool := make(ResPlanPool)
	for k, v := range prodPlanPool {
		if k.PlanType == enumor.PlanTypeHcmInPlan {
			prodMaxAvailablePool[k] = v * 1.2
		} else {
			prodMaxAvailablePool[k] = v
		}
	}

	// matching.
	for prodResPlanKey, consumeCpuCore := range prodConsumePool {
		// zone name should not be empty, it may use resource plan which zone name is equal or zone name is empty.
		plan, ok := prodPlanPool[prodResPlanKey]
		if ok {
			canConsume := math.Max(math.Min(plan, consumeCpuCore), 0)
			prodPlanPool[prodResPlanKey] -= canConsume
			prodMaxAvailablePool[prodResPlanKey] -= canConsume
			consumeCpuCore -= canConsume
		}

		keyWithoutZone := ResPlanPoolKey{
			PlanType:      prodResPlanKey.PlanType,
			AvailableTime: prodResPlanKey.AvailableTime,
			DeviceType:    prodResPlanKey.DeviceType,
			ObsProject:    prodResPlanKey.ObsProject,
			RegionName:    prodResPlanKey.RegionName,
		}

		plan, ok = prodPlanPool[keyWithoutZone]
		if ok {
			canConsume := math.Max(math.Min(plan, consumeCpuCore), 0)
			prodPlanPool[keyWithoutZone] -= canConsume
			prodMaxAvailablePool[keyWithoutZone] -= canConsume
			consumeCpuCore -= canConsume
		}

		if consumeCpuCore > 0 {
			logs.Errorf("record :%v is not enough in op product resource plan pool, rid: %s", prodResPlanKey, kt.Rid)
			return nil, nil, fmt.Errorf("record :%v is not enough in op product resource plan pool", prodResPlanKey)
		}
	}

	return prodPlanPool, prodMaxAvailablePool, nil
}

// compactResPlanPool 紧凑资源预测池，使pool1和pool2的key中，可以通配的机型设置为同一机型。
// 例如：pool1中存在device_type1, device_type2，pool2中存在device_type3, device_type4，
// 假设device_type1和device_type3通配，device_type2和device_type4通配，
// 那么pool2中的device_type3会被修改为device_type1，device_type4会被修改为device_type2。
func (c *Controller) compactResPlanPool(kt *kit.Kit, pool1, pool2 ResPlanPool) error {
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
				newK2 := ResPlanPoolKey{
					PlanType:      k2.PlanType,
					AvailableTime: k2.AvailableTime,
					DeviceType:    k1.DeviceType,
					ObsProject:    k2.ObsProject,
					RegionName:    k2.RegionName,
					ZoneName:      k2.ZoneName,
				}

				pool2[newK2] = v2
				delete(pool2, k2)
			}
		}
	}

	return nil
}

// VerifyProdDemands verify whether the needs of op product can be satisfied.
func (c *Controller) VerifyProdDemands(kt *kit.Kit, prodID, planProdID int64, needs []VerifyResPlanElem) (
	[]VerifyResPlanResElem, error) {

	prodRemain, prodMaxAvailable, err := c.GetProdResRemainPool(kt, prodID, planProdID)
	if err != nil {
		logs.Errorf("failed to get product resource remain pool, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]VerifyResPlanResElem, len(needs))

	// match each need.
	for i, need := range needs {
		if need.IsPrePaid {
			// verify pre paid.
			result[i], err = c.verifyPrePaid(kt, prodRemain, prodMaxAvailable, need)
			if err != nil {
				logs.Errorf("failed to verify pre paid, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		} else {
			// verify post paid by hour.
			result[i], err = c.verifyPostPaidByHour(kt, prodRemain, need)
			if err != nil {
				logs.Errorf("failed to verify post paid by hour, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return result, nil
}

// verifyPrePaid need to be satisfied require two conditions:
// 1. InPlan + OutPlan >= applied.
// 2. InPlan * 120% - consumed >= applied.
func (c *Controller) verifyPrePaid(kt *kit.Kit, prodRemain, prodMaxAvailable ResPlanPool, need VerifyResPlanElem) (
	VerifyResPlanResElem, error) {

	rst, err := c.verifyPostPaidByHour(kt, prodRemain, need)
	if err != nil {
		logs.Errorf("failed to verify post paid by hour, err: %v, rid: %s", err, kt.Rid)
		return VerifyResPlanResElem{}, err
	}

	if rst.VerifyResult != enumor.VerifyResPlanRstPass {
		return rst, nil
	}

	needCpuCore := need.CpuCore
	for key, availableCpuCore := range prodMaxAvailable {
		if key.PlanType != enumor.PlanTypeHcmInPlan ||
			key.AvailableTime != need.AvailableTime ||
			key.ObsProject != need.ObsProject ||
			key.RegionName != need.RegionName ||
			key.ZoneName != need.ZoneName {
			continue
		}

		matched, err := c.IsDeviceMatched(kt, []string{key.DeviceType}, need.DeviceType)
		if err != nil {
			logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
			return VerifyResPlanResElem{}, err
		}

		if !matched[0] {
			continue
		}

		canConsume := math.Max(math.Min(availableCpuCore, needCpuCore), 0)
		prodMaxAvailable[key] -= canConsume
		needCpuCore -= canConsume
	}

	if needCpuCore != 0 {
		return VerifyResPlanResElem{
			VerifyResult: enumor.VerifyResPlanRstFailed,
			Reason:       "in plan resource is not enough",
		}, nil
	}

	return VerifyResPlanResElem{VerifyResult: enumor.VerifyResPlanRstPass}, nil
}

// verifyPostPaidByHour verify whether the post paid by hour need can be satisfied.
// parameter isPriorityInPlan is used to control whether priority to consume InPlan resource plan.
func (c *Controller) verifyPostPaidByHour(kt *kit.Kit, prodRemain ResPlanPool, need VerifyResPlanElem) (
	VerifyResPlanResElem, error) {

	needCpuCore := need.CpuCore
	for _, planType := range enumor.GetPlanTypeHcmMembers() {
		for key, remainCpuCore := range prodRemain {
			if key.PlanType != planType ||
				key.AvailableTime != need.AvailableTime ||
				key.ObsProject != need.ObsProject ||
				key.RegionName != need.RegionName ||
				key.ZoneName != need.ZoneName {
				continue
			}

			matched, err := c.IsDeviceMatched(kt, []string{key.DeviceType}, need.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
				return VerifyResPlanResElem{}, err
			}

			if !matched[0] {
				continue
			}

			canConsume := math.Max(math.Min(remainCpuCore, needCpuCore), 0)
			prodRemain[key] -= canConsume
			needCpuCore -= canConsume
		}
	}

	if needCpuCore != 0 {
		return VerifyResPlanResElem{
			VerifyResult: enumor.VerifyResPlanRstFailed,
			Reason:       "in plan or out plan resource is not enough",
		}, nil
	}

	return VerifyResPlanResElem{VerifyResult: enumor.VerifyResPlanRstPass}, nil
}
