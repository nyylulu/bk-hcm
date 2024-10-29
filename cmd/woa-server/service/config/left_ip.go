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

// Package config left ip config
package config

import (
	"hcm/cmd/woa-server/dal/config/table"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetLeftIP gets zone with left ip config list
func (s *service) GetLeftIP(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetLeftIPParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get zone with left ip, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.LeftIP().GetLeftIP(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateLeftIP creates zone with left ip config
func (s *service) CreateLeftIP(cts *rest.Contexts) (interface{}, error) {
	input := new(table.ZoneLeftIP)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.LeftIP().CreateLeftIP(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateLeftIPProperty updates zone with left ip config property
func (s *service) UpdateLeftIPProperty(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpdateLeftIPPropertyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to update zone with left ip, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cond := map[string]interface{}{
		"id": map[string]interface{}{
			pkg.BKDBIN: input.Ids,
		},
	}

	data := input.Property
	// cannot update device id
	delete(data, "id")

	if err := s.logics.LeftIP().UpdateLeftIPBatch(cts.Kit, cond, input.Property); err != nil {
		logs.Errorf("failed to update zone with left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// SyncLeftIP sync zone left ip from yunti
func (s *service) SyncLeftIP(cts *rest.Contexts) (interface{}, error) {
	input := new(types.SyncLeftIPParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to sync left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to sync left ip, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.LeftIP().SyncLeftIP(cts.Kit, input); err != nil {
		logs.Errorf("failed to sync left ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
