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

package model

import (
	"context"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
)

type initRecord struct {
}

// NextSequence returns next apply order init record sequence id from db
func (i *initRecord) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameInitRecord)
}

// CreateInitRecord creates apply order init record in db
func (i *initRecord) CreateInitRecord(ctx context.Context, inst *types.InitRecord) error {
	return mongodb.Client().Table(common.BKTableNameInitRecord).Insert(ctx, inst)
}

// GetInitRecord gets apply order init record by filter from db
func (i *initRecord) GetInitRecord(ctx context.Context, filter *mapstr.MapStr) (*types.InitRecord, error) {
	inst := new(types.InitRecord)

	if err := mongodb.Client().Table(common.BKTableNameInitRecord).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountInitRecord gets apply order init record count by filter from db
func (i *initRecord) CountInitRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameInitRecord).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyInitRecord gets init record list by filter from db
func (i *initRecord) FindManyInitRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.InitRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(common.BKTableNameInitRecord).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("ip")
	}

	insts := make([]*types.InitRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateInitRecord updates apply order init record by filter and doc in db
func (i *initRecord) UpdateInitRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameInitRecord).Update(ctx, filter, doc)
}

// DeleteInitRecord deletes apply order init record from db
func (i *initRecord) DeleteInitRecord() {

}
