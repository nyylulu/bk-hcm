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

package es

import (
	"fmt"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/querybuilder"
	utils "hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/thirdparty/es"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ES provides interface for operations of es.
type ES interface {
	FindCurHost(kt *kit.Kit, cond map[string][]interface{}, page *core.BasePage) (*dissolve.ListHostDetails, error)
	FindOriginHost(kt *kit.Kit, cond map[string][]interface{}, page *core.BasePage) (*dissolve.ListHostDetails, error)
	ListResDissolveTable(kt *kit.Kit, cond map[string][]interface{}) ([]dissolve.BizDetail, error)
}

type logics struct {
	esbCli     esb.Client
	esCli      *es.EsCli
	originDate string
}

// New create elasticsearch logics.
func New(esbCli esb.Client, esCli *es.EsCli, originDate string) ES {
	return &logics{esbCli: esbCli, esCli: esCli, originDate: originDate}
}

func (l *logics) getOriginHostIndex() string {
	return es.GetIndex(l.originDate)
}

// FindCurHost find current host
func (l *logics) FindCurHost(kt *kit.Kit, cond map[string][]interface{}, page *core.BasePage) (
	*dissolve.ListHostDetails, error) {

	cond, err := l.addCondNotInCCIP(kt, cond)
	if err != nil {
		logs.Errorf("add cond not in cc host ip failed, err: %v, cond: %v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	return l.findHost(kt, cond, es.GetLatestIndex(), page)
}

// FindOriginHost find origin host
func (l *logics) FindOriginHost(kt *kit.Kit, cond map[string][]interface{}, page *core.BasePage) (
	*dissolve.ListHostDetails, error) {

	return l.findHost(kt, cond, l.getOriginHostIndex(), page)
}

func (l *logics) findHost(kt *kit.Kit, cond map[string][]interface{}, index string,
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
	hosts, err := l.esCli.SearchWithCond(kt.Ctx, cond, index, int(page.Start), int(page.Limit), page.Sort)
	if err != nil {
		logs.Errorf("get host by condition failed, err: %v, index: %s, cond: %v, rid: %s", err, index, cond,
			kt.Rid)
		return nil, err
	}

	return &dissolve.ListHostDetails{Details: hosts}, nil
}

// ListResDissolveTable list resource dissolve table
func (l *logics) ListResDissolveTable(kt *kit.Kit, cond map[string][]interface{}) ([]dissolve.BizDetail, error) {
	// 1.获取原始业务和主机数据
	bizMap, err := l.getOriginBizData(kt, cond)
	if err != nil {
		logs.Errorf("get origin business data failed, err: %v, cond: %v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	// 2.补充当前的主机相关数据
	cond, err = l.addCondNotInCCIP(kt, cond)
	if err != nil {
		logs.Errorf("add cond not in cc host ip failed, err: %v, cond: %v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}
	bizMap, err = l.fillCurHostData(kt, cond, bizMap)
	if err != nil {
		logs.Errorf("fill current host data failed, err: %v, cond: %v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	// 3.计算总数以及裁撤进度
	result, err := calculateBizData(bizMap)
	if err != nil {
		logs.Errorf("calculate business data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (l *logics) getOriginBizData(kt *kit.Kit, cond map[string][]interface{}) (map[string]dissolve.BizDetail, error) {
	hosts, err := l.getAllHostFromES(kt, cond, l.getOriginHostIndex())
	if err != nil {
		logs.Errorf("find origin host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	bizMap := make(map[string]dissolve.BizDetail, 0)

	for _, host := range hosts {
		bizDetail, ok := bizMap[host.AppName]
		if !ok {
			bizDetail = dissolve.BizDetail{BizName: host.AppName, ModuleHostCount: make(map[string]int)}
		}

		var hostCount int64 = 1
		if bizDetail.Total.Origin.HostCount != nil {
			hostCount += bizDetail.Total.Origin.HostCount.(int64)
		}
		bizDetail.Total.Origin.HostCount = hostCount
		bizDetail.Total.Origin.CpuCount += host.MaxCPUCoreAmount

		bizMap[host.AppName] = bizDetail
	}

	return bizMap, nil
}

func (l *logics) fillCurHostData(kt *kit.Kit, cond map[string][]interface{}, bizMap map[string]dissolve.BizDetail) (
	map[string]dissolve.BizDetail, error) {

	hosts, err := l.getAllHostFromES(kt, cond, es.GetLatestIndex())
	if err != nil {
		logs.Errorf("find current host failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	for _, host := range hosts {
		bizDetail, ok := bizMap[host.AppName]
		if !ok {
			bizDetail = dissolve.BizDetail{BizName: host.AppName, ModuleHostCount: make(map[string]int)}
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

		bizMap[host.AppName] = bizDetail
	}

	return bizMap, nil
}

// addCondNotInCCIP 由于es中查到的是T-1的数据，不能保证数据的实时性，所以提供此方法，把已经不在cc的ip作为条件，再去查询es
func (l *logics) addCondNotInCCIP(kt *kit.Kit, cond map[string][]interface{}) (map[string][]interface{}, error) {
	hosts, err := l.getAllHostFromES(kt, cond, es.GetLatestIndex())
	if err != nil {
		logs.Errorf("get current host from es failed, err: %v, cond: %+v, rid: %s", err, cond, kt.Rid)
		return nil, err
	}

	ips := make([]string, len(hosts))
	for i, host := range hosts {
		ips[i] = host.InnerIP
	}

	notInCCIPs, err := l.getIPNotInCC(kt, ips)
	if err != nil {
		logs.Errorf("get host not in cc failed, err: %v, ips: %v, rid: %s", err, ips, kt.Rid)
		return nil, err
	}

	for _, ip := range notInCCIPs {
		cond[es.NotInCCIPs] = append(cond[es.NotInCCIPs], ip)
	}

	return cond, nil
}

func (l *logics) getAllHostFromES(kt *kit.Kit, cond map[string][]interface{}, index string) ([]es.Host, error) {
	var pageLimit uint = 1000
	page := &core.BasePage{Start: 0, Limit: pageLimit}

	result := make([]es.Host, 0)
	for {
		hosts, err := l.findHost(kt, cond, index, page)
		if err != nil {
			logs.Errorf("find host failed, err: %v, req: %+v, page: %+v, rid: %s", err, cond, page, kt.Rid)
			return nil, err
		}

		result = append(result, hosts.Details...)

		if len(hosts.Details) < int(pageLimit) {
			break
		}

		page.Start += uint32(pageLimit)
	}

	return result, nil
}

func (l *logics) getIPNotInCC(kt *kit.Kit, ips []string) ([]string, error) {
	result := make([]string, 0)
	if len(ips) == 0 {
		return result, nil
	}

	ips = utils.StrArrayUnique(ips)
	ipMap := make(map[string]struct{})
	for _, ip := range ips {
		ipMap[ip] = struct{}{}
	}

	var start, end int
	for start < len(ips) {
		end = start + common.BKMaxInstanceLimit
		if len(ips)-start < common.BKMaxInstanceLimit {
			end = len(ips)
		}

		req := &cmdb.ListHostReq{
			HostPropertyFilter: &querybuilder.QueryFilter{
				Rule: querybuilder.CombinedRule{
					Condition: querybuilder.ConditionAnd,
					Rules: []querybuilder.Rule{
						querybuilder.AtomRule{
							Field:    common.BKHostInnerIPField,
							Operator: querybuilder.OperatorIn,
							Value:    ips[start:end],
						},
						querybuilder.AtomRule{
							Field:    common.BKCloudIDField,
							Operator: querybuilder.OperatorEqual,
							Value:    0, // 只需要查询管控区域为0的公司的机器
						},
					},
				},
			},
			Fields: []string{common.BKHostInnerIPField},
			Page:   cmdb.BasePage{Start: 0, Limit: common.BKMaxInstanceLimit},
		}

		hosts, err := l.esbCli.Cmdb().ListHost(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, host := range hosts.Data.Info {
			delete(ipMap, host.BkHostInnerIp)
		}

		if len(hosts.Data.Info) < common.BKMaxInstanceLimit {
			break
		}

		start += common.BKMaxInstanceLimit
	}

	for ip := range ipMap {
		result = append(result, ip)
	}

	return result, nil
}

func calculateBizData(bizMap map[string]dissolve.BizDetail) ([]dissolve.BizDetail, error) {
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
