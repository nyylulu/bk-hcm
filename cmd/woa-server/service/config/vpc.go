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

// Package config vpc config
package config

import (
	"strconv"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// GetVpc gets vpc config list
func (s *service) GetVpc(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetVpcParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := mapstr.MapStr{
		"region": input.Region,
	}

	rst, err := s.logics.Vpc().GetVpc(cts.Kit, &cond)
	if err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetVpcList gets vpc id list
func (s *service) GetVpcList(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetVpcListParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := map[string]interface{}{}
	if len(input.Regions) > 0 {
		cond["region"] = map[string]interface{}{
			pkg.BKDBIN: input.Regions,
		}
	}

	rst, err := s.logics.Vpc().GetVpcList(cts.Kit, cond)
	if err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateVpc creates vpc config
func (s *service) CreateVpc(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.Vpc)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Vpc().CreateVpc(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateVpc updates vpc config
func (s *service) UpdateVpc(cts *rest.Contexts) (interface{}, error) {
	inputData := new(mapstr.MapStr)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to update vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Vpc().UpdateVpc(cts.Kit, instId, inputData); err != nil {
		logs.Errorf("failed to update vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteVpc deletes vpc config
func (s *service) DeleteVpc(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Vpc().DeleteVpc(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncVpc sync vpc config from yunti
func (s *service) SyncVpc(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.GetVpcParam)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to sync vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Vpc().SyncVpc(cts.Kit, inputData); err != nil {
		logs.Errorf("failed to sync vpc, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpsertRegionDftVpc upsert region default vpc
func (s *service) UpsertRegionDftVpc(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpsertRegionDftVpcReq)
	if err := cts.DecodeInto(input); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	if err := s.logics.Vpc().UpsertRegionDftVpc(cts.Kit, input.RegionDftVpcInfos); err != nil {
		logs.Errorf("failed to upsert region default vpc, err: %v, input: %v, rid: %s", err, converter.PtrToVal(input),
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
