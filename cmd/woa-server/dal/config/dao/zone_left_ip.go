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

	"hcm/cmd/woa-server/dal/config/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// ZoneLeftIP supplies all the zone with left ip related operations.
type ZoneLeftIP interface {
	// NextSequence returns next zone with left ip sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateZoneLeftIP creates zone with left ip in db
	CreateZoneLeftIP(ctx context.Context, inst *table.ZoneLeftIP) error
	// GetZoneLeftIP gets zone with left ip by filter from db
	GetZoneLeftIP(ctx context.Context, filter *mapstr.MapStr) (*table.ZoneLeftIP, error)
	// CountZoneLeftIP gets zone with left ip count by filter from db
	CountZoneLeftIP(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyZoneLeftIP gets zone with left ip list by filter from db
	FindManyZoneLeftIP(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.ZoneLeftIP, error)
	// UpdateZoneLeftIP updates zone with left ip by filter and doc in db
	UpdateZoneLeftIP(ctx context.Context, filter, doc map[string]interface{}) error
}

var _ ZoneLeftIP = new(zoneLeftIPDao)

type zoneLeftIPDao struct {
}

// NextSequence returns next zone with left ip sequence id from db
func (zo *zoneLeftIPDao) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, table.ZoneLeftIPTable)
}

// CreateZoneLeftIP creates zone with left ip in db
func (zo *zoneLeftIPDao) CreateZoneLeftIP(ctx context.Context, inst *table.ZoneLeftIP) error {
	return mongodb.Client().Table(table.ZoneLeftIPTable).Insert(ctx, inst)
}

// GetZoneLeftIP gets zone with left ip by filter from db
func (zo *zoneLeftIPDao) GetZoneLeftIP(ctx context.Context, filter *mapstr.MapStr) (*table.ZoneLeftIP, error) {
	inst := new(table.ZoneLeftIP)

	if err := mongodb.Client().Table(table.ZoneLeftIPTable).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountZoneLeftIP gets zone with left ip count by filter from db
func (zo *zoneLeftIPDao) CountZoneLeftIP(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.ZoneLeftIPTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyZoneLeftIP gets zone with left ip list by filter from db
func (zo *zoneLeftIPDao) FindManyZoneLeftIP(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.ZoneLeftIP, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.ZoneLeftIPTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("id")
	}

	insts := make([]*table.ZoneLeftIP, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateZoneLeftIP updates zone with left ip by filter and doc in db
func (zo *zoneLeftIPDao) UpdateZoneLeftIP(ctx context.Context, filter, doc map[string]interface{}) error {
	return mongodb.Client().Table(table.ZoneLeftIPTable).Update(ctx, filter, doc)
}
