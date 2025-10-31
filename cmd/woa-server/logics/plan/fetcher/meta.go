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

package fetcher

import (
	dmtypes "hcm/pkg/dal/dao/types/meta"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetMetaMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap.
func (f *ResPlanFetcher) GetMetaMaps(kt *kit.Kit) (map[string]string, map[string]dmtypes.RegionArea,
	map[string]wdt.WoaDeviceTypeTable, error) {

	// get zone id name mapping.
	zoneMap, err := f.dao.WoaZone().GetZoneMap(kt)
	if err != nil {
		logs.Errorf("get zone map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get region area mapping.
	regionAreaMap, err := f.dao.WoaZone().GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("get region area map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get device type mapping.
	deviceTypeMap, err := f.deviceTypesMap.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	return zoneMap, regionAreaMap, deviceTypeMap, nil
}

// GetMetaNameMapsFromIDMap get zone name map and region name map from id map.
func (f *ResPlanFetcher) GetMetaNameMapsFromIDMap(zoneMap map[string]string,
	regionAreaMap map[string]dmtypes.RegionArea) (
	map[string]string, map[string]dmtypes.RegionArea) {

	zoneNameMap := make(map[string]string)
	for id, name := range zoneMap {
		zoneNameMap[name] = id
	}
	regionNameMap := make(map[string]dmtypes.RegionArea)
	for _, item := range regionAreaMap {
		regionNameMap[item.RegionName] = item
	}
	return zoneNameMap, regionNameMap
}
