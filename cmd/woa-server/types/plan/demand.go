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

// Package plan ...
package plan

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	mtypes "hcm/pkg/dal/dao/types/meta"
	"hcm/pkg/tools/times"
)

// ListResPlanDemandReq is list resource plan demand request.
type ListResPlanDemandReq struct {
	BkBizIDs        []int64              `json:"bk_biz_ids" validate:"omitempty,max=100"`
	OpProductIDs    []int64              `json:"op_product_ids" validate:"omitempty,max=100"`
	PlanProductIDs  []int64              `json:"plan_product_ids" validate:"omitempty,max=100"`
	CrpDemandIDs    []int64              `json:"crp_demand_ids" validate:"omitempty,max=100"`
	ObsProjects     []string             `json:"obs_projects" validate:"omitempty,max=100"`
	DemandClasses   []enumor.DemandClass `json:"demand_classes" validate:"omitempty,max=100"`
	DeviceClasses   []string             `json:"device_classes" validate:"omitempty,max=100"`
	DeviceTypes     []string             `json:"device_types" validate:"omitempty,max=100"`
	RegionIDs       []string             `json:"region_ids" validate:"omitempty,max=100"`
	ZoneIDs         []string             `json:"zone_ids" validate:"omitempty,max=100"`
	PlanTypes       []enumor.PlanType    `json:"plan_types" validate:"omitempty,max=100"`
	ExpiringOnly    bool                 `json:"expiring_only" validate:"omitempty"`
	ExpectTimeRange *times.DateRange     `json:"expect_time_range" validate:"required"`
	Page            *core.BasePage       `json:"page" validate:"required"`
}

