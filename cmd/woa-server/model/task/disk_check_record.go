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

// Package model implements disk check record model
package model

import (
	"context"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

type diskCheckRecord struct {
}

// NextSequence returns next apply order disk check record sequence id from db
func (i *diskCheckRecord) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameDiskCheckRecord)
}

// CreateDiskCheckRecord creates apply order disk check record in db
func (i *diskCheckRecord) CreateDiskCheckRecord(ctx context.Context, inst *types.DiskCheckRecord) error {
	return mongodb.Client().Table(pkg.BKTableNameDiskCheckRecord).Insert(ctx, inst)
}

// GetDiskCheckRecord gets apply order disk check record by filter from db
func (i *diskCheckRecord) GetDiskCheckRecord(ctx context.Context, filter *mapstr.MapStr) (*types.DiskCheckRecord,
	error) {

	inst := new(types.DiskCheckRecord)

	if err := mongodb.Client().Table(pkg.BKTableNameDiskCheckRecord).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountDiskCheckRecord gets apply order disk check record count by filter from db
func (i *diskCheckRecord) CountDiskCheckRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(pkg.BKTableNameDiskCheckRecord).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDiskCheckRecord gets disk check record list by filter from db
func (i *diskCheckRecord) FindManyDiskCheckRecord(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*types.DiskCheckRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(pkg.BKTableNameDiskCheckRecord).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("ip")
	}

	insts := make([]*types.DiskCheckRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDiskCheckRecord updates apply order disk check record by filter and doc in db
func (i *diskCheckRecord) UpdateDiskCheckRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameDiskCheckRecord).Update(ctx, filter, doc)
}

// DeleteDiskCheckRecord deletes apply order disk check record from db
func (i *diskCheckRecord) DeleteDiskCheckRecord() {

}
