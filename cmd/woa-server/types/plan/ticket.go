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

// Package plan ...
package plan

import (
	"errors"
	"slices"
	"unicode/utf8"

	"hcm/cmd/woa-server/logics/plan/demand-time"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"
)

// ListResPlanTicketReq is list resource plan ticket request.
type ListResPlanTicketReq struct {
	BkBizIDs        []int64                 `json:"bk_biz_ids" validate:"omitempty"`
	OpProductIDs    []int64                 `json:"op_product_ids" validate:"omitempty"`
	PlanProductIDs  []int64                 `json:"plan_product_ids" validate:"omitempty"`
	TicketIDs       []string                `json:"ticket_ids" validate:"omitempty"`
	Statuses        []enumor.RPTicketStatus `json:"statuses" validate:"omitempty"`
	ObsProjects     []string                `json:"obs_projects" validate:"omitempty"`
	TicketTypes     []enumor.RPTicketType   `json:"ticket_types" validate:"omitempty"`
	Applicants      []string                `json:"applicants" validate:"omitempty"`
	SubmitTimeRange *times.DateRange        `json:"submit_time_range" validate:"omitempty"`
	Page            *core.BasePage          `json:"page" validate:"required"`
}

// Validate whether ListResPlanTicketReq is valid.
func (r *ListResPlanTicketReq) Validate() error {
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

	for _, status := range r.Statuses {
		if err := status.Validate(); err != nil {
			return err
		}
	}

	for _, ticketType := range r.TicketTypes {
		if err := ticketType.Validate(); err != nil {
			return err
		}
	}

	if r.SubmitTimeRange != nil {
		if err := r.SubmitTimeRange.Validate(); err != nil {
			return err
		}
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// GenListOption generate list option by list resource plan ticket request.
func (r *ListResPlanTicketReq) GenListOption() (*types.ListOption, error) {
	rules := make([]filter.RuleFactory, 0)

	if len(r.BkBizIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("bk_biz_id", r.BkBizIDs))
	}

	if len(r.OpProductIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("op_product_id", r.OpProductIDs))
	}

	if len(r.PlanProductIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("plan_product_id", r.PlanProductIDs))
	}

	if len(r.TicketIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("id", r.TicketIDs))
	}

	if len(r.Statuses) > 0 {
		rules = append(rules, tools.ContainersExpression("status", r.Statuses))
	}

	if len(r.TicketTypes) > 0 {
		rules = append(rules, tools.ContainersExpression("type", r.TicketTypes))
	}

	if len(r.Applicants) > 0 {
		rules = append(rules, tools.ContainersExpression("applicant", r.Applicants))
	}

	if r.SubmitTimeRange != nil {
		drOpt, err := tools.DateRangeExpression("submitted_at", r.SubmitTimeRange)
		if err != nil {
			return nil, err
		}
		rules = append(rules, drOpt)
	}

	// copy page for modifying.
	pageCopy := &core.BasePage{
		Count: r.Page.Count,
		Start: r.Page.Start,
		Limit: r.Page.Limit,
		Sort:  r.Page.Sort,
		Order: r.Page.Order,
	}

	// if count == false, default sort by submitted_at desc.
	if !pageCopy.Count {
		if pageCopy.Sort == "" {
			pageCopy.Sort = "submitted_at"
		}
		if pageCopy.Order == "" {
			pageCopy.Order = core.Descending
		}
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: pageCopy,
	}

	return opt, nil
}

// ListBizResPlanTicketReq is list biz resource plan ticket request.
type ListBizResPlanTicketReq struct {
	TicketIDs       []string                `json:"ticket_ids" validate:"omitempty"`
	Statuses        []enumor.RPTicketStatus `json:"statuses" validate:"omitempty"`
	ObsProjects     []string                `json:"obs_projects" validate:"omitempty"`
	TicketTypes     []enumor.RPTicketType   `json:"ticket_types" validate:"omitempty"`
	Applicants      []string                `json:"applicants" validate:"omitempty"`
	SubmitTimeRange *times.DateRange        `json:"submit_time_range" validate:"omitempty"`
	Page            *core.BasePage          `json:"page" validate:"required"`
}

