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

// Package dao is a collection of dao implementation.
package dao

import (
	"context"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// RecycleOrder supplies all the recycle order related operations.
type RecycleOrder interface {
	// NextSequence returns next recycle order sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateRecycleOrder creates recycle order in db
	CreateRecycleOrder(ctx context.Context, inst *table.RecycleOrder) error
	// GetRecycleOrder gets recycle order by filter from db
	GetRecycleOrder(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleOrder, error)
	// CountRecycleOrder gets recycle order count by filter from db
	CountRecycleOrder(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets recycle order list by filter from db
	FindManyRecycleOrder(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.RecycleOrder, error)
	// UpdateRecycleOrder updates recycle order by filter and doc in db
	UpdateRecycleOrder(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
}

var _ RecycleOrder = new(recycleOrderDao)

type recycleOrderDao struct {
}

// NextSequence returns next recycle order sequence id from db
func (ro *recycleOrderDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.RecycleOrderTable)
}

// CreateRecycleOrder creates recycle order in db
func (ro *recycleOrderDao) CreateRecycleOrder(ctx context.Context, inst *table.RecycleOrder) error {
	return mongodb.Client().Table(table.RecycleOrderTable).Insert(ctx, inst)
}

// GetRecycleOrder gets recycle order by filter from db
func (ro *recycleOrderDao) GetRecycleOrder(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleOrder, error) {
	inst := new(table.RecycleOrder)

	if err := mongodb.Client().Table(table.RecycleOrderTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecycleOrder gets recycle order count by filter from db
func (ro *recycleOrderDao) CountRecycleOrder(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecycleOrderTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecycleOrder gets recycle order list by filter from db
func (ro *recycleOrderDao) FindManyRecycleOrder(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecycleOrder, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecycleOrderTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecycleOrder, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecycleOrder updates recycle order by filter and doc in db
func (ro *recycleOrderDao) UpdateRecycleOrder(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.RecycleOrderTable).Update(ctx, filter, doc)
}
