/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package plan

import (
	"fmt"
	"sync"
	"time"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/slice"
)

// DeviceTypesMap cache of device_type, reducing the pressure of MySQL.
type DeviceTypesMap struct {
	lock        sync.RWMutex
	dao         dao.Set
	DeviceTypes map[string]wdt.WoaDeviceTypeTable
	TTL         time.Time
}

// NewDeviceTypesMap ...
func NewDeviceTypesMap(dao dao.Set) *DeviceTypesMap {
	return &DeviceTypesMap{
		dao:         dao,
		DeviceTypes: make(map[string]wdt.WoaDeviceTypeTable),
		TTL:         time.Now(),
	}
}

func (d *DeviceTypesMap) updateDeviceTypesMap(kt *kit.Kit) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	deviceTypeMap, err := d.dao.WoaDeviceType().GetDeviceTypeMap(kt, tools.AllExpression())
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	d.DeviceTypes = deviceTypeMap
	d.TTL = time.Now().Add(1 * time.Minute)
	return nil
}

// GetDeviceTypes get device type map from cache.
func (d *DeviceTypesMap) GetDeviceTypes(kt *kit.Kit) (map[string]wdt.WoaDeviceTypeTable, error) {
	d.lock.RLock()
	res := make(map[string]wdt.WoaDeviceTypeTable)
	if time.Now().After(d.TTL) {
		d.lock.RUnlock()
		err := d.updateDeviceTypesMap(kt)
		if err != nil {
			return nil, err
		}
		d.lock.RLock()
	}

	defer d.lock.RUnlock()
	for k := range d.DeviceTypes {
		res[k] = d.DeviceTypes[k]
	}
	return res, nil
}

// IsDeviceMatched return whether each device type in deviceTypeSlice can use deviceType's resource plan.
func (c *Controller) IsDeviceMatched(kt *kit.Kit, deviceTypeSlice []string, deviceType string) ([]bool, error) {
	// get device type map.
	deviceTypeMap, err := c.deviceTypesMap.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]bool, len(deviceTypeSlice))
	for idx, ele := range deviceTypeSlice {
		// if ele and device type are equal, then they are matched.
		if ele == deviceType {
			result[idx] = true
		}

		if _, ok := deviceTypeMap[ele]; !ok {
			continue
		}

		if _, ok := deviceTypeMap[deviceType]; !ok {
			continue
		}

		// if device family and core type of ele and device type are equal, then they are matched.
		if deviceTypeMap[ele].DeviceFamily == deviceTypeMap[deviceType].DeviceFamily &&
			deviceTypeMap[ele].CoreType == deviceTypeMap[deviceType].CoreType {

			result[idx] = true
		}
	}

	return result, nil
}

