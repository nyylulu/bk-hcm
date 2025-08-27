/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package detector

import (
	"fmt"
	"strings"

	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
)

// CmdbOperator cmdb operator TODO: 通用方法可以下沉到cc client中
type CmdbOperator interface {
	cmdb.Client
	GetContainerParentIp(kt *kit.Kit, host *cmdb.Host) (string, error)
	GetHostBaseInfoByAsset(kt *kit.Kit, assetIds []string) ([]cmdb.Host, error)
	GetHostBaseInfo(kt *kit.Kit, ips []string) ([]cmdb.Host, error)
	GetHostBaseInfoByID(kt *kit.Kit, hostIDs []int64) ([]cmdb.Host, error)
	GetModuleInfo(kt *kit.Kit, bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error)
	GetBizModuleMap(kt *kit.Kit, bizId int64, moduleIDs []int64) (map[int64]*cmdb.ModuleInfo, error)
}

// cmdbOperator cmdb operator
type cmdbOperator struct {
	cmdb.Client
}

// NewCmdbOperator new cmdb operator for recycle
func NewCmdbOperator(cc cmdb.Client) CmdbOperator {
	return &cmdbOperator{Client: cc}
}

// GetBizModuleMap  ...
func (op *cmdbOperator) GetBizModuleMap(kt *kit.Kit, bizId int64, moduleIDs []int64) (
	map[int64]*cmdb.ModuleInfo, error) {

	batchSize := 200
	moduleIDs = slice.Unique(moduleIDs)
	moduleMap := make(map[int64]*cmdb.ModuleInfo, len(moduleIDs))

	for _, moduleIDBatch := range slice.Split(moduleIDs, batchSize) {
		moduleList, err := op.GetModuleInfo(kt, bizId, moduleIDBatch)
		if err != nil {
			logs.Errorf("fail to get module info for recycle precheck, err: %v, biz: %d, module: %v, rid: %s",
				err, bizId, moduleIDBatch, kt.Rid)
			return nil, err

		}

		for _, module := range moduleList {
			moduleMap[module.BkModuleId] = module
		}
	}

	return moduleMap, nil
}

// GetContainerParentIp ...
func (op *cmdbOperator) GetContainerParentIp(kt *kit.Kit, host *cmdb.Host) (string, error) {
	dashIdx := strings.Index(host.BkAssetID, "-")
	if dashIdx < 0 {
		return "", fmt.Errorf("get docker host assetid failed, ip: %s, asset_id: %s",
			host.BkHostInnerIP, host.BkAssetID)
	}

	parentAssetId := host.BkAssetID[:dashIdx]
	if len(parentAssetId) == 0 {
		return "", fmt.Errorf("get docker host assetid failed, ip: %s", host.BkHostInnerIP)
	}

	assetIds := []string{parentAssetId}
	hostBase, err := op.GetHostBaseInfoByAsset(kt, assetIds)
	if err != nil {
		logs.Errorf("get host by asset %s err: %v, rid: %s", parentAssetId, err, kt.Rid)
		return "", fmt.Errorf("get host by asset %s err: %v", parentAssetId, err)
	}

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("failed to container parent ip, for get invalid host num %d != 1, rid: %s", cnt, kt.Rid)
		return "", fmt.Errorf("failed to container parent ip, for get invalid host num %d != 1", cnt)
	}

	return hostBase[0].GetUniqIp(), nil
}

// GetHostBaseInfoByAsset get host base info in cc 3.0
func (op *cmdbOperator) GetHostBaseInfoByAsset(kt *kit.Kit, assetIds []string) ([]cmdb.Host, error) {
	limit := pkg.BKMaxInstanceLimit
	if len(assetIds) > limit {
		return nil, fmt.Errorf("GetHostBaseInfoByAsset: length of ips %d is greater than %d", len(assetIds), limit)
	}
	// getHostBaseInfo get host base info in cc 3.0
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_asset_id",
						Operator: querybuilder.OperatorIn,
						Value:    assetIds,
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
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
			"operator",
			"bk_bak_operator",
			"srv_status",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: int64(limit),
		},
	}

	resp, err := op.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp.Info, nil

}

// GetHostBaseInfo get host base info in cc 3.0
func (op *cmdbOperator) GetHostBaseInfo(kt *kit.Kit, ips []string) ([]cmdb.Host, error) {
	limit := pkg.BKMaxInstanceLimit
	if len(ips) > limit {
		return nil, fmt.Errorf("GetHostBaseInfo: length of ips %d is greater than %d", len(ips), limit)
	}
	// getHostBaseInfo get host base info in cc 3.0
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_host_innerip",
						Operator: querybuilder.OperatorIn,
						Value:    ips,
					},
					// support bk_cloud_id 0 only
					querybuilder.AtomRule{
						Field:    "bk_cloud_id",
						Operator: querybuilder.OperatorEqual,
						Value:    0,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			"bk_host_outerip_v6",
			// 机型
			"svr_device_class",
			// 逻辑区域
			"logic_domain",
			"bk_zone_name",
			"sub_zone",
			"module_name",
			"raid_name",
			"svr_input_time",
			"operator",
			"bk_bak_operator",
			"srv_status",
			"bk_svr_source_type_id",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: int64(limit),
		},
	}

	resp, err := op.ListHost(kt, req)
	if err != nil {
		logs.Errorf("get host base info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Info, nil
}

// GetHostBaseInfoByID get host base info in cc 3.0
func (op *cmdbOperator) GetHostBaseInfoByID(kt *kit.Kit, hostIDs []int64) ([]cmdb.Host, error) {
	limit := pkg.BKMaxInstanceLimit
	if len(hostIDs) > limit {
		return nil, fmt.Errorf("GetHostBaseInfoByID: length of hostIDs %d is greater than %d", len(hostIDs), limit)
	}
	// getHostBaseInfo get host base info in cc 3.0
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_host_id",
						Operator: querybuilder.OperatorIn,
						Value:    hostIDs,
					},
					// support bk_cloud_id 0 only
					querybuilder.AtomRule{
						Field:    "bk_cloud_id",
						Operator: querybuilder.OperatorEqual,
						Value:    0,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			"bk_host_outerip_v6",
			// 机型
			"svr_device_class",
			// 逻辑区域
			"logic_domain",
			"bk_zone_name",
			"sub_zone",
			"module_name",
			"raid_name",
			"svr_input_time",
			"operator",
			"bk_bak_operator",
			"srv_status",
			"bk_svr_source_type_id",
			"bk_cloud_region",
			"bk_cloud_zone",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	resp, err := op.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Info, nil
}

// GetModuleInfo get module info
func (op *cmdbOperator) GetModuleInfo(kt *kit.Kit, bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	limit := 200
	if len(moduleIds) > limit {
		return nil, fmt.Errorf("GetModuleInfo: length of moduleIds %d is greater than limit %d", len(moduleIds), limit)
	}
	req := &cmdb.SearchModuleParams{
		BizID: bizId,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				pkg.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name", "default"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: int64(limit),
		},
	}

	resp, err := op.SearchModule(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Info, nil
}
