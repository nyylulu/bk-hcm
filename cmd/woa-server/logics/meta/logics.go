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

// Package meta is meta logics related package.
package meta

import (
	"hcm/cmd/woa-server/thirdparty/esb"
	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
)

// Logics provides management interface for operations of model and instance and related resources like association
type Logics interface {
	// GetBizsByOpProd get bizs by op product.
	GetBizsByOpProd(kt *kit.Kit, prodID int64) ([]mtypes.Biz, error)
	// GetOpProducts get op products.
	GetOpProducts(kt *kit.Kit) ([]mtypes.OpProduct, error)
	// GetPlanProducts get op products.
	GetPlanProducts(kt *kit.Kit) ([]mtypes.PlanProduct, error)
}

type logics struct {
	esbClient  esb.Client
	authorizer auth.Authorizer
}

// New create a logics manager
func New(esbClient esb.Client, authorizer auth.Authorizer) Logics {
	return &logics{
		esbClient:  esbClient,
		authorizer: authorizer,
	}
}
