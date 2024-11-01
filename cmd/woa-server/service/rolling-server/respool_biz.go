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

package rollingserver

import (
	"errors"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// DeleteResourcePoolBiz delete resource pool biz.
func (s *service) DeleteResourcePoolBiz(cts *rest.Contexts) (any, error) {
	delID := cts.PathParameter("id").String()
	if delID == "" {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("id can't be empty"))
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("delete resource pool business failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.BatchDeleteReq{
		Filter: tools.EqualExpression("id", delID),
	}
	return nil, s.client.DataService().Global.RollingServer.DeleteResPoolBiz(cts.Kit, delReq)
}

// CreateResourcePoolBiz create resource pool biz.
func (s *service) CreateResourcePoolBiz(cts *rest.Contexts) (any, error) {
	req := new(rsproto.ResourcePoolBusinessCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("create resource pool business failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.client.DataService().Global.RollingServer.BatchCreateResPoolBiz(cts.Kit, req)
}

// ListResourcePoolBiz list all resource pool biz.
func (s *service) ListResourcePoolBiz(cts *rest.Contexts) (any, error) {
	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list resource pool business failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	listRst := new(rsproto.ResourcePoolBusinessListResult)
	listReq := &rsproto.ResourcePoolBusinessListReq{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		},
		Page: core.NewDefaultBasePage(),
	}
	for {
		res, err := s.client.DataService().Global.RollingServer.ListResPoolBiz(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("list resource pool business failed, err: %v, req: %+v, rid: %s", err, *listReq, cts.Kit.Rid)
			return nil, err
		}

		listRst.Details = append(listRst.Details, res.Details...)

		if len(res.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return listRst, nil
}
