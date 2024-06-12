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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	rpd "hcm/pkg/dal/table/resource_plan/res-plan-demand"
	"hcm/pkg/tools/times"
)

// ListResPlanTicketReq is list resource plan ticket request.
type ListResPlanTicketReq struct {
	BkBizIDs        []int64          `json:"bk_biz_ids" validate:"omitempty"`
	TicketIDs       []string         `json:"ticket_ids" validate:"omitempty"`
	Applicants      []string         `json:"applicants" validate:"omitempty"`
	SubmitTimeRange *times.DateRange `json:"submit_time_range" validate:"omitempty"`
	Page            *core.BasePage   `json:"page" validate:"required"`
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

	if r.SubmitTimeRange != nil {
		if err := r.SubmitTimeRange.Validate(); err != nil {
			return err
		}
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

	if len(r.Remark) < 20 || len(r.Remark) > 1024 {
		return errors.New("len remark should be >= 20 and < 1024")
	}

	return nil
}

// CreateResPlanDemandReq is create resource plan demand request.
type CreateResPlanDemandReq struct {
	ObsProject   enumor.ObsProject   `json:"obs_project" validate:"required"`
	ExpectTime   string              `json:"expect_time" validate:"required"`
	RegionID     string              `json:"region_id" validate:"required"`
	ZoneID       string              `json:"zone_id" validate:"omitempty"`
	DemandSource enumor.DemandSource `json:"demand_source" validate:"required"`
	Remark       string              `json:"remark" validate:"omitempty"`
	Cvm          *struct {
		ResMode    string `json:"res_mode" validate:"required"`
		DeviceType string `json:"device_type" validate:"required"`
		Os         *int64 `json:"os" validate:"required"`
		CpuCore    *int64 `json:"cpu_core" validate:"required"`
		Memory     *int64 `json:"memory" validate:"required"`
	} `json:"cvm" validate:"omitempty"`
	Cbs *struct {
		DiskType enumor.DiskType `json:"disk_type" validate:"required"`
		DiskIo   *int64          `json:"disk_io" validate:"required"`
		DiskSize *int64          `json:"disk_size" validate:"required"`
	} `json:"cbs" validate:"omitempty"`
}

// Validate whether CreateResPlanDemandReq is valid.
func (r *CreateResPlanDemandReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.ObsProject.Validate(); err != nil {
		return err
	}

	if _, err := times.ParseDay(r.ExpectTime); err != nil {
		return err
	}

	if err := r.DemandSource.Validate(); err != nil {
		return err
	}

	if r.Cvm != nil {
		if *r.Cvm.Os < 0 {
			return errors.New("os should be >= 0")
		}

		if *r.Cvm.CpuCore < 0 {
			return errors.New("cpu core should be >= 0")
		}

		if *r.Cvm.Memory < 0 {
			return errors.New("memory should be >= 0")
		}
	}

	if r.Cbs != nil {
		if err := r.Cbs.DiskType.Validate(); err != nil {
			return err
		}

		if *r.Cbs.DiskIo < 0 {
			return errors.New("disk io should be >= 0")
		}

		if *r.Cbs.DiskSize < 0 {
			return errors.New("disk size should be >= 0")
		}
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
	Applicant       string             `json:"applicant"`
	BkBizID         int64              `json:"bk_biz_id"`
	BkBizName       string             `json:"bk_biz_name"`
	BkProductID     int64              `json:"bk_product_id"`
	BkProductName   string             `json:"bk_product_name"`
	PlanProductID   int64              `json:"plan_product_id"`
	PlanProductName string             `json:"plan_product_name"`
	VirtualDeptID   int64              `json:"virtual_dept_id"`
	VirtualDeptName string             `json:"virtual_dept_name"`
	DemandClass     enumor.DemandClass `json:"demand_class"`
	Remark          string             `json:"remark"`
	SubmittedAt     string             `json:"submitted_at"`
}

// GetRPTicketStatusInfo get resource plan ticket status info.
type GetRPTicketStatusInfo struct {
	Status     enumor.RPTicketStatus `json:"status"`
	StatusName string                `json:"status_name"`
	ItsmSn     string                `json:"itsm_sn"`
	ItsmUrl    string                `json:"itsm_url"`
	CrpSn      string                `json:"crp_sn"`
	CrpUrl     string                `json:"crp_url"`
}

// GetRPTicketDemand get resource plan ticket demand.
type GetRPTicketDemand struct {
	ObsProject   enumor.ObsProject   `json:"obs_project"`
	ExpectTime   string              `json:"expect_time"`
	ZoneID       string              `json:"zone_id"`
	ZoneName     string              `json:"zone_name"`
	RegionID     string              `json:"region_id"`
	RegionName   string              `json:"region_name"`
	AreaID       string              `json:"area_id"`
	AreaName     string              `json:"area_name"`
	DemandSource enumor.DemandSource `json:"demand_source"`
	Remark       string              `json:"remark"`
	Cvm          *rpd.Cvm            `json:"cvm"`
	Cbs          *rpd.Cbs            `json:"cbs"`
}
