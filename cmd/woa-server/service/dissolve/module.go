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
	model "hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateRecycledModule create recycle module
func (s *service) CreateRecycledModule(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleModuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 自研云资源-机房裁撤管理-菜单粒度
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDissolveManage, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	ids, err := s.logics.RecycledModule().Create(cts.Kit, req.Modules)
	if err != nil {
		logs.Errorf("create recycle module failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return model.RecycleModuleCreateResp{IDs: ids}, nil
}

// UpdateRecycledModule update recycle module
func (s *service) UpdateRecycledModule(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleModuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 自研云资源-机房裁撤管理-菜单粒度
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDissolveManage, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	if err := s.logics.RecycledModule().Update(cts.Kit, &req.RecycleModuleTable); err != nil {
		logs.Errorf("update recycle module failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListRecycledModule list recycle module
func (s *service) ListRecycledModule(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleModuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Filter == nil {
		req.Filter = tools.AllExpression()
	}
	data, err := s.logics.RecycledModule().List(cts.Kit,
		&types.ListOption{Fields: req.Field, Filter: req.Filter, Page: req.Page})
	if err != nil {
		logs.Errorf("list recycle module failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return data, nil
}

// DeleteRecycledModule delete recycle module
func (s *service) DeleteRecycledModule(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleModuleDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 自研云资源-机房裁撤管理-菜单粒度
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDissolveManage, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.RecycledModule().Delete(cts.Kit, req.IDs); err != nil {
		logs.Errorf("delete recycle module failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
