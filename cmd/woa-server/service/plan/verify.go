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
	"fmt"
	"slices"
	"time"

	"hcm/cmd/woa-server/logics/plan"
	"hcm/cmd/woa-server/model/config"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types/meta"
	woadevicetype "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// VerifyResPlanDemand verify resource plan demand.
func (s *service) VerifyResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.VerifyResPlanDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode verify resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate verify resource plan demand request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get biz id corresponding op product id and plan product id.
	bizOrgRel, err := s.logics.GetBizOrgRel(cts.Kit, req.BkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	result, err := s.verifyResPlanDemand(cts.Kit, bizOrgRel, req.RequireType.ToObsProject(), req.Suborders)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &ptypes.VerifyResPlanDemandResp{Verifications: result}, nil
}

// verifyResPlanDemand verify resource plan demand.
func (s *service) verifyResPlanDemand(kt *kit.Kit, bizOrgRel *ptypes.BizOrgRel, obsProject enumor.ObsProject,
	suborders []task.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

	// get needed meta maps.
	regionMap, zoneMap, deviceTypeMap, err := s.getVerifyRPDemandNeededMap(kt)
	if err != nil {
		logs.Errorf("failed to get verify resource plan demand needed map, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	result := make([]ptypes.VerifyResPlanDemandElem, len(suborders))
	indexMap := make(map[int]int)
	verifySlice := make([]plan.VerifyResPlanElem, 0)

	for idx, subOrder := range suborders {
		// if resource type is not cvm, set verify result to not involved.
		if subOrder.ResourceType != task.ResourceTypeCvm {
			result[idx] = ptypes.VerifyResPlanDemandElem{
				VerifyResult: enumor.VerifyResPlanRstNotInvolved,
			}
			continue
		}

		// if charge type is postpaid by hour, set is any plan type to true.
		isAnyPlanType := false
		if subOrder.Spec.ChargeType == cvmapi.ChargeTypePostPaidByHour {
			isAnyPlanType = true
		}

		// TODO：可用日期目前设置为当前月份，需改为crp定义的可用年月
		availableTime := plan.NewAvailableTime(time.Now().Year(), int(time.Now().Month()))

		indexMap[len(verifySlice)] = idx
		verifySlice = append(verifySlice, plan.VerifyResPlanElem{
			IsAnyPlanType: isAnyPlanType,
			PlanType:      enumor.PlanTypeHcmInPlan,
			AvailableTime: availableTime,
			DeviceType:    subOrder.Spec.DeviceType,
			ObsProject:    obsProject,
			RegionName:    regionMap[subOrder.Spec.Region].RegionName,
			ZoneName:      zoneMap[subOrder.Spec.Zone],
			CpuCore:       float64(subOrder.Replicas) * float64(deviceTypeMap[subOrder.Spec.DeviceType].CpuCore),
		})
	}

	// call verify resource plan demands to verify each cvm demands.
	rst, err := s.planController.VerifyProdDemands(kt, bizOrgRel.OpProductID, bizOrgRel.PlanProductID, verifySlice)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// set result.
	for idx, pass := range rst {
		verifyResult := enumor.VerifyResPlanRstFailed
		if pass {
			verifyResult = enumor.VerifyResPlanRstPass
		}
		result[indexMap[idx]] = ptypes.VerifyResPlanDemandElem{
			VerifyResult: verifyResult,
		}
	}

	return result, nil
}

// getVerifyRPDemandNeededMap get verify resource plan demand needed region map, zone map, device type map.
func (s *service) getVerifyRPDemandNeededMap(kt *kit.Kit) (map[string]meta.RegionArea, map[string]string,
	map[string]woadevicetype.WoaDeviceTypeTable, error) {
	// get region map.
	regionMap, err := s.dao.WoaZone().GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("failed to get region area map, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get zone map.
	zoneMap, err := s.dao.WoaZone().GetZoneMap(kt)
	if err != nil {
		logs.Errorf("failed to get zone map, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get device type map.
	deviceTypeMap, err := s.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	return regionMap, zoneMap, deviceTypeMap, nil
}

// GetCvmChargeTypeDeviceType get cvm charge type device type.
func (s *service) GetCvmChargeTypeDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.GetCvmChargeTypeDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode get cvm charge type device type request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate get cvm charge type device type request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get biz id corresponding op product id and plan product id.
	bizOrgRel, err := s.logics.GetBizOrgRel(cts.Kit, req.BkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// get op product all remained resource plan.
	prodRemainMap, err := s.planController.GetProdResRemainPool(cts.Kit, bizOrgRel.OpProductID, bizOrgRel.PlanProductID)
	if err != nil {
		logs.Errorf("failed to get op product remained resource plan, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// get needed meta maps.
	regionMap, zoneMap, _, err := s.getVerifyRPDemandNeededMap(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get verify resource plan demand needed map, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	obsProject := req.RequireType.ToObsProject()
	regionName := regionMap[req.Region].RegionName
	zoneName := zoneMap[req.Zone]

	prePaidAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypes(cts.Kit, cvmapi.ChargeTypePrePaid, obsProject,
		regionName, zoneName, prodRemainMap)
	if err != nil {
		logs.Errorf("failed to get pre paid available device types, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	postPaidByHourAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypes(cts.Kit, cvmapi.ChargeTypePostPaidByHour,
		obsProject, regionName, zoneName, prodRemainMap)
	if err != nil {
		logs.Errorf("failed to get post paid by hour available device types, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	prePaidAvailable := slices.ContainsFunc(prePaidAvlDeviceTypes,
		func(deviceAvailable ptypes.DeviceTypeAvailable) bool {
			return deviceAvailable.Available == true
		})

	postPaidByHourAvailable := slices.ContainsFunc(postPaidByHourAvlDeviceTypes,
		func(deviceAvailable ptypes.DeviceTypeAvailable) bool {
			return deviceAvailable.Available == true
		})

	infos := []ptypes.GetCvmChargeTypeDeviceTypeElem{
		{
			ChargeType:  cvmapi.ChargeTypePrePaid,
			Available:   prePaidAvailable,
			DeviceTypes: prePaidAvlDeviceTypes,
		},
		{
			ChargeType:  cvmapi.ChargeTypePostPaidByHour,
			Available:   postPaidByHourAvailable,
			DeviceTypes: postPaidByHourAvlDeviceTypes,
		},
	}

	return &ptypes.GetCvmChargeTypeDeviceTypeRst{Count: int64(len(infos)), Info: infos}, nil
}

// getChargeTypeAvlDeviceTypes get charge type available device types.
func (s *service) getChargeTypeAvlDeviceTypes(kt *kit.Kit, chargeType cvmapi.ChargeType, obsProject enumor.ObsProject,
	regionName, zoneName string, prodRemainMap map[plan.ResPlanPoolKey]float64) ([]ptypes.DeviceTypeAvailable, error) {

	// if charge type is pre paid, get available device types from in plan.
	if chargeType == cvmapi.ChargeTypePrePaid {
		return s.getPlanTypeAvlDeviceTypes(kt, enumor.PlanTypeHcmInPlan, obsProject, regionName, zoneName,
			prodRemainMap)
	}

	// otherwise, get available device types from union in plan and out plan.
	inPlanAvlDeviceTypes, err := s.getPlanTypeAvlDeviceTypes(kt, enumor.PlanTypeHcmInPlan, obsProject, regionName,
		zoneName, prodRemainMap)
	if err != nil {
		logs.Errorf("failed to get in plan available device types, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	outPlanAvlDeviceTypes, err := s.getPlanTypeAvlDeviceTypes(kt, enumor.PlanTypeHcmOutPlan, obsProject, regionName,
		zoneName, prodRemainMap)
	if err != nil {
		logs.Errorf("failed to get out plan available device types, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct device type available map.
	deviceTypeAvlMap := make(map[string]bool)
	for _, ele := range inPlanAvlDeviceTypes {
		deviceTypeAvlMap[ele.DeviceType] = deviceTypeAvlMap[ele.DeviceType] || ele.Available
	}

	for _, ele := range outPlanAvlDeviceTypes {
		deviceTypeAvlMap[ele.DeviceType] = deviceTypeAvlMap[ele.DeviceType] || ele.Available
	}

	result := make([]ptypes.DeviceTypeAvailable, 0, len(deviceTypeAvlMap))
	for deviceType, available := range deviceTypeAvlMap {
		result = append(result, ptypes.DeviceTypeAvailable{DeviceType: deviceType, Available: available})
	}

	return result, nil
}

// getPlanTypeAvlDeviceTypes get plan type available device types.
func (s *service) getPlanTypeAvlDeviceTypes(kt *kit.Kit, planType enumor.PlanType, obsProject enumor.ObsProject,
	regionName, zoneName string, prodRemainMap map[plan.ResPlanPoolKey]float64) ([]ptypes.DeviceTypeAvailable, error) {

	// get region and zone all matched device types from mongodb.
	matchedDeviceTypes, err := s.getMatchedDeviceTypesFromMgo(kt, regionName, zoneName)
	if err != nil {
		logs.Errorf("failed to get matched device types, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get available device type map.
	avlDeviceTypeMap := make(map[string]struct{})

	// TODO：可用日期目前设置为当前月份，需改为crp定义的可用年月
	availableTime := plan.NewAvailableTime(time.Now().Year(), int(time.Now().Month()))

	for _, deviceType := range matchedDeviceTypes {
		key := plan.ResPlanPoolKey{
			PlanType:      planType,
			AvailableTime: availableTime,
			DeviceType:    deviceType,
			ObsProject:    obsProject,
			RegionName:    regionName,
			ZoneName:      zoneName,
		}

		if v := prodRemainMap[key]; v > 0 {
			avlDeviceTypeMap[deviceType] = struct{}{}
		}

		keyWithoutZone := plan.ResPlanPoolKey{
			PlanType:      planType,
			AvailableTime: availableTime,
			DeviceType:    deviceType,
			ObsProject:    obsProject,
			RegionName:    regionName,
		}

		if v := prodRemainMap[keyWithoutZone]; v > 0 {
			avlDeviceTypeMap[deviceType] = struct{}{}
		}
	}

	// get available device type's matched device type.
	for deviceType := range avlDeviceTypeMap {
		matched, err := s.planController.IsDeviceMatched(kt, matchedDeviceTypes, deviceType)
		if err != nil {
			logs.Errorf("failed to check device matched, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for idx, ok := range matched {
			if ok {
				avlDeviceTypeMap[matchedDeviceTypes[idx]] = struct{}{}
			}
		}
	}

	// convert device type map to result.
	result := make([]ptypes.DeviceTypeAvailable, len(matchedDeviceTypes))
	for idx, deviceType := range matchedDeviceTypes {
		if _, ok := avlDeviceTypeMap[deviceType]; ok {
			result[idx] = ptypes.DeviceTypeAvailable{
				DeviceType: deviceType,
				Available:  true,
			}
		} else {
			result[idx] = ptypes.DeviceTypeAvailable{
				DeviceType: deviceType,
				Available:  false,
			}
		}
	}

	return result, nil
}

// getMatchedDeviceTypesFromMgo get matched device types from mongodb.
func (s *service) getMatchedDeviceTypesFromMgo(kt *kit.Kit, regionName, zoneName string) ([]string, error) {
	// get zone name id map and region name id map.
	zoneNameMap, regionNameMap, err := s.getMetaNameMaps(kt)
	if err != nil {
		logs.Errorf("failed to get meta name maps, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// construct mongodb filter.
	mgoFilter := map[string]interface{}{
		"region":       regionNameMap[regionName].RegionID,
		"zone":         zoneNameMap[zoneName],
		"enable_apply": true,
	}

	matchedDeviceTypeInterfaces, err := config.Operation().CvmDevice().FindManyDeviceType(kt.Ctx, mgoFilter)
	if err != nil {
		logs.Errorf("failed to find many device type, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	matchedDeviceTypes := make([]string, len(matchedDeviceTypeInterfaces))
	for idx, deviceTypeInterface := range matchedDeviceTypeInterfaces {
		deviceTypeStr, ok := deviceTypeInterface.(string)
		if !ok {
			logs.Errorf("failed to convert device type interface: %v to string, rid: %s", deviceTypeInterface, kt.Rid)
			return nil, fmt.Errorf("failed to convert device type interface: %v to string", deviceTypeInterface)
		}

		matchedDeviceTypes[idx] = deviceTypeStr
	}

	return matchedDeviceTypes, nil
}
