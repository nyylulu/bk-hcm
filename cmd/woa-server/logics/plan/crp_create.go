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
	"strconv"

	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
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
		sn, err = c.createAdjustCrpTicket(kt, ticket)
	default:
		logs.Errorf("failed to create crp ticket, unsupported ticket type: %s, rid: %s", ticket.Type, kt.Rid)
		return errors.New("unsupported ticket type")
	}
	if err != nil {
		// 这里主要返回的error是crp ticket创建失败，且ticket状态更新失败的日志在函数内已打印，这里可以忽略该错误
		_ = c.updateTicketStatusFailed(kt, ticket, err.Error())
		logs.Errorf("failed to create crp ticket with different ticket type, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// upsertCrpDemand crp tickets are approved and upsert hcm data.
func (c *Controller) upsertCrpDemand(kt *kit.Kit, ticket *TicketInfo) error {
	// call crp api to get ticket corresponding crp demand ids.
	demands, err := c.QueryIEGDemands(kt, &QueryIEGDemandsReq{CrpSns: []string{ticket.CrpSn}})
	if err != nil {
		logs.Errorf("failed to query ieg demands, err: %v, crp_sn: %s, rid: %s", err, ticket.CrpSn, kt.Rid)
		return err
	}

	crpDemandIDs := make([]int64, len(demands))
	for idx, demand := range demands {
		crpDemandID, err := strconv.ParseInt(demand.DemandId, 10, 64)
		if err != nil {
			logs.Errorf("failed to parse crp demand id, err: %v, demand_id: %s, rid: %s", err, demand.DemandId, kt.Rid)
			return err
		}
		crpDemandIDs[idx] = crpDemandID
	}

	// upsert crp demand id and biz relation.
	bizOrgRel := plan.BizOrgRel{
		BkBizID:         ticket.BkBizID,
		BkBizName:       ticket.BkBizName,
		OpProductID:     ticket.OpProductID,
		OpProductName:   ticket.OpProductName,
		PlanProductID:   ticket.PlanProductID,
		PlanProductName: ticket.PlanProductName,
		VirtualDeptID:   ticket.VirtualDeptID,
		VirtualDeptName: ticket.VirtualDeptName,
	}
	if err = c.upsertCrpDemandBizRel(kt, crpDemandIDs, ticket.DemandClass, bizOrgRel, ticket.Applicant); err != nil {
		logs.Errorf("failed to upsert crp demand biz relation, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// unlock all crp demands.
	allCrpDemandIDs := append([]int64{}, crpDemandIDs...)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allCrpDemandIDs = append(allCrpDemandIDs, (*demand.Original).CrpDemandID)
		}
	}
	if err = c.dao.ResPlanCrpDemand().UnlockAllResPlanDemand(kt, allCrpDemandIDs); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// createAddCrpTicket create add crp ticket.
func (c *Controller) createAddCrpTicket(kt *kit.Kit, ticket *TicketInfo) (string, error) {
	addReq, err := c.constructAddReq(kt, ticket)
	if err != nil {
		logs.Errorf("failed to construct add cvm & cbs plan order request, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	resp, err := c.crpCli.AddCvmCbsPlan(kt.Ctx, kt.Header(), addReq)
	if err != nil {
		logs.Errorf("failed to add cvm & cbs plan order, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to add cvm & cbs plan order, code: %d, msg: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, kt.Rid)
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to add cvm & cbs plan order, for return empty order id, rid: %s", kt.Rid)
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

	if err := c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 失败需要释放资源
	allCrpDemandIDs := make([]int64, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allCrpDemandIDs = append(allCrpDemandIDs, (*demand.Original).CrpDemandID)
		}
	}
	if err := c.dao.ResPlanCrpDemand().UnlockAllResPlanDemand(kt, allCrpDemandIDs); err != nil {
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
			DeptName: cvmapi.CvmLaunchDeptName,
			Items:    make([]*cvmapi.AddPlanItem, 0),
		},
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
			CityName:        demand.Updated.RegionName,
			ZoneName:        demand.Updated.ZoneName,
			CoreTypeName:    demand.Updated.Cvm.CoreType,
			InstanceModel:   demand.Updated.Cvm.DeviceType,
			CvmAmount:       demand.Updated.Cvm.Os,
			CoreAmount:      int(demand.Updated.Cvm.CpuCore),
			Desc:            demand.Updated.Remark,
			InstanceIO:      int(demand.Updated.Cbs.DiskIo),
			DiskTypeName:    demand.Updated.Cbs.DiskTypeName,
			DiskAmount:      int(demand.Updated.Cbs.DiskSize),
		}

		addReq.Params.Items = append(addReq.Params.Items, planItem)
	}

	return addReq, nil
}

// createAdjustCrpTicket create adjust crp ticket.
// ticket types RPTicketTypeAdjust and RPTicketTypeDelete are both belonged to adjust crp ticket.
func (c *Controller) createAdjustCrpTicket(kt *kit.Kit, ticket *TicketInfo) (string, error) {
	adjustReq, err := c.constructAdjustReq(kt, ticket)
	if err != nil {
		logs.Errorf("failed to construct adjust cvm & cbs plan order request, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	resp, err := c.crpCli.AdjustCvmCbsPlans(kt.Ctx, kt.Header(), adjustReq)
	if err != nil {
		logs.Errorf("failed to adjust cvm & cbs plan order, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to adjust cvm & cbs plan order, code: %d, msg: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, kt.Rid)
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.Data.OrderId
	if sn == "" {
		logs.Errorf("failed to adjust cvm & cbs plan order, for return empty order id, rid: %s", kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// constructAdjustReq construct cvm cbs plan adjust request.
func (c *Controller) constructAdjustReq(kt *kit.Kit, ticket *TicketInfo) (*cvmapi.CvmCbsPlanAdjustReq, error) {
	adjustReq := &cvmapi.CvmCbsPlanAdjustReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAutoAdjustMethod,
		},
		Params: &cvmapi.CvmCbsPlanAdjustParam{
			BaseInfo: &cvmapi.AdjustBaseInfo{
				DeptId:          cvmapi.CvmDeptId,
				DeptName:        cvmapi.CvmLaunchDeptName,
				PlanProductName: ticket.PlanProductName,
			},
			SrcData:     make([]*cvmapi.AdjustSrcData, 0),
			UpdatedData: make([]*cvmapi.AdjustUpdatedData, 0),
			UserName:    ticket.Applicant,
		},
	}

	crpDemandIDs := make([]int64, len(ticket.Demands))
	for idx, demand := range ticket.Demands {
		if demand.Original == nil {
			logs.Errorf("failed to construct adjust request, demand original is nil, rid: %s", kt.Rid)
			return nil, errors.New("demand original is nil")
		}

		crpDemandIDs[idx] = (*demand.Original).CrpDemandID
	}

	crpDemandMap, err := c.getCrpDemandMap(kt, crpDemandIDs)
	if err != nil {
		logs.Errorf("failed to get crp demand map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, demand := range ticket.Demands {
		crpDemand, ok := crpDemandMap[(*demand.Original).CrpDemandID]
		if !ok {
			logs.Errorf("failed to construct adjust request, crp demand id not found, rid: %s", kt.Rid)
			return nil, errors.New("crp demand id not found")
		}

		var adjustType enumor.CrpAdjustType
		switch ticket.Type {
		case enumor.RPTicketTypeAdjust:
			adjustType = enumor.CrpAdjustTypeUpdate
			if demand.Updated.ExpectTime != demand.Original.ExpectTime {
				adjustType = enumor.CrpAdjustTypeDelay
			}
		case enumor.RPTicketTypeDelete:
			adjustType = enumor.CrpAdjustTypeCancel
		default:
			logs.Errorf("unsupported ticket type: %s", ticket.Type)
			return nil, errors.New("unsupported ticket type")
		}

		srcItem := &cvmapi.AdjustSrcData{
			AdjustType:          string(adjustType),
			CvmCbsPlanQueryItem: crpDemand.Clone(),
		}
		adjustReq.Params.SrcData = append(adjustReq.Params.SrcData, srcItem)

		updatedItem, err := c.constructAdjustUpdatedData(kt, adjustType, ticket.PlanProductName, crpDemand, demand)
		if err != nil {
			logs.Errorf("failed to construct adjust updated data, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if updatedItem != nil {
			adjustReq.Params.UpdatedData = append(adjustReq.Params.UpdatedData, updatedItem)
		}
	}

	return adjustReq, nil
}

// getCrpDemandMap get crp demand id and detail map.
func (c *Controller) getCrpDemandMap(kt *kit.Kit, crpDemandIDs []int64) (map[int64]*cvmapi.CvmCbsPlanQueryItem, error) {
	crpDemands, err := c.QueryIEGDemands(kt, &QueryIEGDemandsReq{CrpDemandIDs: crpDemandIDs})
	if err != nil {
		logs.Errorf("failed to query ieg demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	crpDemandMap := make(map[int64]*cvmapi.CvmCbsPlanQueryItem)
	for _, crpDemand := range crpDemands {
		demandID, err := strconv.ParseInt(crpDemand.DemandId, 10, 64)
		if err != nil {
			logs.Errorf("failed to parse crp demand id, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		crpDemandMap[demandID] = crpDemand
	}

	return crpDemandMap, nil
}

// constructAdjustUpdatedData construct adjust updated data.
// if adjust type is update, updated item are normal.
// if adjust type is delay, updated will fill parameter TimeAdjustCvmAmount with remainOs.
// if adjust type is cancel, updated will be empty.
func (c *Controller) constructAdjustUpdatedData(kt *kit.Kit, adjustType enumor.CrpAdjustType, planProdName string,
	crpSrcDemand *cvmapi.CvmCbsPlanQueryItem, demand rpt.ResPlanDemand) (*cvmapi.AdjustUpdatedData, error) {

	// if adjust type is cancel, updated will be empty.
	if adjustType == enumor.CrpAdjustTypeCancel {
		return nil, nil
	}

	// init updated data.
	updatedData := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(adjustType),
		CvmCbsPlanQueryItem: crpSrcDemand.Clone(),
	}

	// supplement updated data.
	updatedData.CityName = demand.Updated.RegionName
	updatedData.ZoneName = demand.Updated.ZoneName
	updatedData.InstanceModel = demand.Updated.Cvm.DeviceType
	updatedData.CvmAmount = float32(demand.Updated.Cvm.Os)
	updatedData.CoreAmount = float32(demand.Updated.Cvm.CpuCore)
	updatedData.InstanceIO = int(demand.Updated.Cbs.DiskIo)
	updatedData.DiskTypeName = demand.Updated.Cbs.DiskTypeName
	updatedData.AllDiskAmount = float32(demand.Updated.Cbs.DiskSize)
	updatedData.ProjectName = string(demand.Updated.ObsProject)
	updatedData.UseTime = demand.Updated.ExpectTime
	updatedData.PlanProductName = planProdName

	switch adjustType {
	case enumor.CrpAdjustTypeUpdate:
		// do nothing.
	case enumor.CrpAdjustTypeDelay:
		updatedData.TimeAdjustCvmAmount = crpSrcDemand.CvmAmount
	default:
		logs.Errorf("invalid adjust type: %s, rid: %s", adjustType, kt.Rid)
		return nil, errors.New("invalid adjust type")
	}

	return updatedData, nil
}

// upsertCrpDemandBizRel upsert crp demand biz rel.
func (c *Controller) upsertCrpDemandBizRel(kt *kit.Kit, crpDemandIDs []int64, demandClass enumor.DemandClass,
	bizOrgRel plan.BizOrgRel, reviser string) error {

	if len(crpDemandIDs) == 0 {
		logs.Errorf("crp demand ids is empty, rid: %s", kt.Rid)
		return errors.New("crp demand ids is empty")
	}

	listOpt := &types.ListOption{
		Filter: tools.ContainersExpression("crp_demand_id", crpDemandIDs),
		Page:   core.NewDefaultBasePage(),
	}
	rst, err := c.dao.ResPlanCrpDemand().List(kt, listOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	existCrpDemandMap := converter.SliceToMap(rst.Details, func(detail rpcd.ResPlanCrpDemandTable) (int64, struct{}) {
		return detail.CrpDemandID, struct{}{}
	})

	existCrpDemandIDs := make([]int64, 0)
	notExistCrpDemandIDs := make([]int64, 0)
	for _, crpDemandID := range crpDemandIDs {
		if _, exist := existCrpDemandMap[crpDemandID]; exist {
			existCrpDemandIDs = append(existCrpDemandIDs, crpDemandID)
		} else {
			notExistCrpDemandIDs = append(notExistCrpDemandIDs, crpDemandID)
		}
	}

	if len(existCrpDemandIDs) > 0 {
		update := &rpcd.ResPlanCrpDemandTable{
			Locked:          nil,
			DemandClass:     demandClass,
			BkBizID:         bizOrgRel.BkBizID,
			BkBizName:       bizOrgRel.BkBizName,
			OpProductID:     bizOrgRel.OpProductID,
			OpProductName:   bizOrgRel.OpProductName,
			PlanProductID:   bizOrgRel.PlanProductID,
			PlanProductName: bizOrgRel.PlanProductName,
			VirtualDeptID:   bizOrgRel.VirtualDeptID,
			VirtualDeptName: bizOrgRel.VirtualDeptName,
			Reviser:         reviser,
		}
		err = c.dao.ResPlanCrpDemand().Update(kt, tools.ContainersExpression("crp_demand_id", existCrpDemandIDs),
			update)
		if err != nil {
			logs.Errorf("failed to update resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(notExistCrpDemandIDs) > 0 {
		inserts := make([]rpcd.ResPlanCrpDemandTable, len(notExistCrpDemandIDs))
		for idx, crpDemandID := range notExistCrpDemandIDs {
			inserts[idx] = rpcd.ResPlanCrpDemandTable{
				CrpDemandID:     crpDemandID,
				Locked:          converter.ValToPtr(enumor.CrpDemandUnLocked),
				DemandClass:     demandClass,
				BkBizID:         bizOrgRel.BkBizID,
				BkBizName:       bizOrgRel.BkBizName,
				OpProductID:     bizOrgRel.OpProductID,
				OpProductName:   bizOrgRel.OpProductName,
				PlanProductID:   bizOrgRel.PlanProductID,
				PlanProductName: bizOrgRel.PlanProductName,
				VirtualDeptID:   bizOrgRel.VirtualDeptID,
				VirtualDeptName: bizOrgRel.VirtualDeptName,
				Creator:         reviser,
				Reviser:         reviser,
			}
		}
		_, err = c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
			return c.dao.ResPlanCrpDemand().CreateWithTx(kt, txn, inserts)
		})
		if err != nil {
			logs.Errorf("failed to create resource plan crp demand, err: %v, rid: %s", err, kt.Rid)
		}
		return err
	}

	return nil
}
