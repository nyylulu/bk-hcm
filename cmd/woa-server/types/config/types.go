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

// Package config ...
package config

import (
	"errors"
	"fmt"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
)

// Requirement resource requirement type config
type Requirement struct {
	BkInstId    int64  `json:"id" bson:"id"`
	RequireType int64  `json:"require_type" bson:"require_type"`
	RequireName string `json:"require_name" bson:"require_name"`
}

// GetRequirementResult get requirement type list result
type GetRequirementResult struct {
	Count int64          `json:"count"`
	Info  []*Requirement `json:"info"`
}

// Region qcloud resource region config
type Region struct {
	BkInstId       int64  `json:"id" bson:"id"`
	Region         string `json:"region" bson:"region"`
	RegionCn       string `json:"region_cn" bson:"region_cn"`
	CmdbRegionName string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// GetRegionResult get region list result
type GetRegionResult struct {
	Count int64     `json:"count"`
	Info  []*Region `json:"info"`
}

// Zone qcloud resource zone config
type Zone struct {
	BkInstId       int64  `json:"id" bson:"id"`
	Zone           string `json:"zone" bson:"zone"`
	ZoneCn         string `json:"zone_cn" bson:"zone_cn"`
	Region         string `json:"region" bson:"region"`
	RegionCn       string `json:"region_cn" bson:"region_cn"`
	CmdbRegionName string `json:"cmdb_region_name" bson:"cmdb_region_name"`
	CmdbZoneId     int64  `json:"cmdb_zone_id" bson:"cmdb_zone_id"`
	CmdbZoneName   string `json:"cmdb_zone_name" bson:"cmdb_zone_name"`
}

// GetZoneParam get zone list request param
type GetZoneParam struct {
	Region     []string `json:"region" bson:"region"`
	CmdbRegion []string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// GetZoneResult get zone list result
type GetZoneResult struct {
	Count int64   `json:"count"`
	Info  []*Zone `json:"info"`
}

// Vpc cvm vpc config
type Vpc struct {
	BkInstId int64  `json:"id" bson:"id"`
	Region   string `json:"region" bson:"region"`
	VpcId    string `json:"vpc_id" bson:"vpc_id"`
	VpcName  string `json:"vpc_name" bson:"vpc_name"`
}

// GetVpcParam get vpc list request param
type GetVpcParam struct {
	Region string `json:"region" bson:"region"`
}

// GetVpcResult get vpc list result
type GetVpcResult struct {
	Count int64  `json:"count"`
	Info  []*Vpc `json:"info"`
}

// GetVpcListParam get vpc id list request param
type GetVpcListParam struct {
	Regions []string `json:"regions" bson:"regions"`
}

// GetVpcListRst get vpc id list result
type GetVpcListRst struct {
	Info []interface{} `json:"info"`
}

// Subnet cvm subnet config
type Subnet struct {
	BkInstId   int64  `json:"id" bson:"id"`
	Region     string `json:"region" bson:"region"`
	Zone       string `json:"zone" bson:"zone"`
	VpcId      string `json:"vpc_id" bson:"vpc_id"`
	VpcName    string `json:"vpc_name" bson:"vpc_name"`
	SubnetId   string `json:"subnet_id" bson:"subnet_id"`
	SubnetName string `json:"subnet_name" bson:"subnet_name"`
	Enable     bool   `json:"enable" bson:"enable"`
	Comment    string `json:"comment"`
}

// GetSubnetParam get subnet list request param
type GetSubnetParam struct {
	Region string `json:"region" bson:"region"`
	Zone   string `json:"zone" bson:"zone"`
	Vpc    string `json:"vpc" bson:"vpc"`
}

// GetSubnetResult get subnet list result
type GetSubnetResult struct {
	Count int64     `json:"count"`
	Info  []*Subnet `json:"info"`
}

// GetSubnetListParam get subnet detail list request param
type GetSubnetListParam struct {
	Filter *querybuilder.QueryFilter `json:"filter" bson:"filter"`
	Page   metadata.BasePage         `json:"page" bson:"page"`
}

// Validate whether GetSubnetListParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetSubnetListParam) Validate() (errKey string, err error) {
	if key, err := param.Page.Validate(false); err != nil {
		return fmt.Sprintf("page.%s", key), err
	}

	if param.Filter != nil {
		if key, err := param.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if param.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetSubnetListParam) GetFilter() (map[string]interface{}, error) {
	if param.Filter != nil {
		mgoFilter, key, err := param.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		return mgoFilter, nil
	}
	return make(map[string]interface{}), nil
}

// UpdateSubnetPropertyParam update subnet property request param
type UpdateSubnetPropertyParam struct {
	Ids      []int64                `json:"ids" bson:"ids"`
	Property map[string]interface{} `json:"properties"`
}

// Validate whether UpdateSubnetPropertyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateSubnetPropertyParam) Validate() (errKey string, err error) {
	limit := 200
	if len(param.Ids) > limit {
		return "ids", fmt.Errorf("exceed limit %d", limit)
	}

	return "", nil
}

// GetIdcZoneParam get idc zone list request param
type GetIdcZoneParam struct {
	Region []string `json:"cmdb_region_name" bson:"cmdb_region_name"`
}

// IdcZone idc resource zone config
type IdcZone struct {
	BkInstId int64  `json:"id" bson:"id"`
	Region   string `json:"cmdb_region_name" bson:"cmdb_region_name"`
	ZoneName string `json:"cmdb_zone_name" bson:"cmdb_zone_name"`
	ZoneId   int64  `json:"cmdb_zone_id" bson:"cmdb_zone_id"`
}

// GetIdcRegionRst get idc region list result
type GetIdcRegionRst struct {
	Info []interface{} `json:"info"`
}

// GetIdcZoneRst get idc zone list result
type GetIdcZoneRst struct {
	Count int64      `json:"count"`
	Info  []*IdcZone `json:"info"`
}

// DeviceRestrict device restrict config
type DeviceRestrict struct {
	BkInstId int64   `json:"id" bson:"id"`
	Cpu      []int64 `json:"cpu" bson:"cpu"`
	Mem      []int64 `json:"mem" bson:"mem"`
	Disk     []int64 `json:"disk" bson:"disk"`
}

// GetDeviceRestrictResult get device info result
type GetDeviceRestrictResult struct {
	Cpu  []int64 `json:"cpu" bson:"cpu"`
	Mem  []int64 `json:"mem" bson:"mem"`
	Disk []int64 `json:"disk" bson:"disk"`
}

// CvmImage cvm image config
type CvmImage struct {
	BkInstId  int64  `json:"id" bson:"id"`
	Region    string `json:"region" bson:"region"`
	ImageId   string `json:"image_id" bson:"image_id"`
	ImageName string `json:"image_name" bson:"image_name"`
}

// GetCvmImageParam get cvm image list request param
type GetCvmImageParam struct {
	Region []string `json:"region" bson:"region"`
}

// GetCvmImageResult get zone list result
type GetCvmImageResult struct {
	Count int64       `json:"count"`
	Info  []*CvmImage `json:"info"`
}

// DeviceInfo cvm device info
type DeviceInfo struct {
	BkInstId       int64         `json:"id" bson:"id"`
	RequireType    int64         `json:"require_type" bson:"require_type"`
	Region         string        `json:"region" bson:"region"`
	Zone           string        `json:"zone" bson:"zone"`
	DeviceType     string        `json:"device_type" bson:"device_type"`
	Cpu            int64         `json:"cpu" bson:"cpu"`
	Mem            int64         `json:"mem" bson:"mem"`
	Disk           int64         `json:"disk" bson:"disk"`
	Remark         string        `json:"remark" bson:"remark"`
	Label          mapstr.MapStr `json:"label" bson:"label"`
	CapacityFlag   int           `json:"capacity_flag" bson:"capacity_flag"`
	EnableCapacity bool          `json:"enable_capacity" bson:"enable_capacity"`
	EnableApply    bool          `json:"enable_apply" bson:"enable_apply"`
	Score          float64       `json:"score" bson:"score"`
	Comment        string        `json:"comment" bson:"comment"`
}

const (
	// CapLevelUndefined undefined capacity flag
	CapLevelUndefined int = 0
	// CapLevelEmpty "无库存" - 0
	CapLevelEmpty int = 1
	// CapLevelLow "库存紧张" - [1~10]
	CapLevelLow int = 2
	// CapLevelMedium "少量库存" - [11~50]
	CapLevelMedium int = 3
	// CapLevelHigh "库存充足" - [51~oo]
	CapLevelHigh int = 4
)

// GetDeviceParam get device list request param
type GetDeviceParam struct {
	Filter *querybuilder.QueryFilter `json:"filter" bson:"filter"`
	Page   metadata.BasePage         `json:"page" bson:"page"`
}

// Validate whether GetDeviceParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetDeviceParam) Validate() (errKey string, err error) {
	if key, err := param.Page.Validate(false); err != nil {
		return fmt.Sprintf("page.%s", key), err
	}

	if param.Filter != nil {
		if key, err := param.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if param.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetDeviceParam) GetFilter() (map[string]interface{}, error) {
	if param.Filter != nil {
		mgoFilter, key, err := param.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		return mgoFilter, nil
	}
	return make(map[string]interface{}), nil
}

// GetDeviceInfoResult get device info result
type GetDeviceInfoResult struct {
	Count int64         `json:"count"`
	Info  []*DeviceInfo `json:"info"`
}

// GetDeviceTypeResult get device type result
type GetDeviceTypeResult struct {
	Count int64         `json:"count"`
	Info  []interface{} `json:"info"`
}

// DeviceTypeInfo cvm device type info
type DeviceTypeInfo struct {
	DeviceType  string `json:"device_type"`
	DeviceGroup string `json:"device_group"`
}

// GetDeviceTypeDetailResult get device type detail result
type GetDeviceTypeDetailResult struct {
	Count int64             `json:"count"`
	Info  []*DeviceTypeInfo `json:"info"`
}

// CreateManyDeviceParam create device config in batch request param
type CreateManyDeviceParam struct {
	RequireType []int64  `json:"require_type"`
	Zone        []string `json:"zone"`
	DeviceGroup string   `json:"device_group"`
	DeviceType  string   `json:"device_type"`
	Cpu         int64    `json:"cpu"`
	Mem         int64    `json:"mem"`
	Remark      string   `json:"remark"`
}

// Validate whether GetDeviceParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *CreateManyDeviceParam) Validate() (errKey string, err error) {
	if len(param.RequireType) == 0 {
		return "require_type", errors.New("empty or non-exist")
	}

	if len(param.RequireType) > 20 {
		return "require_type", errors.New("exceed limit 20")
	}

	if len(param.Zone) == 0 {
		return "zone", errors.New("empty or non-exist")
	}

	if len(param.Zone) > 100 {
		return "zone", errors.New("exceed limit 100")
	}

	if param.DeviceGroup == "" {
		return "device_group", errors.New("empty or non-exist")
	}

	if param.DeviceType == "" {
		return "device_type", errors.New("empty or non-exist")
	}

	if param.Cpu <= 0 {
		return "cpu", errors.New("should be positive")
	}

	if param.Mem <= 0 {
		return "mem", errors.New("should be positive")
	}

	return "", nil
}

