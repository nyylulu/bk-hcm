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

// Package dao defines all dao operator of recycle return task
package dao

import (
	"context"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
)

// ReturnTask recycle return task operation interface
type ReturnTask interface {
	// CreateRecycleTask creates recycle return task in db
	CreateReturnTask(ctx context.Context, inst *table.ReturnTask) error
	// GetRecycleTask gets recycle return task by filter from db
	GetReturnTask(ctx context.Context, filter *mapstr.MapStr) (*table.ReturnTask, error)
	// CountRecycleTask gets recycle return task count by filter from db
	CountReturnTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleTask gets recycle return task list by filter from db
	FindManyReturnTask(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.ReturnTask, error)
	// UpdateRecycleTask updates recycle return task by filter and doc in db
	UpdateReturnTask(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
}

var _ ReturnTask = new(returnTaskDao)

type returnTaskDao struct {
}

// CreateReturnTask creates recycle return task in db
func (rt *returnTaskDao) CreateReturnTask(ctx context.Context, inst *table.ReturnTask) error {
	return mongodb.Client().Table(table.ReturnTaskTable).Insert(ctx, inst)
}

// GetReturnTask gets recycle return task by filter from db
func (rt *returnTaskDao) GetReturnTask(ctx context.Context, filter *mapstr.MapStr) (*table.ReturnTask, error) {
	inst := new(table.ReturnTask)

	if err := mongodb.Client().Table(table.ReturnTaskTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountReturnTask gets recycle return task count by filter from db
func (rt *returnTaskDao) CountReturnTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.ReturnTaskTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyReturnTask gets recycle return task list by filter from db
func (rt *returnTaskDao) FindManyReturnTask(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.ReturnTask, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.ReturnTaskTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("task_id")
	}

	insts := make([]*table.ReturnTask, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateReturnTask updates recycle return task by filter and doc in db
func (rt *returnTaskDao) UpdateReturnTask(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.ReturnTaskTable).Update(ctx, filter, doc)
}
