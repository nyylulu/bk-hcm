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

// Package dao is a collection of dao implementations.
package dao

import (
	"context"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
)

// RecycleHost supplies all the recycle host related operations.
type RecycleHost interface {
	// NextSequence returns next recycle host sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateRecycleHost creates recycle host in db
	CreateRecycleHost(ctx context.Context, inst *table.RecycleHost) error
	// GetRecycleHost gets recycle host by filter from db
	GetRecycleHost(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleHost, error)
	// CountRecycleHost gets recycle host count by filter from db
	CountRecycleHost(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleHost gets recycle host list by filter from db
	FindManyRecycleHost(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.RecycleHost, error)
	// UpdateRecycleHost updates recycle host by filter and doc in db
	UpdateRecycleHost(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteApplyHost deletes recycle host from db
	DeleteRecycleHost(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// Distinct gets recycle record distinct result from db
	Distinct(ctx context.Context, field string, filter map[string]interface{}) ([]interface{}, error)
}

var _ RecycleHost = new(recycleHostDao)

type recycleHostDao struct {
}

// NextSequence returns next recycle host sequence id from db
func (rh *recycleHostDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.RecycleHostTable)
}

// CreateRecycleHost creates recycle host in db
func (rh *recycleHostDao) CreateRecycleHost(ctx context.Context, inst *table.RecycleHost) error {
	return mongodb.Client().Table(table.RecycleHostTable).Insert(ctx, inst)
}

// GetRecycleHost gets recycle host by filter from db
func (rh *recycleHostDao) GetRecycleHost(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleHost, error) {
	inst := new(table.RecycleHost)

	if err := mongodb.Client().Table(table.RecycleHostTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecycleHost gets recycle host count by filter from db
func (rh *recycleHostDao) CountRecycleHost(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecycleHostTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecycleHost gets recycle host list by filter from db
func (rh *recycleHostDao) FindManyRecycleHost(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecycleHost, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecycleHostTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecycleHost, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecycleHost updates recycle host by filter and doc in db
func (rh *recycleHostDao) UpdateRecycleHost(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.RecycleHostTable).Update(ctx, filter, doc)
}

// DeleteApplyHost deletes recycle host from db
func (rh *recycleHostDao) DeleteRecycleHost(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	return mongodb.Client().Table(table.RecycleHostTable).DeleteMany(ctx, filter)
}

// Distinct gets recycle host distinct result from db
func (rh *recycleHostDao) Distinct(ctx context.Context, field string, filter map[string]interface{}) (
	[]interface{}, error) {
	insts, err := mongodb.Client().Table(table.RecycleHostTable).Distinct(ctx, field, filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}
