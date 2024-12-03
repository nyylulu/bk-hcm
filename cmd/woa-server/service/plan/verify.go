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
	"sort"
	"time"

	"hcm/cmd/woa-server/logics/plan"
	"hcm/cmd/woa-server/logics/plan/demand-time"
	"hcm/cmd/woa-server/model/config"
	mtypes "hcm/cmd/woa-server/types/meta"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
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
func (s *service) verifyResPlanDemand(kt *kit.Kit, bizOrgRel *mtypes.BizOrgRel, obsProject enumor.ObsProject,
	suborders []task.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

	// get meta maps.
	zoneMap, regionAreaMap, deviceTypeMap, err := s.getMetaMaps(kt)
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

		isPrePaid := true
		if subOrder.Spec.ChargeType != cvmapi.ChargeTypePrePaid {
			isPrePaid = false
		}

		nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
		availableTime := plan.NewAvailableTime(nowDemandYear, nowDemandMonth)

		indexMap[len(verifySlice)] = idx
		verifySlice = append(verifySlice, plan.VerifyResPlanElem{
			IsPrePaid:     isPrePaid,
			AvailableTime: availableTime,
			DeviceType:    subOrder.Spec.DeviceType,
			ObsProject:    obsProject,
			RegionName:    regionAreaMap[subOrder.Spec.Region].RegionName,
			ZoneName:      zoneMap[subOrder.Spec.Zone],
			CpuCore:       int64(subOrder.Replicas) * deviceTypeMap[subOrder.Spec.DeviceType].CpuCore,
		})
	}

	// call verify resource plan demands to verify each cvm demands.
	rst, err := s.planController.VerifyProdDemands(kt, bizOrgRel.OpProductID, bizOrgRel.PlanProductID, verifySlice)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// set result.
	for idx, ele := range rst {
		result[indexMap[idx]] = ptypes.VerifyResPlanDemandElem{
			VerifyResult: ele.VerifyResult,
			Reason:       ele.Reason,
		}
	}

	return result, nil
}

