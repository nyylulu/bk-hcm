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
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// GetCpuCoreSummary get cpu core summary.
func (s *service) GetCpuCoreSummary(cts *rest.Contexts) (any, error) {
	return s.getCpuCoreSummary(cts, handler.ListResourceAuthRes, meta.RollingServerManage, meta.Find)
}

// GetBizCpuCoreSummary get biz cpu core summary.
func (s *service) GetBizCpuCoreSummary(cts *rest.Contexts) (any, error) {
	return s.getCpuCoreSummary(cts, handler.ListBizAuthRes, meta.Biz, meta.Find)
}

// listCpuCoreSummary list cpu core summary.
// docs: docs/api-docs/web-server/docs/scr/rolling-server/list_rolling_server_cpu_core_summary.md
func (s *service) getCpuCoreSummary(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	resType meta.ResourceType, action meta.Action) (any, error) {

	req := new(rstypes.CpuCoreSummaryReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server cpu core summary, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server cpu core summary parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: resType, Action: action})
	if err != nil {
		logs.Errorf("list rolling server cpu core summary failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list rolling server cpu core summary no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &rsproto.RollingCpuCoreSummaryItem{}, nil
	}

	return s.rollingServerLogic.GetCpuCoreSummary(cts.Kit, req)
}
