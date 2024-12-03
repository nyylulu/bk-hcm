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
	"time"

	dtime "hcm/cmd/woa-server/logics/plan/demand-time"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	mtypes "hcm/pkg/dal/dao/types/meta"
	wdttablers "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// ListResPlanDemandReq is list resource plan demand request.
type ListResPlanDemandReq struct {
	BkBizIDs        []int64              `json:"bk_biz_ids" validate:"omitempty,max=100"`
	OpProductIDs    []int64              `json:"op_product_ids" validate:"omitempty,max=100"`
	PlanProductIDs  []int64              `json:"plan_product_ids" validate:"omitempty,max=100"`
	DemandIDs       []string             `json:"demand_ids" validate:"omitempty,max=100"`
	ObsProjects     []enumor.ObsProject  `json:"obs_projects" validate:"omitempty,max=100"`
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

	for _, projectName := range r.ObsProjects {
		if err := projectName.ValidateResPlan(); err != nil {
			return err
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
	Overview *ListResPlanDemandOverview `json:"overview" validate:"omitempty"`
	Count    uint64                     `json:"count"`
	Details  []*ListResPlanDemandItem   `json:"details"`
}

// ListDemandIds list crp demand ids
func (l *ListResPlanDemandResp) ListDemandIds() []string {
	res := make([]string, 0, len(l.Details))
	for _, item := range l.Details {
		res = append(res, item.DemandID)
	}
	return res
}

// ListResPlanDemandOverview is list resource plan demand overview
type ListResPlanDemandOverview struct {
	TotalCpuCore          int64 `json:"total_cpu_core"`
	TotalAppliedCore      int64 `json:"total_applied_core"`
	InPlanCpuCore         int64 `json:"in_plan_cpu_core"`
	InPlanAppliedCpuCore  int64 `json:"in_plan_applied_cpu_core"`
	OutPlanCpuCore        int64 `json:"out_plan_cpu_core"`
	OutPlanAppliedCpuCore int64 `json:"out_plan_applied_cpu_core"`
	ExpiringCpuCore       int64 `json:"expiring_cpu_core"`
}

// ListResPlanDemandItem is list resource plan demand detail's item
type ListResPlanDemandItem struct {
	DemandID         string               `json:"demand_id"`
	BkBizID          int64                `json:"bk_biz_id"`
	BkBizName        string               `json:"bk_biz_name"`
	OpProductID      int64                `json:"op_product_id"`
	OpProductName    string               `json:"op_product_name"`
	PlanProductID    int64                `json:"plan_product_id"`
	PlanProductName  string               `json:"plan_product_name"`
	Status           enumor.DemandStatus  `json:"status"`
	StatusName       string               `json:"status_name"`
	DemandClass      enumor.DemandClass   `json:"demand_class"`
	DemandResType    enumor.DemandResType `json:"demand_res_type"`
	ExpectTime       string               `json:"expect_time"`
	DeviceClass      string               `json:"device_class"`
	DeviceType       string               `json:"device_type"`
	TotalOS          decimal.Decimal      `json:"total_os"`
	AppliedOS        decimal.Decimal      `json:"applied_os"`
	RemainedOS       decimal.Decimal      `json:"remained_os"`
	TotalCpuCore     int64                `json:"total_cpu_core"`
	AppliedCpuCore   int64                `json:"applied_cpu_core"`
	RemainedCpuCore  int64                `json:"remained_cpu_core"`
	ExpiringCpuCore  int64                `json:"-"` // ExpiringCpuCore 即将过期核心数，目前仅用于计算overview
	TotalMemory      int64                `json:"total_memory"`
	AppliedMemory    int64                `json:"applied_memory"`
	RemainedMemory   int64                `json:"remained_memory"`
	TotalDiskSize    int64                `json:"total_disk_size"`
	AppliedDiskSize  int64                `json:"applied_disk_size"`
	RemainedDiskSize int64                `json:"remained_disk_size"`
	RegionID         string               `json:"region_id"`
	RegionName       string               `json:"region_name"`
	ZoneID           string               `json:"zone_id"`
	ZoneName         string               `json:"zone_name"`
	PlanType         enumor.PlanType      `json:"plan_type"`
	ObsProject       enumor.ObsProject    `json:"obs_project"`
	DeviceFamily     string               `json:"device_family"`
	DiskType         enumor.DiskType      `json:"disk_type"`
	DiskTypeName     string               `json:"disk_type_name"`
	DiskIO           int64                `json:"disk_io"`
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
	ExpiringCpuCore         float32 `json:"expiring_cpu_core"`
	TotalMemory             float32 `json:"total_memory"`
	AppliedMemory           float32 `json:"applied_memory"`
	RemainedMemory          float32 `json:"remained_memory"`
	TotalDiskSize           float32 `json:"total_disk_size"`
	AppliedDiskSize         float32 `json:"applied_disk_size"`
	RemainedDiskSize        float32 `json:"remained_disk_size"`
}

// GetPlanDemandDetailResp get plan demand detail response
type GetPlanDemandDetailResp struct {
	DemandID        string            `json:"demand_id"`
	ExpectTime      string            `json:"expect_time"`
	BkBizID         int64             `json:"bk_biz_id"`
	BkBizName       string            `json:"bk_biz_name"`
	DeptID          int64             `json:"dept_id"`
	DeptName        string            `json:"dept_name"`
	PlanProductID   int64             `json:"plan_product_id"`
	PlanProductName string            `json:"plan_product_name"`
	OpProductID     int64             `json:"op_product_id"`
	OpProductName   string            `json:"op_product_name"`
	ObsProject      enumor.ObsProject `json:"obs_project"`
	AreaID          string            `json:"area_id"`
	AreaName        string            `json:"area_name"`
	RegionID        string            `json:"region_id"`
	RegionName      string            `json:"region_name"`
	ZoneID          string            `json:"zone_id"`
	ZoneName        string            `json:"zone_name"`
	PlanType        enumor.PlanType   `json:"plan_type"`
	CoreType        string            `json:"core_type"`
	DeviceFamily    string            `json:"device_family"`
	DeviceClass     string            `json:"device_class"`
	DeviceType      string            `json:"device_type"`
	OS              decimal.Decimal   `json:"os"`
	Memory          int64             `json:"memory"`
	CpuCore         int64             `json:"cpu_core"`
	DiskSize        int64             `json:"disk_size"`
	DiskIO          int64             `json:"disk_io"`
	DiskType        enumor.DiskType   `json:"disk_type"`
	DiskTypeName    string            `json:"disk_type_name"`
	ResMode         enumor.ResMode    `json:"res_mode"`
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
	DemandID string         `json:"demand_id" validate:"required"`
	Page     *core.BasePage `json:"page" validate:"required"`
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
	Count   uint64                     `json:"count"`
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
	ID                string            `json:"id"`
	DemandId          string            `json:"demand_id"`
	ExpectTime        string            `json:"expect_time"`
	ObsProject        enumor.ObsProject `json:"obs_project"`
	RegionName        string            `json:"region_name"`
	ZoneName          string            `json:"zone_name"`
	DeviceType        string            `json:"device_type"`
	ChangeCvmAmount   decimal.Decimal   `json:"change_cvm_amount"`
	ChangeCoreAmount  int64             `json:"change_core_amount"`
	ChangeRamAmount   int64             `json:"change_ram_amount"`
	ChangedDiskAmount int64             `json:"changed_disk_amount"`
	DemandSource      string            `json:"demand_source"`
	TicketID          string            `json:"ticket_id"`
	CrpSn             string            `json:"crp_sn"`
	SuborderID        string            `json:"suborder_id"`
	CreateTime        string            `json:"create_time"`
	Remark            string            `json:"remark"`
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
	DemandID     string                    `json:"demand_id" validate:"required"`
	CrpDemandID  int64                     `json:"crp_demand_id" validate:"omitempty"`
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

	if len(e.DemandID) <= 0 {
		return errors.New("invalid demand id, should be > 0")
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
	CrpDemandIDs  []int64                 `json:"crp_demand_ids" validate:"omitempty"`
	CancelDemands []CancelRPDemandReqElem `json:"cancel_demands" validate:"required,max=100"`
}

// Validate whether CancelRPDemandReq is valid.
func (r *CancelRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, demand := range r.CancelDemands {
		if err := demand.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CancelRPDemandReqElem is cancel resource plan demand request element.
type CancelRPDemandReqElem struct {
	DemandID        string `json:"demand_id" validate:"required"`
	RemainedCpuCore int64  `json:"remained_cpu_core" validate:"required"`
}

// Validate whether CancelRPDemandReqElem is valid.
func (r CancelRPDemandReqElem) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.RemainedCpuCore <= 0 {
		return errors.New("remained cpu core should be > 0")
	}

	return nil
}

// RepairRPDemandReq is repair resource plan demand request.
type RepairRPDemandReq struct {
	BkBizIDs          []int64         `json:"bk_biz_ids" validate:"omitempty,max=100,dive,gt=0"`
	RepairTicketRange times.DateRange `json:"repair_ticket_range" validate:"required"`
}

// Validate whether RepairRPDemandReq is valid.
func (c RepairRPDemandReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	if err := c.RepairTicketRange.Validate(); err != nil {
		return err
	}

	return nil
}

// CalcPenaltyBaseReq is request of calc penalty base.
type CalcPenaltyBaseReq struct {
	BkBizIDs []int64 `json:"bk_biz_ids" validate:"omitempty,max=100"`
	// PenaltyBaseDay is any day of the penalty base week. Format is YYYY-MM-DD.
	PenaltyBaseDay string `json:"penalty_base_day" validate:"required"`
}

// Validate whether CalcPenaltyBaseReq is valid.
func (c CalcPenaltyBaseReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, c.PenaltyBaseDay)
	if err != nil {
		return err
	}

	return nil
}

// CalcAndPushPenaltyRatioReq is request of calc and push penalty ratio.
type CalcAndPushPenaltyRatioReq struct {
	// PenaltyTime is any day of the month which penalty ratio be calculated.
	// Note that the first week may not be part of the month.
	// Format is YYYY-MM-DD.
	PenaltyTime string `json:"penalty_time" validate:"required"`
}

// Validate whether CalcAndPushPenaltyRatioReq is valid.
func (c *CalcAndPushPenaltyRatioReq) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	_, err := time.Parse(constant.DateLayout, c.PenaltyTime)
	if err != nil {
		return err
	}

	return nil
}

// CrpOrderChangeInfo is response of crp order change info.
type CrpOrderChangeInfo struct {
	OrderID       string               `json:"order_id"`
	ExpectTime    string               `json:"expect_time"`
	ObsProject    enumor.ObsProject    `json:"obs_project"`
	DemandResType enumor.DemandResType `json:"demand_res_type"`
	ResMode       enumor.ResModeCode   `json:"res_mode"`
	PlanType      enumor.PlanTypeCode  `json:"plan_type"`
	AreaID        string               `json:"area_id"`
	AreaName      string               `json:"area_name"`
	RegionID      string               `json:"region_id"`
	RegionName    string               `json:"region_name"`
	ZoneID        string               `json:"zone_id"`
	ZoneName      string               `json:"zone_name"`
	DeviceFamily  string               `json:"device_family"`
	DeviceClass   string               `json:"device_class"`
	DeviceType    string               `json:"device_type"`
	CoreType      string               `json:"core_type"`
	DiskType      enumor.DiskType      `json:"disk_type"`
	DiskTypeName  string               `json:"disk_type_name"`
	DiskIO        int64                `json:"disk_io"`

	ChangeOs       decimal.Decimal `json:"change_os"`
	ChangeCpuCore  int64           `json:"change_cpu_core"`
	ChangeMemory   int64           `json:"change_memory"`
	ChangeDiskSize int64           `json:"change_disk_size"`
}

// SetRegionAreaAndZoneID set region area and zone id.
func (c *CrpOrderChangeInfo) SetRegionAreaAndZoneID(zoneNameMap map[string]string,
	regionNameMap map[string]mtypes.RegionArea) error {

	regionArea, exists := regionNameMap[c.RegionName]
	if !exists {
		return fmt.Errorf("region name: %s not found in woa_zone", c.RegionName)
	}
	c.RegionID = regionArea.RegionID
	c.AreaID = regionArea.AreaID
	c.AreaName = regionArea.AreaName

	zoneID, exists := zoneNameMap[c.ZoneName]
	if !exists {
		return fmt.Errorf("zone name: %s not found in woa_zone", c.ZoneName)
	}
	c.ZoneID = zoneID
	return nil
}

// GetKey get key of crp order change info.
func (c *CrpOrderChangeInfo) GetKey(bkBizID int64, demandClass enumor.DemandClass) ResPlanDemandKey {
	key := ResPlanDemandKey{
		BkBizID:       bkBizID,
		DemandClass:   demandClass,
		DemandResType: c.DemandResType,
		ResMode:       c.ResMode,
		ObsProject:    c.ObsProject,
		ExpectTime:    c.ExpectTime,
		PlanType:      c.PlanType,
		RegionID:      c.RegionID,
		ZoneID:        c.ZoneID,
		DeviceType:    c.DeviceType,
		DiskType:      c.DiskType,
		DiskIO:        c.DiskIO,
	}

	return key
}

// GetAggregateKey get aggregate key of crp order change info.
func (c *CrpOrderChangeInfo) GetAggregateKey(bkBizID int64,
	deviceTypes map[string]wdttablers.WoaDeviceTypeTable) (ResPlanDemandAggregateKey, error) {

	deviceInfo, ok := deviceTypes[c.DeviceType]
	if !ok {
		return ResPlanDemandAggregateKey{}, fmt.Errorf("device type: %s not found", c.DeviceType)
	}

	expectTimeT, err := time.Parse(constant.DateLayout, c.ExpectTime)
	if err != nil {
		return ResPlanDemandAggregateKey{}, fmt.Errorf("failed to parse expect time: %s", c.ExpectTime)
	}

	key := ResPlanDemandAggregateKey{
		BkBizID:         bkBizID,
		RegionID:        c.RegionID,
		ExpectTimeRange: dtime.GetDemandDateRangeInMonth(expectTimeT),
		DeviceFamily:    deviceInfo.DeviceFamily,
		CoreType:        deviceInfo.CoreType,
		PlanType:        c.PlanType,
		ObsProject:      c.ObsProject,
		ResType:         c.DemandResType,
	}

	return key, nil
}

// ResPlanDemandKey is key of res plan demand.
type ResPlanDemandKey struct {
	BkBizID       int64
	DemandClass   enumor.DemandClass
	DemandResType enumor.DemandResType
	ResMode       enumor.ResModeCode
	ObsProject    enumor.ObsProject
	ExpectTime    string
	PlanType      enumor.PlanTypeCode
	RegionID      string
	ZoneID        string
	DeviceType    string
	DiskType      enumor.DiskType
	DiskIO        int64
}

// ResPlanDemandAggregateKey 聚合key
// 为解决CRP模糊调整导致数据出现负数的问题，demandKey需要按照模糊范围查找多条进行调整，避免负数出现
// 模糊范围：城市、可用范围（当前是整个月，未来可能精确到周）、机型族、核心类型、预测内外、项目类型、资源类型
type ResPlanDemandAggregateKey struct {
	BkBizID         int64
	RegionID        string
	ExpectTimeRange times.DateRange
	DeviceFamily    string
	CoreType        string
	PlanType        enumor.PlanTypeCode
	ObsProject      enumor.ObsProject
	ResType         enumor.DemandResType
}

// DemandPenaltyBaseKey is key of demand penalty.
// bk_biz_id / area_name / device_family
type DemandPenaltyBaseKey struct {
	BkBizID      int64
	AreaName     string
	DeviceFamily string
}

// ResPlanDemandExpendKey is key of res plan demand expend.
type ResPlanDemandExpendKey struct {
	// DemandClass enumor.DemandClass // 目前暂时不考虑CVM和CA的区别
	// DiskType      enumor.DiskType
	BkBizID       int64
	PlanType      enumor.PlanTypeCode
	AvailableTime AvailableMonth
	// 为了快速处理通配的情况，这里通过机型族和大小核心作为key
	// DeviceType    string
	DeviceFamily string
	CoreType     string
	ObsProject   enumor.ObsProject
	RegionID     string
}

// AvailableMonth available time.
type AvailableMonth string

// NewAvailableMonth new an available month.
func NewAvailableMonth(expectTime string) (AvailableMonth, error) {
	t, err := time.Parse(constant.DateLayout, expectTime)
	if err != nil {
		return "", err
	}

	return AvailableMonth(t.Format(constant.YearMonthLayout)), nil
}
