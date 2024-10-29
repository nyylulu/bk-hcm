/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dao

import (
	"context"

	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// GradeCfg supplies all the resource grade config related operations.
type GradeCfg interface {
	// NextSequence returns next resource grade config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateGradeCfg creates resource grade config in db
	CreateGradeCfg(ctx context.Context, inst *table.GradeCfg) error
	// GetRecycleOrder gets resource grade config by filter from db
	GetGradeCfg(ctx context.Context, filter *mapstr.MapStr) (*table.GradeCfg, error)
	// CountRecycleOrder gets resource grade config count by filter from db
	CountGradeCfg(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource grade config list by filter from db
	FindManyGradeCfg(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*table.GradeCfg,
		error)
	// UpdateGradeCfg updates resource grade config by filter and doc in db
	UpdateGradeCfg(ctx context.Context, filter, doc map[string]interface{}) error
	// Distinct gets resource grade config distinct result from db
	Distinct(ctx context.Context, field string, filter map[string]interface{}) ([]interface{}, error)
}

var _ GradeCfg = new(gradeCfgDao)

type gradeCfgDao struct {
}

// NextSequence returns next resource grade config sequence id from db
func (gc *gradeCfgDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.PoolGradeCfgTable)
}

// CreateGradeCfg creates resource grade config in db
func (gc *gradeCfgDao) CreateGradeCfg(ctx context.Context, inst *table.GradeCfg) error {
	return mongodb.Client().Table(table.PoolGradeCfgTable).Insert(ctx, inst)
}

// GetGradeCfg gets resource grade config by filter from db
func (gc *gradeCfgDao) GetGradeCfg(ctx context.Context, filter *mapstr.MapStr) (*table.GradeCfg, error) {
	inst := new(table.GradeCfg)

	if err := mongodb.Client().Table(table.PoolGradeCfgTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountGradeCfg gets resource grade config count by filter from db
func (gc *gradeCfgDao) CountGradeCfg(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.PoolGradeCfgTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyGradeCfg gets resource grade config list by filter from db
func (gc *gradeCfgDao) FindManyGradeCfg(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.GradeCfg, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.PoolGradeCfgTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("id")
	}

	insts := make([]*table.GradeCfg, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateGradeCfg updates resource grade config by filter and doc in db
func (gc *gradeCfgDao) UpdateGradeCfg(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.PoolGradeCfgTable).Update(ctx, filter, doc)
}

// Distinct gets resource grade config distinct result from db
func (gc *gradeCfgDao) Distinct(ctx context.Context, field string, filter map[string]interface{}) (
	[]interface{}, error) {
	insts, err := mongodb.Client().Table(table.PoolGradeCfgTable).Distinct(ctx, field, filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}
