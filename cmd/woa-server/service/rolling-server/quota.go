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
	"fmt"

	rstypes "hcm/cmd/woa-server/types/rolling-server"
	dataproto "hcm/pkg/api/data-service"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// AdjustQuotaOffsets adjust rolling quota offset configs
func (s *service) AdjustQuotaOffsets(cts *rest.Contexts) (any, error) {
	req := new(rstypes.AdjustQuotaOffsetsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to adjust quota offset configs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate adjust quota offset configs parameter, err: %v, req: %v, rid: %s", err,
			*req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// adjust也使用平台管理权限
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Update}})
	if err != nil {
		logs.Errorf("adjust quota offset configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 计算调整值
	var quotaOffset int64
	switch req.AdjustType {
	case enumor.IncreaseOffsetAdjustType:
		quotaOffset = req.QuotaOffset
	case enumor.DecreaseOffsetAdjustType:
		quotaOffset = -req.QuotaOffset
	default:
		return nil, fmt.Errorf("unsupported adjust type: %s", req.AdjustType)
	}

	effectIDs, err := s.rollingServerLogic.AdjustQuotaOffsetConfigs(cts.Kit, req.BkBizIDs, req.AdjustMonth, quotaOffset)
	if err != nil {
		return nil, err
	}

	// 审计记录
	err = s.rollingServerLogic.BatchCreateQuotaOffsetConfigAudit(cts.Kit, effectIDs.IDs, quotaOffset)
	if err != nil {
		logs.Warnf("failed to create quota offset config audit, err: %v, user: %s, app_code: %s, rid: %s", err,
			cts.Kit.User, cts.Kit.AppCode, cts.Kit.Rid)
	}

	return effectIDs, nil
}

// CreateBizQuotaConfigs create biz quota configs
func (s *service) CreateBizQuotaConfigs(cts *rest.Contexts) (any, error) {
	req := new(rstypes.CreateBizQuotaConfigsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create biz quota configs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create biz quota configs parameter, err: %v, req: %v, rid: %s", err, *req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create也使用平台管理权限
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Create}})
	if err != nil {
		logs.Errorf("create biz quota configs failed, err: %v, user: %s, app_code: %s, rid: %s", err,
			cts.Kit.User, cts.Kit.AppCode, cts.Kit.Rid)
		return nil, err
	}

	// 请求不带业务ID时，根据全局配置表尝试为所有业务创建基础配额
	if len(req.BkBizIDs) == 0 {
		return s.rollingServerLogic.CreateBizQuotaConfigsForAllBiz(cts.Kit, req.QuotaMonth)
	}
	return s.rollingServerLogic.CreateBizQuotaConfigs(cts.Kit, req)
}

// DeleteGlobalQuotaConfig delete global quota config
func (s *service) DeleteGlobalQuotaConfig(cts *rest.Contexts) (any, error) {
	delID := cts.PathParameter("id").String()
	if delID == "" {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("id can't be empty"))
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("delete global quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.BatchDeleteReq{
		Filter: tools.EqualExpression("id", delID),
	}
	return nil, s.client.DataService().Global.RollingServer.DeleteGlobalConfig(cts.Kit, delReq)
}

// CreateGlobalQuotaConfigs create global quota configs
func (s *service) CreateGlobalQuotaConfigs(cts *rest.Contexts) (any, error) {
	createReq := new(rsproto.RollingGlobalConfigCreateReq)
	if err := cts.DecodeInto(createReq); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := createReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("create global quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.client.DataService().Global.RollingServer.BatchCreateGlobalConfig(cts.Kit, createReq)
}

// GetGlobalQuotaConfigs get global quota configs.
func (s *service) GetGlobalQuotaConfigs(cts *rest.Contexts) (any, error) {
	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("get global quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	listOne, err := s.rollingServerLogic.GetGlobalQuotaConfig(cts.Kit)
	if err != nil {
		logs.Errorf("get global quota config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst := rstypes.GetGlobalQuotaConfigResp{
		ID:          listOne.ID,
		GlobalQuota: cvt.PtrToVal(listOne.GlobalQuota),
		BizQuota:    cvt.PtrToVal(listOne.BizQuota),
		UnitPrice:   cvt.PtrToVal(listOne.UnitPrice),
		Creator:     listOne.Creator,
		Reviser:     listOne.Reviser,
		CreatedAt:   listOne.CreatedAt,
		UpdatedAt:   listOne.UpdatedAt,
	}

	return rst, nil
}

// ListBizsWithExistQuota list biz with exist quota.
func (s *service) ListBizsWithExistQuota(cts *rest.Contexts) (any, error) {
	req := new(rstypes.ListBizsWithExistQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list businesses with exist quota, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate businesses with exist quota parameter, err: %v, req: %v, rid: %s", err,
			*req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list businesses with exist quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.rollingServerLogic.ListBizsWithExistQuota(cts.Kit, req)
}

// ListBizQuotaConfigs list biz quota configs.
func (s *service) ListBizQuotaConfigs(cts *rest.Contexts) (any, error) {
	req := new(rstypes.ListBizQuotaConfigsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list biz quota configs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate biz quota configs parameter, err: %v, req: %v, rid: %s", err, *req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list biz quota configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.rollingServerLogic.ListBizQuotaConfigs(cts.Kit, req)
}

// ListQuotaOffsetsAdjustRecords list quota offsets adjust records.
func (s *service) ListQuotaOffsetsAdjustRecords(cts *rest.Contexts) (any, error) {
	req := new(rstypes.ListQuotaOffsetsAdjustRecordsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list quota offsets adjust records, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate quota offsets adjust records parameter, err: %v, req: %v, rid: %s", err, *req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorized
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list quota offsets adjust records failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.rollingServerLogic.ListQuotaOffsetAdjustRecords(cts.Kit, req.OffsetConfigIds, req.Page)
}
