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
	"fmt"

	logicshost "hcm/cmd/woa-server/logics/dissolve/host"
	logicsmodule "hcm/cmd/woa-server/logics/dissolve/module"
	"hcm/cmd/woa-server/thirdparty/es"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Table provides interface for operations of dissolve table.
type Table interface {
	FindOriginHost(kt *kit.Kit, req *dissolve.HostListReq) (*dissolve.ListHostDetails, error)
	FindCurHost(kt *kit.Kit, req *dissolve.HostListReq) (*dissolve.ListHostDetails, error)
	ListResDissolveTable(kt *kit.Kit, req *dissolve.ResDissolveReq) ([]dissolve.BizDetail, error)
}

type logics struct {
	recycledHost   logicshost.RecycledHost
	recycledModule logicsmodule.RecycledModule
	esbCli         esb.Client
	esCli          *es.EsCli
	originDate     string
	blacklist      string
}

// New create resource dissolve table logics.
func New(recycledModule logicsmodule.RecycledModule, recycledHost logicshost.RecycledHost, esbCli esb.Client,
	esCli *es.EsCli, originDate string, blacklist string) Table {

	return &logics{
		recycledHost:   recycledHost,
		recycledModule: recycledModule,
		esbCli:         esbCli,
		esCli:          esCli,
		originDate:     originDate,
		blacklist:      blacklist,
	}
}

