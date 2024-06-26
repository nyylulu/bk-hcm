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
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/config"
)

type subnet struct {
}

// NextSequence returns next subnet config sequence id from db
func (s *subnet) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgSubnet)
}

// CreateSubnet creates subnet config in db
func (s *subnet) CreateSubnet(ctx context.Context, inst *types.Subnet) error {
	return mongodb.Client().Table(common.BKTableNameCfgSubnet).Insert(ctx, inst)
}

// GetSubnet gets subnet config by filter from db
func (s *subnet) GetSubnet(ctx context.Context, filter *mapstr.MapStr) (*types.Subnet, error) {
	inst := new(types.Subnet)

	if err := mongodb.Client().Table(common.BKTableNameCfgSubnet).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountSubnet gets subnet count by filter from db
func (v *subnet) CountSubnet(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameCfgSubnet).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManySubnet gets subnet config list by filter from db
func (s *subnet) FindManySubnet(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.Subnet, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(common.BKTableNameCfgSubnet).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("id")
	}

	insts := make([]*types.Subnet, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateSubnet updates subnet config by filter and doc in db
func (s *subnet) UpdateSubnet(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(common.BKTableNameCfgSubnet).Update(ctx, filter, doc)
}

// DeleteSubnet deletes subnet config from db
func (s *subnet) DeleteSubnet(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgSubnet).Delete(ctx, filter)
}
