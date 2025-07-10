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
	"errors"

	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListAppliedRecords list applied records.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_applied_record.md
func (s *service) ListAppliedRecords(cts *rest.Contexts) (any, error) {
	req := new(rsproto.RollingAppliedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server applied records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server applied records parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list applied records auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.listAppliedRecords(cts.Kit, req.Filter, req.Page)
}

// ListBizAppliedRecords list biz applied records.
func (s *service) ListBizAppliedRecords(cts *rest.Contexts) (any, error) {
	req := new(rsproto.RollingAppliedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list biz rolling server applied records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate biz rolling server applied records parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: meta.Biz, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list biz rolling server applied records failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list biz rolling server applied records no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	return s.listAppliedRecords(cts.Kit, expr, req.Page)
}

// listAppliedRecords lists applied records.
func (s *service) listAppliedRecords(kt *kit.Kit, filter *filter.Expression, page *core.BasePage) (any, error) {
	listReq := &rsproto.RollingAppliedRecordListReq{
		Filter: filter,
		Page:   page,
	}
	return s.client.DataService().Global.RollingServer.ListAppliedRecord(kt, listReq)
}

// ListReturnedRecords list returned records.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_returned_record.md
func (s *service) ListReturnedRecords(cts *rest.Contexts) (any, error) {
	req := new(rsproto.RollingReturnedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server returned records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server returned records parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list returned records auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.listReturnedRecords(cts.Kit, req.Filter, req.Page)
}

// ListBizReturnedRecords list biz returned records.
func (s *service) ListBizReturnedRecords(cts *rest.Contexts) (any, error) {
	req := new(rsproto.RollingReturnedRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list biz rolling server returned records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate biz rolling server returned records parameter, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: meta.Biz, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list biz rolling server returned records failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list biz rolling server returned records no perm, req: %v, rid: %s",
			cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	return s.listReturnedRecords(cts.Kit, expr, req.Page)
}

// listReturnedRecords lists returned records.
func (s *service) listReturnedRecords(kt *kit.Kit, filter *filter.Expression, page *core.BasePage) (any, error) {
	listReq := &rsproto.RollingReturnedRecordListReq{
		Filter: filter,
		Page:   page,
	}
	return s.client.DataService().Global.RollingServer.ListReturnedRecord(kt, listReq)
}

// UpdateAppliedRecordsNoticeState update applied records notice state.
func (s *service) UpdateAppliedRecordsNoticeState(cts *rest.Contexts) (any, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("update applied records notice state auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.updateAppliedRecordsNoticeState(cts)
}

// UpdateBizAppliedRecordsNoticeState update biz applied records notice state.
func (s *service) UpdateBizAppliedRecordsNoticeState(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	})
	if err != nil {
		logs.Errorf("update biz applied records notice state auth failed, err: %v, bizID: %d, rid: %s", err, bizID,
			cts.Kit.Rid)
		return nil, err
	}

	return s.updateAppliedRecordsNoticeState(cts)
}

func (s *service) updateAppliedRecordsNoticeState(cts *rest.Contexts) (any, error) {
	req := new(rsproto.AppliedRecordUpdateNoticeStateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	state := enumor.RsAppliedRecordNoticeState(cts.PathParameter("state").String())
	if err := state.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	notNotice := cvt.ValToPtr(state.IsNotNotice())

	countReq := rsproto.RollingAppliedRecordListReq{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewCountPage(),
	}
	countResult, err := s.client.DataService().Global.RollingServer.ListAppliedRecord(cts.Kit, &countReq)
	if err != nil {
		logs.Errorf("failed to list applied records, err: %v, req: %+v, rid: %s", err, countReq, cts.Kit.Rid)
		return nil, err
	}
	if countResult.Count != uint64(len(req.IDs)) {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("some applied records not found"))
	}

	appliedRecords := make([]rsproto.RollingAppliedRecordUpdateReq, 0)
	for _, id := range req.IDs {
		appliedRecords = append(appliedRecords, rsproto.RollingAppliedRecordUpdateReq{
			ID:        id,
			NotNotice: notNotice,
		})
	}

	// add audit
	updateLogs := make([]protoaudit.CloudResourceUpdateInfo, 0)
	for _, id := range req.IDs {
		updateFields := map[string]interface{}{"not_notice": notNotice}
		updateLogs = append(updateLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.RsAppliedRecordAuditResType,
			ResID:        id,
			UpdateFields: updateFields,
		})
	}
	auditReq := protoaudit.CloudResourceUpdateAuditReq{Updates: updateLogs}
	err = s.client.DataService().Global.Audit.CloudResourceUpdateAudit(cts.Kit.Ctx, cts.Kit.Header(), &auditReq)
	if err != nil {
		logs.Errorf("failed to add audit, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	updateReq := rsproto.BatchUpdateRollingAppliedRecordReq{
		AppliedRecords: appliedRecords,
	}
	if err = s.client.DataService().Global.RollingServer.UpdateAppliedRecord(cts.Kit, &updateReq); err != nil {
		logs.Errorf("failed to update applied records notice state, err: %v, req: %+v, rid: %s", err, updateReq,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
