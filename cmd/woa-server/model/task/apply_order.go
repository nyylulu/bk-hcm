/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package model implements all db operations of apply order
package model

import (
	"context"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	daltypes "hcm/cmd/woa-server/storage/dal/types"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
)

type applyOrder struct {
}

// NextSequence returns next apply order sequence id from db
func (a *applyOrder) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameApplyOrder)
}

// CreateApplyOrder creates apply order in db
func (a *applyOrder) CreateApplyOrder(ctx context.Context, inst *types.ApplyOrder) error {
	err := mongodb.Client().Table(common.BKTableNameApplyOrder).Insert(ctx, inst)
	if err != nil {
		logs.Errorf("create suborder for table %s failed, err: %v", common.BKTableNameApplyOrder, err)
		return err
	}

	return nil
}

// GetApplyOrder gets apply order by filter from db
func (a *applyOrder) GetApplyOrder(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyOrder, error) {
	inst := new(types.ApplyOrder)

	if err := mongodb.Client().Table(common.BKTableNameApplyOrder).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountApplyOrder gets apply order count by filter from db
func (a *applyOrder) CountApplyOrder(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameApplyOrder).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyApplyOrder gets apply order list by filter from db
func (a *applyOrder) FindManyApplyOrder(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.ApplyOrder, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(common.BKTableNameApplyOrder).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("suborder_id")
	}

	insts := make([]*types.ApplyOrder, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateApplyOrder updates apply order by filter and doc in db
func (a *applyOrder) UpdateApplyOrder(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameApplyOrder).Update(ctx, filter, doc)
}

// DeleteApplyOrder deletes apply order from db
func (a *applyOrder) DeleteApplyOrder() {
	// TODO
}

// AggregateAll apply order aggregate all operation
func (a *applyOrder) AggregateAll(ctx context.Context, pipeline interface{}, result interface{},
	opts ...*daltypes.AggregateOpts) error {

	if err := mongodb.Client().Table(common.BKTableNameApplyOrder).AggregateAll(ctx, pipeline, result,
		opts...); err != nil {
		return err
	}

	return nil
}
