/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package config requirement config
package config

import (
	"strconv"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetRequirement gets requirement type config list
func (s *service) GetRequirement(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.Requirement().GetRequirement(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get requirement list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateRequirement creates requirement type config
func (s *service) CreateRequirement(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.Requirement)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create requirement type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Requirement().CreateRequirement(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create requirement type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateRequirement updates requirement type config
func (s *service) UpdateRequirement(cts *rest.Contexts) (interface{}, error) {
	inputData := new(mapstr.MapStr)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to update requirement type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Requirement().UpdateRequirement(cts.Kit, instId, inputData); err != nil {
		logs.Errorf("failed to update requirement type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteRequirement deletes requirement type config
func (s *service) DeleteRequirement(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Requirement().DeleteRequirement(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete requirement type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
