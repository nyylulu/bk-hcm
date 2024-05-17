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
	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
)

// ZoneIf provides management interface for operations of zone config
type ZoneIf interface {
	// GetZone get zone type config list
	GetZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetZoneResult, error)
	// CreateZone creates zone type config
	CreateZone(kt *kit.Kit, input *types.Zone) (mapstr.MapStr, error)
	// UpdateZone updates zone type config
	UpdateZone(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteZone deletes zone type config
	DeleteZone(kt *kit.Kit, instId int64) error

	// GetIdcZone get idc zone type config list
	GetIdcZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetIdcZoneRst, error)
	// CreateIdcZone creates idc zone type config
	CreateIdcZone(kt *kit.Kit, input *types.IdcZone) (mapstr.MapStr, error)
}

// NewZoneOp creates a zone interface
func NewZoneOp() ZoneIf {
	return &zone{}
}

type zone struct {
}

// GetZone get zone type config list
func (z *zone) GetZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetZoneResult, error) {
	insts, err := config.Operation().Zone().FindManyZone(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetZoneResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// CreateZone creates zone type config
func (z *zone) CreateZone(kt *kit.Kit, input *types.Zone) (mapstr.MapStr, error) {
	id, err := config.Operation().Zone().NextSequence(kt.Ctx)
	if err != nil {
		blog.Errorf("failed to create zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Zone().CreateZone(kt.Ctx, input); err != nil {
		blog.Errorf("failed to create zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateZone updates zone type config
func (z *zone) UpdateZone(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Zone().UpdateZone(kt.Ctx, filter, input); err != nil {
		blog.Errorf("failed to update zone, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteZone deletes zone type config
func (z *zone) DeleteZone(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Zone().DeleteZone(kt.Ctx, filter); err != nil {
		blog.Errorf("failed to delete zone, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetIdcZone get idc zone type config list
func (z *zone) GetIdcZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetIdcZoneRst, error) {
	insts, err := config.Operation().IdcZone().FindManyZone(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetIdcZoneRst{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// CreateIdcZone creates idc zone type config
func (z *zone) CreateIdcZone(kt *kit.Kit, input *types.IdcZone) (mapstr.MapStr, error) {
	id, err := config.Operation().IdcZone().NextSequence(kt.Ctx)
	if err != nil {
		blog.Errorf("failed to create idc zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().IdcZone().CreateZone(kt.Ctx, input); err != nil {
		blog.Errorf("failed to create idc zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}
