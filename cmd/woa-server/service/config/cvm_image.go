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

// Package config cvm image config
package config

import (
	"strconv"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCvmImage gets cvm image config list
func (s *service) GetCvmImage(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetCvmImageParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get cvm image list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cond := mapstr.MapStr{}
	if len(input.Region) > 0 {
		cond["region"] = mapstr.MapStr{pkg.BKDBIN: input.Region}
	}

	rst, err := s.logics.CvmImage().GetCvmImage(cts.Kit, &cond)
	if err != nil {
		logs.Errorf("failed to get cvm image list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateCvmImage creates cvm image config
func (s *service) CreateCvmImage(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.CvmImage)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create cvm image, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.CvmImage().CreateCvmImage(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create cvm image, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateCvmImage updates cvm image config
func (s *service) UpdateCvmImage(cts *rest.Contexts) (interface{}, error) {
	inputData := new(mapstr.MapStr)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to update cvm image, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.CvmImage().UpdateCvmImage(cts.Kit, instId, inputData); err != nil {
		logs.Errorf("failed to update cvm image, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteCvmImage deletes cvm image config
func (s *service) DeleteCvmImage(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.CvmImage().DeleteCvmImage(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete cvm image, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
