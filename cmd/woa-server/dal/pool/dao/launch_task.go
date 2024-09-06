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

// LaunchTask supplies all the resource launch task related operations.
type LaunchTask interface {
	// NextSequence returns next resource launch task sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateLaunchTask creates resource launch task in db
	CreateLaunchTask(ctx context.Context, inst *table.LaunchTask) error
	// GetRecycleOrder gets resource launch task by filter from db
	GetLaunchTask(ctx context.Context, filter *mapstr.MapStr) (*table.LaunchTask, error)
	// CountRecycleOrder gets resource launch task count by filter from db
	CountLaunchTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource launch task list by filter from db
	FindManyLaunchTask(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*table.LaunchTask,
		error)
	// UpdateLaunchTask updates resource launch task by filter and doc in db
	UpdateLaunchTask(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ LaunchTask = new(launchTaskDao)

type launchTaskDao struct {
}

// NextSequence returns next resource launch task sequence id from db
func (lt *launchTaskDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.LaunchTaskTable)
}

// CreateLaunchTask creates resource launch task in db
func (lt *launchTaskDao) CreateLaunchTask(ctx context.Context, inst *table.LaunchTask) error {
	return mongodb.Client().Table(table.LaunchTaskTable).Insert(ctx, inst)
}

// GetLaunchTask gets resource launch task by filter from db
func (lt *launchTaskDao) GetLaunchTask(ctx context.Context, filter *mapstr.MapStr) (*table.LaunchTask, error) {
	inst := new(table.LaunchTask)

	if err := mongodb.Client().Table(table.LaunchTaskTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountLaunchTask gets resource launch task count by filter from db
func (lt *launchTaskDao) CountLaunchTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.LaunchTaskTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyLaunchTask gets resource launch task list by filter from db
func (lt *launchTaskDao) FindManyLaunchTask(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.LaunchTask, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.LaunchTaskTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-id")
	}

	insts := make([]*table.LaunchTask, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateLaunchTask updates resource launch task by filter and doc in db
func (lt *launchTaskDao) UpdateLaunchTask(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.LaunchTaskTable).Update(ctx, filter, doc)
}
