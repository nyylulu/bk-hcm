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

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/config"
)

type idcZone struct {
}

// NextSequence returns next zone config sequence id from db
func (z *idcZone) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgIdcZone)
}

// CreateZone creates zone config in db
func (z *idcZone) CreateZone(ctx context.Context, inst *types.IdcZone) error {
	return mongodb.Client().Table(common.BKTableNameCfgIdcZone).Insert(ctx, inst)
}

// GetZone gets zone config by filter from db
func (z *idcZone) GetZone(ctx context.Context, filter *mapstr.MapStr) (*types.IdcZone, error) {
	inst := new(types.IdcZone)

	if err := mongodb.Client().Table(common.BKTableNameCfgIdcZone).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyZone gets zone config list by filter from db
func (z *idcZone) FindManyZone(ctx context.Context, filter *mapstr.MapStr) ([]*types.IdcZone, error) {
	insts := make([]*types.IdcZone, 0)

	if err := mongodb.Client().Table(common.BKTableNameCfgIdcZone).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// GetRegionList gets region list by filter from db
func (z *idcZone) GetRegionList(ctx context.Context, filter map[string]interface{}) ([]interface{}, error) {
	insts, err := mongodb.Client().Table(common.BKTableNameCfgIdcZone).Distinct(ctx, "cmdb_region_name", filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateZone updates zone config by filter and doc in db
func (z *idcZone) UpdateZone(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgIdcZone).Update(ctx, filter, doc)
}

// DeleteZone deletes zone config from db
func (z *idcZone) DeleteZone(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgIdcZone).Delete(ctx, filter)
}
