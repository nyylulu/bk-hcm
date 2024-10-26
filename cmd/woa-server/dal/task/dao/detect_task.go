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

// Package dao contains all the dao layer api of recycle detection
package dao

import (
	"context"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// DetectTask supplies all the recycle detection related operations.
type DetectTask interface {
	// CreateRecycleTask creates recycle detection task in db
	CreateDetectTask(ctx context.Context, inst *table.DetectTask) error
	// GetRecycleTask gets recycle detection task by filter from db
	GetDetectTask(ctx context.Context, filter *mapstr.MapStr) (*table.DetectTask, error)
	// CountRecycleTask gets recycle detection task count by filter from db
	CountDetectTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleTask gets recycle detection task list by filter from db
	FindManyDetectTask(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.DetectTask, error)
	// UpdateRecycleTask updates recycle detection task by filter and doc in db
	UpdateDetectTask(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteRecycleTask deletes recycle detection task from db
	DeleteDetectTask(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// GetRecycleHostList gets recycle host list by filter from db
	GetRecycleHostList(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
	// UpdateDetectTasks updates recycle detection tasks by filter and doc in db
	UpdateDetectTasks(kt *kit.Kit, filter *mapstr.MapStr, doc *mapstr.MapStr) error
}

var _ DetectTask = new(detectTaskDao)

type detectTaskDao struct {
}

// CreateDetectTask creates recycle detection task in db
func (dt *detectTaskDao) CreateDetectTask(ctx context.Context, inst *table.DetectTask) error {
	return mongodb.Client().Table(table.DetectTaskTable).Insert(ctx, inst)
}

// GetDetectTask gets recycle detection task by filter from db
func (dt *detectTaskDao) GetDetectTask(ctx context.Context, filter *mapstr.MapStr) (*table.DetectTask, error) {
	inst := new(table.DetectTask)

	if err := mongodb.Client().Table(table.DetectTaskTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountDetectTask gets recycle detection task count by filter from db
func (dt *detectTaskDao) CountDetectTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.DetectTaskTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDetectTask gets recycle detection task list by filter from db
func (dt *detectTaskDao) FindManyDetectTask(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.DetectTask, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.DetectTaskTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("task_id")
	}

	insts := make([]*table.DetectTask, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDetectTasks updates recycle detection tasks by filter and doc in db
func (dt *detectTaskDao) UpdateDetectTasks(kt *kit.Kit, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	if _, err := mongodb.Client().Table(table.DetectTaskTable).UpdateMany(kt.Ctx, filter, doc); err != nil {
		logs.Errorf("failed to update detect task, err: %v, filter: %v, doc: %v, rid: %s", err, filter, doc, kt.Rid)
		return err
	}
	return nil
}

// UpdateDetectTask updates recycle detection task by filter and doc in db
func (dt *detectTaskDao) UpdateDetectTask(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.DetectTaskTable).Update(ctx, filter, doc)
}

// DeleteDetectTask deletes recycle detection task from db
func (dt *detectTaskDao) DeleteDetectTask(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	return mongodb.Client().Table(table.DetectTaskTable).DeleteMany(ctx, filter)
}

// GetRecycleHostList gets recycle host list by filter from db
func (dt *detectTaskDao) GetRecycleHostList(ctx context.Context, filter map[string]interface{}) ([]interface{}, error) {
	insts, err := mongodb.Client().Table(table.DetectTaskTable).Distinct(ctx, "ip", filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}
