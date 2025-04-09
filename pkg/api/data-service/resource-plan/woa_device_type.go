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

// Package resourceplan ...
package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	wdttable "hcm/pkg/dal/table/resource-plan/woa-device-type"
)

// WoaDeviceTypeListReq list request
type WoaDeviceTypeListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *WoaDeviceTypeListReq) Validate() error {
	return r.ListReq.Validate()
}

// WoaDeviceTypeListResult list result
type WoaDeviceTypeListResult types.ListResult[wdttable.WoaDeviceTypeTable]

// WoaDeviceTypeBatchCreateReq create request
type WoaDeviceTypeBatchCreateReq struct {
	DeviceTypes []wdttable.WoaDeviceTypeTable `json:"device_types" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *WoaDeviceTypeBatchCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// WoaDeviceTypeBatchUpdateReq batch update request
type WoaDeviceTypeBatchUpdateReq struct {
	DeviceTypes []wdttable.WoaDeviceTypeTable `json:"device_types" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *WoaDeviceTypeBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// WoaDeviceTypeSyncReq sync request
type WoaDeviceTypeSyncReq struct {
	DeviceTypes []string `json:"device_types" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *WoaDeviceTypeSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}
