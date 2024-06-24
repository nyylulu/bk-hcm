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
	"fmt"
	"strconv"
	"strings"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// getBizIDNameByName 此方法如果没有传name,那么会查询所有的业务
func (l *logics) getBizIDNameByName(kt *kit.Kit, names []string, groupIDs []string) (map[int64]string, error) {
	groupIDMap := make(map[int64]struct{})
	for _, groupID := range groupIDs {
		id, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("group id:%s is invalid, err: %v", groupID, err)
		}

		groupIDMap[id] = struct{}{}
	}

	if len(names) == 0 {
		return l.getAllBizIDName(kt, groupIDMap)
	}

	names = util.StrArrayUnique(names)
	bizIDName := make(map[int64]string, 0)

	req := &cmdb.SearchBizReq{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_oper_grp_name_id"},
		Page:   cmdb.BasePage{Start: 0, Limit: common.BKMaxInstanceLimit},
	}
	start := 0
	end := len(names)
	if len(names) > common.BKMaxInstanceLimit {
		end = common.BKMaxInstanceLimit
	}

	for {
		req.Filter = &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_name",
						Operator: querybuilder.OperatorIn,
						Value:    names[start:end],
					},
				},
			},
		}

		resp, err := l.esbCli.Cmdb().SearchBiz(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		if !resp.Result || resp.Code != 0 {
			return nil, fmt.Errorf("failed to get biz id, err: %s", resp.ErrMsg)
		}

		if resp.Data == nil || len(resp.Data.Info) == 0 {
			return bizIDName, nil
		}

		for _, info := range resp.Data.Info {
			if _, ok := groupIDMap[info.BkOperGrpNameID]; !ok && len(groupIDMap) != 0 {
				continue
			}

			bizIDName[info.BkBizId] = info.BkBizName
		}

		if len(resp.Data.Info) < common.BKMaxInstanceLimit {
			break
		}

		start = end
		if end+common.BKMaxInstanceLimit > len(names) {
			end = len(names)
			continue
		}

		end += common.BKMaxInstanceLimit
	}

	return bizIDName, nil
}

// getBizIDNameByID 此方法如果没有传id,那么会查询所有的业务
func (l *logics) getBizIDNameByID(kt *kit.Kit, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return l.getAllBizIDName(kt, make(map[int64]struct{}))
	}

	ids = util.IntArrayUnique(ids)
	bizIDName := make(map[int64]string, 0)

	req := &cmdb.SearchBizReq{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page:   cmdb.BasePage{Start: 0, Limit: common.BKMaxInstanceLimit},
	}
	start := 0
	end := len(ids)
	if len(ids) > common.BKMaxInstanceLimit {
		end = common.BKMaxInstanceLimit
	}

	for {
		req.Filter = &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_id",
						Operator: querybuilder.OperatorIn,
						Value:    ids[start:end],
					},
				},
			},
		}

		resp, err := l.esbCli.Cmdb().SearchBiz(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		if !resp.Result || resp.Code != 0 {
			return nil, fmt.Errorf("failed to get biz id, err: %s", resp.ErrMsg)
		}

		if resp.Data == nil || len(resp.Data.Info) == 0 {
			return bizIDName, nil
		}

		for _, info := range resp.Data.Info {
			bizIDName[info.BkBizId] = info.BkBizName
		}

		if len(resp.Data.Info) < common.BKMaxInstanceLimit {
			break
		}

		start = end
		if end+common.BKMaxInstanceLimit > len(ids) {
			end = len(ids)
			continue
		}

		end += common.BKMaxInstanceLimit
	}

	return bizIDName, nil
}