// UpdateDevicePropertyParam update device property request param
type UpdateDevicePropertyParam struct {
	Ids      []int64                `json:"ids" bson:"ids"`
	Property map[string]interface{} `json:"properties"`
}

// Validate whether UpdateDevicePropertyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateDevicePropertyParam) Validate() (errKey string, err error) {
	limit := 200
	if len(param.Ids) > limit {
		return "ids", fmt.Errorf("exceed limit %d", limit)
	}

	return "", nil
}

// DvmDeviceInfo dvm device info
type DvmDeviceInfo struct {
	BkInstId    int64         `json:"id" bson:"id"`
	DeviceType  string        `json:"device_type" bson:"device_type"`
	Cpu         int64         `json:"cpu" bson:"cpu"`
	Mem         int64         `json:"mem" bson:"mem"`
	Disk        int64         `json:"disk" bson:"disk"`
	NetWork     string        `json:"network" bson:"network"`
	CpuProvider string        `json:"cpu_provider" bson:"cpu_provider"`
	Remark      string        `json:"remark" bson:"remark"`
	Label       mapstr.MapStr `json:"label" bson:"label"`
}

// GetDvmDeviceRst get dvm device result
type GetDvmDeviceRst struct {
	Count int64            `json:"count"`
	Info  []*DvmDeviceInfo `json:"info"`
}

