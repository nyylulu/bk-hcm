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
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// RequirementIf provides management interface for operations of requirement config
type RequirementIf interface {
	// GetRequirement get requirement type config list
	GetRequirement(kt *kit.Kit) (*types.GetRequirementResult, error)
	// CreateRequirement creates requirement type config
	CreateRequirement(kt *kit.Kit, input *types.Requirement) (mapstr.MapStr, error)
	// UpdateRequirement updates requirement type config
	UpdateRequirement(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteRequirement deletes requirement type config
	DeleteRequirement(kt *kit.Kit, instId int64) error
}

// NewRequirementOp creates a requirement interface
func NewRequirementOp() RequirementIf {
	return &requirement{}
}

type requirement struct {
}

// GetRequirement get requirement type config list
func (r *requirement) GetRequirement(kt *kit.Kit) (*types.GetRequirementResult, error) {
	filter := new(mapstr.MapStr)
	insts, err := config.Operation().Requirement().FindManyRequirement(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRequirementResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// CreateRequirement creates requirement type config
func (r *requirement) CreateRequirement(kt *kit.Kit, input *types.Requirement) (mapstr.MapStr, error) {
	id, err := config.Operation().Requirement().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create requirement, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Requirement().CreateRequirement(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create requirement, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateRequirement updates requirement type config
func (r *requirement) UpdateRequirement(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Requirement().UpdateRequirement(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update requirement, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteRequirement deletes requirement type config
func (r *requirement) DeleteRequirement(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Requirement().DeleteRequirement(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete requirement, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
