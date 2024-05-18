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

// RecallTask supplies all the resource recall task related operations.
type RecallTask interface {
	// NextSequence returns next resource recall task sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateRecallTask creates resource recall task in db
	CreateRecallTask(ctx context.Context, inst *table.RecallTask) error
	// GetRecycleOrder gets resource recall task by filter from db
	GetRecallTask(ctx context.Context, filter *mapstr.MapStr) (*table.RecallTask, error)
	// CountRecycleOrder gets resource recall task count by filter from db
	CountRecallTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource recall task list by filter from db
	FindManyRecallTask(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*table.RecallTask,
		error)
	// UpdateRecallTask updates resource recall task by filter and doc in db
	UpdateRecallTask(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ RecallTask = new(recallTaskDao)

type recallTaskDao struct {
}

// NextSequence returns next resource recall task sequence id from db
func (rt *recallTaskDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.RecallTaskTable)
}

// CreateRecallTask creates resource recall task in db
func (rt *recallTaskDao) CreateRecallTask(ctx context.Context, inst *table.RecallTask) error {
	return mongodb.Client().Table(table.RecallTaskTable).Insert(ctx, inst)
}

// GetRecallTask gets resource recall task by filter from db
func (rt *recallTaskDao) GetRecallTask(ctx context.Context, filter *mapstr.MapStr) (*table.RecallTask, error) {
	inst := new(table.RecallTask)

	if err := mongodb.Client().Table(table.RecallTaskTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecallTask gets resource recall task count by filter from db
func (rt *recallTaskDao) CountRecallTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecallTaskTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecallTask gets resource recall task list by filter from db
func (rt *recallTaskDao) FindManyRecallTask(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecallTask, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecallTaskTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecallTask, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecallTask updates resource recall task by filter and doc in db
func (rt *recallTaskDao) UpdateRecallTask(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.RecallTaskTable).Update(ctx, filter, doc)
}