// PmDeviceInfo physical machine device info
type PmDeviceInfo struct {
	BkInstId   int64         `json:"id" bson:"id"`
	DeviceType string        `json:"device_type" bson:"device_type"`
	Cpu        int64         `json:"cpu" bson:"cpu"`
	Mem        int64         `json:"mem" bson:"mem"`
	Raid       string        `json:"raid" bson:"raid"`
	NetWork    string        `json:"network" bson:"network"`
	Remark     string        `json:"remark" bson:"remark"`
	Label      mapstr.MapStr `json:"label" bson:"label"`
}

// GetPmDeviceRst get dvm device result
type GetPmDeviceRst struct {
	Count int64           `json:"count"`
	Info  []*PmDeviceInfo `json:"info"`
}

// GetCapacityParam get resource apply capacity request param
type GetCapacityParam struct {
	RequireType int64  `json:"require_type"`
	DeviceType  string `json:"device_type"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
	Vpc         string `json:"vpc"`
	Subnet      string `json:"subnet"`
}

// Validate whether GetCapacityParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetCapacityParam) Validate() (string, error) {
	if param.RequireType <= 0 {
		key := "require_type"
		return key, fmt.Errorf("invalid %s <= 0", key)
	}

	if len(param.DeviceType) == 0 {
		key := "device_type"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Region) == 0 {
		key := "region"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Zone) == 0 {
		key := "zone"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	return "", nil
}

// GetCapacityRst get resource apply capacity result
type GetCapacityRst struct {
	Count int64           `json:"count"`
	Info  []*CapacityInfo `json:"info"`
}

// CapacityInfo resource apply capacity info
type CapacityInfo struct {
	Region  string             `json:"region"`
	Zone    string             `json:"zone"`
	Vpc     string             `json:"vpc"`
	Subnet  string             `json:"subnet"`
	MaxNum  int64              `json:"max_num"`
	MaxInfo []*CapacityMaxInfo `json:"max_info"`
}

// CapacityMaxInfo resource apply capacity max info
type CapacityMaxInfo struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

// UpdateCapacityParam update resource capacity info param
type UpdateCapacityParam struct {
	RequireType int64  `json:"require_type"`
	DeviceType  string `json:"device_type"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
}

