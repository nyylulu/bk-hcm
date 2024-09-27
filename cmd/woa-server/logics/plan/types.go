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

// TicketBriefInfo resource plan ticket brief info
type TicketBriefInfo struct {
	ID              string                `json:"id"`
	Type            enumor.RPTicketType   `json:"type"`
	Applicant       string                `json:"applicant"`
	BkBizID         int64                 `json:"bk_biz_id"`
	BkBizName       string                `json:"bk_biz_name"`
	BkProductName   string                `json:"bk_product_name"`
	PlanProductName string                `json:"plan_product_name"`
	DemandClass     enumor.DemandClass    `json:"demand_class"`
	CpuCore         int64                 `json:"cpu_core"`
	Memory          int64                 `json:"memory"`
	DiskSize        int64                 `json:"disk_size"`
	SubmittedAt     string                `json:"submitted_at"`
	Status          enumor.RPTicketStatus `json:"status"`
	ItsmSn          string                `json:"itsm_sn"`
	ItsmUrl         string                `json:"itsm_url"`
	CrpSn           string                `json:"crp_sn"`
	CrpUrl          string                `json:"crp_url"`
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
