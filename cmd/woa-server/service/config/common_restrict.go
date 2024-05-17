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
	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/common/mapstr"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/rest"
)

// GetAffinity gets anti affinity level config list
func (s *service) GetAffinity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetAffinityParam)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get affinity list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst := new(types.GetAffinityRst)
	switch input.ResourceType {
	case types.ResourceTypePm, types.ResourceTypeIdcDvm:
		rst.Info = s.getIdcAffinity(input.HasZone)
	case types.ResourceTypeQcloudDvm:
		rst.Info = s.getQcloudAffinity(input.HasZone)
	default:
		rst.Info = s.getAllAffinity()
	}

	return rst, nil
}

// getIdcAffinity gets idc anti affinity level config list
func (s *service) getIdcAffinity(hasZone bool) []*types.AffinityInfo {
	if hasZone {
		return []*types.AffinityInfo{
			&types.AffinityInfo{
				Level:       types.AntiNone,
				Description: types.Description[types.AntiNone],
			},
			&types.AffinityInfo{
				Level:       types.AntiModule,
				Description: types.Description[types.AntiModule],
			},
			&types.AffinityInfo{
				Level:       types.AntiRack,
				Description: types.Description[types.AntiRack],
			},
		}
	}

	return []*types.AffinityInfo{
		&types.AffinityInfo{
			Level:       types.AntiNone,
			Description: types.Description[types.AntiNone],
		},
		&types.AffinityInfo{
			Level:       types.AntiCampus,
			Description: types.Description[types.AntiCampus],
		},
	}
}

// getQcloudAffinity gets qcloud anti affinity level config list
func (s *service) getQcloudAffinity(hasZone bool) []*types.AffinityInfo {
	if hasZone {
		return []*types.AffinityInfo{
			&types.AffinityInfo{
				Level:       types.AntiNone,
				Description: types.Description[types.AntiNone],
			},
		}
	}

	return []*types.AffinityInfo{
		&types.AffinityInfo{
			Level:       types.AntiNone,
			Description: types.Description[types.AntiNone],
		},
		&types.AffinityInfo{
			Level:       types.AntiCampus,
			Description: types.Description[types.AntiCampus],
		},
	}
}

// getAllAffinity gets all anti affinity level config list
func (s *service) getAllAffinity() []*types.AffinityInfo {
	return []*types.AffinityInfo{
		&types.AffinityInfo{
			Level:       types.AntiNone,
			Description: types.Description[types.AntiNone],
		},
		&types.AffinityInfo{
			Level:       types.AntiCampus,
			Description: types.Description[types.AntiCampus],
		},
		&types.AffinityInfo{
			Level:       types.AntiModule,
			Description: types.Description[types.AntiModule],
		},
		&types.AffinityInfo{
			Level:       types.AntiRack,
			Description: types.Description[types.AntiRack],
		},
	}
}

// GetApplyStage gets apply stage config list
func (s *service) GetApplyStage(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			mapstr.MapStr{
				"stage":       "UNCOMMIT",
				"description": "未提交",
			},
			mapstr.MapStr{
				"stage":       "AUDIT",
				"description": "待审核",
			},
			mapstr.MapStr{
				"stage":       "TERMINATE",
				"description": "终止",
			},
			mapstr.MapStr{
				"stage":       "RUNNING",
				"description": "备货中",
			},
			mapstr.MapStr{
				"stage":       "SUSPEND",
				"description": "备货异常",
			},
			mapstr.MapStr{
				"stage":       "DONE",
				"description": "完成",
			},
		},
	}

	return rst, nil
}
