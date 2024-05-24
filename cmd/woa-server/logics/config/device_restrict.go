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
	"fmt"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// DeviceRestrictIf provides management interface for operations of device restrict config
type DeviceRestrictIf interface {
	// GetDeviceRestrict get device restrict type config list
	GetDeviceRestrict(kt *kit.Kit) (*types.GetDeviceRestrictResult, error)
	// CreateDeviceRestrict creates device restrict type config
	CreateDeviceRestrict(kt *kit.Kit, input *types.DeviceRestrict) (mapstr.MapStr, error)
	// UpdateDeviceRestrict updates device restrict type config
	UpdateDeviceRestrict(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteDeviceRestrict deletes device restrict type config
	DeleteDeviceRestrict(kt *kit.Kit, instId int64) error
}

// NewDeviceRestrictOp creates a device restrict interface
func NewDeviceRestrictOp() DeviceRestrictIf {
	return &deviceRestrict{}
}

type deviceRestrict struct {
}

// GetDeviceRestrict get device restrict type config list
func (d *deviceRestrict) GetDeviceRestrict(kt *kit.Kit) (*types.GetDeviceRestrictResult, error) {
	filter := &mapstr.MapStr{}
	inst, err := config.Operation().DeviceRestrict().GetDeviceRestrict(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	if inst == nil {
		return nil, fmt.Errorf("get no device restrict")
	}

	rst := &types.GetDeviceRestrictResult{
		Cpu:  inst.Cpu,
		Mem:  inst.Mem,
		Disk: inst.Disk,
	}

	return rst, nil
}

// CreateDeviceRestrict creates device restrict type config
func (d *deviceRestrict) CreateDeviceRestrict(kt *kit.Kit, input *types.DeviceRestrict) (mapstr.MapStr, error) {
	id, err := config.Operation().DeviceRestrict().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create device restrict, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().DeviceRestrict().CreateDeviceRestrict(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create device restrict, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateDeviceRestrict updates device restrict type config
func (d *deviceRestrict) UpdateDeviceRestrict(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().DeviceRestrict().UpdateDeviceRestrict(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update device restrict, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteDeviceRestrict deletes device restrict type config
func (d *deviceRestrict) DeleteDeviceRestrict(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().DeviceRestrict().DeleteDeviceRestrict(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete device restrict, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