// Validate whether ListResPlanDemandReq is valid.
func (r *ListResPlanDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, bkBizID := range r.BkBizIDs {
		if bkBizID <= 0 {
			return errors.New("bk biz id should be > 0")
		}
	}
	for _, opProductID := range r.OpProductIDs {
		if opProductID <= 0 {
			return errors.New("op product id should be > 0")
		}
	}
	for _, planProductID := range r.PlanProductIDs {
		if planProductID <= 0 {
			return errors.New("plan product id should be > 0")
		}
	}

	for _, class := range r.DemandClasses {
		if err := class.Validate(); err != nil {
			return err
		}
	}

	for _, planType := range r.PlanTypes {
		if err := planType.Validate(); err != nil {
			return err
		}
	}

	if r.ExpectTimeRange != nil {
		if err := r.ExpectTimeRange.Validate(); err != nil {
			return err
		}
	}

	if r.Page != nil {
		if err := r.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ListResPlanDemandResp is list resource plan demand response.
type ListResPlanDemandResp struct {
	Overview *ListResPlanDemandOverview `json:"overview"`
	Count    int                        `json:"count"`
	Details  []*ListResPlanDemandItem   `json:"details"`
}

// SortAndPage 排序并按照分页需求截断数据，默认根据submitted_at(提单时间)倒序排序，
// 可排序字段：
//
//	cpu_core(CPU核心数)、memory(内存大小)、disk_size(云盘大小)、expect_time(期望交付时间)
func (l *ListResPlanDemandResp) SortAndPage(page *core.BasePage) error {
	if page == nil {
		return errors.New("page is nil")
	}
	if err := page.Validate(); err != nil {
		return err
	}

	sort.Slice(l.Details, func(i, j int) bool {
		var less bool
		switch page.Sort {
		case "cpu_core":
			less = l.Details[i].TotalCpuCore < l.Details[j].TotalCpuCore
		case "memory":
			less = l.Details[i].TotalMemory < l.Details[j].TotalMemory
		case "disk_size":
			less = l.Details[i].TotalDiskSize < l.Details[j].TotalDiskSize
		case "expect_time":
			less = l.Details[i].GetExpectTimeStr().Before(l.Details[j].GetExpectTimeStr())
		default:
			less = l.Details[i].GetExpectTimeStr().Before(l.Details[j].GetExpectTimeStr())
		}
		if page.Order == core.Descending {
			return !less
		}
		return less
	})

	// start超出范围
	if int(page.Start) >= len(l.Details) {
		l.Details = l.Details[:0]
		return nil
	}

	end := int(page.Start) + int(page.Limit)
	// end超出范围
	if end > len(l.Details) {
		end = len(l.Details)
	}
	// 按page需求截断slice
	l.Details = l.Details[int(page.Start):end]
	return nil
}

// ListDemandIds list crp demand ids
func (l *ListResPlanDemandResp) ListDemandIds() []int64 {
	res := make([]int64, 0, len(l.Details))
	for _, item := range l.Details {
		res = append(res, item.CrpDemandID)
	}
	return res
}

// ListResPlanDemandOverview is list resource plan demand overview
type ListResPlanDemandOverview struct {
	TotalCpuCore          float32 `json:"total_cpu_core"`
	TotalAppliedCore      float32 `json:"total_applied_core"`
	InPlanCpuCore         float32 `json:"in_plan_cpu_core"`
	InPlanAppliedCpuCore  float32 `json:"in_plan_applied_cpu_core"`
	OutPlanCpuCore        float32 `json:"out_plan_cpu_core"`
	OutPlanAppliedCpuCore float32 `json:"out_plan_applied_cpu_core"`
	ExpiringCpuCore       float32 `json:"expiring_cpu_core"`
}

// ListResPlanDemandItem is list resource plan demand detail's item
type ListResPlanDemandItem struct {
	CrpDemandID        int64               `json:"crp_demand_id"`
	BkBizID            int64               `json:"bk_biz_id"`
	BkBizName          string              `json:"bk_biz_name"`
	OpProductID        int64               `json:"op_product_id"`
	OpProductName      string              `json:"op_product_name"`
	PlanProductID      int64               `json:"plan_product_id"`
	PlanProductName    string              `json:"plan_product_name"`
	Status             enumor.DemandStatus `json:"status"`
	StatusName         string              `json:"status_name"`
	DemandClass        enumor.DemandClass  `json:"demand_class"`
	AvailableYearMonth string              `json:"available_year_month"`
	ExpectTime         string              `json:"expect_time"`
	expectTimeStr      time.Time           // expectTimeStr 用于排序
	DeviceClass        string              `json:"device_class"`
	DeviceType         string              `json:"device_type"`
	TotalOS            float32             `json:"total_os"`
	AppliedOS          float32             `json:"applied_os"`
	RemainedOS         float32             `json:"remained_os"`
	TotalCpuCore       float32             `json:"total_cpu_core"`
	AppliedCpuCore     float32             `json:"applied_cpu_core"`
	RemainedCpuCore    float32             `json:"remained_cpu_core"`
	ExpiredCpuCore     float32             `json:"-"` // ExpiredCpuCore 目前仅用于计算overview
	TotalMemory        float32             `json:"total_memory"`
	AppliedMemory      float32             `json:"applied_memory"`
	RemainedMemory     float32             `json:"remained_memory"`
	TotalDiskSize      float32             `json:"total_disk_size"`
	AppliedDiskSize    float32             `json:"applied_disk_size"`
	RemainedDiskSize   float32             `json:"remained_disk_size"`
	RegionID           string              `json:"region_id"`
	RegionName         string              `json:"region_name"`
	ZoneID             string              `json:"zone_id"`
	ZoneName           string              `json:"zone_name"`
	PlanType           enumor.PlanType     `json:"plan_type"`
	ObsProject         string              `json:"obs_project"`
	GenerationType     string              `json:"generation_type"`
	DeviceFamily       string              `json:"device_family"`
	DiskType           enumor.DiskType     `json:"disk_type"`
	DiskTypeName       string              `json:"disk_type_name"`
	DiskIO             int                 `json:"disk_io"`
}

// ParseExpectTime parse expect_time
func (l *ListResPlanDemandItem) ParseExpectTime() error {
	t, err := time.Parse(constant.DateLayout, l.ExpectTime)
	if err != nil {
		return err
	}
	l.expectTimeStr = t
	return nil
}

// GetExpectTimeStr get expect_time for sort
func (l *ListResPlanDemandItem) GetExpectTimeStr() time.Time {
	return l.expectTimeStr
}

// SetStatus set demand status
func (l *ListResPlanDemandItem) SetStatus(status enumor.DemandStatus) {
	l.Status = status
	// spent_all（已耗尽）优先级更高
	if l.AppliedCpuCore == l.TotalCpuCore {
		l.Status = enumor.DemandStatusSpentAll
	}
}

// SetRegionAndZoneID set region and zone id
func (l *ListResPlanDemandItem) SetRegionAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[l.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", l.RegionName)
	}
	l.RegionID = regionArea.RegionID

	zoneID, exists := zoneNameMap[l.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", l.ZoneName)
	}
	l.ZoneID = zoneID
	return nil
}

// SetDiskType set disk type
func (l *ListResPlanDemandItem) SetDiskType() error {
	diskType, err := enumor.GetDiskTypeFromName(l.DiskTypeName)
	if err != nil {
		return err
	}
	l.DiskType = diskType
	return nil
}

// PlanDemandDetail crp demand detail的本地格式化
type PlanDemandDetail struct {
	GetPlanDemandDetailResp `json:",inline"`
	Year                    int     `json:"year"`
	Month                   int     `json:"month"`
	Week                    int     `json:"week"`
	TotalOS                 float32 `json:"total_os"`
	AppliedOS               float32 `json:"applied_os"`
	RemainedOS              float32 `json:"remained_os"`
	TotalCpuCore            float32 `json:"total_cpu_core"`
	AppliedCpuCore          float32 `json:"applied_cpu_core"`
	RemainedCpuCore         float32 `json:"remained_cpu_core"`
	ExpiredCpuCore          float32 `json:"expired_cpu_core"`
	TotalMemory             float32 `json:"total_memory"`
	AppliedMemory           float32 `json:"applied_memory"`
	RemainedMemory          float32 `json:"remained_memory"`
	TotalDiskSize           float32 `json:"total_disk_size"`
	AppliedDiskSize         float32 `json:"applied_disk_size"`
	RemainedDiskSize        float32 `json:"remained_disk_size"`
}

// GetPlanDemandDetailResp get plan demand detail response
type GetPlanDemandDetailResp struct {
	CrpDemandID        string          `json:"crp_demand_id"`
	YearMonthWeek      string          `json:"year_month_week"`
	ExpectStartDate    string          `json:"expect_start_date"`
	ExpectEndDate      string          `json:"expect_end_date"`
	ExpectTime         string          `json:"expect_time"`
	BkBizID            int64           `json:"bk_biz_id"`
	BkBizName          string          `json:"bk_biz_name"`
	BgID               int64           `json:"bg_id"`
	BgName             string          `json:"bg_name"`
	DeptID             int64           `json:"dept_id"`
	DeptName           string          `json:"dept_name"`
	PlanProductID      int64           `json:"plan_product_id"`
	PlanProductName    string          `json:"plan_product_name"`
	OpProductID        int64           `json:"op_product_id"`
	OpProductName      string          `json:"op_product_name"`
	ObsProject         string          `json:"obs_project"`
	AreaID             string          `json:"area_id"`
	AreaName           string          `json:"area_name"`
	RegionID           string          `json:"region_id"`
	RegionName         string          `json:"region_name"`
	ZoneID             string          `json:"zone_id"`
	ZoneName           string          `json:"zone_name"`
	PlanType           enumor.PlanType `json:"plan_type"`
	PlanAdvanceWeek    int             `json:"plan_advance_week"`
	ExpeditedPostponed string          `json:"expedited_postponed"`
	CoreTypeID         int             `json:"core_type_id"`
	CoreType           string          `json:"core_type"`
	DeviceFamily       string          `json:"device_family"`
	DeviceClass        string          `json:"device_class"`
	DeviceType         string          `json:"device_type"`
	OS                 float32         `json:"os"`
	Memory             float32         `json:"memory"`
	CpuCore            float32         `json:"cpu_core"`
	DiskSize           float32         `json:"disk_size"`
	DiskIO             int             `json:"disk_io"`
	DiskType           enumor.DiskType `json:"disk_type"`
	DiskTypeName       string          `json:"disk_type_name"`
	DemandWeek         string          `json:"demand_week"`
	ResPoolType        int             `json:"res_pool_type"`
	ResPool            string          `json:"res_pool"`
	ResMode            string          `json:"res_mode"`
	GenerationType     string          `json:"generation_type"`
}

// SetDiskType set disk type
func (g *GetPlanDemandDetailResp) SetDiskType() error {
	diskTypes := enumor.GetDiskTypeMembers()
	for _, diskType := range diskTypes {
		if g.DiskTypeName == diskType.Name() {
			g.DiskType = diskType
			return nil
		}
	}
	return fmt.Errorf("invalid disk type name: %s", g.DiskTypeName)
}

// SetRegionAreaAndZoneID set region/area and zone id
func (g *GetPlanDemandDetailResp) SetRegionAreaAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[g.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", g.RegionName)
	}
	g.RegionID = regionArea.RegionID
	g.AreaID = regionArea.AreaID
	g.AreaName = regionArea.AreaName

	zoneID, exists := zoneNameMap[g.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", g.ZoneName)
	}
	g.ZoneID = zoneID
	return nil
}

