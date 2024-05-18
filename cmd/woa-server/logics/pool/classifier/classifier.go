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

// Package classifier implements device classifier which helps to tell host resource type.
package classifier

import (
	"strings"

	"hcm/cmd/woa-server/dal/pool/table"
)

// GetResType get host resource type
func GetResType(assetId string) table.ResourceType {
	if IsUnsupportedDevice(assetId) {
		return table.ResourceTypeUnsupported
	}

	if IsQcloudCvm(assetId) {
		return table.ResourceTypeCvm
	}

	if isIdcPm(assetId) {
		return table.ResourceTypePm
	}

	return table.ResourceTypeOthers
}

// IsUnsupportedDevice verify if given host is unsupported device
// 固资号非 TC、TYSV或TDKIEG 开头的均为不支持资源类型
func IsUnsupportedDevice(assetId string) bool {
	// cvm
	if strings.HasPrefix(assetId, "TC") {
		return false
	}

	// idc physical machine
	if strings.HasPrefix(assetId, "TYSV") {
		return false
	}

	// 算力特殊机型
	if strings.HasPrefix(assetId, "TDKIEG") {
		return false
	}

	return true
}

// IsQcloudCvm verify if given host is qcloud cvm device
// 固资号为 TC*** （排除掉 TC***-VM****) 的是CVM机型
func IsQcloudCvm(assetId string) bool {
	if !strings.HasPrefix(assetId, "TC") {
		return false
	}

	dashIdx := strings.Index(assetId, "-")
	if dashIdx < 0 {
		return true
	}

	// exclude qcloud docker vm
	if strings.HasPrefix(assetId[dashIdx+1:], "VM") {
		return false
	}

	return true
}

// isIdcPm verify if given host is idc physical machine
// 固资号为 TYSV*** （排除掉 TYSV***-VM****) 的是物理机机型
func isIdcPm(assetId string) bool {
	if !strings.HasPrefix(assetId, "TYSV") {
		return false
	}

	dashIdx := strings.Index(assetId, "-")
	if dashIdx < 0 {
		return true
	}

	// exclude idc docker vm and kvm
	if strings.HasPrefix(assetId[dashIdx+1:], "VM") {
		return false
	}

	return true
}
