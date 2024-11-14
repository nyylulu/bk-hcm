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
	rstypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCpuCoreSummary get cpu core summary.
func (s *service) GetCpuCoreSummary(cts *rest.Contexts) (any, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("get cpu core summary auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(rstypes.CpuCoreSummaryReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server cpu core summary, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server cpu core summary parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.getCpuCoreSummary(cts.Kit, req)
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

	req := new(rstypes.CpuCoreSummaryReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server cpu core summary, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server cpu core summary parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 业务下查询summary时，bk_biz_ids只能传当前业务ID 或 传空，不能传入其他业务的ID
	if len(req.BkBizIDs) > 0 && (len(req.BkBizIDs) != 1 || req.BkBizIDs[0] != bizID) {
		logs.Errorf("failed to validate rolling server parameter, only bk_biz_id of own business can be passed in, "+
			"bkBizIDs: %v, rid: %s", req.BkBizIDs, cts.Kit.Rid)
		return nil, errf.Newf(errf.InvalidParameter, "only bk_biz_id of own business can be passed in")
	}

	return s.getCpuCoreSummary(cts.Kit, req)
}

// listCpuCoreSummary list cpu core summary.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_cpu_core_summary.md
func (s *service) getCpuCoreSummary(kt *kit.Kit, req *rstypes.CpuCoreSummaryReq) (any, error) {
	return s.rollingServerLogic.GetCpuCoreSummary(kt, req)
}