// FindOriginHost find origin host
func (l *logics) FindOriginHost(kt *kit.Kit, req *dissolve.HostListReq) (
	*dissolve.ListHostDetails, error) {

	moduleAssetIDMap, err := l.getAssetIDByModule(kt, req.ModuleNames)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	// 注意：请求参数中可以不传业务条件，所以这里的bizIDName可能会是空
	bizIDName, err := l.getBizIDNameByName(kt, req.BizNames, make([]string, 0))
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	blackBizIDName, err := l.getBlackBizIDName(kt)
	if err != nil {
		logs.Errorf("get black biz ids failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	cond, err := req.GetESCond(moduleAssetIDMap, bizIDName, blackBizIDName)
	if err != nil {
		logs.Errorf("get es cond failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	res, err := l.findHostFromES(kt, cond, l.getOriginHostIndex(), req.Page)
	if err != nil {
		logs.Errorf("find host from es failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return res, nil
	}

	originHosts := res.Details
	res.Details, err = l.getCurBizName(kt, originHosts)
	if err != nil {
		logs.Errorf("get host current biz name failed, err: %v, host: %+v, rid: %s", err, originHosts, kt.Rid)
		return nil, err
	}

	return res, nil
}

func (l *logics) getOriginHostIndex() string {
	return es.GetIndex(l.originDate)
}

func (l *logics) getCurBizName(kt *kit.Kit, hosts []dissolve.Host) ([]dissolve.Host, error) {
	if len(hosts) == 0 {
		return make([]dissolve.Host, 0), nil
	}

	hostBizIDMap := make(map[string]int64)
	bizIDs := make([]int64, 0)
	for _, host := range hosts {
		hostBizIDMap[host.ServerAssetID] = host.BizID
		bizIDs = append(bizIDs, host.BizID)
	}

	bizIDName, err := l.getBizIDNameByID(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, ids: %v, rid: %s", err, bizIDs, kt.Rid)
		return nil, err
	}

	for i, host := range hosts {
		bizID, ok := hostBizIDMap[host.ServerAssetID]
		if !ok {
			logs.Errorf("can not find biz id, host asset id: %s, map: %+v, rid: %s", host.ServerAssetID, hostBizIDMap,
				kt.Rid)
			return nil, errors.New("host is invalid")
		}

		bizName, ok := bizIDName[bizID]
		if !ok {
			logs.Errorf("can not find biz name, biz id: %d, map: %+v,rid: %s", bizID, bizIDName, kt.Rid)
			return nil, errors.New("biz is invalid")
		}

		hosts[i].AppName = bizName
	}

	return hosts, nil
}

// FindCurHost find current host
func (l *logics) FindCurHost(kt *kit.Kit, req *dissolve.HostListReq) (
	*dissolve.ListHostDetails, error) {

	moduleAssetIDMap, err := l.getAssetIDByModule(kt, req.ModuleNames)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	// 1. 由于有些条件值不在cc的主机字段上，所以先根据主机上有的字段，查出host id
	cond := req.GetCCHostCond(moduleAssetIDMap)
	originHostIDs, err := l.getAllHostIDFromCC(kt, cond)
	if err != nil {
		logs.Errorf("get host id from cc failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}
	if len(originHostIDs) == 0 {
		return &dissolve.ListHostDetails{Details: []dissolve.Host{}}, nil
	}

	// 2. 根据业务条件筛选host id
	bizIDName, err := l.getBizIDNameByName(kt, req.BizNames, req.GroupIDs)
	if err != nil {
		logs.Errorf("get biz id and name failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	blackBizIDName, err := l.getBlackBizIDName(kt)
	if err != nil {
		logs.Errorf("get black biz ids failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	hostIDs, hostBizIDMap, err := l.getHostIDByBizCond(kt, originHostIDs, bizIDName, blackBizIDName)
	if err != nil {
		logs.Errorf("get host id by biz cond failed, err: %v, hostIDs: %v, bizIDName: %v, blackBizIDName: %v, rid: %s",
			err, originHostIDs, bizIDName, blackBizIDName, kt.Rid)
		return nil, err
	}

	// 3. 根据过滤出来的hostIDs查询cc中的主机
	hosts, count, err := l.getHostByIDFromCC(kt, hostIDs, req.Page)
	if err != nil {
		logs.Errorf("get host from cc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &dissolve.ListHostDetails{Count: count}, nil
	}

	// 4. 从es中填充主机缺少的字段
	data, err := l.fillHostDataByES(kt, hosts, hostBizIDMap)
	if err != nil {
		logs.Errorf("fill host data by es failed, err: %v, host: %+v, rid: %s", err, hosts, kt.Rid)
		return nil, err
	}

	return &dissolve.ListHostDetails{Details: data}, nil
}

// ListResDissolveTable list resource dissolve table
func (l *logics) ListResDissolveTable(kt *kit.Kit, req *dissolve.ResDissolveReq) ([]dissolve.BizDetail, error) {
	// 1.获取原始业务和主机数据
	bizMap, err := l.getOriginBizData(kt, req)
	if err != nil {
		logs.Errorf("get origin business data failed, err: %v, cond: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	// 2.补充当前的主机相关数据
	bizMap, err = l.fillCurHostData(kt, req, bizMap)
	if err != nil {
		logs.Errorf("fill current host data failed, err: %v, cond: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if len(bizMap) == 0 {
		return make([]dissolve.BizDetail, 0), nil
	}

	// 3.计算总数以及裁撤进度
	result, err := calculateBizData(bizMap)
	if err != nil {
		logs.Errorf("calculate business data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (l *logics) getOriginBizData(kt *kit.Kit, cond *dissolve.ResDissolveReq) (map[int64]dissolve.BizDetail, error) {
	req := &dissolve.HostListReq{ResDissolveReq: *cond, Page: &core.BasePage{Limit: noLimit}}
	res, err := l.FindOriginHost(kt, req)
	if err != nil {
		logs.Errorf("find origin host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	bizMap := make(map[int64]dissolve.BizDetail, 0)
	for _, host := range res.Details {
		bizDetail, ok := bizMap[host.BizID]
		if !ok {
			bizDetail = dissolve.BizDetail{
				BizID: host.BizID, BizName: host.AppName, ModuleHostCount: make(map[string]int),
			}
		}

		var hostCount int64 = 1
		if bizDetail.Total.Origin.HostCount != nil {
			hostCount += bizDetail.Total.Origin.HostCount.(int64)
		}
		bizDetail.Total.Origin.HostCount = hostCount
		bizDetail.Total.Origin.CpuCount += host.MaxCPUCoreAmount

		bizMap[host.BizID] = bizDetail
	}

	return bizMap, nil
}

func (l *logics) fillCurHostData(kt *kit.Kit, cond *dissolve.ResDissolveReq, bizMap map[int64]dissolve.BizDetail) (
	map[int64]dissolve.BizDetail, error) {

	req := &dissolve.HostListReq{ResDissolveReq: *cond, Page: &core.BasePage{Limit: noLimit}}
	res, err := l.FindCurHost(kt, req)
	if err != nil {
		logs.Errorf("find current host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	for _, host := range res.Details {
		bizDetail, ok := bizMap[host.BizID]
		if !ok {
			bizDetail = dissolve.BizDetail{
				BizID: host.BizID, BizName: host.AppName, ModuleHostCount: make(map[string]int),
			}
		}

		var hostCount int64 = 1
		if bizDetail.Total.Current.HostCount != nil {
			hostCount += bizDetail.Total.Current.HostCount.(int64)
		}
		bizDetail.Total.Current.HostCount = hostCount
		bizDetail.Total.Current.CpuCount += host.MaxCPUCoreAmount

		if _, ok = bizDetail.ModuleHostCount[host.ModuleName]; !ok {
			bizDetail.ModuleHostCount[host.ModuleName] = 0
		}
		bizDetail.ModuleHostCount[host.ModuleName]++

		bizMap[host.BizID] = bizDetail
	}

	return bizMap, nil
}

func calculateBizData(bizMap map[int64]dissolve.BizDetail) ([]dissolve.BizDetail, error) {
	result := make([]dissolve.BizDetail, 0)
	var curHostNum, originHostNum, curCpuNum, originCpuNum int64
	moduleHostCount := make(map[string]int)

	for _, data := range bizMap {
		var curBizHostNum, originBizHostNum int64
		if data.Total.Current.HostCount != nil {
			curBizHostNum = data.Total.Current.HostCount.(int64)
		}
		if data.Total.Origin.HostCount != nil {
			originBizHostNum = data.Total.Origin.HostCount.(int64)
		}

		curHostNum += curBizHostNum
		originHostNum += originBizHostNum
		curCpuNum += data.Total.Current.CpuCount
		originCpuNum += data.Total.Origin.CpuCount

		for module, count := range data.ModuleHostCount {
			if _, ok := moduleHostCount[module]; !ok {
				moduleHostCount[module] = count
				continue
			}

			moduleHostCount[module] += count
		}

		data.Progress = getProgress(originBizHostNum, curBizHostNum)
		result = append(result, data)
	}

	total := dissolve.BizDetail{
		BizID:           "total",
		BizName:         "总数",
		ModuleHostCount: moduleHostCount,
		Total: dissolve.Total{
			Origin:  dissolve.TotalData{HostCount: originHostNum, CpuCount: originCpuNum},
			Current: dissolve.TotalData{HostCount: curHostNum, CpuCount: curCpuNum},
		},
	}
	result = append(result, total)

	val := getProgress(originHostNum, curHostNum)
	progress := dissolve.BizDetail{
		BizID:   "recycle-progress",
		BizName: "裁撤进度",
		Total: dissolve.Total{
			Origin:  dissolve.TotalData{HostCount: val},
			Current: dissolve.TotalData{HostCount: val},
		},
		Progress: val,
	}
	result = append(result, progress)

	return result, nil
}

func getProgress(origin, current int64) string {
	if origin == 0 {
		return ""
	}

	return fmt.Sprintf("%.2f%%", (float64(origin-current))/float64(origin)*100)
}
