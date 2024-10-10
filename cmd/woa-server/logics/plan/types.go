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
	"errors"
	"fmt"
	"unicode/utf8"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/tools/times"
)

// TicketInfo resource plan ticket info.
type TicketInfo struct {
	ID               string                `json:"id"`
	Type             enumor.RPTicketType   `json:"type"`
	Applicant        string                `json:"applicant"`
	BkBizID          int64                 `json:"bk_biz_id"`
	BkBizName        string                `json:"bk_biz_name"`
	OpProductID      int64                 `json:"op_product_id"`
	OpProductName    string                `json:"op_product_name"`
	PlanProductID    int64                 `json:"plan_product_id"`
	PlanProductName  string                `json:"plan_product_name"`
	VirtualDeptID    int64                 `json:"virtual_dept_id"`
	VirtualDeptName  string                `json:"virtual_dept_name"`
	DemandClass      enumor.DemandClass    `json:"demand_class"`
	OriginalCpuCore  float64               `json:"original_cpu_core"`
	OriginalMemory   float64               `json:"original_memory"`
	OriginalDiskSize float64               `json:"original_disk_size"`
	UpdatedCpuCore   float64               `json:"updated_cpu_core"`
	UpdatedMemory    float64               `json:"updated_memory"`
	UpdatedDiskSize  float64               `json:"updated_disk_size"`
	Demands          rpt.ResPlanDemands    `json:"demands"`
	SubmittedAt      string                `json:"submitted_at"`
	Status           enumor.RPTicketStatus `json:"status"`
	ItsmSn           string                `json:"itsm_sn"`
	ItsmUrl          string                `json:"itsm_url"`
	CrpSn            string                `json:"crp_sn"`
	CrpUrl           string                `json:"crp_url"`
}

// CreateResPlanTicketReq is create resource plan ticket request.
type CreateResPlanTicketReq struct {
	TicketType  enumor.RPTicketType `json:"ticket_type" validate:"required"`
	DemandClass enumor.DemandClass  `json:"demand_class" validate:"required"`
	BizOrgRel   ptypes.BizOrgRel    `json:"biz_org_rel" validate:"required"`
	Demands     rpt.ResPlanDemands  `json:"demands" validate:"required"`
	Remark      string              `json:"remark" validate:"required"`
}

// Validate whether CreateResPlanTicketReq is valid.
func (r *CreateResPlanTicketReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if err := r.TicketType.Validate(); err != nil {
		return err
	}

	switch r.TicketType {
	case enumor.RPTicketTypeAdd:
		for _, demand := range r.Demands {
			if demand.Original != nil {
				return errors.New("original demand of add ticket should be empty")
			}

			if demand.Updated == nil {
				return errors.New("updated demand of add ticket can not be empty")
			}
		}
	case enumor.RPTicketTypeAdjust:
		for _, demand := range r.Demands {
			if demand.Original == nil {
				return errors.New("original demand of adjust ticket can not be empty")
			}

			if demand.Updated == nil {
				return errors.New("updated demand of adjust ticket can not be empty")
			}
		}
	case enumor.RPTicketTypeDelete:
		for _, demand := range r.Demands {
			if demand.Original == nil {
				return errors.New("original demand of delete ticket can not be empty")
			}

			if demand.Updated != nil {
				return errors.New("updated demand of delete ticket should be empty")
			}
		}
	default:
		return fmt.Errorf("unsupported resource plan ticket type: %s", r.TicketType)
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

// QueryAllDemandsReq query all demands request.
type QueryAllDemandsReq struct {
	ExpectTimeRange *times.DateRange
	CrpDemandIDs    []int64
	CrpSns          []string
	DeviceClasses   []string
	PlanProdNames   []string
	ObsProjects     []string
	RegionNames     []string
	ZoneNames       []string
}

// AvailableTime available time.
type AvailableTime string

// NewAvailableTime new an available time.
// TODO: 目前只关注年和月，未来会添加周
func NewAvailableTime(year, month int) AvailableTime {
	return AvailableTime(fmt.Sprintf("%04d-%02d", year, month))
}

// VerifyResPlanElem verify resource plan element.
type VerifyResPlanElem struct {
	// IsAnyPlanType if IsAnyPlanType is true, it will examine both InPlan and OutPlan plan types.
	IsAnyPlanType bool
	PlanType      enumor.PlanType
	AvailableTime AvailableTime
	DeviceType    string
	ObsProject    enumor.ObsProject
	RegionName    string
	ZoneName      string
	CpuCore       float64
}

// ResPlanElem resource plan element.
type ResPlanElem struct {
	PlanType      enumor.PlanType
	AvailableTime AvailableTime
	DeviceType    string
	ObsProject    enumor.ObsProject
	RegionName    string
	ZoneName      string
	CpuCore       float64
}

// ResPlanPoolKey resource plan pool key.
type ResPlanPoolKey struct {
	PlanType      enumor.PlanType
	AvailableTime AvailableTime
	DeviceType    string
	ObsProject    enumor.ObsProject
	RegionName    string
	ZoneName      string
}

// ResPlanPool resource plan pool.
type ResPlanPool map[ResPlanPoolKey]float64

// StrUnionFind string union find struct.
type StrUnionFind struct {
	parent map[string]string
}

// NewStrUnionFind news a string union find.
func NewStrUnionFind() *StrUnionFind {
	return &StrUnionFind{parent: make(map[string]string)}
}

// Add adds a new element x.
func (uf *StrUnionFind) Add(x string) {
	uf.parent[x] = x
}

// Elements return all elements in StrUnionFind.
func (uf *StrUnionFind) Elements() []string {
	var res []string
	for k := range uf.parent {
		res = append(res, k)
	}

	return res
}

// Find finds the root parent of x.
func (uf *StrUnionFind) Find(x string) string {
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}

	return uf.parent[x]
}

// Union unions the unions where x and y are.
func (uf *StrUnionFind) Union(x, y string) {
	parentX := uf.Find(x)
	parentY := uf.Find(y)
	if parentX != parentY {
		uf.parent[parentY] = parentX
	}
}

// Connected judges whether x and y are connected.
func (uf *StrUnionFind) Connected(x, y string) bool {
	return uf.Find(x) == uf.Find(y)
}
