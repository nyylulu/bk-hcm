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

package config

import (
	"context"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
)

type cvmImage struct {
}

// NextSequence returns next cvm image config sequence id from db
func (i *cvmImage) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameCfgCvmImage)
}

// CreateCvmImage creates cvm image config in db
func (i *cvmImage) CreateCvmImage(ctx context.Context, inst *types.CvmImage) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmImage).Insert(ctx, inst)
}

// GetCvmImage gets cvm image config by filter from db
func (i *cvmImage) GetCvmImage(ctx context.Context, filter *mapstr.MapStr) (*types.CvmImage, error) {
	inst := new(types.CvmImage)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgCvmImage).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// FindManyCvmImage gets cvm image config list by filter from db
func (i *cvmImage) FindManyCvmImage(ctx context.Context, filter *mapstr.MapStr) ([]*types.CvmImage, error) {
	insts := make([]*types.CvmImage, 0)

	if err := mongodb.Client().Table(pkg.BKTableNameCfgCvmImage).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateCvmImage updates cvm image config by filter and doc in db
func (i *cvmImage) UpdateCvmImage(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmImage).Update(ctx, filter, doc)
}

// DeleteCvmImage deletes cvm image config from db
func (i *cvmImage) DeleteCvmImage(ctx context.Context, filter *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameCfgCvmImage).Delete(ctx, filter)
}
