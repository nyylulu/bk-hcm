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

// Package dao supplies all the apply order modify record related operations.
package dao

import (
	"context"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// ModifyRecord supplies all the apply order modify record related operations.
type ModifyRecord interface {
	// NextSequence returns next apply order modify record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateModifyRecord creates apply order modify record in db
	CreateModifyRecord(ctx context.Context, inst *table.ModifyRecord) error
	// GetModifyRecord gets apply order modify record by filter from db
	GetModifyRecord(ctx context.Context, filter *mapstr.MapStr) (*table.ModifyRecord, error)
	// CountModifyRecord gets apply order modify record count by filter from db
	CountModifyRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyModifyRecord gets apply order modify record list by filter from db
	FindManyModifyRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.ModifyRecord, error)
	// UpdateModifyRecord updates apply order modify record by filter and doc in db
	UpdateModifyRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
}

var _ ModifyRecord = new(modifyRecordDao)

type modifyRecordDao struct {
}

// NextSequence returns next apply order modify record sequence id from db
func (mr *modifyRecordDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.ModifyRecordTable)
}

// CreateModifyRecord creates apply order modify record in db
func (mr *modifyRecordDao) CreateModifyRecord(ctx context.Context, inst *table.ModifyRecord) error {
	return mongodb.Client().Table(table.ModifyRecordTable).Insert(ctx, inst)
}

// GetModifyRecord gets apply order modify record by filter from db
func (mr *modifyRecordDao) GetModifyRecord(ctx context.Context, filter *mapstr.MapStr) (*table.ModifyRecord, error) {
	inst := new(table.ModifyRecord)

	if err := mongodb.Client().Table(table.ModifyRecordTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountModifyRecord gets apply order modify record count by filter from db
func (mr *modifyRecordDao) CountModifyRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.ModifyRecordTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyModifyRecord gets apply order modify record list by filter from db
func (mr *modifyRecordDao) FindManyModifyRecord(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.ModifyRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.ModifyRecordTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.ModifyRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateModifyRecord updates apply order modify record by filter and doc in db
func (mr *modifyRecordDao) UpdateModifyRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.ModifyRecordTable).Update(ctx, filter, doc)
}
