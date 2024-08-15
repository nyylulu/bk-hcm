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

// Package detector ...
package detector

import (
	"encoding/json"
	"fmt"
	"strings"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/pkg/logs"

	"go.uber.org/ratelimit"
)

var ccLimiter ratelimit.Limiter
var cvmLimiter ratelimit.Limiter

func init() {
	// set CC ratelimit to 100
	ccLimiter = ratelimit.New(100)
	// set cvm ratelimit to 50
	cvmLimiter = ratelimit.New(50)
}

func (d *Detector) isTcDevice(host *cmdb.HostInfo) bool {
	return strings.HasPrefix(host.BkAssetId, "TC")
}

func (d *Detector) isDockerVM(host *cmdb.HostInfo) bool {
	dashIdx := strings.Index(host.BkAssetId, "-")
	if dashIdx < 0 {
		return false
	}

	if !strings.HasPrefix(host.BkAssetId[dashIdx+1:], "VM") {
		return false
	}

	if !strings.HasPrefix(host.SvrDeviceClass, "D") {
		return false
	}

	return true
}

func (d *Detector) getContainerParentIp(host *cmdb.HostInfo) (string, error) {
	dashIdx := strings.Index(host.BkAssetId, "-")
	if dashIdx < 0 {
		return "", fmt.Errorf("get docker host assetid failed, ip: %s", host.BkHostInnerIp)
	}

	parentAssetId := host.BkAssetId[:dashIdx]
	if len(parentAssetId) == 0 {
		return "", fmt.Errorf("get docker host assetid failed, ip: %s", host.BkHostInnerIp)
	}

	assetIds := []string{parentAssetId}
	hostBase, err := d.getHostBaseInfoByAsset(assetIds)
	if err != nil {
		logs.Errorf("get host by asset %s err: %v", parentAssetId, err)
		return "", fmt.Errorf("get host by asset %s err: %v", parentAssetId, err)
	}

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("failed to check log4j, for get invalid host num %d != 1", cnt)
		return "", fmt.Errorf("failed to check log4j, for get invalid host num %d != 1", cnt)
	}

	return hostBase[0].GetUniqIp(), nil
}

// getHostBaseInfo get host base info in cc 3.0
func (d *Detector) getHostBaseInfo(ips []string) ([]*cmdb.HostInfo, error) {
	// getHostBaseInfo get host base info in cc 3.0
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
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
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: common.BKMaxInstanceLimit,
		},
	}

	// set rate limit to avoid cc api error "API rate limit exceeded by stage/resource strategy"
	ccLimiter.Take()
	resp, err := d.cc.ListHost(nil, nil, req)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:getHostBaseInfo:failed, failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("recycler:logics:cvm:getHostBaseInfo:failed, failed to get cc host info, code: %d, msg: %s",
			resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// getHostBaseInfo get host base info in cc 3.0
func (d *Detector) getHostBaseInfoByAsset(assetIds []string) ([]*cmdb.HostInfo, error) {
	// getHostBaseInfo get host base info in cc 3.0
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
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
			Limit: common.BKMaxInstanceLimit,
		},
	}

	// set rate limit to avoid cc api error "API rate limit exceeded by stage/resource strategy"
	ccLimiter.Take()
	resp, err := d.cc.ListHost(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// getHostTopoInfo get host topo info in cc 3.0
func (d *Detector) getHostTopoInfo(hostIds []int64) ([]*cmdb.HostBizRel, error) {
	req := &cmdb.HostBizRelReq{
		BkHostId: hostIds,
	}

	// set rate limit to avoid cc api error "API rate limit exceeded by stage/resource strategy"
	ccLimiter.Take()
	resp, err := d.cc.FindHostBizRelation(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host topo info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host topo info, err: %s", resp.ErrMsg)
	}

	return resp.Data, nil
}

// getModuleInfo get module info in cc 3.0
func (d *Detector) getModuleInfo(bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleReq{
		BkBizId: bizId,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				common.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	// set rate limit to avoid cc api error "API rate limit exceeded by stage/resource strategy"
	ccLimiter.Take()
	resp, err := d.cc.SearchModule(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc module info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc module info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

func (d *Detector) structToStr(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		logs.Warnf("failed to convert struct to string: %+v", v)
		return ""
	}

	return string(b)
}
