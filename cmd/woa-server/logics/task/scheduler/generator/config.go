/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package generator generate task
package generator

import (
	"errors"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/common/util"
	cfgtype "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
)

// getAvailableZoneInfo get available cvm zone info
func (g *Generator) getAvailableZoneInfo(kt *kit.Kit, requireType int64, deviceType, region string) (
	[]*cfgtype.Zone, error) {

	allZones, err := g.getZoneList(kt, region)
	if err != nil {
		return nil, err
	}

	availZoneIds, err := g.getAvailableZoneIds(kt, requireType, deviceType, region)
	if err != nil {
		return nil, err
	}

	availZones := make([]*cfgtype.Zone, 0)
	for _, zone := range allZones {
		for _, zoneId := range availZoneIds {
			if zone.Zone == zoneId {
				availZones = append(availZones, zone)
				break
			}
		}
	}

	return availZones, nil
}

// getAvailableZoneIds get available cvm zone id
func (g *Generator) getAvailableZoneIds(kt *kit.Kit, requireType int64, deviceType, region string) ([]string, error) {
	param := &cfgtype.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					&querybuilder.AtomRule{
						Field:    "require_type",
						Operator: querybuilder.OperatorEqual,
						Value:    requireType,
					},
					&querybuilder.AtomRule{
						Field:    "device_type",
						Operator: querybuilder.OperatorEqual,
						Value:    deviceType,
					},
					&querybuilder.AtomRule{
						Field:    "region",
						Operator: querybuilder.OperatorEqual,
						Value:    region,
					},
				},
			},
		},
	}
	zoneResp, err := g.configLogics.Device().GetDevice(kt, param)
	if err != nil {
		return nil, err
	}

	zoneIds := make([]string, 0)
	for _, device := range zoneResp.Info {
		zoneIds = append(zoneIds, device.Zone)
	}

	zoneIds = util.StrArrayUnique(zoneIds)
	return zoneIds, nil
}

// getZoneList get zone info in certain region
func (g *Generator) getZoneList(kt *kit.Kit, region string) ([]*cfgtype.Zone, error) {
	cond := mapstr.MapStr{}
	// if input region is empty list, return all zone info
	if len(region) > 0 {
		cond["region"] = mapstr.MapStr{
			common.BKDBIN: []string{region},
		}
	}
	zoneResp, err := g.configLogics.Zone().GetZone(kt, &cond)
	if err != nil {
		return nil, err
	}

	return zoneResp.Info, nil
}

// getCapacity get resource apply capacity info
func (g *Generator) getCapacity(kt *kit.Kit, requireType int64, deviceType, region, zone, vpc, subnet string) (
	map[string]int64, error) {

	param := &cfgtype.GetCapacityParam{
		RequireType: requireType,
		DeviceType:  deviceType,
		Region:      region,
		Zone:        zone,
		Vpc:         vpc,
		Subnet:      subnet,
	}

	rst, err := g.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return nil, err
	}

	zoneCapacity := make(map[string]int64)
	for _, capInfo := range rst.Info {
		zoneCapacity[capInfo.Zone] = capInfo.MaxNum
	}

	return zoneCapacity, nil
}

// getCapacityDetail get resource apply capacity detail info
func (g *Generator) getCapacityDetail(kt *kit.Kit, requireType int64, deviceType, region, zone, vpc, subnet string) (
	*cfgtype.CapacityInfo, error) {

	param := &cfgtype.GetCapacityParam{
		RequireType: requireType,
		DeviceType:  deviceType,
		Region:      region,
		Zone:        zone,
		Vpc:         vpc,
		Subnet:      subnet,
	}

	rst, err := g.configLogics.Capacity().GetCapacity(kt, param)
	if err != nil {
		return nil, err
	}

	if len(rst.Info) == 0 {
		return nil, errors.New("get no capacity info")
	}

	return rst.Info[0], nil
}