// ListDemandChangeLogReq is list demand change log request.
type ListDemandChangeLogReq struct {
	CrpDemandId int64          `json:"crp_demand_id" validate:"required"`
	Page        *core.BasePage `json:"page" validate:"required"`
}

// Validate whether ListDemandChangeLogReq is valid.
func (r *ListDemandChangeLogReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// ListDemandChangeLogResp is list demand change log response
type ListDemandChangeLogResp struct {
	Count   int                        `json:"count"`
	Details []*ListDemandChangeLogItem `json:"details"`
}

// Page 对details结果分页
func (l *ListDemandChangeLogResp) Page(page *core.BasePage) {
	// start超出范围
	if int(page.Start) >= len(l.Details) {
		l.Details = l.Details[:0]
		return
	}

	end := int(page.Start) + int(page.Limit)
	// end超出范围
	if end > len(l.Details) {
		end = len(l.Details)
	}
	// 按page需求截断slice
	l.Details = l.Details[int(page.Start):end]
	return
}

// ListDemandChangeLogItem is list demand change log detail's item
type ListDemandChangeLogItem struct {
	CrpDemandId       int64   `json:"crp_demand_id"`
	ExpectTime        string  `json:"expect_time"`
	BgName            string  `json:"bg_name"`
	DeptName          string  `json:"dept_name"`
	PlanProductName   string  `json:"plan_product_name"`
	OpProductName     string  `json:"op_product_name"`
	ObsProject        string  `json:"obs_project"`
	RegionName        string  `json:"region_name"`
	ZoneName          string  `json:"zone_name"`
	DemandWeek        string  `json:"demand_week"`
	ResPoolType       int     `json:"res_pool_type"`
	DeviceClass       string  `json:"device_class"`
	DeviceType        string  `json:"device_type"`
	ChangeCvmAmount   float32 `json:"change_cvm_amount"`
	AfterCvmAmount    float32 `json:"after_cvm_amount"`
	ChangeCoreAmount  float32 `json:"change_core_amount"`
	AfterCoreAmount   float32 `json:"after_core_amount"`
	ChangeRamAmount   float32 `json:"change_ram_amount"`
	AfterRamAmount    float32 `json:"after_ram_amount"`
	DiskType          string  `json:"disk_type"`
	DiskIo            int     `json:"disk_io"`
	ChangedDiskAmount float32 `json:"changed_disk_amount"`
	AfterDiskAmount   float32 `json:"after_disk_amount"`
	DemandSource      string  `json:"demand_source"`
	CrpSn             string  `json:"crp_sn"`
	CreateTime        string  `json:"create_time"`
	Remark            string  `json:"remark"`
	ResPool           string  `json:"res_pool"`
}

// AdjustRPDemandReq is adjust resource plan demand request.
type AdjustRPDemandReq struct {
	Adjusts []AdjustRPDemandReqElem `json:"adjusts" validate:"required,max=100"`
}

// Validate whether AdjustRPDemandReq is valid.
func (r *AdjustRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, adjust := range r.Adjusts {
		if err := adjust.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AdjustRPDemandReqElem is adjust resource plan demand request element.
type AdjustRPDemandReqElem struct {
	CrpDemandID  int64                     `json:"crp_demand_id" validate:"required"`
	AdjustType   enumor.RPDemandAdjustType `json:"adjust_type" validate:"required"`
	DemandSource enumor.DemandSource       `json:"demand_source" validate:"omitempty"`
	OriginalInfo *CreateResPlanDemandReq   `json:"original_info" validate:"omitempty"`
	UpdatedInfo  *CreateResPlanDemandReq   `json:"updated_info" validate:"omitempty"`
	ExpectTime   string                    `json:"expect_time" validate:"omitempty"`
	// TODO: 目前DelayOs没有使用
	DelayOs *int64 `json:"delay_os" validate:"omitempty"`
}

// Validate whether AdjustRPDemandReqElem is valid.
func (e *AdjustRPDemandReqElem) Validate() error {
	if err := validator.Validate.Struct(e); err != nil {
		return err
	}

	if e.CrpDemandID <= 0 {
		return errors.New("invalid crp demand id, should be > 0")
	}

	switch e.AdjustType {
	case enumor.RPDemandAdjustTypeUpdate:
		if e.DemandSource != "" {
			if err := e.DemandSource.Validate(); err != nil {
				return err
			}
		}

		if e.OriginalInfo == nil {
			return errors.New("original info of update demand can not be empty")
		}

		if err := e.OriginalInfo.Validate(); err != nil {
			return err
		}

		if e.UpdatedInfo == nil {
			return errors.New("updated info of update demand can not be empty")
		}

		if err := e.UpdatedInfo.Validate(); err != nil {
			return err
		}
	case enumor.RPDemandAdjustTypeDelay:
		if len(e.ExpectTime) == 0 {
			return errors.New("expect time of delay demand can not be empty")
		}

		// TODO：目前DelayOs没有使用，因此未做校验
	default:
		return fmt.Errorf("unsupported resource plan demand adjust type: %s", e.AdjustType)
	}

	return nil
}

// CancelRPDemandReq is cancel resource plan demand request.
type CancelRPDemandReq struct {
	CrpDemandIDs []int64 `json:"crp_demand_ids" validate:"required,max=100"`
}

// Validate whether CancelRPDemandReq is valid.
func (r *CancelRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, crpDemandID := range r.CrpDemandIDs {
		if crpDemandID <= 0 {
			return errors.New("invalid crp demand id, should be > 0")
		}
	}

	return nil
}
