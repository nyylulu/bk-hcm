/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package config

import (
	"context"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

type cvmApplyOrderStatisticsConfig struct{}

// NextSequence returns next config sequence id from db
func (c *cvmApplyOrderStatisticsConfig) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig)
}

// Create creates config in db
func (c *cvmApplyOrderStatisticsConfig) Create(ctx context.Context, inst *types.CvmApplyOrderStatisticsConfig) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig).Insert(ctx, inst)
}

// Get gets config by filter from db
func (c *cvmApplyOrderStatisticsConfig) Get(ctx context.Context, filter *mapstr.MapStr) (*types.CvmApplyOrderStatisticsConfig, error) {
	inst := new(types.CvmApplyOrderStatisticsConfig)
	table := mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig)
	if err := table.Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}
	return inst, nil
}

// Count gets config count by filter from db
func (c *cvmApplyOrderStatisticsConfig) Count(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig).Find(filter).Count(ctx)
}

// FindMany gets config list by filter from db
func (c *cvmApplyOrderStatisticsConfig) FindMany(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*types.CvmApplyOrderStatisticsConfig, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	table := mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig)
	query := table.Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("id")
	}

	insts := make([]*types.CvmApplyOrderStatisticsConfig, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}
	return insts, nil
}

// Update updates config by filter and doc in db
func (c *cvmApplyOrderStatisticsConfig) Update(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig).Update(ctx, filter, doc)
}

// Delete deletes config from db
func (c *cvmApplyOrderStatisticsConfig) Delete(ctx context.Context, filter map[string]interface{}) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmApplyOrderStatisticsConfig).Delete(ctx, filter)
}
