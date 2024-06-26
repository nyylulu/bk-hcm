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

type deviceRestrict struct {
}

// NextSequence returns next device restrict config sequence id from db
func (d *deviceRestrict) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgDeviceRestrict)
}

// CreateDeviceRestrict creates device restrict config in db
func (d *deviceRestrict) CreateDeviceRestrict(ctx context.Context, inst *types.DeviceRestrict) error {
	return mongodb.Client().Table(common.BKTableNameCfgDeviceRestrict).Insert(ctx, inst)
}

// GetDeviceRestrict gets device restrict config by filter from db
func (d *deviceRestrict) GetDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) (*types.DeviceRestrict, error) {
	inst := new(types.DeviceRestrict)

	if err := mongodb.Client().Table(common.BKTableNameCfgDeviceRestrict).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyDeviceRestrict gets device restrict config list by filter from db
func (d *deviceRestrict) FindManyDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) ([]*types.DeviceRestrict,
	error) {

	insts := make([]*types.DeviceRestrict, 0)

	if err := mongodb.Client().Table(common.BKTableNameCfgDeviceRestrict).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDeviceRestrict updates device restrict config by filter and doc in db
func (d *deviceRestrict) UpdateDeviceRestrict(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgDeviceRestrict).Update(ctx, filter, doc)
}

// DeleteDeviceRestrict deletes device restrict config from db
func (d *deviceRestrict) DeleteDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgDeviceRestrict).Delete(ctx, filter)
}
