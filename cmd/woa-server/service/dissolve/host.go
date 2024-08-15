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

// CreateRecycledHost create recycle host
func (s *service) CreateRecycledHost(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleHostCreateReq)
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

	ids, err := s.logics.RecycledHost().Create(cts.Kit, req.Hosts)
	if err != nil {
		logs.Errorf("create recycle host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return model.RecycleHostCreateResp{IDs: ids}, nil
}

// UpdateRecycledHost update recycle host
func (s *service) UpdateRecycledHost(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleHostUpdateReq)
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

	if err := s.logics.RecycledHost().Update(cts.Kit, &req.RecycleHostTable); err != nil {
		logs.Errorf("update recycle host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListRecycledHost list recycle host
func (s *service) ListRecycledHost(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleHostListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Filter == nil {
		req.Filter = tools.AllExpression()
	}
	data, err := s.logics.RecycledHost().List(cts.Kit,
		&types.ListOption{Fields: req.Field, Filter: req.Filter, Page: req.Page})
	if err != nil {
		logs.Errorf("list recycle host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return data, nil
}

// DeleteRecycledHost delete recycle host
func (s *service) DeleteRecycledHost(cts *rest.Contexts) (interface{}, error) {
	req := new(model.RecycleHostDeleteReq)
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

	if err := s.logics.RecycledHost().Delete(cts.Kit, req.IDs); err != nil {
		logs.Errorf("delete recycle host failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
