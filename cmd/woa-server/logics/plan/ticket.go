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
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	dtypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// CreateResPlanTicket create resource plan ticket.
func (c *Controller) CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error) {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create resource plan ticket request, err: %s, rid: %s", err, kt.Rid)
		return "", err
	}

	// construct resource plan ticket.
	ticket, err := constructResPlanTicket(req, kt.User)
	if err != nil {
		logs.Errorf("failed to construct resource plan ticket, err: %s, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketID, err := c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ticketIDs, err := c.dao.ResPlanTicket().CreateWithTx(kt, txn, []rpt.ResPlanTicketTable{*ticket})
		if err != nil {
			logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		if len(ticketIDs) != 1 {
			logs.Errorf("create resource plan ticket, but len ticketIDs != 1, rid: %s", kt.Rid)
			return "", errors.New("create resource plan ticket, but len ticketIDs != 1")
		}

		ticketID := ticketIDs[0]

		// create resource plan ticket status.
		statuses := []rpts.ResPlanTicketStatusTable{{
			TicketID: ticketID,
			Status:   enumor.RPTicketStatusInit,
		}}
		if err = c.dao.ResPlanTicketStatus().CreateWithTx(kt, txn, statuses); err != nil {
			logs.Errorf("create resource plan ticket status failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}

		return ticketID, nil
	})

	if err != nil {
		logs.Errorf("create resource plan ticket failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	ticketIDStr, ok := ticketID.(string)
	if !ok {
		logs.Errorf("convert resource plan ticket id %v from interface to string failed, err: %v, rid: %s",
			ticketID, err, kt.Rid)
		return "", fmt.Errorf("convert resource plan ticket id %v from interface to string failed", ticketID)
	}

	return ticketIDStr, nil
}

// constructResPlanTicket construct resource plan ticket.
func constructResPlanTicket(req *CreateResPlanTicketReq, applicant string) (*rpt.ResPlanTicketTable, error) {
	var originalOs, originalCpuCore, originalMemory, originalDiskSize int64
	var updatedOs, updatedCpuCore, updatedMemory, updatedDiskSize int64
	for _, demand := range req.Demands {
		if demand.Original != nil {
			originalOs += (*demand.Original).Cvm.Os
			originalCpuCore += (*demand.Original).Cvm.CpuCore
			originalMemory += (*demand.Original).Cvm.Memory
			originalDiskSize += (*demand.Original).Cbs.DiskSize
		}

		if demand.Updated != nil {
			updatedOs += (*demand.Updated).Cvm.Os
			updatedCpuCore += (*demand.Updated).Cvm.CpuCore
			updatedMemory += (*demand.Updated).Cvm.Memory
			updatedDiskSize += (*demand.Updated).Cbs.DiskSize
		}
	}

	demandsJson, err := dtypes.NewJsonField(req.Demands)
	if err != nil {
		return nil, err
	}

	result := &rpt.ResPlanTicketTable{
		Type:             req.TicketType,
		Demands:          demandsJson,
		Applicant:        applicant,
		BkBizID:          req.BizOrgRel.BkBizID,
		BkBizName:        req.BizOrgRel.BkBizName,
		OpProductID:      req.BizOrgRel.OpProductID,
		OpProductName:    req.BizOrgRel.OpProductName,
		PlanProductID:    req.BizOrgRel.PlanProductID,
		PlanProductName:  req.BizOrgRel.PlanProductName,
		VirtualDeptID:    req.BizOrgRel.VirtualDeptID,
		VirtualDeptName:  req.BizOrgRel.VirtualDeptName,
		DemandClass:      req.DemandClass,
		OriginalOS:       originalOs,
		OriginalCpuCore:  originalCpuCore,
		OriginalMemory:   originalMemory,
		OriginalDiskSize: originalDiskSize,
		UpdatedOS:        updatedOs,
		UpdatedCpuCore:   updatedCpuCore,
		UpdatedMemory:    updatedMemory,
		UpdatedDiskSize:  updatedDiskSize,
		Remark:           req.Remark,
		Creator:          applicant,
		Reviser:          applicant,
		SubmittedAt:      time.Now().Format(constant.DateTimeLayout),
	}

	return result, nil
}