// Validate whether UpdateCapacityParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateCapacityParam) Validate() (string, error) {
	if param.RequireType <= 0 {
		key := "require_type"
		return key, fmt.Errorf("invalid %s <= 0", key)
	}

	if len(param.DeviceType) == 0 {
		key := "device_type"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Region) == 0 {
		key := "region"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	if len(param.Zone) == 0 {
		key := "zone"
		return key, fmt.Errorf("invalid %s, empty or non-exist", key)
	}

	return "", nil
}

// GetAffinityParam get affinity request parameter
type GetAffinityParam struct {
	ResourceType string `json:"resource_type" bson:"resource_type"`
	HasZone      bool   `json:"has_zone" bson:"has_zone"`
}

// GetAffinityRst get affinity result
type GetAffinityRst struct {
	Info []*AffinityInfo `json:"info"`
}

// AffinityInfo affinity info
type AffinityInfo struct {
	Level       string `json:"level" bson:"level"`
	Description string `json:"description" bson:"description"`
}

const (
	ResourceTypePm        string = "IDCPM"
	ResourceTypeCvm       string = "QCLOUDCVM"
	ResourceTypeIdcDvm    string = "IDCDVM"
	ResourceTypeQcloudDvm string = "QCLOUDDVM"
)

const (
	AntiNone   string = "ANTI_NONE"   //无要求
	AntiRack   string = "ANTI_RACK"   //分机架
	AntiModule string = "ANTI_MODULE" //分Module
	AntiCampus string = "ANTI_CAMPUS" //分Campus
)

// Description description of terms
var Description = map[string]string{
	AntiNone:   "无要求",
	AntiRack:   "分机架",
	AntiModule: "分Module",
	AntiCampus: "分Campus",
}
