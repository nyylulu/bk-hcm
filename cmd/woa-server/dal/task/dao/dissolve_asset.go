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

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/driver/mongodb"
)

// DissolveAsset resource dissolve asset operation interface
type DissolveAsset interface {
	// CountDissolveAsset gets dissolve asset count by filter from db
	CountDissolveAsset(ctx context.Context, filter map[string]interface{}) (uint64, error)
}

var _ DissolveAsset = new(dissolveAssetDao)

type dissolveAssetDao struct {
}

// CountDissolveAsset gets dissolve asset count by filter from db
func (da *dissolveAssetDao) CountDissolveAsset(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(table.DissolveAssetTable).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}
