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

// Package plan ...
package plan

import (
	dataproto "hcm/pkg/api/data-service"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListDeviceType list device type.
func (s *service) ListDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.WoaDeviceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	//// 权限校验
	//authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	//if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
	//	return nil, err
	//}

	result, err := s.client.DataService().Global.ResourcePlan.ListWoaDeviceType(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// CreateDeviceType create device type.
func (s *service) CreateDeviceType(cts *rest.Contexts) (any, error) {
	req := new(rpproto.WoaDeviceTypeBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	result, err := s.client.DataService().Global.ResourcePlan.BatchCreateWoaDeviceType(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to create device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	return result.IDs, nil
}

// UpdateDeviceType update device type.
func (s *service) UpdateDeviceType(cts *rest.Contexts) (any, error) {
	req := new(rpproto.WoaDeviceTypeBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to update device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate update device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	err := s.client.DataService().Global.ResourcePlan.BatchUpdateWoaDeviceType(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to update device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// DeleteDeviceType delete device type.
func (s *service) DeleteDeviceType(cts *rest.Contexts) (any, error) {
	req := new(dataproto.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to delete device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate delete device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	err := s.client.DataService().Global.ResourcePlan.BatchDeleteWoaDeviceType(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to delete device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// SyncDeviceType sync device type.
func (s *service) SyncDeviceType(cts *rest.Contexts) (any, error) {
	req := new(rpproto.WoaDeviceTypeSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to sync device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate sync device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlan, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	err := s.planController.SyncDeviceTypesFromCRP(cts.Kit, req.DeviceTypes)
	if err != nil {
		logs.Errorf("failed to sync device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
