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
	"fmt"
	"sort"
	"time"

	"hcm/cmd/woa-server/model/config"
	ptypes "hcm/cmd/woa-server/types/plan"
	ttypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

// VerifyResPlanDemandV2 verify resource plan demand for subOrders.
func (c *Controller) VerifyResPlanDemandV2(kt *kit.Kit, bkBizID int64, requireType enumor.RequireType,
	subOrders []ttypes.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

	obsProject := requireType.ToObsProject()
	// get all device type maps.
	deviceTypeMap, err := c.GetAllDeviceTypeMap(kt)
	if err != nil {
		logs.Errorf("get all device type map failed, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, err
	}

	result := make([]ptypes.VerifyResPlanDemandElem, len(subOrders))
	indexMap := make(map[int]int)
	verifySlice := make([]VerifyResPlanElemV2, 0)

	for idx, subOrder := range subOrders {
		// if resource type is not cvm, set verify result to not involved.
		if subOrder.ResourceType != ttypes.ResourceTypeCvm {
			result[idx] = ptypes.VerifyResPlanDemandElem{
				VerifyResult: enumor.VerifyResPlanRstNotInvolved,
			}
			continue
		}

		// 是否包年包月
		isPrePaid := true
		if subOrder.Spec.ChargeType.GetWithDefault() != cvmapi.ChargeTypePrePaid {
			isPrePaid = false
		}

		nowDemandYear, nowDemandMonth, err := c.demandTime.GetDemandYearMonth(kt, time.Now())
		if err != nil {
			logs.Errorf("failed to get demand year month, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}
		availableTime := NewAvailableTime(nowDemandYear, nowDemandMonth)

		var cpuCore int64
		if deviceInfo, ok := deviceTypeMap[subOrder.Spec.DeviceType]; ok {
			cpuCore = deviceInfo.CpuCore
		}

		indexMap[len(verifySlice)] = idx
		verifySlice = append(verifySlice, VerifyResPlanElemV2{
			IsPrePaid:     isPrePaid,
			AvailableTime: availableTime,
			DeviceType:    subOrder.Spec.DeviceType,
			ObsProject:    obsProject,
			BkBizID:       bkBizID,
			DemandClass:   enumor.DemandClassCVM,
			RegionID:      subOrder.Spec.Region,
			ZoneID:        subOrder.Spec.Zone,
			DiskType:      subOrder.Spec.DiskType.GetWithDefault(),
			CpuCore:       int64(subOrder.Replicas) * cpuCore,
		})
	}

	// call verify resource plan demands to verify each cvm demands.
	rst, err := c.VerifyProdDemandsV2(kt, bkBizID, requireType, verifySlice)
	if err != nil {
		logs.Errorf("failed to verify resource plan demand v2, err: %v, bkBizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if len(rst) != len(verifySlice) {
		logs.Errorf("verify resource plan demand v2 failed, rst len: %d, verifySlice len: %d, bkBizID: %d, rid: %s",
			len(rst), len(verifySlice), bkBizID, kt.Rid)
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

// GetPlanTypeAvlDeviceTypesV2 get plan type available device types v2.
func (c *Controller) GetPlanTypeAvlDeviceTypesV2(kt *kit.Kit, planType enumor.PlanTypeCode,
	req *ptypes.GetCvmChargeTypeDeviceTypeReq, prodRemainMap map[ResPlanPoolKeyV2]map[string]int64) (
	[]ptypes.DeviceTypeAvailable, error) {

	// get region and zone all matched device types from mongodb.
	matchedDeviceTypes, err := c.getMatchedDeviceTypesFromMgoV2(kt, req.Region, req.Zone)
	if err != nil {
		logs.Errorf("failed to get matched device types v2, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}

	// get available device type map.
	avlDeviceTypeMap, err := c.getProdRemainAvlDeviceTypeMap(kt, req, prodRemainMap, planType)
	if err != nil {
		return nil, err
	}

	// get available device type's matched device type.
	for deviceType := range avlDeviceTypeMap {
		matched, err := c.IsDeviceMatched(kt, matchedDeviceTypes, deviceType)
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

func (c *Controller) getProdRemainAvlDeviceTypeMap(kt *kit.Kit, req *ptypes.GetCvmChargeTypeDeviceTypeReq,
	prodRemainMap map[ResPlanPoolKeyV2]map[string]int64, planType enumor.PlanTypeCode) (map[string]struct{}, error) {

	nowDemandYear, nowDemandMonth, err := c.demandTime.GetDemandYearMonth(kt, time.Now())
	if err != nil {
		logs.Errorf("failed to get demand year month, err: %v, planType: %s, rid: %s", err, planType, kt.Rid)
		return nil, err
	}

	availableTime := NewAvailableTime(nowDemandYear, nowDemandMonth)
	obsProject := req.RequireType.ToObsProject()
	avlDeviceTypeMap := make(map[string]struct{})

	for key, remainCoreMap := range prodRemainMap {
		//  机房裁撤需要忽略预测内、预测外 --story=121848852
		if req.RequireType == enumor.RequireTypeDissolve || key.PlanType == planType {
			avlDeviceTypeMap = getAvlDeviceTypeMap(req, key, remainCoreMap, availableTime, obsProject, avlDeviceTypeMap)
		}
	}
	return avlDeviceTypeMap, nil
}

func getAvlDeviceTypeMap(req *ptypes.GetCvmChargeTypeDeviceTypeReq, key ResPlanPoolKeyV2,
	remainCoreMap map[string]int64, availableTime AvailableTime, obsProject enumor.ObsProject,
	avlDeviceTypeMap map[string]struct{}) map[string]struct{} {

	if key.AvailableTime == availableTime && key.ObsProject == obsProject &&
		key.BkBizID == req.BkBizID && key.RegionID == req.Region {
		for _, remain := range remainCoreMap {
			if remain > 0 {
				avlDeviceTypeMap[key.DeviceType] = struct{}{}
				break
			}
		}
	}
	return avlDeviceTypeMap
}

// getMatchedDeviceTypesFromMgoV2 get matched device types from mongodb.
func (c *Controller) getMatchedDeviceTypesFromMgoV2(kt *kit.Kit, regionID, zoneID string) ([]string, error) {
	// construct mongodb filter.
	mgoFilter := map[string]interface{}{
		"region":       regionID,
		"enable_apply": true,
	}

	// zone name may be empty, if it is not empty, supplement it into filter.
	if zoneID != "" && zoneID != cvmapi.CvmSeparateCampus {
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
