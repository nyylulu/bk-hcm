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

package table

import (
	"errors"

	"hcm/cmd/woa-server/thirdparty/es"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (l *logics) findHostFromES(kt *kit.Kit, cond map[string][]interface{}, index string,
	page *core.BasePage) (*dissolve.ListHostDetails, error) {

	if page.Count {
		count, err := l.esCli.CountWithCond(kt.Ctx, cond, index)
		if err != nil {
			logs.Errorf("get host count by condition failed, err: %v, index: %s, cond: %v, rid: %s", err, index, cond,
				kt.Rid)
			return nil, err
		}

		return &dissolve.ListHostDetails{Count: count}, nil
	}

	if page.Sort == "" {
		page.Sort = "_id"
	}

	res := make([]es.Host, 0)
	var pageLimit uint = 10000
	limit := page.Limit
	if limit > pageLimit {
		page.Limit = pageLimit
	}
	for {
		hosts, err := l.esCli.SearchWithCond(kt.Ctx, cond, index, int(page.Start), int(page.Limit), page.Sort)
		if err != nil {
			logs.Errorf("get host by condition failed, err: %v, index: %s, cond: %v, rid: %s", err, index, cond,
				kt.Rid)
			return nil, err
		}

		res = append(res, hosts...)

		if len(hosts) < int(page.Limit) {
			break
		}

		if len(res) >= int(limit) {
			break
		}

		page.Start += uint32(pageLimit)

		if page.Start+uint32(page.Limit) > uint32(limit) {
			page.Limit = limit - uint(page.Start)
		}
	}

	details := make([]dissolve.Host, len(res))
	for i, host := range res {
		detail, err := dissolve.ConvertHost(&host)
		if err != nil {
			logs.Errorf("convert host failed, err: %v, host: %v, rid: %s", err, host, kt.Rid)
			return nil, err
		}

		details[i] = *detail
	}

	return &dissolve.ListHostDetails{Details: details}, nil
}

func (l *logics) fillHostDataByES(kt *kit.Kit, ccHosts []cmdb.HostInfo, esHostMap map[string]dissolve.Host,
	hostBizIDMap map[int64]int64) ([]dissolve.Host, error) {

	bizIDs := make([]int64, 0)
	for _, bizID := range hostBizIDMap {
		bizIDs = append(bizIDs, bizID)
	}
	bizIDName, err := l.getBizIDNameByID(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, ids: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	result := make([]dissolve.Host, len(ccHosts))
	for idx, ccHost := range ccHosts {
		bizID, ok := hostBizIDMap[ccHost.BkHostId]
		if !ok {
			logs.Errorf("can not find biz id, host id: %d, map: %+v,rid: %s", ccHost.BkHostId, hostBizIDMap, kt.Rid)
			return nil, errors.New("host is invalid")
		}

		bizName, ok := bizIDName[bizID]
		if !ok {
			logs.Errorf("can not find biz name, biz id: %d, map: %+v,rid: %s", bizID, bizIDName, kt.Rid)
			return nil, errors.New("biz is invalid")
		}

		esHost, ok := esHostMap[ccHost.BkAssetId]
		if ok {
			esHost.AppName = bizName
			esHost.ModuleName = ccHost.ModuleName
			esHost.BizID = bizID
			esHost.ServerOperator = ccHost.Operator
			esHost.ServerBakOperator = ccHost.BakOperator
			result[idx] = esHost
			continue
		}

		result[idx] = dissolve.Host{
			ServerAssetID:     ccHost.BkAssetId,
			InnerIP:           ccHost.BkHostInnerIP,
			OuterIP:           ccHost.BkHostOuterIP,
			AppName:           bizName,
			BizID:             bizID,
			DeviceType:        ccHost.SvrDeviceClass,
			ModuleName:        ccHost.ModuleName,
			IdcUnitName:       ccHost.IdcUnitName,
			SfwNameVersion:    ccHost.BkOsVersion,
			GoUpDate:          ccHost.SvrInputTime,
			RaidName:          ccHost.RaidName,
			LogicArea:         ccHost.LogicDomain,
			ServerBakOperator: ccHost.BakOperator,
			ServerOperator:    ccHost.Operator,
			DiskTotal:         ccHost.BkDisk,
			MaxCPUCoreAmount:  ccHost.BkCpu,
		}
	}

	return result, nil
}
