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
	"sync"
	"time"

	model "hcm/cmd/woa-server/model/task"
	demandtime "hcm/cmd/woa-server/service/plan/demand-time"
	tasktypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	rpdc "hcm/pkg/dal/table/resource-plan/res-plan-demand-changelog"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/bkbase"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/concurrence"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"

	"github.com/jmoiron/sqlx"
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
		wildcardSource := strUnionFind.Elements()
		matches, err := c.IsDeviceMatched(kt, wildcardSource, demand.InstanceModel)
		if err != nil {
			logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for idx, match := range matches {
			if match {
				strUnionFind.Union(wildcardSource[idx], demand.InstanceModel)
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

		pool[key] += int64(demand.PlanCoreAmount)
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

	// get plan product all apply order consume pool map.
	orderConsumePoolMap, err := c.getApplyOrderConsumePoolMap(kt, demands)
	if err != nil {
		logs.Errorf("failed to get apply order consume pool map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get order ids of op product in order ids.
	planProdOrderIDs := cvt.MapKeyToSlice(orderConsumePoolMap)
	prodOrderIDs, err := c.getProdOrders(kt, prodID, planProdOrderIDs)
	if err != nil {
		logs.Errorf("failed to get op product orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	prodConsumePool := make(ResPlanPool)
	strUnionFind := NewStrUnionFind()
	for _, consumePool := range orderConsumePoolMap {
		for consumePoolKey := range consumePool {
			strUnionFind.Add(consumePoolKey.DeviceType)
		}
	}

	for _, prodOrderID := range prodOrderIDs {
		consumePool := orderConsumePoolMap[prodOrderID]
		for consumePoolKey, consumeCpuCore := range consumePool {
			deviceType := consumePoolKey.DeviceType
			for _, ele := range strUnionFind.Elements() {
				matched, err := c.IsDeviceMatched(kt, []string{ele}, consumePoolKey.DeviceType)
				if err != nil {
					logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
					return nil, err
				}

				if matched[0] {
					strUnionFind.Union(ele, consumePoolKey.DeviceType)
					deviceType = strUnionFind.Find(ele)
					break
				}
			}

			consumePoolKey.DeviceType = deviceType
			prodConsumePool[consumePoolKey] += consumeCpuCore
		}
	}

	return prodConsumePool, nil
}

// getApplyOrderConsumePoolMap get crp demand ids corresponding apply order consume resource plan pool map.
func (c *Controller) getApplyOrderConsumePoolMap(kt *kit.Kit, demands []*cvmapi.CvmCbsPlanQueryItem) (
	map[string]ResPlanPool, error) {

	orderConsumePoolMap := make(map[string]ResPlanPool)
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

			consumePoolKey := ResPlanPoolKey{
				PlanType:      enumor.PlanType(demand.InPlan).ToAnotherPlanType(),
				AvailableTime: NewAvailableTime(demand.Year, time.Month(demand.Month)),
				DeviceType:    demand.InstanceModel,
				ObsProject:    enumor.ObsProject(demand.ProjectName),
				RegionName:    demand.CityName,
				ZoneName:      demand.ZoneName,
			}

			// 消耗预测CRP changelog中ChangeCoreAmount对应负值，因此需要乘-1取反
			consumeCpuCore := -int64(changelog.ChangeCoreAmount)

			mutex.Lock()
			if _, ok := orderConsumePoolMap[changelog.OrderId]; !ok {
				orderConsumePoolMap[changelog.OrderId] = make(ResPlanPool)
			}
			orderConsumePoolMap[changelog.OrderId][consumePoolKey] += consumeCpuCore
			mutex.Unlock()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return orderConsumePoolMap, nil
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
	req := cvmapi.NewOrderQueryReq(&cvmapi.OrderQueryParam{OrderId: orderIDs})
	resp, err := c.crpCli.QueryCvmOrders(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query cvm orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Result == nil {
		logs.Errorf("query cvm orders, but result is nil, trace id: %s, rid: %s", resp.TraceId, kt.Rid)
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
			// TODO: 预测内的总预测需要 * 120%，目前没整清楚120%的逻辑，先按100%计算
			prodMaxAvailablePool[k] = v
		} else {
			prodMaxAvailablePool[k] = v
		}
	}

	// matching.
	for prodResPlanKey, consumeCpuCore := range prodConsumePool {
		// zone name should not be empty, it may use resource plan which zone name is equal or zone name is empty.
		plan, ok := prodPlanPool[prodResPlanKey]
		if ok {
			canConsume := max(min(plan, consumeCpuCore), 0)
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
			canConsume := max(min(plan, consumeCpuCore), 0)
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
			key.RegionName != need.RegionName {
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

		canConsume := max(min(availableCpuCore, needCpuCore), 0)
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
				key.RegionName != need.RegionName {
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

			canConsume := max(min(remainCpuCore, needCpuCore), 0)
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

// listAllPlanDemandsByBkBizID list all plan demand by bk biz id.
func (c *Controller) listAllPlanDemandsByBkBizID(kt *kit.Kit, bkBizID int64, startDay, endDay time.Time) (
	[]rpd.ResPlanDemandTable, error) {

	startDate, err := strconv.Atoi(startDay.Format(bkbase.DateLayout))
	if err != nil {
		logs.Errorf("failed to converting stary day to integer: %v", err)
		return nil, err
	}
	endDate, err := strconv.Atoi(endDay.Format(bkbase.DateLayout))
	if err != nil {
		logs.Errorf("failed to converting end day to integer: %v", err)
		return nil, err
	}

	planDemandDetails := make([]rpd.ResPlanDemandTable, 0)
	listOpt := &rpproto.ResPlanDemandListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("locked", int8(enumor.CrpDemandUnLocked)),
				tools.RuleEqual("bk_biz_id", bkBizID),
				tools.RuleGreaterThanEqual("expect_time", startDate),
				tools.RuleLessThanEqual("expect_time", endDate),
			),
			Page: core.NewDefaultBasePage(),
		},
	}
	for {
		rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listOpt)
		if err != nil {
			logs.Errorf("failed to list resource plan demand, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, detail := range rst.Details {
			planDemandDetails = append(planDemandDetails, detail)
		}

		if len(rst.Details) < int(listOpt.Page.Limit) {
			break
		}

		listOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return planDemandDetails, nil
}

// GetProdResConsumePoolV2 get biz resource consume pool v2.
func (c *Controller) GetProdResConsumePoolV2(kt *kit.Kit, bkBizIDs []int64, startDay, endDay time.Time) (
	ResPlanConsumePool, error) {

	// list apply order from db by bk biz id.
	subOrders, err := c.listApplyOrder(kt, bkBizIDs, startDay, endDay)
	if err != nil {
		logs.Errorf("failed to list apply order details, err: %v, bkBizIDs: %v, startDay: %s, endDay: %s, rid: %s",
			err, bkBizIDs, startDay.Format(constant.TimeStdFormat), endDay.Format(constant.TimeStdFormat), kt.Rid)
		return nil, err
	}

	// get plan product all apply order consume pool map.
	orderConsumePoolMap, err := c.getApplyOrderConsumePoolMapV2(kt, subOrders)
	if err != nil {
		logs.Errorf("failed to get apply order consume pool map v2, err: %v, subOrders: %+v, rid: %s",
			err, cvt.PtrToSlice(subOrders), kt.Rid)
		return nil, err
	}

	prodConsumePool := make(ResPlanConsumePool)
	strUnionFind := NewStrUnionFind()
	for consumePoolKey := range orderConsumePoolMap {
		strUnionFind.Add(consumePoolKey.DeviceType)
	}

	for consumePoolKey, consumeCpuCore := range orderConsumePoolMap {
		deviceType := consumePoolKey.DeviceType
		for _, ele := range strUnionFind.Elements() {
			matched, err := c.IsDeviceMatched(kt, []string{ele}, consumePoolKey.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device is matched v2, err: %v, consumePoolKey: %+v, rid: %s",
					err, consumePoolKey, kt.Rid)
				return nil, err
			}

			if matched[0] {
				strUnionFind.Union(ele, consumePoolKey.DeviceType)
				deviceType = strUnionFind.Find(ele)
				break
			}
		}

		consumePoolKey.DeviceType = deviceType
		prodConsumePool[consumePoolKey] += consumeCpuCore
	}
	logs.Infof("get biz resource consume pool v2, bkBizIDs: %v, startDay: %s, endDay: %s, pool: %+v, rid: %s", bkBizIDs,
		startDay.Format(constant.TimeStdFormat), endDay.Format(constant.TimeStdFormat), prodConsumePool, kt.Rid)

	return prodConsumePool, nil
}

// listApplyOrder list apply order from db by bk biz ids.
func (c *Controller) listApplyOrder(kt *kit.Kit, bkBizIDs []int64, startDay, endDay time.Time) (
	[]*tasktypes.ApplyOrder, error) {

	listFilter := map[string]interface{}{
		"bk_biz_id": mapstr.MapStr{
			pkg.BKDBIN: bkBizIDs,
		},
		"stage":  tasktypes.TicketStageDone,
		"status": tasktypes.ApplyStatusDone,
		"create_at": mapstr.MapStr{
			pkg.BKDBGTE: startDay,
			pkg.BKDBLTE: endDay,
		},
	}
	page := metadata.BasePage{
		Limit: pkg.BKNoLimit,
	}

	orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, listFilter)
	if err != nil {
		logs.Errorf("failed to list apply order by bkBizIDs: %v, rid: %s", bkBizIDs, kt.Rid)
		return nil, err
	}

	return orders, nil
}

// getApplyOrderConsumePoolMapV2 get apply order consume resource plan pool map.
func (c *Controller) getApplyOrderConsumePoolMapV2(kt *kit.Kit, subOrders []*tasktypes.ApplyOrder) (
	ResPlanConsumePool, error) {

	orderConsumePoolMap := make(ResPlanConsumePool)
	for _, subOrderInfo := range subOrders {
		planType, err := c.GetPlanTypeByChargeType(subOrderInfo.Spec.ChargeType)
		if err != nil {
			logs.Errorf("failed to get plan type by charge type, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		var demandClass enumor.DemandClass
		if subOrderInfo.ResourceType == tasktypes.ResourceTypeCvm {
			demandClass = enumor.DemandClassCVM
		}

		consumePoolKey := ResPlanPoolKeyV2{
			PlanType:      planType,
			AvailableTime: NewAvailableTime(subOrderInfo.CreateAt.Year(), subOrderInfo.CreateAt.Month()),
			DeviceType:    subOrderInfo.Spec.DeviceType,
			ObsProject:    subOrderInfo.RequireType.ToObsProject(),
			BkBizID:       subOrderInfo.BkBizId,
			DemandClass:   demandClass,
			RegionID:      subOrderInfo.Spec.Region,
			ZoneID:        subOrderInfo.Spec.Zone,
			DiskType:      subOrderInfo.Spec.DiskType,
		}

		// 交付的核心数量(消耗预测CRP的核心数)
		consumeCpuCore := int64(subOrderInfo.DeliveredCore)
		orderConsumePoolMap[consumePoolKey] += consumeCpuCore
	}

	return orderConsumePoolMap, nil
}

// VerifyProdDemandsV2 verify whether the needs of biz can be satisfied.
func (c *Controller) VerifyProdDemandsV2(kt *kit.Kit, bkBizID int64, needs []VerifyResPlanElemV2) (
	[]VerifyResPlanResElem, error) {

	prodRemain, prodMaxAvailable, err := c.GetProdResRemainPoolMatch(kt, bkBizID)
	if err != nil {
		logs.Errorf("failed to get product resource remain pool match, bkBizID: %d, err: %v, rid: %s",
			bkBizID, err, kt.Rid)
		return nil, err
	}

	result := make([]VerifyResPlanResElem, len(needs))
	// match each need.
	for i, need := range needs {
		if need.IsPrePaid {
			// verify pre paid.
			result[i], err = c.getPrePaidDemandMatch(kt, prodMaxAvailable, need)
			if err != nil {
				logs.Errorf("failed to loop verify pre paid match, err: %v, need: %+v, rid: %s", err, need, kt.Rid)
				return nil, err
			}
		} else {
			// verify post paid by hour.
			result[i], err = c.getPostPaidByHourDemandMatch(kt, prodRemain, need)
			if err != nil {
				logs.Errorf("failed to loop verify post paid by hour, err: %v, need: %+v, rid: %s", err, need, kt.Rid)
				return nil, err
			}
		}
	}

	return result, nil
}

// GetProdResRemainPoolMatch get biz resource remain pool match.
// @param bkBizID is the bk biz id.
// @return prodRemainedPool is the biz in plan and out plan remained resource plan pool.
// @return prodMaxAvailablePool is the biz in plan and out plan remained max available resource plan pool.
// NOTE: maxAvailableInPlanPool = totalInPlan * 120% - consumeInPlan, because the special rules of the crp system.
func (c *Controller) GetProdResRemainPoolMatch(kt *kit.Kit, bkBizID int64) (ResPlanPoolMatch, ResPlanPoolMatch, error) {
	prodPlanPool, prodConsumePool, err := c.getCurrMonthPlanConsumePool(kt, bkBizID)
	if err != nil {
		return nil, nil, err
	}

	// construct product max available resource plan pool.
	prodMaxAvailablePool := make(ResPlanPoolMatch)
	for k, v := range prodPlanPool {
		if k.PlanType == enumor.PlanTypeCodeInPlan {
			// 预测内的总预测需要 * 120%，目前没整清楚120%的逻辑，先按100%计算
			prodMaxAvailablePool[k] = v
		} else {
			prodMaxAvailablePool[k] = v
		}
	}

	// matching.
	for prodResPlanKey, consumeCpuCore := range prodConsumePool {
		// zone name should not be empty, it may use resource plan which zone name is equal or zone name is empty.
		planMap, ok := prodPlanPool[prodResPlanKey]
		if ok {
			for demandID, planCore := range planMap {
				canConsume := max(min(planCore, consumeCpuCore), 0)
				prodPlanPool[prodResPlanKey][demandID] -= canConsume
				prodMaxAvailablePool[prodResPlanKey][demandID] -= canConsume
				consumeCpuCore -= canConsume
			}
		}

		// 主机申领的时候允许模糊可用区，例如用户可以在南京一区申领机器，消耗南京二区的预测
		zoneIDMap, err := c.GetZoneMapByRegionIDs(kt, []string{prodResPlanKey.RegionID})
		if err != nil {
			return nil, nil, err
		}

		for zoneID := range zoneIDMap {
			keyLoopZone := ResPlanPoolKeyV2{
				PlanType:      prodResPlanKey.PlanType,
				AvailableTime: prodResPlanKey.AvailableTime,
				DeviceType:    prodResPlanKey.DeviceType,
				ObsProject:    prodResPlanKey.ObsProject,
				BkBizID:       prodResPlanKey.BkBizID,
				DemandClass:   prodResPlanKey.DemandClass,
				RegionID:      prodResPlanKey.RegionID,
				DiskType:      prodResPlanKey.DiskType,
				ZoneID:        zoneID,
			}

			planMap, ok = prodPlanPool[keyLoopZone]
			if ok {
				for demandID, planCore := range planMap {
					canConsume := max(min(planCore, consumeCpuCore), 0)
					prodPlanPool[keyLoopZone][demandID] -= canConsume
					prodMaxAvailablePool[keyLoopZone][demandID] -= canConsume
					consumeCpuCore -= canConsume
				}
			}
		}
		logs.Infof("biz resource plan pool is loop matched, bkBizID: %d, record: %+v, ok: %v, plan: %+v, "+
			"zoneIDMap: %+v, prodPlanPool: %+v, maxAvailablePool: %+v, consumeCpuCore: %d, rid: %s", bkBizID,
			prodResPlanKey, ok, planMap, zoneIDMap, prodPlanPool, prodMaxAvailablePool, consumeCpuCore, kt.Rid)
	}

	return prodPlanPool, prodMaxAvailablePool, nil
}

func (c *Controller) getCurrMonthPlanConsumePool(kt *kit.Kit, bkBizID int64) (
	ResPlanPoolMatch, ResPlanConsumePool, error) {

	nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
	startDay := time.Date(nowDemandYear, nowDemandMonth, 1, 0, 0, 0, 0, time.UTC)
	nextMonthDay := time.Date(nowDemandYear, nowDemandMonth+1, 1, 0, 0, 0, 0, time.UTC)
	endDay := nextMonthDay.AddDate(0, 0, -1)

	// get biz resource plan pool.
	prodPlanPool, err := c.GetProdResPlanPoolMatch(kt, bkBizID, startDay, endDay)
	if err != nil {
		logs.Errorf("failed to get biz resource plan pool match, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, nil, err
	}

	// get biz resource consume pool.
	prodConsumePool, err := c.GetProdResConsumePoolV2(kt, []int64{bkBizID}, startDay, endDay)
	if err != nil {
		logs.Errorf("failed to get biz resource consume pool v2, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, nil, err
	}

	// compact keys of prodPlanPool and prodConsumePool, set their matched device type to the same.
	if err = c.compactResPlanPoolMatch(kt, prodPlanPool, prodConsumePool); err != nil {
		logs.Errorf("failed to compact resource plan pool match, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, nil, err
	}
	return prodPlanPool, prodConsumePool, nil
}

// GetProdResPlanPoolMatch get prod resource plan pool match.
func (c *Controller) GetProdResPlanPoolMatch(kt *kit.Kit, bkBizID int64, startDay, endDay time.Time) (
	ResPlanPoolMatch, error) {

	// get biz all unlocked demand details.
	demands, err := c.listAllPlanDemandsByBkBizID(kt, bkBizID, startDay, endDay)
	if err != nil {
		logs.Errorf("failed to list all demand details, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, err
	}

	// construct resource plan pool.
	pool := make(ResPlanPoolMatch)
	strUnionFind := NewStrUnionFind()
	for _, demand := range demands {
		strUnionFind.Add(demand.DeviceType)
	}

	for _, demand := range demands {
		// merge match device type.
		deviceType := demand.DeviceType
		wildcardSource := strUnionFind.Elements()
		matches, err := c.IsDeviceMatched(kt, wildcardSource, demand.DeviceType)
		if err != nil {
			logs.Errorf("failed to check device is matched, err: %v, demand: %+v, rid: %s", err, demand, kt.Rid)
			return nil, err
		}

		for idx, match := range matches {
			if match {
				strUnionFind.Union(wildcardSource[idx], demand.DeviceType)
				deviceType = strUnionFind.Find(demand.DeviceType)
				break
			}
		}

		expectTime, err := time.Parse(bkbase.DateLayout, strconv.Itoa(demand.ExpectTime))
		if err != nil {
			return nil, fmt.Errorf("conv expect time failed, expectTime: %d, err: %v", demand.ExpectTime, err)
		}

		key := ResPlanPoolKeyV2{
			PlanType:      demand.PlanType,
			AvailableTime: NewAvailableTime(expectTime.Year(), expectTime.Month()),
			DeviceType:    deviceType,
			ObsProject:    demand.ObsProject,
			BkBizID:       demand.BkBizID,
			DemandClass:   demand.DemandClass,
			RegionID:      demand.RegionID,
			ZoneID:        demand.ZoneID,
			DiskType:      demand.DiskType,
		}
		if _, ok := pool[key]; !ok {
			pool[key] = make(map[string]int64, 0)
		}
		pool[key][demand.ID] += cvt.PtrToVal(demand.CpuCore)
	}
	// 记录日志方便排查问题
	logs.Infof("get res plan demand pool match success, bkBizID: %d, pool: %+v, rid: %s", bkBizID, pool, kt.Rid)
	return pool, nil
}

// compactResPlanPoolMatch 紧凑资源预测池，使pool1和pool2的key中，可以通配的机型设置为同一机型。
// 例如：pool1中存在device_type1, device_type2，pool2中存在device_type3, device_type4，
// 假设device_type1和device_type3通配，device_type2和device_type4通配，
// 那么pool2中的device_type3会被修改为device_type1，device_type4会被修改为device_type2。
func (c *Controller) compactResPlanPoolMatch(kt *kit.Kit, pool1 ResPlanPoolMatch, pool2 ResPlanConsumePool) error {
	for k1 := range pool1 {
		for k2, v2 := range pool2 {
			if k1.DeviceType == k2.DeviceType {
				continue
			}

			matched, err := c.IsDeviceMatched(kt, []string{k1.DeviceType}, k2.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched v2, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			if matched[0] {
				newK2 := ResPlanPoolKeyV2{
					PlanType:      k2.PlanType,
					AvailableTime: k2.AvailableTime,
					DeviceType:    k1.DeviceType,
					ObsProject:    k2.ObsProject,
					DemandClass:   k2.DemandClass,
					RegionID:      k2.RegionID,
					ZoneID:        k2.ZoneID,
					DiskType:      k2.DiskType,
				}

				if _, ok := pool2[newK2]; !ok {
					pool2[newK2] = v2
					delete(pool2, k2)
				}
			}
		}
	}
	return nil
}

// getPrePaidDemandMatch get prepaid demand match.
func (c *Controller) getPrePaidDemandMatch(kt *kit.Kit, prodMaxAvailable ResPlanPoolMatch, need VerifyResPlanElemV2) (
	VerifyResPlanResElem, error) {

	matchDemandIDs := make([]string, 0)
	// 检查[预测内]是否有余量
	needCpuCore := need.CpuCore
	for key, availableCpuCoreMap := range prodMaxAvailable {
		// 已经匹配完需要的核心数
		if needCpuCore == 0 {
			break
		}

		if key.PlanType != enumor.PlanTypeCodeInPlan || key.AvailableTime != need.AvailableTime ||
			key.ObsProject != need.ObsProject || key.BkBizID != need.BkBizID || key.DemandClass != need.DemandClass ||
			key.RegionID != need.RegionID || key.DiskType != need.DiskType {
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

		for demandID, availableCpuCore := range availableCpuCoreMap {
			if availableCpuCore <= 0 || needCpuCore <= 0 {
				continue
			}
			canConsume := max(min(availableCpuCore, needCpuCore), 0)
			needCpuCore -= canConsume
			prodMaxAvailable[key][demandID] -= canConsume
			matchDemandIDs = append(matchDemandIDs, demandID)
			logs.Infof("get pre paid match loop, key: %+v, demandID: %s, prodMaxAvailableKey: %+v, canConsume: %f, "+
				"needCpuCore: %f, need: %+v, rid: %s", key, demandID, prodMaxAvailable[key], canConsume,
				needCpuCore, need, kt.Rid)
		}
	}

	if needCpuCore != 0 {
		return VerifyResPlanResElem{
			VerifyResult: enumor.VerifyResPlanRstFailed,
			Reason:       "in plan resource is not enough",
		}, nil
	}

	return VerifyResPlanResElem{VerifyResult: enumor.VerifyResPlanRstPass, MatchDemandIDs: matchDemandIDs}, nil
}

// getPostPaidByHourDemandMatch get post paid by hour demand match.
func (c *Controller) getPostPaidByHourDemandMatch(kt *kit.Kit, prodRemain ResPlanPoolMatch, need VerifyResPlanElemV2) (
	VerifyResPlanResElem, error) {

	matchDemandIDs := make([]string, 0)
	needCpuCore := need.CpuCore
	for _, planType := range enumor.GetPlanTypeCodeHcmMembers() {
		for key, remainCpuCoreMap := range prodRemain {
			// 已经匹配完需要的核心数
			if needCpuCore == 0 {
				break
			}

			if key.PlanType != planType || key.AvailableTime != need.AvailableTime ||
				key.ObsProject != need.ObsProject || key.BkBizID != need.BkBizID ||
				key.DemandClass != need.DemandClass || key.RegionID != need.RegionID || key.DiskType != need.DiskType {
				continue
			}

			matched, err := c.IsDeviceMatched(kt, []string{key.DeviceType}, need.DeviceType)
			if err != nil {
				logs.Errorf("failed to check device matched match, err: %v, key: %+v, remainCpuCoreMap: %+v, "+
					"need: %+v, rid: %s", err, key, remainCpuCoreMap, need, kt.Rid)
				return VerifyResPlanResElem{}, err
			}

			if !matched[0] {
				continue
			}

			for demandID, remainCpuCore := range remainCpuCoreMap {
				if remainCpuCore <= 0 || needCpuCore <= 0 {
					continue
				}
				canConsume := max(min(remainCpuCore, needCpuCore), 0)
				needCpuCore -= canConsume
				prodRemain[key][demandID] -= canConsume
				matchDemandIDs = append(matchDemandIDs, demandID)
			}
			logs.Infof("get post paid by hour match loop, key: %+v, remainCpuCoreMap: %+v, prodRemain[key]: %+v, "+
				"needCpuCore: %f, need: %+v, rid: %s", key, remainCpuCoreMap, prodRemain[key],
				needCpuCore, need, kt.Rid)
		}
	}

	if needCpuCore != 0 {
		return VerifyResPlanResElem{
			VerifyResult: enumor.VerifyResPlanRstFailed,
			Reason:       "in plan or out plan resource is not enough",
		}, nil
	}

	return VerifyResPlanResElem{VerifyResult: enumor.VerifyResPlanRstPass, MatchDemandIDs: matchDemandIDs}, nil
}

// getMatchedPlanDemandIDs get matched plan demand ids.
func (c *Controller) getMatchedPlanDemandIDs(kt *kit.Kit, bkBizID int64, subOrder *tasktypes.ApplyOrder) (
	[]string, error) {

	// 是否包年包月
	isPrePaid := true
	if subOrder.Spec.ChargeType != cvmapi.ChargeTypePrePaid {
		isPrePaid = false
	}

	nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
	availableTime := NewAvailableTime(nowDemandYear, nowDemandMonth)

	verifySlice := make([]VerifyResPlanElemV2, 0)
	verifySlice = append(verifySlice, VerifyResPlanElemV2{
		IsPrePaid:     isPrePaid,
		AvailableTime: availableTime,
		DeviceType:    subOrder.Spec.DeviceType,
		ObsProject:    subOrder.RequireType.ToObsProject(),
		BkBizID:       bkBizID,
		DemandClass:   enumor.DemandClassCVM,
		RegionID:      subOrder.Spec.Region,
		ZoneID:        subOrder.Spec.Zone,
		DiskType:      subOrder.Spec.DiskType,
		CpuCore:       int64(subOrder.AppliedCore),
	})

	// call verify resource plan demands to verify each cvm demands.
	ret, err := c.VerifyProdDemandsV2(kt, bkBizID, verifySlice)
	if err != nil {
		logs.Errorf("failed to get matched resource plan demand ids, err: %v, bkBizID: %d, subOrder: %+v, rid: %s",
			err, bkBizID, cvt.PtrToVal(subOrder), kt.Rid)
		return nil, err
	}

	if len(ret) != 1 {
		return nil, errf.Newf(errf.InvalidParameter, "get matched plan demand result length is not 1, "+
			"verify result: %+v", ret)
	}

	return ret[0].MatchDemandIDs, nil
}

// AddMatchedPlanDemandExpendLogs add matched plan demand expend logs.
func (c *Controller) AddMatchedPlanDemandExpendLogs(kt *kit.Kit, bkBizID int64, subOrder *tasktypes.ApplyOrder) error {
	// if resource type is not cvm,	return success.
	if subOrder.ResourceType != tasktypes.ResourceTypeCvm {
		return nil
	}

	demandIDs, err := c.getMatchedPlanDemandIDs(kt, bkBizID, subOrder)
	if err != nil {
		logs.Errorf("failed to get matched plan demand ids, err: %v, bkBizID: %d, subOrder: %+v, rid: %s",
			err, bkBizID, cvt.PtrToVal(subOrder), kt.Rid)
		return err
	}

	if len(demandIDs) == 0 {
		logs.Infof("get matched plan demand ids empty, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
			bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
		return nil
	}

	// 记录日志方便排查问题
	logs.Infof("get matched plan demand ids success, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
		bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)

	demadOpt := &types.ListOption{
		Filter: tools.ContainersExpression("id", demandIDs),
		Page:   core.NewDefaultBasePage(),
	}
	demandListResp, err := c.dao.ResPlanDemand().List(kt, demadOpt)
	if err != nil {
		logs.Errorf("list resource plan demand by ids failed, err: %v, demandIDs: %v, rid: %s", err, demandIDs, kt.Rid)
		return fmt.Errorf("list resource plan demand by ids failed, err: %v, demandIDs: %v", err, demandIDs)
	}
	if len(demandListResp.Details) == 0 {
		logs.Infof("matched list plan demand from db ids empty, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
			bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
		return nil
	}

	inserts := make([]rpdc.DemandChangelogTable, len(demandIDs))
	for idx, demandItem := range demandListResp.Details {
		inserts[idx] = rpdc.DemandChangelogTable{
			DemandID:      demandItem.ID,
			SuborderID:    subOrder.SubOrderId,
			Type:          enumor.DemandChangelogTypeExpend,
			ExpectTime:    demandItem.ExpectTime,
			ObsProject:    subOrder.RequireType.ToObsProject(),
			RegionName:    demandItem.RegionName,
			ZoneName:      demandItem.ZoneName,
			DeviceType:    subOrder.Spec.DeviceType,
			CpuCoreChange: cvt.ValToPtr(int64(subOrder.AppliedCore)),
			Remark:        subOrder.Remark,
		}
	}
	_, err = c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return c.dao.ResPlanDemandChangelog().CreateWithTx(kt, txn, inserts)
	})
	if err != nil {
		logs.Errorf("failed to create plan crp demand log, err: %v, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
			err, bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
	}
	return nil
}

// GetAllDeviceTypeMap get all device type map.
func (c *Controller) GetAllDeviceTypeMap(kt *kit.Kit) (map[string]wdt.WoaDeviceTypeTable, error) {
	// get all device type maps.
	deviceTypeMap, err := c.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("get all device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return deviceTypeMap, nil
}
