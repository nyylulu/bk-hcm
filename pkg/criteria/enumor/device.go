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

// DiskType is disk type.
type DiskType string

const (
	// DiskPremium disk cloud premium.
	DiskPremium DiskType = "CLOUD_PREMIUM"
	// DiskSSD disk cloud ssd.
	DiskSSD DiskType = "CLOUD_SSD"
)

// Validate DiskType.
func (t DiskType) Validate() error {
	switch t {
	case DiskPremium:
	case DiskSSD:
	default:
		return fmt.Errorf("unsupported disk type: %s", t)
	}

	return nil
}

// diskTypeNameMap records disk type corresponding name.
var diskTypeNameMap = map[DiskType]string{
	DiskPremium: "高性能云硬盘",
	DiskSSD:     "SSD云硬盘",
}

// Name return disk type name.
func (t DiskType) Name() string {
	return diskTypeNameMap[t]
}

// GetDiskTypeMembers get DiskType's members.
func GetDiskTypeMembers() []DiskType {
	return []DiskType{DiskPremium, DiskSSD}
}
