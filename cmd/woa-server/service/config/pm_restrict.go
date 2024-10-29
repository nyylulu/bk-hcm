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

// Package config pm restrict config
package config

import (
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/rest"
)

// GetPmOstype gets physical machine os type config list
func (s *service) GetPmOstype(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"Tencent tlinux release 1.2 (tkernel2)",
			"Tencent tlinux release 2.2 (Final)",
			"Tencent tlinux release 2.6 (tkernel4)",
			"TencentOS Server 3 (Final)",
			"XServer V08_64",
			"XServer V12_64",
			"XServer V16_64",
			"Tencent tlinux release 2.4 for ARM64",
		},
	}

	return rst, nil
}

// GetPmIsp gets physical machine isp config list
func (s *service) GetPmIsp(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"电信",
			"联通",
			"移动",
			"CAP",
		},
	}

	return rst, nil
}

// GetPmRaidtype gets physical machine raid type config list
func (s *service) GetPmRaidtype(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []string{
			"NORAID",
			"RAID1",
			"RAID5",
		},
	}

	return rst, nil
}
