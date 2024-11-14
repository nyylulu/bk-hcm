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

// Package greenchannel ...
package greenchannel

import (
	greenchannel "hcm/cmd/woa-server/types/green-channel"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCpuCoreSummary get cpu core summary.
func (s *service) GetCpuCoreSummary(cts *rest.Contexts) (any, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.GreenChannel, Action: meta.Find}})
	if err != nil {
		logs.Errorf("get cpu core summary auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(greenchannel.CpuCoreSummaryReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list green channel cpu core summary, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate green channel cpu core summary parameter, err: %v, req: %+v, rid: %s", err, req,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.gcLogics.GetCpuCoreSummary(cts.Kit, req)
}

// GetBizCpuCoreSummary get biz cpu core summary.
func (s *service) GetBizCpuCoreSummary(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Find}, BizID: bizID})
	if err != nil {
		logs.Errorf("get biz cpu core summary auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(greenchannel.CpuCoreSummaryReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list green channel cpu core summary, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate green channel cpu core summary parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req.BkBizIDs = []int64{bizID}
	return s.gcLogics.GetCpuCoreSummary(cts.Kit, req)
}

// ListStatisticalRecord list statistical record.
func (s *service) ListStatisticalRecord(cts *rest.Contexts) (any, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.GreenChannel, Action: meta.Find}})
	if err != nil {
		logs.Errorf("list green channel statistical record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(greenchannel.StatisticalRecordReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list green channel statistical record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate green channel parameter, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.gcLogics.ListStatisticalRecord(cts.Kit, req)
}
