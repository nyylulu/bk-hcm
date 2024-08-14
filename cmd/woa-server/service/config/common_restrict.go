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

// Package config implements config service
package config

import (
	"hcm/cmd/woa-server/common/mapstr"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetAffinity gets anti affinity level config list
func (s *service) GetAffinity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetAffinityParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get affinity list, err: %v, rid: %s", err, cts.Kit.Rid)
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
		return []*types.AffinityInfo{{
			Level:       types.AntiNone,
			Description: types.Description[types.AntiNone],
		}, {
			Level:       types.AntiModule,
			Description: types.Description[types.AntiModule],
		}, {
			Level:       types.AntiRack,
			Description: types.Description[types.AntiRack],
		},
		}
	}

	return []*types.AffinityInfo{{
		Level:       types.AntiNone,
		Description: types.Description[types.AntiNone],
	}, {
		Level:       types.AntiCampus,
		Description: types.Description[types.AntiCampus],
	},
	}
}

// getQcloudAffinity gets qcloud anti affinity level config list
func (s *service) getQcloudAffinity(hasZone bool) []*types.AffinityInfo {
	if hasZone {
		return []*types.AffinityInfo{{
			Level:       types.AntiNone,
			Description: types.Description[types.AntiNone],
		},
		}
	}

	return []*types.AffinityInfo{{
		Level:       types.AntiNone,
		Description: types.Description[types.AntiNone],
	}, {
		Level:       types.AntiCampus,
		Description: types.Description[types.AntiCampus],
	},
	}
}

// getAllAffinity gets all anti affinity level config list
func (s *service) getAllAffinity() []*types.AffinityInfo {
	return []*types.AffinityInfo{{
		Level:       types.AntiNone,
		Description: types.Description[types.AntiNone],
	}, {
		Level:       types.AntiCampus,
		Description: types.Description[types.AntiCampus],
	}, {
		Level:       types.AntiModule,
		Description: types.Description[types.AntiModule],
	}, {
		Level:       types.AntiRack,
		Description: types.Description[types.AntiRack],
	},
	}
}

// GetApplyStage gets apply stage config list
func (s *service) GetApplyStage(_ *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{{
			"stage":       "UNCOMMIT",
			"description": "未提交",
		}, {
			"stage":       "AUDIT",
			"description": "待审核",
		}, {
			"stage":       "TERMINATE",
			"description": "终止",
		}, {
			"stage":       "RUNNING",
			"description": "备货中",
		}, {
			"stage":       "SUSPEND",
			"description": "备货异常",
		}, {
			"stage":       "DONE",
			"description": "完成",
		},
		},
	}

	return rst, nil
}
