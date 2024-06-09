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

package dissolve

import (
	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/dissolve/module"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListOriginHost list resource dissolve origin host
func (s *service) ListOriginHost(cts *rest.Contexts) (interface{}, error) {
	req := new(dissolve.HostListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	moduleAssetIDMap, err := s.getAssetIDByModule(cts.Kit, req.ModuleNames)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	result, err := s.logics.ES().FindOriginHost(cts.Kit, req.GetESCond(moduleAssetIDMap), req.Page)
	if err != nil {
		logs.Errorf("find origin host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListCurHost list resource dissolve current host
func (s *service) ListCurHost(cts *rest.Contexts) (interface{}, error) {
	req := new(dissolve.HostListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	moduleAssetIDMap, err := s.getAssetIDByModule(cts.Kit, req.ModuleNames)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	result, err := s.logics.ES().FindCurHost(cts.Kit, req.GetESCond(moduleAssetIDMap), req.Page)
	if err != nil {
		logs.Errorf("find current host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListResDissolveTable list resource dissolve table
func (s *service) ListResDissolveTable(cts *rest.Contexts) (interface{}, error) {
	req := new(dissolve.ResDissolveReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	moduleAssetIDMap, err := s.getAssetIDByModule(cts.Kit, req.ModuleNames)
	if err != nil {
		logs.Errorf("get host asset id by module name failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	table, err := s.logics.ES().ListResDissolveTable(cts.Kit, req.GetESCond(moduleAssetIDMap))
	if err != nil {
		logs.Errorf("list resource dissolve table failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return dissolve.ResDissolveTable{Items: table}, nil
}

// getAssetIDByModule 当模块的裁撤状态为为「部分裁」时，返回它所需要进行裁撤主机的固资产号
func (s *service) getAssetIDByModule(kt *kit.Kit, modules []string) (map[string][]string, error) {
	result := make(map[string][]string)
	if len(modules) == 0 {
		return result, nil
	}

	filter, err := tools.And(
		tools.ContainersExpression("name", modules),
		tools.EqualExpression("recycle_type", int(module.Part)),
	)
	if err != nil {
		return nil, err
	}
	page := &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}
	opt := &types.ListOption{Fields: []string{"name"}, Filter: filter, Page: page}

	partTypeModule := make([]string, 0)
	for {
		list, err := s.logics.RecycledModule().List(kt, opt)
		if err != nil {
			logs.Errorf("list recycle module failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		for _, v := range list.Details {
			partTypeModule = append(partTypeModule, *v.Name)
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	if len(partTypeModule) == 0 {
		return result, nil
	}

	for _, name := range partTypeModule {
		result[name] = make([]string, 0)
	}

	filter = tools.ContainersExpression("module", partTypeModule)
	page = &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}
	opt = &types.ListOption{Fields: []string{"module", "asset_id"}, Filter: filter, Page: page}
	for {
		list, err := s.logics.RecycledHost().List(kt, opt)
		if err != nil {
			logs.Errorf("list recycle host failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		for _, v := range list.Details {
			result[*v.Module] = append(result[*v.Module], *v.AssetID)
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return result, nil
}
