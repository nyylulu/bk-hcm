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

// Package generator implements the generator of task
package generator

import (
	"fmt"

	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/querybuilder"
)

// syncHostByAsset sync host info to cc 3.0 by asset id
func (g *Generator) syncHostByAsset(kt *kit.Kit, assetIds []string) error {
	// once 10 hosts at most
	maxNum := 10
	begin := 0
	end := begin
	len := len(assetIds)

	for begin < len {
		end += maxNum
		if end > len {
			end = len
		}

		req := &cmdb.AddHostReq{
			AssetIDs: assetIds[begin:end],
		}

		err := g.cc.AddHost(kt, req)
		if err != nil {
			logs.Errorf("failed to call cc api to sync host, err: %v", err)
			return fmt.Errorf("failed to call cc api to sync host, err: %v", err)
		}
		begin = end
	}

	return nil
}

// syncHostByIp sync host info to cc 3.0 by ip
func (g *Generator) syncHostByIp(ips []string) error {
	// once 10 hosts at most
	maxNum := 10
	begin := 0
	end := begin
	len := len(ips)

	for begin < len {
		end += maxNum
		if end > len {
			end = len
		}

		req := &cmdb.AddHostReq{
			InnerIps: ips[begin:end],
		}

		kt := core.NewBackendKit()
		err := g.cc.AddHost(kt, req)
		if err != nil {
			logs.Errorf("failed to call cc api to sync host, err: %v", err)
			return fmt.Errorf("failed to call cc api to sync host, err: %v", err)
		}
		begin = end
	}

	return nil
}

// listDeviceTopo list device topology info
func (g *Generator) listDeviceTopo(kt *kit.Kit, ips []string) ([]*cmdb.DeviceTopoInfo, error) {
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					querybuilder.AtomRule{
						Field:    "bk_host_innerip",
						Operator: querybuilder.OperatorIn,
						Value:    ips,
					},
					// support bk_cloud_id 0 only
					cmdb.AtomRule{
						Field:    "bk_cloud_id",
						Operator: cmdb.OperatorEqual,
						Value:    0,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			// 机型
			"svr_device_class",
			"bk_os_name",
			"bk_os_version",
			// idc区域
			"bk_idc_area",
			// 可用区
			"sub_zone",
			"module_name",
			// 机架号，string类型
			"rack_id",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	resp, err := g.cc.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	topoInfos := make([]*cmdb.DeviceTopoInfo, 0)
	for _, host := range resp.Info {
		topo := &cmdb.DeviceTopoInfo{
			InnerIP:      host.GetUniqIp(),
			AssetID:      host.BkAssetID,
			DeviceClass:  host.SvrDeviceClass,
			Raid:         host.RaidName,
			OSName:       host.BkOSName,
			OSVersion:    host.BkOsVersion,
			IdcArea:      host.BkIdcArea,
			SZone:        host.SubZone,
			ModuleName:   host.ModuleName,
			Equipment:    host.RackId,
			IdcLogicArea: host.LogicDomain,
		}
		topoInfos = append(topoInfos, topo)
	}

	return topoInfos, nil
}
