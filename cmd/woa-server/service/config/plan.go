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

// Package config plan config
package config

import (
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/rest"
)

// GetPlanCoreType gets plan core type config list
func (s *service) GetPlanCoreType(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"小核心",
			"大核心",
		},
	}

	return rst, nil
}

// GetPlanDiskType gets plan disk type config list
func (s *service) GetPlanDiskType(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"高性能云硬盘",
			"SSD云硬盘",
		},
	}

	return rst, nil
}

// GetPlanOrderType gets plan order type config list
func (s *service) GetPlanOrderType(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"需求追加",
		},
	}

	return rst, nil
}

// GetPlanDeviceGroup gets plan device group config list
func (s *service) GetPlanDeviceGroup(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"标准型",
			"高IO型",
		},
	}

	return rst, nil
}