// Validate whether ListBizResPlanTicketReq is valid.
func (r *ListBizResPlanTicketReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, status := range r.Statuses {
		if err := status.Validate(); err != nil {
			return err
		}
	}

	for _, ticketType := range r.TicketTypes {
		if err := ticketType.Validate(); err != nil {
			return err
		}
	}

	if r.SubmitTimeRange != nil {
		if err := r.SubmitTimeRange.Validate(); err != nil {
			return err
		}
	}

	if err := r.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// CreateResPlanTicketReq is create resource plan ticket request.
type CreateResPlanTicketReq struct {
	BkBizID     int64                    `json:"bk_biz_id" validate:"required"`
	DemandClass enumor.DemandClass       `json:"demand_class" validate:"required"`
	Demands     []CreateResPlanDemandReq `json:"demands" validate:"required"`
	Remark      string                   `json:"remark" validate:"required"`
}

// Validate whether CreateResPlanTicketReq is valid.
func (r *CreateResPlanTicketReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if err := r.DemandClass.Validate(); err != nil {
		return err
	}

	for _, demand := range r.Demands {
		if err := demand.Validate(); err != nil {
			return err
		}
	}

	lenRemark := utf8.RuneCountInString(r.Remark)
	if lenRemark < 20 || lenRemark > 1024 {
		return errors.New("len remark should be >= 20 and < 1024")
	}

	return nil
}

// CreateResPlanDemandReq is create resource plan demand request.
type CreateResPlanDemandReq struct {
	ObsProject     enumor.ObsProject      `json:"obs_project" validate:"required"`
	ExpectTime     string                 `json:"expect_time" validate:"required"`
	RegionID       string                 `json:"region_id" validate:"required"`
	ZoneID         string                 `json:"zone_id" validate:"omitempty"`
	DemandSource   enumor.DemandSource    `json:"demand_source" validate:"omitempty"`
	Remark         string                 `json:"remark" validate:"omitempty"`
	DemandResTypes []enumor.DemandResType `json:"demand_res_types" validate:"required"`
	Cvm            *struct {
		ResMode    string   `json:"res_mode"`
		DeviceType string   `json:"device_type"`
		Os         *float64 `json:"os"`
		CpuCore    *float64 `json:"cpu_core"`
		Memory     *float64 `json:"memory"`
	} `json:"cvm" validate:"omitempty"`
	Cbs *struct {
		DiskType enumor.DiskType `json:"disk_type"`
		DiskIo   *int64          `json:"disk_io"`
		DiskSize *float64        `json:"disk_size"`
	} `json:"cbs" validate:"omitempty"`
}

// Validate whether CreateResPlanDemandReq is valid.
func (r *CreateResPlanDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.ObsProject.ValidateResPlan(); err != nil {
		return err
	}

	et, err := times.ParseDay(r.ExpectTime)
	if err != nil {
		return err
	}
	if demandtime.IsDayCrossMonth(et) {
		return errors.New("expect_time should not be cross month")
	}

	if r.DemandSource != "" {
		if err := r.DemandSource.Validate(); err != nil {
			return err
		}
	}

	lenRemark := utf8.RuneCountInString(r.Remark)
	if lenRemark > 255 {
		return errors.New("len remark should <= 255")
	}

	for _, demandResType := range r.DemandResTypes {
		if err := demandResType.Validate(); err != nil {
			return err
		}
	}

	if slices.Contains(r.DemandResTypes, enumor.DemandResTypeCVM) {
		if err := r.cvmValidate(); err != nil {
			return err
		}
	}

	if slices.Contains(r.DemandResTypes, enumor.DemandResTypeCBS) {
		if err := r.cbsValidate(); err != nil {
			return err
		}
	}

	return nil
}

func (r *CreateResPlanDemandReq) cvmValidate() error {
	if r.Cvm == nil {
		return errors.New("demand includes cvm, cvm should not be nil")
	}

	if len(r.Cvm.ResMode) == 0 {
		return errors.New("cvm res mode should not be empty")
	}

	if len(r.Cvm.DeviceType) == 0 {
		return errors.New("cvm device type should not be empty")
	}

	if *r.Cvm.Os < 0 {
		return errors.New("os should be >= 0")
	}

	if *r.Cvm.CpuCore < 0 {
		return errors.New("cpu core should be >= 0")
	}

	if *r.Cvm.Memory < 0 {
		return errors.New("memory should be >= 0")
	}

	return nil
}

func (r *CreateResPlanDemandReq) cbsValidate() error {
	if r.Cbs == nil {
		return errors.New("demand includes cbs, cbs should not be nil")
	}

	if err := r.Cbs.DiskType.Validate(); err != nil {
		return err
	}

	if *r.Cbs.DiskIo < 0 {
		return errors.New("disk io should be >= 0")
	}

	if *r.Cbs.DiskSize < 0 {
		return errors.New("disk size should be >= 0")
	}

	return nil
}

// GetResPlanTicketResp is get resource plan ticket response.
type GetResPlanTicketResp struct {
	ID         string                 `json:"id"`
	BaseInfo   *GetRPTicketBaseInfo   `json:"base_info"`
	StatusInfo *GetRPTicketStatusInfo `json:"status_info"`
	Demands    []GetRPTicketDemand    `json:"demands"`
}

// GetRPTicketBaseInfo get resource plan ticket base info.
type GetRPTicketBaseInfo struct {
	Type            enumor.RPTicketType `json:"type"`
	TypeName        string              `json:"type_name"`
	Applicant       string              `json:"applicant"`
	BkBizID         int64               `json:"bk_biz_id"`
	BkBizName       string              `json:"bk_biz_name"`
	OpProductID     int64               `json:"op_product_id"`
	OpProductName   string              `json:"op_product_name"`
	PlanProductID   int64               `json:"plan_product_id"`
	PlanProductName string              `json:"plan_product_name"`
	VirtualDeptID   int64               `json:"virtual_dept_id"`
	VirtualDeptName string              `json:"virtual_dept_name"`
	DemandClass     enumor.DemandClass  `json:"demand_class"`
	Remark          string              `json:"remark"`
	SubmittedAt     string              `json:"submitted_at"`
}

// GetRPTicketStatusInfo get resource plan ticket status info.
type GetRPTicketStatusInfo struct {
	Status     enumor.RPTicketStatus `json:"status"`
	StatusName string                `json:"status_name"`
	ItsmSn     string                `json:"itsm_sn"`
	ItsmUrl    string                `json:"itsm_url"`
	CrpSn      string                `json:"crp_sn"`
	CrpUrl     string                `json:"crp_url"`
	Message    string                `json:"message"`
}

// GetRPTicketDemand get resource plan ticket demand.
type GetRPTicketDemand struct {
	DemandClass  enumor.DemandClass        `json:"demand_class"`
	OriginalInfo *rpt.OriginalRPDemandItem `json:"original_info"`
	UpdatedInfo  *rpt.UpdatedRPDemandItem  `json:"updated_info"`
}

// GetResPlanTicketAuditResp is get resource plan ticket audit response.
type GetResPlanTicketAuditResp struct {
	TicketID  string                `json:"ticket_id"`
	ItsmAudit *GetRPTicketItsmAudit `json:"itsm_audit"`
	CrpAudit  *GetRPTicketCrpAudit  `json:"crp_audit"`
}

// GetRPTicketItsmAudit get resource plan ticket itsm audit.
type GetRPTicketItsmAudit struct {
	ItsmSn       string                `json:"itsm_sn"`
	ItsmUrl      string                `json:"itsm_url"`
	Status       enumor.RPTicketStatus `json:"status"`
	StatusName   string                `json:"status_name"`
	Message      string                `json:"message"`
	CurrentSteps []*ItsmAuditStep      `json:"current_steps"`
	Logs         []*ItsmAuditLog       `json:"logs"`
}

// GetRPTicketCrpAudit get resource plan ticket crp audit.
type GetRPTicketCrpAudit struct {
	CrpSn        string                `json:"crp_sn"`
	CrpUrl       string                `json:"crp_url"`
	Status       enumor.RPTicketStatus `json:"status"`
	StatusName   string                `json:"status_name"`
	Message      string                `json:"message"`
	CurrentSteps []*CrpAuditStep       `json:"current_steps"`
	Logs         []*CrpAuditLog        `json:"logs"`
}

// ItsmAuditStep is itsm audit step.
type ItsmAuditStep struct {
	StateID    int64    `json:"state_id"`
	Name       string   `json:"name"`
	Processors []string `json:"processors"`
}

// ItsmAuditLog is itsm audit log.
type ItsmAuditLog struct {
	Operator  string `json:"operator"`
	OperateAt string `json:"operate_at"`
	Message   string `json:"message"`
}

// CrpAuditStep is crp audit step.
type CrpAuditStep struct {
	StateID    string   `json:"state_id"`
	Name       string   `json:"name"`
	Processors []string `json:"processors"`
}

// CrpAuditLog is crp audit log.
type CrpAuditLog struct {
	Operator  string `json:"operator"`
	OperateAt string `json:"operate_at"`
	Message   string `json:"message"`
	Name      string `json:"name"`
}

// TransferResPlanTicketReq is transfer demand ticket request.
type TransferResPlanTicketReq struct {
	TicketIDs []string `json:"ticket_ids" binding:"omitempty,max=100"`
	BkBizIDs  []int64  `json:"bk_biz_ids" binding:"omitempty,max=100"`
}

// Validate validate transfer demand ticket request.
func (t *TransferResPlanTicketReq) Validate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.TicketIDs) == 0 && len(t.BkBizIDs) == 0 {
		return errors.New("params ticket_ids or bk_biz_ids is required")
	}

	return nil
}

// GenListTicketsOption generate list option for transfer demand ticket.
func (t *TransferResPlanTicketReq) GenListTicketsOption(page *core.BasePage) *types.ListOption {
	return t.genListOption("id", page)
}

// GenListDemandsOption generate list option for transfer demand ticket.
func (t *TransferResPlanTicketReq) GenListDemandsOption(page *core.BasePage) *types.ListOption {
	return t.genListOption("ticket_id", page)
}

func (t *TransferResPlanTicketReq) genListOption(ticketIDKey string, page *core.BasePage) *types.ListOption {
	var rules []filter.RuleFactory

	if len(t.BkBizIDs) > 0 {
		rules = append(rules, tools.ContainersExpression("bk_biz_id", t.BkBizIDs))
	}
	if len(t.TicketIDs) > 0 {
		rules = append(rules, tools.ContainersExpression(ticketIDKey, t.TicketIDs))
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		},
		Page: page,
	}

	return opt
}
