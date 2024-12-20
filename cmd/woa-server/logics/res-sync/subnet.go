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

// Package ressync ...
package ressync

import (
	"fmt"
	"time"

	configTypes "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
)

// SyncSubnet sync subnet
func (l *logics) SyncSubnet() error {
	startTime := time.Now()
	kt := core.NewBackendKit()
	logs.Infof("start to sync subnet, startTime: %v, rid: %s", startTime, kt.Rid)

	zoneCond := mapstr.MapStr{}
	zones, err := l.configLogics.Zone().GetZone(kt, &zoneCond)
	if err != nil {
		logs.Errorf("failed to get all zone list, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("failed to get all zone list, err: %v", err)
	}

	mapRegionZone := make(map[string][]*configTypes.Zone)
	for _, zone := range zones.Info {
		if _, ok := mapRegionZone[zone.Region]; !ok {
			mapRegionZone[zone.Region] = make([]*configTypes.Zone, 0)
		}

		mapRegionZone[zone.Region] = append(mapRegionZone[zone.Region], zone)
	}

	success := 0
	failed := 0
	for region, zoneList := range mapRegionZone {
		vpcCond := mapstr.MapStr{
			"region": region,
		}
		vpcs, err := l.configLogics.Vpc().GetVpc(kt, &vpcCond)
		if err != nil {
			logs.Warnf("failed to get vpc info, region: %s, err: %v, rid: %s", region, err, kt.Rid)
			// continue when error occurs
			continue
		}

		for _, zone := range zoneList {
			for _, vpc := range vpcs.Info {
				reqSync := &configTypes.GetSubnetParam{
					Region: region,
					Zone:   zone.Zone,
					Vpc:    vpc.VpcId,
				}
				if err = l.configLogics.Subnet().SyncSubnet(kt, reqSync); err != nil {
					failed++
					logs.Warnf("failed to sync subnet, region: %s, zone: %s, vpc: %s, err: %v, rid: %s",
						region, zone.Zone, vpc.VpcId, err, kt.Rid)
					// continue when error occurs
					continue
				}
				success++
			}
		}
	}
	endTime := time.Now()
	logs.Infof("end to sync subnet, success: %d, failed: %d, endTime: %v, cost: %fs, rid: %s", success, failed,
		endTime, endTime.Sub(startTime).Seconds(), kt.Rid)

	return nil
}
