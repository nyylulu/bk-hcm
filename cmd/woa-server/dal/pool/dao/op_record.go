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

// OpRecord supplies all the resource operation record related operations.
type OpRecord interface {
	// NextSequence returns next resource operation record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateOpRecord creates resource operation record in db
	CreateOpRecord(ctx context.Context, inst *table.OpRecord) error
	// GetRecycleOrder gets resource operation record by filter from db
	GetOpRecord(ctx context.Context, filter *mapstr.MapStr) (*table.OpRecord, error)
	// CountRecycleOrder gets resource operation record count by filter from db
	CountOpRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource operation record list by filter from db
	FindManyOpRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*table.OpRecord,
		error)
	// UpdateOpRecord updates resource operation record by filter and doc in db
	UpdateOpRecord(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ OpRecord = new(opRecordDao)

type opRecordDao struct {
}

// NextSequence returns next resource operation record sequence id from db
func (or *opRecordDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.PoolOpRecordTable)
}

// CreateOpRecord creates resource operation record in db
func (or *opRecordDao) CreateOpRecord(ctx context.Context, inst *table.OpRecord) error {
	return mongodb.Client().Table(table.PoolOpRecordTable).Insert(ctx, inst)
}

// GetOpRecord gets resource operation record by filter from db
func (or *opRecordDao) GetOpRecord(ctx context.Context, filter *mapstr.MapStr) (*table.OpRecord, error) {
	inst := new(table.OpRecord)

	if err := mongodb.Client().Table(table.PoolOpRecordTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountOpRecord gets resource operation record count by filter from db
func (or *opRecordDao) CountOpRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.PoolOpRecordTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyOpRecord gets resource operation record list by filter from db
func (or *opRecordDao) FindManyOpRecord(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.OpRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.PoolOpRecordTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.OpRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateOpRecord updates resource operation record by filter and doc in db
func (or *opRecordDao) UpdateOpRecord(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.PoolOpRecordTable).Update(ctx, filter, doc)
}
