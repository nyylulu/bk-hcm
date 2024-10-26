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
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListFineDetails list fine details.
func (s *service) ListFineDetails(cts *rest.Contexts) (any, error) {
	return s.listAppliedRecords(cts, handler.ListResourceAuthRes, meta.RollingServerManage, meta.Find)
}

// listFineDetails lists fine details.
func (s *service) listFineDetails(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	resType meta.ResourceType, action meta.Action) (any, error) {

	req := new(rsproto.RollingFineDetailListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list rolling server fine details, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate rolling server fine details parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: s.authorizer,
		ResType: resType, Action: action, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list rolling server fine details failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list rolling server fine details no perm, req: %v, rid: %s", cvt.PtrToVal(req), cts.Kit.Rid)
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &rsproto.RollingFineDetailListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return s.client.DataService().Global.RollingServer.ListFineDetail(cts.Kit, listReq)
}
