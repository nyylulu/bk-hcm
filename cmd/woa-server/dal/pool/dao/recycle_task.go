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

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
)

// RecycleTask supplies all the resource recycle task related operations.
type RecycleTask interface {
	// NextSequence returns next resource recycle task sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateRecycleTask creates resource recycle task in db
	CreateRecycleTask(ctx context.Context, inst *table.RecycleTask) error
	// GetRecycleOrder gets resource recycle task by filter from db
	GetRecycleTask(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleTask, error)
	// CountRecycleOrder gets resource recycle task count by filter from db
	CountRecycleTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource recycle task list by filter from db
	FindManyRecycleTask(ctx context.Context, page metadata.BasePage,
		filter map[string]interface{}) ([]*table.RecycleTask,
		error)
	// UpdateRecycleTask updates resource recycle task by filter and doc in db
	UpdateRecycleTask(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ RecycleTask = new(recycleTaskDao)

type recycleTaskDao struct {
}

// NextSequence returns next resource recycle task sequence id from db
func (rt *recycleTaskDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.RecycleTaskTable)
}

// CreateRecycleTask creates resource recycle task in db
func (rt *recycleTaskDao) CreateRecycleTask(ctx context.Context, inst *table.RecycleTask) error {
	return mongodb.Client().Table(table.RecycleTaskTable).Insert(ctx, inst)
}

// GetRecycleTask gets resource recycle task by filter from db
func (rt *recycleTaskDao) GetRecycleTask(ctx context.Context, filter *mapstr.MapStr) (*table.RecycleTask, error) {
	inst := new(table.RecycleTask)

	if err := mongodb.Client().Table(table.RecycleTaskTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecycleTask gets resource recycle task count by filter from db
func (rt *recycleTaskDao) CountRecycleTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecycleTaskTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecycleTask gets resource recycle task list by filter from db
func (rt *recycleTaskDao) FindManyRecycleTask(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecycleTask, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecycleTaskTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecycleTask, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecycleTask updates resource recycle task by filter and doc in db
func (rt *recycleTaskDao) UpdateRecycleTask(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.RecycleTaskTable).Update(ctx, filter, doc)
}
