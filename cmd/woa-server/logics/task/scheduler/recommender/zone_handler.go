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

// Package recommender ...
package recommender

import (
	"hcm/cmd/woa-server/logics/config"
	cfgtype "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// ZoneHandler apply order zone modification recommend handler
type ZoneHandler struct {
	handler      Handler
	cvm          cvmapi.CVMClientInterface
	configLogics config.Logics
}

// Handle make modification recommendation of apply order
func (zh *ZoneHandler) Handle(order *types.ApplyOrder) (*Recommendation, bool) {
	kt := kit.New()
	// cvm_separate_campus cannot be modified
	if order.Spec.Zone == "" || order.Spec.Zone == cvmapi.CvmSeparateCampus {
		return nil, false
	}

	//  get available zones
	requireType := order.RequireType
	// 小额绿通均使用常规项目的机型
	if requireType == enumor.RequireTypeGreenChannel {
		requireType = enumor.RequireTypeRegular
	}

	// 3. get capacity
	zoneCapacity, err := zh.getCapacity(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region,
		cvmapi.CvmSeparateCampus, order.Spec.ChargeType)
	if err != nil {
		logs.Errorf("failed to get zone capacity err: %v, order id: %s", err, order.SubOrderId)
		return nil, false
	}
	logs.Infof("zone capacity: %+v, order id: %s", zoneCapacity, order.SubOrderId)

	maxNum := int64(0)
	bestChoice := ""
	for zone, capacity := range zoneCapacity {
		if capacity > maxNum {
			maxNum = capacity
			bestChoice = zone
		}
	}

	if maxNum < int64(order.PendingNum) {
		return nil, false
	}

	rst := &Recommendation{
		Zone:       bestChoice,
		DeviceType: order.Spec.DeviceType,
	}

	return rst, true
}

// getCapacity get resource apply capacity info
func (zh *ZoneHandler) getCapacity(kt *kit.Kit, requireType enumor.RequireType, deviceType, region, zone string,
	chargeType cvmapi.ChargeType) (map[string]int64, error) {

	param := &cfgtype.GetCapacityParam{
		RequireType: requireType,
		DeviceType:  deviceType,
		Region:      region,
		Zone:        zone,
	}
	// 计费模式,默认包年包月
	if len(chargeType) > 0 {
		param.ChargeType = chargeType
	}

	rst, err := zh.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return nil, err
	}

	zoneCapacity := make(map[string]int64)
	for _, capInfo := range rst.Info {
		zoneCapacity[capInfo.Zone] = capInfo.MaxNum
	}

	return zoneCapacity, nil
}
