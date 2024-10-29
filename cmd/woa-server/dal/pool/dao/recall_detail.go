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

// Package dao ...
package dao

import (
	"context"

	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// RecallDetail supplies all the resource recall detail related operations.
type RecallDetail interface {
	// CreateRecallDetail creates resource recall task in db
	CreateRecallDetail(ctx context.Context, inst *table.RecallDetail) error
	// GetRecallDetail gets resource recall task by filter from db
	GetRecallDetail(ctx context.Context, filter *mapstr.MapStr) (*table.RecallDetail, error)
	// CountRecallDetail gets resource recall task count by filter from db
	CountRecallDetail(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyRecallDetail gets resource recall task list by filter from db
	FindManyRecallDetail(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.RecallDetail, error)
	// UpdateRecallDetail updates resource recall task by filter and doc in db
	UpdateRecallDetail(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ RecallDetail = new(recallDetailDao)

type recallDetailDao struct {
}

// CreateRecallDetail creates resource recall task in db
func (rd *recallDetailDao) CreateRecallDetail(ctx context.Context, inst *table.RecallDetail) error {
	return mongodb.Client().Table(table.RecallDetailTable).Insert(ctx, inst)
}

// GetRecallDetail gets resource recall task by filter from db
func (rd *recallDetailDao) GetRecallDetail(ctx context.Context, filter *mapstr.MapStr) (*table.RecallDetail, error) {
	inst := new(table.RecallDetail)

	if err := mongodb.Client().Table(table.RecallDetailTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountRecallDetail gets resource recall task count by filter from db
func (rd *recallDetailDao) CountRecallDetail(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.RecallDetailTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyRecallDetail gets resource recall task list by filter from db
func (rd *recallDetailDao) FindManyRecallDetail(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.RecallDetail, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.RecallDetailTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-create_at")
	}

	insts := make([]*table.RecallDetail, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateRecallDetail updates resource recall task by filter and doc in db
func (rd *recallDetailDao) UpdateRecallDetail(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.RecallDetailTable).Update(ctx, filter, doc)
}
