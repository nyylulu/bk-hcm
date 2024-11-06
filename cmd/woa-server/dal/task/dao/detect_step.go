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

// Package dao implements the object storage access of recycle detection step.
package dao

import (
	"context"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// DetectStep supplies all the recycle detection step related operations.
type DetectStep interface {
	// CreateDetectStep creates recycle detection step in db
	CreateDetectStep(ctx context.Context, inst *table.DetectStep) error
	// GetDetectStep gets recycle detection step by filter from db
	GetDetectStep(ctx context.Context, filter *mapstr.MapStr) (*table.DetectStep, error)
	// CountDetectStep gets recycle detection step count by filter from db
	CountDetectStep(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDetectStep gets recycle detection step list by filter from db
	FindManyDetectStep(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.DetectStep, error)
	// UpdateDetectStep updates recycle detection step by filter and doc in db
	UpdateDetectStep(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDetectStep deletes recycle detection step from db
	DeleteDetectStep(ctx context.Context, filter map[string]interface{}) (uint64, error)
}

var _ DetectStep = new(detectStepDao)

type detectStepDao struct {
}

// CreateDetectStep creates recycle detection step in db
func (ds *detectStepDao) CreateDetectStep(ctx context.Context, inst *table.DetectStep) error {
	return mongodb.Client().Table(table.DetectStepTable).Insert(ctx, inst)
}

// GetDetectStep gets recycle detection step by filter from db
func (ds *detectStepDao) GetDetectStep(ctx context.Context, filter *mapstr.MapStr) (*table.DetectStep, error) {
	inst := new(table.DetectStep)

	if err := mongodb.Client().Table(table.DetectStepTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountDetectStep gets recycle detection step count by filter from db
func (ds *detectStepDao) CountDetectStep(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.DetectStepTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDetectStep gets recycle detection step list by filter from db
func (ds *detectStepDao) FindManyDetectStep(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.DetectStep, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.DetectStepTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("-step_id")
	}

	insts := make([]*table.DetectStep, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDetectStep updates recycle detection step by filter and doc in db
func (ds *detectStepDao) UpdateDetectStep(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(table.DetectStepTable).Update(ctx, filter, doc)
}

// DeleteDetectStep deletes recycle detection step from db
func (ds *detectStepDao) DeleteDetectStep(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	return mongodb.Client().Table(table.DetectStepTable).DeleteMany(ctx, filter)
}
