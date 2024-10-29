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

// Package config zone config
package config

import (
	"strconv"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetQcloudZone gets qcloud zone config list
func (s *service) GetQcloudZone(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetZoneParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := mapstr.MapStr{}
	// if input region is empty list, return all zone info
	if len(input.Region) > 0 {
		cond["region"] = mapstr.MapStr{
			pkg.BKDBIN: input.Region,
		}
	} else if len(input.CmdbRegion) > 0 {
		cond["cmdb_region_name"] = mapstr.MapStr{
			pkg.BKDBIN: input.CmdbRegion,
		}
	}

	rst, err := s.logics.Zone().GetZone(cts.Kit, &cond)
	if err != nil {
		logs.Errorf("failed to get zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateQcloudZone creates qcloud zone config
func (s *service) CreateQcloudZone(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.Zone)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Zone().CreateZone(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateQcloudZone updates qcloud zone config
func (s *service) UpdateQcloudZone(cts *rest.Contexts) (interface{}, error) {
	inputData := new(mapstr.MapStr)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to update zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Zone().UpdateZone(cts.Kit, instId, inputData); err != nil {
		logs.Errorf("failed to update zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteQcloudZone deletes qcloud zone config
func (s *service) DeleteQcloudZone(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.Zone().DeleteZone(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetIdcZone gets idc zone config list
func (s *service) GetIdcZone(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetIdcZoneParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get idc zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := mapstr.MapStr{}
	// if input region is empty list, return all zone info
	if len(input.Region) > 0 {
		cond["cmdb_region_name"] = mapstr.MapStr{
			pkg.BKDBIN: input.Region,
		}
	}

	rst, err := s.logics.Zone().GetIdcZone(cts.Kit, &cond)
	if err != nil {
		logs.Errorf("failed to get idc zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateIdcZone creates idc zone config
func (s *service) CreateIdcZone(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.IdcZone)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create idc zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Zone().CreateIdcZone(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create idc zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
