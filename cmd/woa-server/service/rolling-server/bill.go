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
	rolling_server "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListBills list bills.
func (s *service) ListBills(cts *rest.Contexts) (any, error) {
	return s.listBills(cts, handler.ListResourceAuthRes, meta.RollingServerManage, meta.Find)
}

// listBills lists bills.
func (s *service) listBills(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	resType meta.ResourceType, action meta.Action) (any, error) {

	req := new(rsproto.RollingBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server bills, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server bills parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: resType, Action: action, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list rolling server bills failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list rolling server bills no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &rsproto.RollingBillListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return s.client.DataService().Global.RollingServer.ListBill(cts.Kit, listReq)
}

// SyncBills sync bills.
func (s *service) SyncBills(cts *rest.Contexts) (any, error) {
	req := new(rolling_server.RollingBillSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to sync rolling server bills, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("validate rolling server bills param failed, err: %v, req: %+v, rid: %s", err, req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{Basic: &meta.Basic{
		Type: meta.RollingServerManage, Action: meta.Find}}); err != nil {
		logs.Errorf("no permission to sync rolling server bill, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.rollingServerLogic.SyncBills(cts.Kit, req); err != nil {
		logs.Errorf("sync bills failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
