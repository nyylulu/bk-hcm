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
	"hcm/pkg/tools/metadata"
)

// PoolHost supplies all the pool host related operations.
type PoolHost interface {
	// NextSequence returns next pool host sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreatePoolHost creates pool host in db
	CreatePoolHost(ctx context.Context, inst *table.PoolHost) error
	// GetPoolHost gets pool host by filter from db
	GetPoolHost(ctx context.Context, filter map[string]interface{}) (*table.PoolHost, error)
	// CountPoolHost gets pool host count by filter from db
	CountPoolHost(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyPoolHost gets pool host list by filter from db
	FindManyPoolHost(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*table.PoolHost,
		error)
	// UpdatePoolHost updates pool host by filter and doc in db
	UpdatePoolHost(ctx context.Context, filter, doc map[string]interface{}) error
	// UpsertPoolHost updates or inserts pool host by filter in db
	UpsertPoolHost(ctx context.Context, filter map[string]interface{}, inst *table.PoolHost) error
}

var _ PoolHost = new(poolHostDao)

type poolHostDao struct {
}

// NextSequence returns next pool host sequence id from db
func (ph *poolHostDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.PoolHostTable)
}

// CreatePoolHost creates pool host in db
func (ph *poolHostDao) CreatePoolHost(ctx context.Context, inst *table.PoolHost) error {
	return mongodb.Client().Table(table.PoolHostTable).Insert(ctx, inst)
}

// GetPoolHost gets pool host by filter from db
func (ph *poolHostDao) GetPoolHost(ctx context.Context, filter map[string]interface{}) (*table.PoolHost, error) {
	inst := new(table.PoolHost)

	if err := mongodb.Client().Table(table.PoolHostTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountPoolHost gets pool host count by filter from db
func (ph *poolHostDao) CountPoolHost(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.PoolHostTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyPoolHost gets pool host list by filter from db
func (ph *poolHostDao) FindManyPoolHost(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.PoolHost, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.PoolHostTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("bk_host_id")
	}

	insts := make([]*table.PoolHost, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdatePoolHost updates pool host by filter and doc in db
func (ph *poolHostDao) UpdatePoolHost(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.PoolHostTable).Update(ctx, filter, doc)
}

// UpsertPoolHost updates or inserts pool host by filter in db
func (ph *poolHostDao) UpsertPoolHost(ctx context.Context, filter map[string]interface{}, inst *table.PoolHost) error {
	return mongodb.Client().Table(table.PoolHostTable).Upsert(ctx, filter, inst)
}
