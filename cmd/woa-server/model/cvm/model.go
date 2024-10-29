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

// Package model ...
package model

import (
	"context"

	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// model all model operation interface
type model struct {
	applyOrder ApplyOrder
	cvmInfo    CvmInfo
}

// ApplyOrder get apply order operation interface
func (m *model) ApplyOrder() ApplyOrder {
	return m.applyOrder
}

// CvmInfo get cvm info operation interface
func (m *model) CvmInfo() CvmInfo {
	return m.cvmInfo
}

var operation *model

func init() {
	operation = &model{
		applyOrder: &applyOrder{},
		cvmInfo:    &cvmInfo{},
	}
}

// Operation return all model operation interface
func Operation() *model {
	return operation
}

// Model provides storage interface for operations of models
type Model interface {
	ApplyOrder() ApplyOrder
	CvmInfo() CvmInfo
}

// ApplyOrder apply order operation interface
type ApplyOrder interface {
	// NextSequence returns next apply order sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateApplyOrder creates apply order in db
	CreateApplyOrder(ctx context.Context, inst *types.ApplyOrder) error
	// GetApplyOrder gets apply order by filter from db
	GetApplyOrder(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyOrder, error)
	// CountApplyOrder gets apply order count by filter from db
	CountApplyOrder(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyApplyOrder gets apply order list by filter from db
	FindManyApplyOrder(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*types.ApplyOrder, error)
	// UpdateApplyOrder updates apply order by filter and doc in db
	UpdateApplyOrder(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteApplyOrder deletes apply order from db
	DeleteApplyOrder()
}

// CvmInfo cvm info operation interface
type CvmInfo interface {
	// CreateCvmInfo creates cvm info in db
	CreateCvmInfo(ctx context.Context, inst *types.CvmInfo) error
	// GetCvmInfo gets cvm info by filter from db
	GetCvmInfo(ctx context.Context, filter *mapstr.MapStr) ([]*types.CvmInfo, error)
	// UpdateCvmInfo updates cvm info by filter and doc in db
	UpdateCvmInfo(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteCvmInfo deletes cvm info from db
	DeleteCvmInfo()
}
