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
	"context"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
)

type requirement struct {
}

// NextSequence returns next resource requirement type config sequence id from db
func (r *requirement) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameCfgRequirement)
}

// CreateRequirement creates resource requirement type config in db
func (r *requirement) CreateRequirement(ctx context.Context, inst *types.Requirement) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgRequirement).Insert(ctx, inst)
}

// GetRequirement gets resource requirement type config by filter from db
func (r *requirement) GetRequirement(ctx context.Context, filter *mapstr.MapStr) (*types.Requirement, error) {
	inst := new(types.Requirement)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgRequirement).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyRequirement gets resource requirement type config list by filter from db
func (r *requirement) FindManyRequirement(ctx context.Context, filter *mapstr.MapStr) ([]*types.Requirement, error) {
	insts := make([]*types.Requirement, 0)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgRequirement).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRequirement updates resource requirement type config by filter and doc in db
func (r *requirement) UpdateRequirement(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgRequirement).Update(ctx, filter, doc)
}

// DeleteRequirement deletes resource requirement type config from db
func (r *requirement) DeleteRequirement(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgRequirement).Delete(ctx, filter)
}
