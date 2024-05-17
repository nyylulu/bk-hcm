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

type region struct {
}

// NextSequence returns next region config sequence id from db
func (r *region) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgQcloudRegion)
}

// CreateRegion creates region config in db
func (r *region) CreateRegion(ctx context.Context, inst *types.Region) error {
	return mongodb.Client().Table(common.BKTableNameCfgQcloudRegion).Insert(ctx, inst)
}

// GetRegion gets resource region config by filter from db
func (r *region) GetRegion(ctx context.Context, filter *mapstr.MapStr) (*types.Region, error) {
	inst := new(types.Region)

	if err := mongodb.Client().Table(common.BKTableNameCfgQcloudRegion).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyRegion gets region config list by filter from db
func (r *region) FindManyRegion(ctx context.Context, filter *mapstr.MapStr) ([]*types.Region, error) {
	insts := make([]*types.Region, 0)

	if err := mongodb.Client().Table(common.BKTableNameCfgQcloudRegion).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRegion updates region config by filter and doc in db
func (r *region) UpdateRegion(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgQcloudRegion).Update(ctx, filter, doc)
}

// DeleteRegion deletes region config from db
func (r *region) DeleteRegion(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgQcloudRegion).Delete(ctx, filter)
}
