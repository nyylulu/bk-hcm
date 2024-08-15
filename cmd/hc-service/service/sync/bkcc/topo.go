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

package bkcc

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/slice"
)

func (s *Syncer) listIEGBizIDs(kt *kit.Kit) ([]int64, error) {
	iegRule := &cmdb.AtomRule{
		Field:    "bk_operate_dept_id",
		Operator: "equal",
		Value:    3,
	}

	params := &cmdb.SearchBizParams{
		BizPropertyFilter: &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: []cmdb.Rule{iegRule}}},
		Fields:            []string{"bk_biz_id"},
	}
	resp, err := s.EsbCli.Cmdb().SearchBusiness(kt, params)
	if err != nil {
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	bizIDs := make([]int64, 0)
	for _, biz := range resp.Info {
		bizIDs = append(bizIDs, biz.BizID)
	}

	return bizIDs, nil
}

func (s *Syncer) getHostBizID(kt *kit.Kit, hostIDs []int64) (map[int64]int64, error) {
	if len(hostIDs) == 0 {
		return make(map[int64]int64), nil
	}

	hostBizIDMap := make(map[int64]int64)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		req := &cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := s.EsbCli.Cmdb().FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("fail to find cmdb topo relation, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, relation := range *relationRes {
			hostBizIDMap[relation.HostID] = relation.BizID
		}
	}

	return hostBizIDMap, nil
}
