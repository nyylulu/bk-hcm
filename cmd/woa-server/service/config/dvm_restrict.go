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
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/pkg/rest"
)

// GetDvmImage gets dvm image config list
func (s *service) GetDvmImage(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{"hub.oa.com/library/tlinux1.2:v1.17", "hub.oa.com/library/tlinux2.2:v1.6"},
	}

	return rst, nil
}

// GetDvmKernel gets dvm kernel config list
func (s *service) GetDvmKernel(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{"3.10.107-1-tlinux2-0048"},
	}

	return rst, nil
}

// GetDvmMountPath gets dvm mount path config list
func (s *service) GetDvmMountPath(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{"/data1", "/data"},
	}

	return rst, nil
}

// GetDvmIdcDeviceGroup gets dvm idc device group config list
func (s *service) GetDvmIdcDeviceGroup(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			mapstr.MapStr{
				"type":        "GAMESERVER",
				"description": "通用计算型",
			},
			mapstr.MapStr{
				"type":        "DBSERVICE",
				"description": "IO存储型",
			},
			mapstr.MapStr{
				"type":        "HIGHFREQ",
				"description": "高性能计算型",
			},
		},
	}

	return rst, nil
}

// GetDvmQcloudDeviceGroup gets dvm qcloud device group config list
func (s *service) GetDvmQcloudDeviceGroup(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			mapstr.MapStr{
				"type":         "GAMESERVER",
				"description":  "通用计算型",
				"cpu_provider": []string{"Intel", "AMD", "无需求"},
			},
			mapstr.MapStr{
				"type":         "DBSERVICE",
				"description":  "IO存储型",
				"cpu_provider": []string{"Intel"},
			},
			mapstr.MapStr{
				"type":         "HIGHFREQ",
				"description":  "高性能计算型",
				"cpu_provider": []string{"Intel"},
			},
		},
	}

	return rst, nil
}
