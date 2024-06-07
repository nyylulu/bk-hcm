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

package meta

import (
	"hcm/cmd/woa-server/types/meta"
	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	rtypes "hcm/pkg/dal/dao/types/resource-plan"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListDiskType lists disk type.
func (s *service) ListDiskType(_ *rest.Contexts) (interface{}, error) {
	// get disk type members.
	diskTypes := enumor.GetDiskTypeMembers()
	// convert to meta.DiskTypeItem slice.
	details := make([]meta.DiskTypeItem, 0, len(diskTypes))
	for _, diskType := range diskTypes {
		details = append(details, meta.DiskTypeItem{
			DiskType:     diskType,
			DiskTypeName: diskType.Name(),
		})
	}
	return &core.ListResultT[meta.DiskTypeItem]{Details: details}, nil
}

// ListObsProject lists obs project.
func (s *service) ListObsProject(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.ObsProject]{Details: enumor.GetObsProjectMembers()}, nil
}

// ListRegion lists region.
func (s *service) ListRegion(cts *rest.Contexts) (interface{}, error) {
	details, err := s.dao.WoaZone().GetRegionList(cts.Kit, tools.AllExpression())
	if err != nil {
		logs.Errorf("failed to get region list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[rtypes.RegionElem]{Details: details}, nil
}

// ListZone lists zone.
func (s *service) ListZone(cts *rest.Contexts) (interface{}, error) {
	req := new(mtypes.ListZoneReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list zone, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list zone parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := tools.AllExpression()
	if len(req.RegionIDs) > 0 {
		opt = tools.ContainersExpression("region_id", req.RegionIDs)
	}
	details, err := s.dao.WoaZone().GetZoneList(cts.Kit, opt)
	if err != nil {
		logs.Errorf("failed to get zone list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[rtypes.ZoneElem]{Details: details}, nil
}

// ListDeviceClass lists region.
func (s *service) ListDeviceClass(cts *rest.Contexts) (interface{}, error) {
	details, err := s.dao.WoaDeviceType().GetDeviceClassList(cts.Kit, tools.AllExpression())
	if err != nil {
		logs.Errorf("failed to get device class list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &core.ListResultT[string]{Details: details}, nil
}

// ListDeviceType lists device type.
func (s *service) ListDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(mtypes.ListDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to list device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := tools.AllExpression()
	if len(req.DeviceClasses) > 0 {
		opt = tools.ContainersExpression("device_class", req.DeviceClasses)
	}
	devTypeMap, err := s.dao.WoaDeviceType().GetDeviceTypeMap(cts.Kit, opt)
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	details := make([]mtypes.ListDeviceTypeRst, 0, len(devTypeMap))
	for _, v := range devTypeMap {
		details = append(details, mtypes.ListDeviceTypeRst{
			DeviceType: v.DeviceType,
			CoreType:   v.CoreType,
			CpuCore:    v.CpuCore,
			Memory:     v.Memory,
		})
	}

	return &core.ListResultT[mtypes.ListDeviceTypeRst]{Details: details}, nil
}
