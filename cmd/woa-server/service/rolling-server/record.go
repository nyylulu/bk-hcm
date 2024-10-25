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

// Package rollingserver ...
package rollingserver

import (
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListAppliedRecords list applied records.
func (s *service) ListAppliedRecords(cts *rest.Contexts) (any, error) {
	return s.listAppliedRecords(cts, handler.ListResourceAuthRes, meta.RollingServerManage, meta.Find)
}

// ListBizAppliedRecords list biz applied records.
func (s *service) ListBizAppliedRecords(cts *rest.Contexts) (any, error) {
	return s.listAppliedRecords(cts, handler.ListBizAuthRes, meta.Biz, meta.Find)
}

// ListAppliedRecords lists applied records.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_applied_record.md
func (s *service) listAppliedRecords(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	resType meta.ResourceType, action meta.Action) (any, error) {

	req := new(rsproto.RollingAppliedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server applied records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server applied records parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: resType, Action: action, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list rolling server applied records failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list rolling server applied records no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &rsproto.RollingAppliedRecordListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return s.client.DataService().Global.RollingServer.ListAppliedRecord(cts.Kit, listReq)
}

// ListReturnedRecords list returned records.
func (s *service) ListReturnedRecords(cts *rest.Contexts) (any, error) {
	return s.listReturnedRecords(cts, handler.ListResourceAuthRes, meta.RollingServerManage, meta.Find)
}

// ListBizReturnedRecords list biz returned records.
func (s *service) ListBizReturnedRecords(cts *rest.Contexts) (any, error) {
	return s.listReturnedRecords(cts, handler.ListBizAuthRes, meta.Biz, meta.Find)
}

// ListReturnedRecords lists returned records.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_returned_record.md
func (s *service) listReturnedRecords(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	resType meta.ResourceType, action meta.Action) (any, error) {

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server returned records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server returned records parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: resType, Action: action, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list rolling server returned records failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list rolling server returned records no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &rsproto.RollingReturnedRecordListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return s.client.DataService().Global.RollingServer.ListReturnedRecord(cts.Kit, listReq)
}
