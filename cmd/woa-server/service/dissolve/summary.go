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

	"hcm/cmd/woa-server/types/dissolve"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListDissolveCpuCoreSummary list dissolve cpu core summary
func (s *service) ListDissolveCpuCoreSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(dissolve.ListDissolveCpuCoreSummaryReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 服务请求-机房裁撤-菜单粒度
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ServiceResDissolve, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	return s.listDissolveCpuCoreSummary(cts, req.BizID)
}

// ListBizDissolveCpuCoreSummary list business dissolve cpu core summary
func (s *service) ListBizDissolveCpuCoreSummary(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	})
	if err != nil {
		return nil, err
	}

	return s.listDissolveCpuCoreSummary(cts, bizID)
}

func (s *service) listDissolveCpuCoreSummary(cts *rest.Contexts, bizID int64) (interface{}, error) {
	bizSummaryMap, err := s.logics.Table().ListBizCpuCoreSummary(cts.Kit, []int64{bizID})
	if err != nil {
		logs.Errorf("list biz dissolve cpu core summary failed, err: %v, bizID: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}
	summary, ok := bizSummaryMap[bizID]
	if !ok {
		logs.Errorf("can not find biz dissolve cpu core summary, bizID: %d, rid: %s", bizID, cts.Kit.Rid)
		return nil, fmt.Errorf("can not find biz dissolve cpu core summary, bizID: %d", bizID)
	}

	return summary, nil
}
