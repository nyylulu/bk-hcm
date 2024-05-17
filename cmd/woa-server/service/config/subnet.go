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

package config

import (
	"strconv"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/blog"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// GetSubnet gets subnet config list
func (s *service) GetSubnet(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetSubnetParam)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := map[string]interface{}{
		"region": input.Region,
		"zone":   input.Zone,
		"vpc_id": input.Vpc,
	}
	// get subnet with enable flag only
	cond["enable"] = true

	rst, err := s.logics.Subnet().GetSubnet(cts.Kit, cond)
	if err != nil {
		blog.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetSubnetList gets subnet detail config list
func (s *service) GetSubnetList(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetSubnetListParam)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get subnet list, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Subnet().GetSubnetList(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateSubnet creates subnet config
func (s *service) CreateSubnet(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.Subnet)
	if err := cts.DecodeInto(inputData); err != nil {
		blog.Errorf("failed to create subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Subnet().CreateSubnet(cts.Kit, inputData)
	if err != nil {
		blog.Errorf("failed to create subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateSubnet updates subnet config
func (s *service) UpdateSubnet(cts *rest.Contexts) (interface{}, error) {
	input := make(map[string]interface{})
	if err := cts.DecodeInto(&input); err != nil {
		blog.Errorf("failed to update subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		blog.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Subnet().UpdateSubnet(cts.Kit, instId, input); err != nil {
		blog.Errorf("failed to update subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateSubnetProperty updates subnet config property
func (s *service) UpdateSubnetProperty(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpdateSubnetPropertyParam)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to update subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to update subnet, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cond := map[string]interface{}{
		"id": map[string]interface{}{
			common.BKDBIN: input.Ids,
		},
	}

	data := input.Property
	// cannot update device id
	delete(data, "id")

	if err := s.logics.Subnet().UpdateSubnetBatch(cts.Kit, cond, input.Property); err != nil {
		blog.Errorf("failed to update subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteSubnet deletes subnet config
func (s *service) DeleteSubnet(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		blog.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Subnet().DeleteSubnet(cts.Kit, instId); err != nil {
		blog.Errorf("failed to delete subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncSubnet sync subnet config from yunti
func (s *service) SyncSubnet(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.GetSubnetParam)
	if err := cts.DecodeInto(inputData); err != nil {
		blog.Errorf("failed to sync subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Subnet().SyncSubnet(cts.Kit, inputData); err != nil {
		blog.Errorf("failed to sync subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