// TODO
// SyncDeviceTypesFromCRP sync device types from crp.
func (c *Controller) SyncDeviceTypesFromCRP(kt *kit.Kit, deviceTypes []string) error {
	// 1.从本地获取这些机型
	localDeviceTypeMap := make(map[string]wdt.WoaDeviceTypeTable)
	for _, batch := range slice.Split(deviceTypes, int(core.DefaultMaxPageLimit)) {
		listReq := &rpproto.WoaDeviceTypeListReq{
			ListReq: core.ListReq{
				Filter: tools.ContainersExpression("device_type", batch),
				Page:   core.NewDefaultBasePage(),
				Fields: []string{"id", "device_type"},
			},
		}
		batchRst, err := c.client.DataService().Global.ResourcePlan.ListWoaDeviceType(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list cvm device type from crp, err: %v, deviceTypes: %v, rid: %s", err,
				batch, kt.Rid)
			return err
		}

		for _, item := range batchRst.Details {
			localDeviceTypeMap[item.DeviceType] = item
		}
	}

	// 2.从crp平台获取机型
	crpDeviceTypeMap, err := c.listCvmInstanceTypeFromCrp(kt, deviceTypes)
	if err != nil {
		logs.Errorf("failed to list cvm device type from crp, err: %v, deviceTypes: %v, rid: %s", err,
			deviceTypes, kt.Rid)
		return err
	}

	// 3. (临时) 从CRP获取技术分类
	crpTechnicalClassMap, err := c.listCvmTechnicalClassFromCrp(kt)
	logs.Infof("list cvm technical class from crp, deviceTypes: %v, crpTechnicalClassMap: %v, rid: %s",
		deviceTypes, crpTechnicalClassMap, kt.Rid)
	if err != nil {
		logs.Errorf("failed to list cvm technical class from crp, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	needCreate := make([]wdt.WoaDeviceTypeTable, 0)
	needUpdate := make([]wdt.WoaDeviceTypeTable, 0)
	for deviceType, item := range crpDeviceTypeMap {
		techClass, ok := crpTechnicalClassMap[deviceType]
		if !ok {
			return fmt.Errorf("technical class not found for device type: %s", deviceType)
		}
		item.TechnicalClass = techClass

		if localItem, ok := localDeviceTypeMap[deviceType]; ok {
			item.ID = localItem.ID
			needUpdate = append(needUpdate, item)
			continue
		}
		needCreate = append(needCreate, item)
	}

	// 4.本地已存在时更新
	for _, batch := range slice.Split(needUpdate, constant.BatchOperationMaxLimit) {
		updateReq := &rpproto.WoaDeviceTypeBatchUpdateReq{
			DeviceTypes: batch,
		}
		err = c.client.DataService().Global.ResourcePlan.BatchUpdateWoaDeviceType(kt, updateReq)
		if err != nil {
			logs.Errorf("failed to batch update cvm device type, err: %v, deviceTypes: %v, rid: %s", err,
				batch, kt.Rid)
			return err
		}
	}

	// 5.本地不存在时创建
	for _, batch := range slice.Split(needCreate, constant.BatchOperationMaxLimit) {
		createReq := &rpproto.WoaDeviceTypeBatchCreateReq{
			DeviceTypes: batch,
		}
		_, err = c.client.DataService().Global.ResourcePlan.BatchCreateWoaDeviceType(kt, createReq)
		if err != nil {
			logs.Errorf("failed to batch create cvm device type, err: %v, deviceTypes: %v, rid: %s", err,
				batch, kt.Rid)
			return err
		}
	}

	return nil
}

// listCvmInstanceTypeFromCrp 从Crp平台获取机型
func (c *Controller) listCvmInstanceTypeFromCrp(kt *kit.Kit, deviceTypes []string) (map[string]wdt.WoaDeviceTypeTable,
	error) {

	req := &cvmapi.QueryCvmInstanceTypeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.QueryCvmInstanceType,
		},
		Params: &cvmapi.QueryCvmInstanceTypeParams{InstanceType: deviceTypes},
	}

	resp, err := c.crpCli.QueryCvmInstanceType(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm device type failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if resp.Result == nil {
		logs.Errorf("query cvm device type error, deviceTypes: %v, resp: %+v, rid: %s", deviceTypes, resp, kt.Rid)
		return nil, err
	}

	deviceTypeMap := make(map[string]wdt.WoaDeviceTypeTable)
	for _, item := range resp.Result.Data {
		if _, ok := deviceTypeMap[item.InstanceType]; ok {
			continue
		}
		deviceTypeMap[item.InstanceType] = wdt.WoaDeviceTypeTable{
			DeviceType:   item.InstanceType,
			DeviceClass:  item.InstanceClassDesc,
			DeviceFamily: item.InstanceGroup,
			CoreType:     string(enumor.GetCoreTypeByCRPCoreTypeID(item.CoreType)),
			// CPU和内存都是整数值，可直接转换
			CpuCore:         int64(item.CPUAmount),
			Memory:          int64(item.RamAmount),
			DeviceTypeClass: item.InstanceTypeClass,
		}
	}

	return deviceTypeMap, nil
}

// listCvmTechnicalClassFromCrp 从Crp平台获取机型
func (c *Controller) listCvmTechnicalClassFromCrp(kt *kit.Kit) (map[string]string, error) {

	req := &cvmapi.QueryTechnicalClassReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmQueryTecTechnicalClass,
		},
	}

	resp, err := c.crpCli.QueryTechnicalClass(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm device type failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if resp.Result == nil {
		logs.Errorf("query cvm device type error, resp: %+v, rid: %s", resp, kt.Rid)
		return nil, err
	}

	technicalClassMap := make(map[string]string)
	for _, item := range resp.Result {
		technicalClassMap[item.CvmInstanceModel] = item.TechnicalClass
	}

	return technicalClassMap, nil
}
