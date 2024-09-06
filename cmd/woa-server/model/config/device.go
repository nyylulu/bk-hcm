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

type cvmDevice struct {
}

// NextSequence returns next resource device type config sequence id from db
func (d *cvmDevice) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameCfgDevice)
}

// CreateDevice creates resource device type config in db
func (d *cvmDevice) CreateDevice(ctx context.Context, inst *types.DeviceInfo) error {
	return mongodb.Client().Table(common.BKTableNameCfgDevice).Insert(ctx, inst)
}

// GetDevice gets resource device type config by filter from db
func (d *cvmDevice) GetDevice(ctx context.Context, filter *mapstr.MapStr) (*types.DeviceInfo, error) {
	inst := new(types.DeviceInfo)

	if err := mongodb.Client().Table(common.BKTableNameCfgDevice).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountDevice gets resource device count by filter from db
func (d *cvmDevice) CountDevice(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameCfgDevice).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDevice gets resource device detail config list by filter from db
func (d *cvmDevice) FindManyDevice(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.DeviceInfo, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(common.BKTableNameCfgDevice).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("id")
	}

	insts := make([]*types.DeviceInfo, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// FindManyDeviceType gets resource device type config list by filter from db
func (d *cvmDevice) FindManyDeviceType(ctx context.Context, filter map[string]interface{}) ([]interface{}, error) {
	insts, err := mongodb.Client().Table(common.BKTableNameCfgDevice).Distinct(ctx, "device_type", filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDevice updates resource device type config by filter and doc in db
func (d *cvmDevice) UpdateDevice(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(common.BKTableNameCfgDevice).Update(ctx, filter, doc)
}

// DeleteDevice deletes resource device type config from db
func (d *cvmDevice) DeleteDevice(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameCfgDevice).Delete(ctx, filter)
}
