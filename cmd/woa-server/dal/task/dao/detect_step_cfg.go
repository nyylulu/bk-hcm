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

// Package dao implements the access to the underlying storage service
package dao

import (
	"context"

	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
)

// DetectStepCfg supplies all the recycle detection step config related operations.
type DetectStepCfg interface {
	// GetDetectStepConfig gets recycle detection step config by filter from db
	GetDetectStepConfig(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*table.DetectStepCfg, error)
}

var _ DetectStepCfg = new(detectStepCfgDao)

type detectStepCfgDao struct {
}

// GetDetectStepConfig gets recycle detection step config by filter from db
func (dsc *detectStepCfgDao) GetDetectStepConfig(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*table.DetectStepCfg, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(table.DetectStepCfgTable).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("sequence")
	}

	insts := make([]*table.DetectStepCfg, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}
