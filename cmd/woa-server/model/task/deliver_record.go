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

// Package model implements all db related operations.
package model

import (
	"context"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
)

type deliverRecord struct {
}

// NextSequence returns next apply order deliver record sequence id from db
func (d *deliverRecord) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, common.BKTableNameDeliverRecord)
}

// CreateDeliverRecord creates apply order deliver record in db
func (d *deliverRecord) CreateDeliverRecord(ctx context.Context, inst *types.DeliverRecord) error {
	return mongodb.Client().Table(common.BKTableNameDeliverRecord).Insert(ctx, inst)
}

// GetDeliverRecord gets apply order deliver record by filter from db
func (d *deliverRecord) GetDeliverRecord(ctx context.Context, filter *mapstr.MapStr) (*types.DeliverRecord, error) {
	inst := new(types.DeliverRecord)

	if err := mongodb.Client().Table(common.BKTableNameDeliverRecord).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountDeliverRecord gets apply order deliver record count by filter from db
func (d *deliverRecord) CountDeliverRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(common.BKTableNameDeliverRecord).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDeliverRecord gets deliver record list by filter from db
func (d *deliverRecord) FindManyDeliverRecord(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*types.DeliverRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(common.BKTableNameDeliverRecord).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("ip")
	}

	insts := make([]*types.DeliverRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDeliverRecord updates apply order deliver record by filter and doc in db
func (d *deliverRecord) UpdateDeliverRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(common.BKTableNameDeliverRecord).Update(ctx, filter, doc)
}

// DeleteDeliverRecord deletes apply order deliver record from db
func (d *deliverRecord) DeleteDeliverRecord() {

}
