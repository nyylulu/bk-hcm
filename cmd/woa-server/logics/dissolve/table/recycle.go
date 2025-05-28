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
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// getAssetIDByModule 查询模块需要进行裁撤的主机固资号，onlyIncomplete参数表示只返回当前阶段为未裁撤的主机
func (l *logics) getAssetIDByModule(kt *kit.Kit, modules []string, onlyIncomplete bool) ([]string, error) {
	result := make([]string, 0)
	if len(modules) == 0 {
		return result, nil
	}

	var err error
	filter := tools.ContainersExpression("module", modules)
	if onlyIncomplete {
		filter, err = tools.And(filter,
			tools.ContainersExpression("abolish_phase", []enumor.AbolishPhase{enumor.Incomplete, enumor.BsiComplete,
				enumor.Retain}))
		if err != nil {
			logs.Errorf("build filter with abolish_phase failed, err: %v, filter: %+v, rid: %s", err, filter, kt.Rid)
			return nil, err
		}
	}
	listExcludedProjectNames := cc.WoaServer().ResDissolve.ListExcludedProjectNames
	if len(listExcludedProjectNames) != 0 {
		filter, err = tools.And(filter, tools.RuleNotIn("project_name", listExcludedProjectNames))
		if err != nil {
			logs.Errorf("build filter with project_name failed, err: %v, filter: %+v, rid: %s", err, filter, kt.Rid)
			return nil, err
		}
	}

	page := &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}
	opt := &types.ListOption{Fields: []string{"asset_id"}, Filter: filter, Page: page}
	for {
		list, err := l.recycledHost.List(kt, opt)
		if err != nil {
			logs.Errorf("list recycle host failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		for _, v := range list.Details {
			result = append(result, cvt.PtrToVal(v.AssetID))
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return result, nil
}
