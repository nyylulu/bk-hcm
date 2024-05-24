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

// Package config implements cvm restrict config
package config

import (
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/pkg/rest"
)

// GetCvmDiskType gets cvm disk type config list
func (s *service) GetCvmDiskType(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"count": 2,
		"info": []mapstr.MapStr{
			{
				"disk_type": "CLOUD_SSD",
				"disk_name": "SSD云硬盘",
			},
			{
				"disk_type": "CLOUD_PREMIUM",
				"disk_name": "高性能云盘",
			},
		},
	}

	return rst, nil
}
