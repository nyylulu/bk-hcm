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
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// RegionIf provides management interface for operations of region config
type RegionIf interface {
	// GetRegion get region type config list
	GetRegion(kt *kit.Kit) (*types.GetRegionResult, error)
	// CreateRegion creates region type config
	CreateRegion(kt *kit.Kit, input *types.Region) (mapstr.MapStr, error)
	// UpdateRegion updates region type config
	UpdateRegion(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteRegion deletes region type config
	DeleteRegion(kt *kit.Kit, instId int64) error
	// GetIdcRegion get region type config list
	GetIdcRegion(kt *kit.Kit) (*types.GetIdcRegionRst, error)
}

// NewRegionOp creates a region interface
func NewRegionOp() RegionIf {
	return &region{}
}

type region struct {
}

// GetRegion get region type config list
func (r *region) GetRegion(kt *kit.Kit) (*types.GetRegionResult, error) {
	filter := new(mapstr.MapStr)
	insts, err := config.Operation().Region().FindManyRegion(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRegionResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// CreateRegion creates region type config
func (r *region) CreateRegion(kt *kit.Kit, input *types.Region) (mapstr.MapStr, error) {
	id, err := config.Operation().Region().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create region, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Region().CreateRegion(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create region, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateRegion updates region type config
func (r *region) UpdateRegion(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Region().UpdateRegion(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update region, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteRegion deletes region type config
func (r *region) DeleteRegion(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Region().DeleteRegion(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete region, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetIdcRegion get idc region list
func (r *region) GetIdcRegion(kt *kit.Kit) (*types.GetIdcRegionRst, error) {
	filter := make(map[string]interface{})

	insts, err := config.Operation().IdcZone().GetRegionList(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetIdcRegionRst{
		Info: insts,
	}

	return rst, nil
}
