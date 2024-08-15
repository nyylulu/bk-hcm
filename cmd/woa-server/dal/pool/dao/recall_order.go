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

// RecallOrder supplies all the resource recall order related operations.
type RecallOrder interface {
	// CreateRecallOrder creates resource recall order in db
	CreateRecallOrder(ctx context.Context, inst *table.RecallOrder) error
	// GetRecycleOrder gets resource recall order by filter from db
	GetRecallOrder(ctx context.Context, filter *mapstr.MapStr) (*table.RecallOrder, error)
	// CountRecycleOrder gets resource recall order count by filter from db
	CountRecallOrder(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecycleOrder gets resource recall order list by filter from db
	FindManyRecallOrder(ctx context.Context, page metadata.BasePage,
		filter map[string]interface{}) ([]*table.RecallOrder,
		error)
	// UpdateRecallOrder updates resource recall order by filter and doc in db
	UpdateRecallOrder(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ RecallOrder = new(recallOrderDao)

type recallOrderDao struct {
}

// CreateRecallOrder creates resource recall order in db
func (ro *recallOrderDao) CreateRecallOrder(ctx context.Context, inst *table.RecallOrder) error {
	return mongodb.Client().Table(table.RecallOrderTable).Insert(ctx, inst)
}

// GetRecallOrder gets resource recall order by filter from db
func (ro *recallOrderDao) GetRecallOrder(ctx context.Context, filter *mapstr.MapStr) (*table.RecallOrder, error) {
	inst := new(table.RecallOrder)

	if err := mongodb.Client().Table(table.RecallOrderTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecallOrder gets resource recall order count by filter from db
func (ro *recallOrderDao) CountRecallOrder(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecallOrderTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecallOrder gets resource recall order list by filter from db
func (ro *recallOrderDao) FindManyRecallOrder(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecallOrder, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecallOrderTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecallOrder, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecallOrder updates resource recall order by filter and doc in db
func (ro *recallOrderDao) UpdateRecallOrder(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.RecallOrderTable).Update(ctx, filter, doc)
}
