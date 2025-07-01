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
	"strings"

	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// createCrpTicket create crp ticket.
func (c *Controller) createCrpTicket(kt *kit.Kit, ticket *TicketInfo) error {
	if ticket == nil {
		logs.Errorf("failed to create crp ticket, ticket is nil, rid: %s", kt.Rid)
		return errors.New("ticket is nil")
	}

	// call crp api to create crp ticket.
	var sn string
	var err error
	switch ticket.Type {
	case enumor.RPTicketTypeAdd:
		sn, err = c.createAddCrpTicket(kt, ticket)
	case enumor.RPTicketTypeAdjust, enumor.RPTicketTypeDelete:
		adjustCreator := NewCrpAdjustTicketCreator(c, c.crpCli)
		sn, err = adjustCreator.CreateCRPTicket(kt, ticket)
		// sn, err = c.createAdjustCrpTicket(kt, ticket)
	default:
		logs.Errorf("failed to create crp ticket, unsupported ticket type: %s, ticket_id: %s, rid: %s", ticket.Type,
			ticket.ID, kt.Rid)
		return errors.New("unsupported ticket type")
	}
	if err != nil {
		// 因CRP单据修改冲突导致的提单失败，不返回报错，记录日志后返回队列继续等待
		if strings.Contains(err.Error(), constant.CRPResPlanDemandIsInProcessing) {
			logs.Warnf("failed to create crp ticket, as crp res plan demand is in processing, err: %v, "+
				"ticket_id: %s, rid: %s", err, ticket.ID, kt.Rid)
			return nil
		}

		// 这里主要返回的error是crp ticket创建失败，且ticket状态更新失败的日志在函数内已打印，这里可以忽略该错误
		_ = c.updateTicketStatusFailed(kt, ticket, err.Error())
		logs.Errorf("failed to create crp ticket with different ticket type, err: %v, ticket_id: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return err
	}

	// save crp sn and crp url to resource plan ticket status table.
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSn:   ticket.ItsmSn,
		ItsmUrl:  ticket.ItsmUrl,
		CrpSn:    sn,
		CrpUrl:   cvmapi.CvmPlanLinkPrefix + sn,
	}

	if err = c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, ticket_id: %s, rid: %s", err, ticket.ID,
			kt.Rid)
		return err
	}

	return nil
}

// createAddCrpTicket create add crp ticket.
func (c *Controller) createAddCrpTicket(kt *kit.Kit, ticket *TicketInfo) (string, error) {
	addReq, err := c.constructAddReq(kt, ticket)
	if err != nil {
		logs.Errorf("failed to construct add cvm & cbs plan order request, err: %v, ticket_id: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return "", err
	}

	resp, err := c.crpCli.AddCvmCbsPlan(kt.Ctx, kt.Header(), addReq)
	if err != nil {
		logs.Errorf("failed to add cvm & cbs plan order, err: %v, ticket_id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to add cvm & cbs plan order, code: %d, msg: %s, crp_tran_id: %s, ticket_id: %s, rid: %s",
			resp.Error.Code,
			resp.Error.Message, resp.TraceId, ticket.ID, kt.Rid)
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to add cvm & cbs plan order, for return empty order id, ticket_id: %s, rid: %s", ticket.ID,
			kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// updateTicketStatusFailed update ticket status to failed.
func (c *Controller) updateTicketStatusFailed(kt *kit.Kit, ticket *TicketInfo, msg string) error {
	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusFailed,
		ItsmSn:   ticket.ItsmSn,
		ItsmUrl:  ticket.ItsmUrl,
		CrpSn:    ticket.CrpSn,
		CrpUrl:   ticket.CrpUrl,
		Message:  msg,
	}

	if len(msg) > 255 {
		logs.Warnf("failure message is truncated to 255, origin message: %s, rid: %s", msg, kt.Rid)
		update.Message = msg[:255]
	}

	if err := c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 失败需要释放资源
	allDemandIDs := make([]string, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allDemandIDs = append(allDemandIDs, (*demand.Original).DemandID)
		}
	}
	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(allDemandIDs, 0)
	if err := c.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// constructAddReq construct cvm cbs plan add request.
func (c *Controller) constructAddReq(kt *kit.Kit, ticket *TicketInfo) (*cvmapi.AddCvmCbsPlanReq, error) {
	addReq := &cvmapi.AddCvmCbsPlanReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAddMethod,
		},
		Params: &cvmapi.AddCvmCbsPlanParam{
			Operator: ticket.Applicant,
			DeptName: ticket.VirtualDeptName,
			Desc:     "",
			Items:    make([]*cvmapi.AddPlanItem, 0),
		},
	}

	switch ticket.DemandClass {
	case enumor.DemandClassCVM:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCvmDesc
	case enumor.DemandClassCA:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCADesc
	default:
		logs.Warnf("failed to construct add desc, unsupported demand class: %s, rid: %s", ticket.DemandClass, kt.Rid)
	}

	for _, demand := range ticket.Demands {
		if demand.Updated == nil {
			logs.Errorf("failed to create add crp ticket, demand updated is nil, rid: %s", kt.Rid)
			return nil, errors.New("demand updated is nil")
		}

		planItem := &cvmapi.AddPlanItem{
			UseTime:         demand.Updated.ExpectTime,
			ProjectName:     string(demand.Updated.ObsProject),
			PlanProductName: ticket.PlanProductName,
			ProductName:     ticket.OpProductName,
			CityName:        demand.Updated.RegionName,
			ZoneName:        demand.Updated.ZoneName,
			CoreTypeName:    demand.Updated.Cvm.CoreType,
			InstanceModel:   demand.Updated.Cvm.DeviceType,
			CvmAmount:       demand.Updated.Cvm.Os.InexactFloat64(),
			CoreAmount:      int(demand.Updated.Cvm.CpuCore),
			Desc:            demand.Updated.Remark,
			InstanceIO:      int(demand.Updated.Cbs.DiskIo),
			DiskTypeName:    demand.Updated.Cbs.DiskType.Name(),
			DiskAmount:      int(demand.Updated.Cbs.DiskSize),
		}

		addReq.Params.Items = append(addReq.Params.Items, planItem)
	}

	return addReq, nil
}
