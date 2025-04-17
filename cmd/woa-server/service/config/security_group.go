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

package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// UpsertRegionDftSg upsert region default security group.
func (s *service) UpsertRegionDftSg(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpsertRegionDftSgReq)
	if err := cts.DecodeInto(input); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	if err := s.logics.Sg().UpsertRegionDftSg(cts.Kit, input.RegionDftSgInfos); err != nil {
		logs.Errorf("failed to upsert region default security group, err: %v, input: %v, rid: %s", err,
			converter.PtrToVal(input), cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
