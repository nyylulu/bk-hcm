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
	"time"

	demandtime "hcm/cmd/woa-server/logics/plan/demand-time"
	ptypes "hcm/cmd/woa-server/types/plan"
	ttypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// VerifyResPlanDemandV2 verify resource plan demand for subOrders.
func (c *Controller) VerifyResPlanDemandV2(kt *kit.Kit, bkBizID int64, obsProject enumor.ObsProject,
	subOrders []ttypes.Suborder) ([]ptypes.VerifyResPlanDemandElem, error) {

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

		nowDemandYear, nowDemandMonth := demandtime.GetDemandYearMonth(time.Now())
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
	rst, err := c.VerifyProdDemandsV2(kt, bkBizID, verifySlice)
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
