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

// Package model implements the object model for the service.
package model

import (
	"context"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
)

type applyStep struct {
}

// NextSequence returns next apply order step sequence id from db
func (a *applyStep) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameApplyStep)
}

// CreateApplyStep creates apply order step info in db
func (a *applyStep) CreateApplyStep(ctx context.Context, inst *types.ApplyStep) error {
	return mongodb.Client().Table(pkg.BKTableNameApplyStep).Insert(ctx, inst)
}

// GetApplyStep gets apply order step info by filter from db
func (a *applyStep) GetApplyStep(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyStep, error) {
	inst := new(types.ApplyStep)

	if err := mongodb.Client().Table(pkg.BKTableNameApplyStep).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountApplyStep gets apply step count by filter from db
func (a *applyStep) CountApplyStep(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(pkg.BKTableNameApplyStep).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyApplyStep gets apply order step info list by filter from db
func (a *applyStep) FindManyApplyStep(ctx context.Context, filter *mapstr.MapStr) ([]*types.ApplyStep, error) {
	insts := make([]*types.ApplyStep, 0)

	if err := mongodb.Client().Table(pkg.BKTableNameApplyStep).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateApplyStep updates apply order step info by filter and doc in db
func (a *applyStep) UpdateApplyStep(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameApplyStep).Update(ctx, filter, doc)
}

// DeleteApplyStep deletes apply order step info from db
func (a *applyStep) DeleteApplyStep() {
	// TODO
}
