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

type zone struct {
}

// NextSequence returns next zone config sequence id from db
func (z *zone) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameCfgQcloudZone)
}

// CreateZone creates zone config in db
func (z *zone) CreateZone(ctx context.Context, inst *types.Zone) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgQcloudZone).Insert(ctx, inst)
}

// GetZone gets zone config by filter from db
func (z *zone) GetZone(ctx context.Context, filter *mapstr.MapStr) (*types.Zone, error) {
	inst := new(types.Zone)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgQcloudZone).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyZone gets zone config list by filter from db
func (z *zone) FindManyZone(ctx context.Context, filter *mapstr.MapStr) ([]*types.Zone, error) {
	insts := make([]*types.Zone, 0)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgQcloudZone).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateZone updates zone config by filter and doc in db
func (z *zone) UpdateZone(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgQcloudZone).Update(ctx, filter, doc)
}

// DeleteZone deletes zone config from db
func (z *zone) DeleteZone(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgQcloudZone).Delete(ctx, filter)
}
