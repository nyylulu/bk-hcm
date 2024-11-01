/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package enumor

import "fmt"

// AppliedType is rolling applied record type.
type AppliedType string

const (
	// NormalAppliedType is rolling applied record type normal.
	NormalAppliedType AppliedType = "normal"
	// ResourcePoolAppliedType is rolling applied record type resource pool.
	ResourcePoolAppliedType AppliedType = "resource_pool"
	// CvmProduceAppliedType is rolling applied record type cvm product.
	CvmProduceAppliedType AppliedType = "cvm_product"
)

// Validate AppliedType.
func (t AppliedType) Validate() error {
	switch t {
	case NormalAppliedType, ResourcePoolAppliedType, CvmProduceAppliedType:
	default:
		return fmt.Errorf("unsupported rolling applied record type: %s", t)
	}

	return nil
}

// ReturnedWay is rolling returned way.
type ReturnedWay string

const (
	// CrpReturnedWay is rolling returned way crp.
	CrpReturnedWay ReturnedWay = "crp"
	// ResourcePoolReturnedWay is rolling returned way resource pool.
	ResourcePoolReturnedWay ReturnedWay = "resource_pool"
)

// Validate ReturnedWay.
func (t ReturnedWay) Validate() error {
	switch t {
	case CrpReturnedWay, ResourcePoolReturnedWay:
	default:
		return fmt.Errorf("unsupported rolling returned way: %s", t)
	}

	return nil
}

// QuotaOffsetAdjustType is quota offset adjust type.
type QuotaOffsetAdjustType string

const (
	// IncreaseOffsetAdjustType is increase quota offset adjust type.
	IncreaseOffsetAdjustType QuotaOffsetAdjustType = "increase"
	// DecreaseOffsetAdjustType is decrease quota offset adjust type.
	DecreaseOffsetAdjustType QuotaOffsetAdjustType = "decrease"
)

// Validate QuotaOffsetAdjustType
func (q QuotaOffsetAdjustType) Validate() error {
	switch q {
	case IncreaseOffsetAdjustType, DecreaseOffsetAdjustType:
	default:
		return fmt.Errorf("unsupported quota offset adjust type: %s", q)
	}

	return nil
}
