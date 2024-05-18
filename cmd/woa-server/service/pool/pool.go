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

package pool

import (
	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/pool/table"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// CreateLaunchTask creates resource pool launch task
func (s *service) CreateLaunchTask(cts *rest.Contexts) (interface{}, error) {
	input := new(types.LaunchReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to create pool launch task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to create pool launch task, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().CreateLaunchTask(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to create pool launch task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateRecallTask creates resource pool recall task
func (s *service) CreateRecallTask(cts *rest.Contexts) (interface{}, error) {
	input := new(types.RecallReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to create pool recall task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to create pool recall task, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().CreateRecallTask(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to create pool recall task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetLaunchTask gets resource pool launch task
func (s *service) GetLaunchTask(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetLaunchTaskReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool launch task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool launch task, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetLaunchTask(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to create pool recall task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallTask gets resource pool recall task
func (s *service) GetRecallTask(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecallTaskReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool recall task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool recall task, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecallTask(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool recall task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetLaunchHost gets resource pool launch host
func (s *service) GetLaunchHost(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetLaunchHostReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool launch host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool launch host, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetLaunchHost(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool launch host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallHost gets resource pool recall host
func (s *service) GetRecallHost(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecallHostReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool recall host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool recall host, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecallHost(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool recall host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetIdleHost gets resource pool idle host
func (s *service) GetIdleHost(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetPoolHostReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool idle host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// set phase filter to IDLE to retrieve idle host only
	input.Phase = []table.PoolHostPhase{table.PoolHostPhaseIdle}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool idle host, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetPoolHost(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool idle host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// DrawHost draw hosts from resource pool
func (s *service) DrawHost(cts *rest.Contexts) (interface{}, error) {
	input := new(types.DrawHostReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to draw host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to draw host, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Pool().DrawHost(cts.Kit, input); err != nil {
		blog.Errorf("failed to draw host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ReturnHost return hosts to resource pool
func (s *service) ReturnHost(cts *rest.Contexts) (interface{}, error) {
	input := new(types.ReturnHostReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to return host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to return host, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Pool().ReturnHost(cts.Kit, input); err != nil {
		blog.Errorf("failed to return host, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateRecallOrder creates resource pool recall order
func (s *service) CreateRecallOrder(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CreateRecallOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to create pool recall order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to create pool recall order, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().CreateRecallOrder(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to create pool recall order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallOrder gets resource pool recall order
func (s *service) GetRecallOrder(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecallOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool recall order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool recall order, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecallOrder(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool recall order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecalledInstance gets resource pool recalled instances
func (s *service) GetRecalledInstance(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecalledInstReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get pool recalled instances, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get pool recalled instances, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecalledInstance(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get pool recalled instances, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallDetail get resource recall task detail info
func (s *service) GetRecallDetail(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecallDetailReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get recall task detail info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get get recall task detail info, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecallDetail(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get recall task detail info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetLaunchMatchDevice get resource launch match devices
func (s *service) GetLaunchMatchDevice(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetLaunchMatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get launch match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get launch match device info, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetLaunchMatchDevice(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get launch match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallMatchDevice get resource recall match devices
func (s *service) GetRecallMatchDevice(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetRecallMatchDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to get recall match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to get recall match device info, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().GetRecallMatchDevice(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to get recall match device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// ResumeRecycleTask resumes recycle task
func (s *service) ResumeRecycleTask(cts *rest.Contexts) (interface{}, error) {
	input := new(types.ResumeRecycleTaskReq)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to resumes recycle task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to resumes recycle task, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Pool().ResumeRecycleTask(cts.Kit, input); err != nil {
		blog.Errorf("failed to resume recycle task, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateGradeCfg create pool grade config
func (s *service) CreateGradeCfg(cts *rest.Contexts) (interface{}, error) {
	input := new(table.GradeCfg)
	if err := cts.DecodeInto(input); err != nil {
		blog.Errorf("failed to create pool grade config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		blog.Errorf("failed to create pool grade config, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Pool().CreateGradeCfg(cts.Kit, input)
	if err != nil {
		blog.Errorf("failed to create pool grade config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetGradeCfg get pool grade config
func (s *service) GetGradeCfg(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.Pool().GetGradeCfg(cts.Kit)
	if err != nil {
		blog.Errorf("failed to get pool grade config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetRecallStatusCfg get recall status config
func (s *service) GetRecallStatusCfg(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"status":      table.RecallStatusReturned,
				"description": table.RecallStatusDescReturned,
			},
			{
				"status":      table.RecallStatusPreChecking,
				"description": table.RecallStatusDescPreChecking,
			},
			{
				"status":      table.RecallStatusPreCheckFailed,
				"description": table.RecallStatusDescPreCheckFailed,
			},
			{
				"status":      table.RecallStatusClearChecking,
				"description": table.RecallStatusDescClearChecking,
			},
			{
				"status":      table.RecallStatusClearCheckFailed,
				"description": table.RecallStatusDescClearCheckFailed,
			},
			{
				"status":      table.RecallStatusReinstalling,
				"description": table.RecallStatusDescReinstalling,
			},
			{
				"status":      table.RecallStatusReinstallFailed,
				"description": table.RecallStatusDescReinstallFailed,
			},
			{
				"status":      table.RecallStatusInitializing,
				"description": table.RecallStatusDescInitializing,
			},
			{
				"status":      table.RecallStatusInitializeFailed,
				"description": table.RecallStatusDescInitializeFailed,
			},
			{
				"status":      table.RecallStatusDataDeleting,
				"description": table.RecallStatusDescDataDeleting,
			},
			{
				"status":      table.RecallStatusDataDeleteFailed,
				"description": table.RecallStatusDescDataDeleteFailed,
			},
			{
				"status":      table.RecallStatusConfChecking,
				"description": table.RecallStatusDescConfChecking,
			},
			{
				"status":      table.RecallStatusConfCheckFailed,
				"description": table.RecallStatusDescConfCheckFailed,
			},
			{
				"status":      table.RecallStatusTransiting,
				"description": table.RecallStatusDescTransiting,
			},
			{
				"status":      table.RecallStatusTransitFailed,
				"description": table.RecallStatusDescTransitFailed,
			},
			{
				"status":      table.RecallStatusDone,
				"description": table.RecallStatusDescDone,
			},
			{
				"status":      table.RecallStatusTerminate,
				"description": table.RecallStatusDescTerminate,
			},
		},
	}

	return rst, nil
}

// GetTaskStatusCfg get task status config
func (s *service) GetTaskStatusCfg(cts *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"status":      table.OpTaskPhaseInit,
				"description": table.OpTaskPhaseDescInit,
			},
			{
				"status":      table.OpTaskPhaseRunning,
				"description": table.OpTaskPhaseDescRunning,
			},
			{
				"status":      table.OpTaskPhaseSuccess,
				"description": table.OpTaskPhaseDescSuccess,
			},
			{
				"status":      table.OpTaskPhaseFailed,
				"description": table.OpTaskPhaseDescFailed,
			},
		},
	}

	return rst, nil
}

// GetDeviceType get supported device type list
func (s *service) GetDeviceType(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.Pool().GetDeviceType(cts.Kit)
	if err != nil {
		blog.Errorf("failed to get pool supported device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
