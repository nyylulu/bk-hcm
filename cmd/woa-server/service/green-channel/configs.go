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

// GetConfigs ...
func (s *service) GetConfigs(cts *rest.Contexts) (interface{}, error) {
	result, err := s.gcLogics.GetConfigs(cts.Kit)
	if err != nil {
		logs.Errorf("get green channel configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateConfigs ...
func (s *service) UpdateConfigs(cts *rest.Contexts) (interface{}, error) {
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.GreenChannel, Action: meta.Find}})
	if err != nil {
		logs.Errorf("update green channel configs auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(greenchannel.UpdateConfigsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update green channel configs decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("update green channel configs validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.gcLogics.UpdateConfigs(cts.Kit, req); err != nil {
		logs.Errorf("update green channel configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
