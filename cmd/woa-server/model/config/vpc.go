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

type vpc struct {
}

// NextSequence returns next vpc config sequence id from db
func (v *vpc) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgVpc)
}

// CreateVpc creates vpc config in db
func (v *vpc) CreateVpc(ctx context.Context, inst *types.Vpc) error {
	return mongodb.Client().Table(common.BKTableNameCfgVpc).Insert(ctx, inst)
}

// GetVpc gets vpc config by filter from db
func (v *vpc) GetVpc(ctx context.Context, filter *mapstr.MapStr) (*types.Vpc, error) {
	inst := new(types.Vpc)

	if err := mongodb.Client().Table(common.BKTableNameCfgVpc).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountVpc gets vpc count by filter from db
func (v *vpc) CountVpc(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameCfgVpc).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyVpc gets vpc config list by filter from db
func (v *vpc) FindManyVpc(ctx context.Context, filter *mapstr.MapStr) ([]*types.Vpc, error) {
	insts := make([]*types.Vpc, 0)

	if err := mongodb.Client().Table(common.BKTableNameCfgVpc).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// FindManyVpcId gets vpc id list by filter from db
func (v *vpc) FindManyVpcId(ctx context.Context, filter map[string]interface{}) ([]interface{}, error) {
	insts, err := mongodb.Client().Table(common.BKTableNameCfgVpc).Distinct(ctx, "vpc_id", filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateVpc updates vpc config by filter and doc in db
func (v *vpc) UpdateVpc(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgVpc).Update(ctx, filter, doc)
}

// DeleteVpc deletes vpc config from db
func (v *vpc) DeleteVpc(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgVpc).Delete(ctx, filter)
}
