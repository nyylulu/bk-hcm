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
	"fmt"

	model "hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/maps"
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

// SyncRecycledHost sync recycle host
func (s *service) SyncRecycledHost(cts *rest.Contexts) (interface{}, error) {
	// 自研云资源-机房裁撤管理-菜单粒度
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDissolveManage, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	if err = s.logics.RecycledHost().Sync(cts.Kit); err != nil {
		logs.Errorf("sync recycle host failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CheckHostDissolveStatus check host dissolve status
func (s *service) CheckHostDissolveStatus(cts *rest.Contexts) (interface{}, error) {
	return s.checkHostDissolveStatus(cts, handler.ListBizAuthRes)
}

func (s *service) checkHostDissolveStatus(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(model.HostDissolveStatusCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authFilter, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: meta.Cvm, Action: meta.Find, Filter: tools.AllExpression()})
	if err != nil {
		return nil, err
	}
	if noPermFlag {
		return nil, errf.NewFromErr(errf.PermissionDenied, fmt.Errorf("no permission"))
	}

	rules := make([]filter.RuleFactory, 0)
	rules = append(rules, tools.RuleEqual("vendor", enumor.TCloudZiyan))
	rules = append(rules, tools.RuleJsonIn("bk_host_id", req.HostIDs))
	rules = append(rules, authFilter)
	listFilter := &filter.Expression{
		Op:    filter.And,
		Rules: rules,
	}
	listReq := &dataproto.CvmListReq{
		Field:  []string{"extension"},
		Filter: listFilter,
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := s.client.DataService().TCloudZiyan.Cvm.ListCvmExt(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("failed to list host by bk_host_id, err: %v, ids: %v, rid: %s", err, req.HostIDs, cts.Kit.Rid)
		return nil, err
	}

	if len(req.HostIDs) != len(rst.Details) {
		logs.Errorf("get host info not match with bk_host_ids, count: %d, host_id count: %d, rid: %s",
			len(rst.Details), len(req.HostIDs), cts.Kit.Rid)
		return nil, fmt.Errorf("host count not match, there could be invalid bk_host_id or no permissions")
	}

	hostIDAssetIDMap := make(map[int64]string)
	for _, host := range rst.Details {
		if host.Extension == nil {
			logs.Errorf("host extension is nil, host: %v, rid: %s", host, cts.Kit.Rid)
			return nil, errf.New(errf.InvalidParameter, "host info is invalid, can not find host asset id")
		}
		hostIDAssetIDMap[host.BkHostID] = host.Extension.BkAssetID
	}

	assetIDStatusMap, err := s.logics.RecycledHost().IsDissolveHost(cts.Kit, maps.Values(hostIDAssetIDMap))
	if err != nil {
		logs.Errorf("check host dissolve status failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	info := make([]model.HostDissolveStatusCheckInfo, 0, len(assetIDStatusMap))
	for _, hostID := range req.HostIDs {
		info = append(info, model.HostDissolveStatusCheckInfo{
			HostID: hostID,
			Status: assetIDStatusMap[hostIDAssetIDMap[hostID]],
		})
	}

	return model.HostDissolveStatusCheckResp{Info: info}, nil
}
