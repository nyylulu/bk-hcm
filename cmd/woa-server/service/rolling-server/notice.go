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
	ptypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// PushReturnNotification push rolling server return notification.
func (s *service) PushReturnNotification(cts *rest.Contexts) (any, error) {
	req := new(ptypes.PushReturnNoticeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to push rolling server return notification, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate push rolling server return notification parameter, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RollingServerManage, Action: meta.Find}})
	if err != nil {
		logs.Errorf("push rolling server return notification auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err = s.rollingServerLogic.PushReturnNotifications(cts.Kit, req.BizIDs, req.Receivers); err != nil {
		logs.Errorf("failed to push rolling server return notification, err: %v, req: %v, rid: %s", err, *req,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
