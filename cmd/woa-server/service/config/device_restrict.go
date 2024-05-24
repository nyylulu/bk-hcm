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

// Package config device restrict config
package config

import (
	"strconv"

	"hcm/cmd/woa-server/common/mapstr"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetDeviceRestrict gets device restrict config list
func (s *service) GetDeviceRestrict(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.DeviceRestrict().GetDeviceRestrict(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get device restrict list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateDeviceRestrict creates device restrict config
func (s *service) CreateDeviceRestrict(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.DeviceRestrict)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create device restrict, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.DeviceRestrict().CreateDeviceRestrict(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create device restrict, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateDeviceRestrict updates device restrict config
func (s *service) UpdateDeviceRestrict(cts *rest.Contexts) (interface{}, error) {
	inputData := new(mapstr.MapStr)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to update device restrict, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.DeviceRestrict().UpdateDeviceRestrict(cts.Kit, instId, inputData); err != nil {
		logs.Errorf("failed to update device restrict, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteDeviceRestrict deletes device restrict config
func (s *service) DeleteDeviceRestrict(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.logics.DeviceRestrict().DeleteDeviceRestrict(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete device restrict, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