func (l *logics) getAllBizIDName(kt *kit.Kit, groupIDMap map[int64]struct{}) (map[int64]string, error) {
	req := &cmdb.SearchBizReq{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_oper_grp_name_id"},
		Page:   cmdb.BasePage{Start: 0, Limit: noLimit},
	}

	resp, err := l.esbCli.Cmdb().SearchBiz(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("call cmdb search business api failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if resp.Data == nil {
		return make(map[int64]string), nil
	}

	bizIDName := make(map[int64]string, 0)
	for _, info := range resp.Data.Info {
		if _, ok := groupIDMap[info.BkOperGrpNameID]; !ok && len(groupIDMap) != 0 {
			continue
		}
		bizIDName[info.BkBizId] = info.BkBizName
	}

	return bizIDName, nil
}

func (l *logics) getBlackBizIDName(kt *kit.Kit) (map[int64]string, error) {
	if len(l.blacklist) == 0 {
		return make(map[int64]string), nil
	}

	bizNames := make([]string, 0)
	for _, v := range strings.Split(l.blacklist, ",") {
		bizNames = append(bizNames, v)
	}

	return l.getBizIDNameByName(kt, bizNames, make([]string, 0))
}

func (l *logics) getHostByIDFromCC(kt *kit.Kit, hostIDs []int64, page *core.BasePage) ([]cmdb.HostInfo, int64, error) {
	req := &cmdb.ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field: common.BKHostIDField, Operator: querybuilder.OperatorIn, Value: hostIDs,
					},
				},
			},
		},
	}

	if !page.Count {
		req.Page = cmdb.BasePage{Start: int(page.Start), Limit: int(page.Limit), Sort: common.BKHostIDField}
		hosts, err := l.getHostFromCC(kt, req)
		if err != nil {
			logs.Errorf("get host from cc failed, err: %v, cond: %+v, rid: %s", err, req, kt.Rid)
			return nil, 0, err
		}

		return hosts, 0, nil
	}

	count, err := l.getHostCountFromCC(kt, req)
	if err != nil {
		logs.Errorf("get host count from cc failed, err: %v, cond: %+v, rid: %s", err, req, kt.Rid)
		return nil, 0, err
	}

	return nil, count, nil
}

const noLimit = 999999999

func (l *logics) getAllHostIDFromCC(kt *kit.Kit, filter *querybuilder.QueryFilter) ([]int64, error) {
	req := &cmdb.ListHostReq{
		Fields:             []string{common.BKHostIDField},
		HostPropertyFilter: filter,
		Page:               cmdb.BasePage{Start: 0, Limit: noLimit, Sort: common.BKHostIDField},
	}
	hosts, err := l.getHostFromCC(kt, req)
	if err != nil {
		logs.Errorf("get host id from cc failed, err: %v, cond: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		hostIDs = append(hostIDs, host.BkHostId)
	}

	return hostIDs, nil
}

func (l *logics) getHostFromCC(kt *kit.Kit, req *cmdb.ListHostReq) ([]cmdb.HostInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("call cc req is nil")
	}

	limit := req.Page.Limit
	if limit > common.BKMaxInstanceLimit {
		req.Page.Limit = common.BKMaxInstanceLimit
	}

	result := make([]cmdb.HostInfo, 0)
	for {
		hosts, err := l.esbCli.Cmdb().ListHost(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, host := range hosts.Data.Info {
			result = append(result, *host)
		}

		// 小于单页大小表示查至最后一页，可返回
		if len(hosts.Data.Info) < req.Page.Limit {
			break
		}

		if len(result) >= limit {
			break
		}

		req.Page.Start += common.BKMaxInstanceLimit

		if req.Page.Start+req.Page.Limit > limit {
			req.Page.Limit = limit - req.Page.Start
		}
	}

	return result, nil
}

func (l *logics) getHostCountFromCC(kt *kit.Kit, req *cmdb.ListHostReq) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("call cc req is nil")
	}

	req.Page = cmdb.BasePage{Start: 0, Limit: 1, Sort: common.BKHostIDField}
	hosts, err := l.esbCli.Cmdb().ListHost(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return 0, err
	}

	return int64(hosts.Data.Count), nil
}

// getHostIDByBizCond 筛选hostIDs, 获取在bizIDName对应的业务中，但是不在blackBizIDName业务的host id结果
func (l *logics) getHostIDByBizCond(kt *kit.Kit, originHostIDs []int64, bizIDName, blackBizIDName map[int64]string) (
	[]int64, map[int64]int64, error) {

	originHostBizIDMap, err := l.esbCli.Cmdb().GetHostBizIds(kt.Ctx, kt.Header(), originHostIDs)
	if err != nil {
		logs.Errorf("get host biz id failed, err: %v, originHostIDs: %v, rid: %s", err, originHostIDs, kt.Rid)
		return nil, nil, err
	}

	bizMap := make(map[int64]struct{})
	for bizID := range bizIDName {
		bizMap[bizID] = struct{}{}
	}

	blackBizIDMap := make(map[int64]struct{})
	for bizID := range blackBizIDName {
		blackBizIDMap[bizID] = struct{}{}
	}

	hostIDs := make([]int64, 0)
	hostBizIDMap := make(map[int64]int64)

	for hostID, bizID := range originHostBizIDMap {
		if _, ok := bizMap[bizID]; !ok {
			continue
		}

		if _, ok := blackBizIDMap[bizID]; ok {
			continue
		}

		hostIDs = append(hostIDs, hostID)
		hostBizIDMap[hostID] = bizID
	}

	return hostIDs, hostBizIDMap, nil
}