// VerifyResPlanDemandV2 verify resource plan demand.
func (s *service) VerifyResPlanDemandV2(cts *rest.Contexts) (any, error) {
	req := new(ptypes.VerifyResPlanDemandReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode verify resource plan demand v2 request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate verify resource plan demand v2 request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := s.verifyResPlanDemandV2(cts.Kit, req.BkBizID, req.RequireType.ToObsProject(), req.Suborders)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand v2, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &ptypes.VerifyResPlanDemandResp{Verifications: result}, nil
}

// verifyResPlanDemandV2 verify resource plan demand.
func (s *service) verifyResPlanDemandV2(kt *kit.Kit, bkBizID int64, obsProject enumor.ObsProject,
	subOrders []task.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

	// get all device type maps.
	deviceTypeMap, err := s.planController.GetAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type map failed, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, err
	}

	result := make([]ptypes.VerifyResPlanDemandElem, len(subOrders))
	indexMap := make(map[int]int)
	verifySlice := make([]plan.VerifyResPlanElemV2, 0)

	for idx, subOrder := range subOrders {
		// if resource type is not cvm, set verify result to not involved.
		if subOrder.ResourceType != task.ResourceTypeCvm {
			result[idx] = ptypes.VerifyResPlanDemandElem{
				VerifyResult: enumor.VerifyResPlanRstNotInvolved,
			}
			continue
		}

		// 是否包年包月
		isPrePaid := true
		if subOrder.Spec.ChargeType != cvmapi.ChargeTypePrePaid {
			isPrePaid = false
		}

		nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
		availableTime := plan.NewAvailableTime(nowDemandYear, nowDemandMonth)

		var cpuCore int64
		if deviceInfo, ok := deviceTypeMap[subOrder.Spec.DeviceType]; ok {
			cpuCore = deviceInfo.CpuCore
		}

		indexMap[len(verifySlice)] = idx
		verifySlice = append(verifySlice, plan.VerifyResPlanElemV2{
			IsPrePaid:     isPrePaid,
			AvailableTime: availableTime,
			DeviceType:    subOrder.Spec.DeviceType,
			ObsProject:    obsProject,
			BkBizID:       bkBizID,
			DemandClass:   enumor.DemandClassCVM,
			RegionID:      subOrder.Spec.Region,
			ZoneID:        subOrder.Spec.Zone,
			DiskType:      subOrder.Spec.DiskType,
			CpuCore:       int64(subOrder.Replicas) * cpuCore,
		})
	}

	// call verify resource plan demands to verify each cvm demands.
	rst, err := s.planController.VerifyProdDemandsV2(kt, bkBizID, verifySlice)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand v2, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// set result.
	for idx, ele := range rst {
		result[indexMap[idx]] = ptypes.VerifyResPlanDemandElem{
			VerifyResult: ele.VerifyResult,
			Reason:       ele.Reason,
		}
	}

	logs.Infof("verify res plan demand v2 end, bkBizID: %d, verifySlice: %+v, result: %+v, rid: %s",
		bkBizID, verifySlice, result, kt.Rid)
	return result, nil
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

	// get op product remained resource plan.
	_, prodMaxAvailable, err := s.planController.GetProdResRemainPool(cts.Kit, bizOrgRel.OpProductID,
		bizOrgRel.PlanProductID)
	if err != nil {
		logs.Errorf("failed to get op product remained resource plan, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// get meta maps.
	zoneMap, regionAreaMap, _, err := s.getMetaMaps(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get verify resource plan demand needed map, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	obsProject := req.RequireType.ToObsProject()
	regionName := regionAreaMap[req.Region].RegionName
	zoneName := zoneMap[req.Zone]

	prePaidAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypes(cts.Kit, cvmapi.ChargeTypePrePaid, obsProject,
		regionName, zoneName, prodMaxAvailable)
	if err != nil {
		logs.Errorf("failed to get pre paid available device types, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	postPaidByHourAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypes(cts.Kit, cvmapi.ChargeTypePostPaidByHour,
		obsProject, regionName, zoneName, prodMaxAvailable)
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
	regionName, zoneName string, prodRemainMap map[plan.ResPlanPoolKey]int64) ([]ptypes.DeviceTypeAvailable, error) {

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
	deviceTypeAvlMap := cvt.SliceToMap(inPlanAvlDeviceTypes, func(ele ptypes.DeviceTypeAvailable) (string, bool) {
		return ele.DeviceType, ele.Available
	})

	for _, ele := range outPlanAvlDeviceTypes {
		deviceTypeAvlMap[ele.DeviceType] = deviceTypeAvlMap[ele.DeviceType] || ele.Available
	}

	result := cvt.MapToSlice(deviceTypeAvlMap, func(deviceType string, avl bool) ptypes.DeviceTypeAvailable {
		return ptypes.DeviceTypeAvailable{DeviceType: deviceType, Available: avl}
	})

	// sort result, put available of true to the head.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Available
	})

	return result, nil
}

// getPlanTypeAvlDeviceTypes get plan type available device types.
func (s *service) getPlanTypeAvlDeviceTypes(kt *kit.Kit, planType enumor.PlanType, obsProject enumor.ObsProject,
	regionName, zoneName string, prodRemainMap map[plan.ResPlanPoolKey]int64) ([]ptypes.DeviceTypeAvailable, error) {

	// get region and zone all matched device types from mongodb.
	matchedDeviceTypes, err := s.getMatchedDeviceTypesFromMgo(kt, regionName, zoneName)
	if err != nil {
		logs.Errorf("failed to get matched device types, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get available device type map.
	avlDeviceTypeMap := make(map[string]struct{})

	nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
	availableTime := plan.NewAvailableTime(nowDemandYear, nowDemandMonth)

	for key, remain := range prodRemainMap {
		if key.PlanType == planType &&
			key.AvailableTime == availableTime &&
			key.ObsProject == obsProject &&
			key.RegionName == regionName &&
			remain > 0 {

			avlDeviceTypeMap[key.DeviceType] = struct{}{}
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
		available := false
		if _, ok := avlDeviceTypeMap[deviceType]; ok {
			available = true
		}

		result[idx] = ptypes.DeviceTypeAvailable{
			DeviceType: deviceType,
			Available:  available,
		}
	}

	// sort result, put available of true to the head.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Available
	})

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
		"enable_apply": true,
	}

	// zone name may be empty, if it is not empty, supplement it into filter.
	if zoneName != "" {
		mgoFilter["zone"] = zoneNameMap[zoneName]
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

// getChargeTypeAvlDeviceTypesV2 get charge type available device types v2.
func (s *service) getChargeTypeAvlDeviceTypesV2(kt *kit.Kit, chargeType cvmapi.ChargeType,
	req *ptypes.GetCvmChargeTypeDeviceTypeReq, prodRemainMap map[plan.ResPlanPoolKeyV2]map[string]int64) (
	[]ptypes.DeviceTypeAvailable, error) {

	// if charge type is pre paid, get available device types from in plan.
	if chargeType == cvmapi.ChargeTypePrePaid {
		return s.getPlanTypeAvlDeviceTypesV2(kt, enumor.PlanTypeCodeInPlan, req, prodRemainMap)
	}

	// 按量计费只消耗预测外的预测
	result, err := s.getPlanTypeAvlDeviceTypesV2(kt, enumor.PlanTypeCodeOutPlan, req, prodRemainMap)
	if err != nil {
		logs.Errorf("failed to get out plan available device types v2, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// sort result, put available of true to the head.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Available
	})

	return result, nil
}

// getPlanTypeAvlDeviceTypesV2 get plan type available device types v2.
func (s *service) getPlanTypeAvlDeviceTypesV2(kt *kit.Kit, planType enumor.PlanTypeCode,
	req *ptypes.GetCvmChargeTypeDeviceTypeReq, prodRemainMap map[plan.ResPlanPoolKeyV2]map[string]int64) (
	[]ptypes.DeviceTypeAvailable, error) {

	// get region and zone all matched device types from mongodb.
	matchedDeviceTypes, err := s.getMatchedDeviceTypesFromMgoV2(kt, req.Region, req.Zone)
	if err != nil {
		logs.Errorf("failed to get matched device types v2, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}

	// get available device type map.
	avlDeviceTypeMap := make(map[string]struct{})

	nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
	availableTime := plan.NewAvailableTime(nowDemandYear, nowDemandMonth)
	obsProject := req.RequireType.ToObsProject()

	for key, remainCoreMap := range prodRemainMap {
		if key.PlanType == planType && key.AvailableTime == availableTime &&
			key.ObsProject == obsProject && key.BkBizID == req.BkBizID && key.RegionID == req.Region {
			for _, remain := range remainCoreMap {
				if remain > 0 {
					avlDeviceTypeMap[key.DeviceType] = struct{}{}
					break
				}
			}
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
		available := false
		if _, ok := avlDeviceTypeMap[deviceType]; ok {
			available = true
		}

		result[idx] = ptypes.DeviceTypeAvailable{
			DeviceType: deviceType,
			Available:  available,
		}
	}

	// sort result, put available of true to the head.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Available
	})

	return result, nil
}

// getMatchedDeviceTypesFromMgo get matched device types from mongodb.
func (s *service) getMatchedDeviceTypesFromMgoV2(kt *kit.Kit, regionID, zoneID string) ([]string, error) {
	// construct mongodb filter.
	mgoFilter := map[string]interface{}{
		"region":       regionID,
		"enable_apply": true,
	}

	// zone name may be empty, if it is not empty, supplement it into filter.
	if zoneID != "" {
		mgoFilter["zone"] = zoneID
	}

	matchedDeviceTypeInterfaces, err := config.Operation().CvmDevice().FindManyDeviceType(kt.Ctx, mgoFilter)
	if err != nil {
		logs.Errorf("failed to find many device type v2, err: %v, rid: %s", err, kt.Rid)
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

// GetCvmChargeTypeDeviceTypeV2 get cvm charge type device type v2.
func (s *service) GetCvmChargeTypeDeviceTypeV2(cts *rest.Contexts) (any, error) {
	req := new(ptypes.GetCvmChargeTypeDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode get cvm charge type device type v2 request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate get cvm charge type device type v2 request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get biz remained resource plan.
	_, prodMaxAvailable, err := s.planController.GetProdResRemainPoolMatch(cts.Kit, req.BkBizID)
	if err != nil {
		logs.Errorf("failed to get biz remained resource plan v2, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// 包年包月
	prePaidAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypesV2(cts.Kit, cvmapi.ChargeTypePrePaid,
		req, prodMaxAvailable)
	if err != nil {
		logs.Errorf("failed to get pre paid available device types v2, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// 按量计费
	postPaidByHourAvlDeviceTypes, err := s.getChargeTypeAvlDeviceTypesV2(cts.Kit, cvmapi.ChargeTypePostPaidByHour,
		req, prodMaxAvailable)
	if err != nil {
		logs.Errorf("failed to get post paid by hour available device types v2, err: %v, rid: %s", err, cts.Kit.Rid)
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
