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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	tasktypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/times"
	"hcm/pkg/tools/utils/wait"
)

// Logics provides management interface for resource plan.
type Logics interface {
	// CreateAuditFlow creates an audit flow for resource plan ticket.
	CreateAuditFlow(kt *kit.Kit, ticketID string) error
	// CreateResPlanTicket create resource plan ticket.
	CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error)
	// QueryIEGDemands query IEG crp demands.
	QueryIEGDemands(kt *kit.Kit, req *QueryIEGDemandsReq) ([]*cvmapi.CvmCbsPlanQueryItem, error)
	// ExamineDemandClass examine whether all demands are the same demand class, and return the demand class.
	ExamineDemandClass(kt *kit.Kit, crpDemandIDs []int64) (enumor.DemandClass, error)
	// IsDeviceMatched return whether each device type in deviceTypeSlice can use deviceType's resource plan.
	IsDeviceMatched(kt *kit.Kit, deviceTypeSlice []string, deviceType string) ([]bool, error)
	// GetProdResPlanPool get op product resource plan pool.
	GetProdResPlanPool(kt *kit.Kit, prodID int64) (ResPlanPool, error)
	// GetProdResConsumePool get op product resource consume pool.
	GetProdResConsumePool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, error)
	// GetProdResRemainPool get op product resource remain pool.
	// @param prodID is the op product id.
	// @param planProdID is the corresponding plan product id of the op product id.
	// @return prodRemainedPool is the op product in plan and out plan remained resource plan pool.
	// @return prodMaxAvailablePool is the op product in plan and out plan remained max available resource plan pool.
	// NOTE: maxAvailableInPlanPool = totalInPlan * 120% - consumeInPlan, because the special rules of the crp system.
	GetProdResRemainPool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, ResPlanPool, error)
	// VerifyProdDemands verify whether the needs of op product can be satisfied.
	VerifyProdDemands(kt *kit.Kit, prodID, planProdID int64, needs []VerifyResPlanElem) ([]VerifyResPlanResElem, error)
	// GetProdResConsumePoolV2 get biz resource consume pool.
	GetProdResConsumePoolV2(kt *kit.Kit, bkBizIDs []int64, startDay, endDay time.Time) (ResPlanConsumePool, error)
	// VerifyProdDemandsV2 verify whether the needs of biz can be satisfied.
	VerifyProdDemandsV2(kt *kit.Kit, bkBizID int64, needs []VerifyResPlanElemV2) ([]VerifyResPlanResElem, error)
	// AddMatchedPlanDemandExpendLogs add matched plan demand expend logs.
	AddMatchedPlanDemandExpendLogs(kt *kit.Kit, bkBizID int64, subOrder *tasktypes.ApplyOrder) error
	// GetAllDeviceTypeMap get all device type map.
	GetAllDeviceTypeMap(kt *kit.Kit) (map[string]wdt.WoaDeviceTypeTable, error)
}

// Controller motivates the resource plan ticket status flow.
type Controller struct {
	dao          dao.Set
	sd           serviced.State
	client       *client.ClientSet
	itsmCli      itsm.Client
	itsmFlow     cc.ItsmFlow
	crpAuditNode cc.StateNode
	crpCli       cvmapi.CVMClientInterface
	workQueue    *UniQueue
	ctx          context.Context
}

const (
	// TicketSvcNameResPlan 资源预测在ITSM的服务
	TicketSvcNameResPlan = "res_plan"
	// TicketNodeNameCrpAudit 资源预测在ITSM流程中的CRP审批节点
	TicketNodeNameCrpAudit = "crp_audit"
	// TicketOperatorNameCrpAudit 资源预测在ITSM流程中的CRP审批节点操作人
	TicketOperatorNameCrpAudit = "icr"
	// AuditFlowTimeoutDay 审批流超时时间，单位天
	AuditFlowTimeoutDay int = 5
	// PendingTicketTraceDay 带处理的单据历史追溯时间，单位天
	PendingTicketTraceDay int = 7
)

// New creates a resource plan ticket controller instance.
func New(sd serviced.State, client *client.ClientSet, dao dao.Set, itsmCli itsm.Client,
	crpCli cvmapi.CVMClientInterface) (*Controller, error) {

	q := NewUniQueue()

	var itsmFlowCfg cc.ItsmFlow
	for _, itsmFlow := range cc.WoaServer().ItsmFlows {
		if itsmFlow.ServiceName == TicketSvcNameResPlan {
			itsmFlowCfg = itsmFlow
			break
		}
	}

	var crpAuditNode cc.StateNode
	for _, node := range itsmFlowCfg.StateNodes {
		if node.NodeName == TicketNodeNameCrpAudit {
			crpAuditNode = node
		}
	}

	ctrl := &Controller{
		dao:          dao,
		sd:           sd,
		client:       client,
		itsmCli:      itsmCli,
		itsmFlow:     itsmFlowCfg,
		crpAuditNode: crpAuditNode,
		crpCli:       crpCli,
		workQueue:    q,
		ctx:          context.Background(),
	}

	go ctrl.Run()

	return ctrl, nil
}

// Run starts dispatcher
func (c *Controller) Run() {
	// TODO: get interval from config
	// list and watch tickets every 2 minutes
	go wait.JitterUntil(c.listAndWatchTickets, 2*time.Minute, 0.5, true, c.ctx)

	// TODO: get worker num from config
	for i := 0; i < 10; i++ {
		// get and handle tickets every 2 minutes
		go wait.JitterUntil(c.runWorker, 2*time.Minute, 0.5, true, c.ctx)
	}

	select {
	case <-c.ctx.Done():
		logs.Infof("resource plan ticket controller exits")
	}
}

func (c *Controller) listAndWatchTickets() error {
	logs.Infof("ready to list and watch tickets")
	if !c.sd.IsMaster() {
		// pop all pending orders
		c.workQueue.Clear()
		return nil
	}

	// list pending orders
	kt := kit.New()
	pendingTkIDs, err := c.listAllPendingTickets(kt)
	if err != nil {
		logs.Errorf("failed to list pending resource plan tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// enqueue pending orders
	for _, tkID := range pendingTkIDs {
		c.workQueue.Enqueue(tkID)
	}

	return nil
}

func (c *Controller) listAllPendingTickets(kt *kit.Kit) ([]string, error) {
	// list tickets of recent 7 days.
	dr := &times.DateRange{
		Start: time.Now().AddDate(0, 0, -PendingTicketTraceDay).Format(constant.DateLayout),
		End:   time.Now().Format(constant.DateLayout),
	}

	drOpt, err := tools.DateRangeExpression("submitted_at", dr)
	if err != nil {
		return nil, err
	}

	// TODO: 当单据数量超过500时，可能会漏单据。这里改为分页查询
	recentOpt := &types.ListOption{
		Fields: []string{"id"},
		Filter: drOpt,
		Page:   core.NewDefaultBasePage(),
	}

	tkRst, err := c.dao.ResPlanTicket().List(kt, recentOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	recentTkIDs := make([]string, 0)
	for _, ticket := range tkRst.Details {
		recentTkIDs = append(recentTkIDs, ticket.ID)
	}

	// list tickets with auditing status
	auditOpt := &types.ListOption{
		Fields: []string{"ticket_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("ticket_id", recentTkIDs),
			tools.RuleEqual("status", enumor.RPTicketStatusAuditing),
		),
		Page: core.NewDefaultBasePage(),
	}

	statusRst, err := c.dao.ResPlanTicketStatus().List(kt, auditOpt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pendingTkIDs := make([]string, 0)
	for _, ticket := range statusRst.Details {
		pendingTkIDs = append(pendingTkIDs, ticket.TicketID)
	}

	return pendingTkIDs, nil
}

func (c *Controller) runWorker() error {
	logs.Infof("ready to run worker")

	// only master node handle plan tickets.
	if !c.sd.IsMaster() {
		return nil
	}

	// get one ticket from the work queue
	tkID, ok := c.workQueue.Pop()
	if !ok {
		return nil
	}

	logs.Infof("ready to handle ticket %s", tkID)

	// check the status of the ticket
	kt := kit.New()
	tkInfo, err := c.getTicketInfo(kt, tkID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v, id: %s, rid: %s", err, tkID, kt.Rid)
		return err
	}

	if tkInfo.Status != enumor.RPTicketStatusAuditing {
		logs.Warnf("need not handle ticket for its status %s != %s, id: %s, rid: %s", tkInfo.Status,
			enumor.RPTicketStatusAuditing, tkID, kt.Rid)
		return nil
	}

	if tkInfo.ItsmSn == "" {
		logs.Errorf("failed to handle ticket for itsm sn is empty, id: %s, rid: %s", tkID, kt.Rid)
		return errors.New("failed to handle ticket for itsm sn is empty")
	}

	if tkInfo.CrpSn != "" {
		return c.checkCrpTicket(kt, tkInfo)
	}

	return c.checkItsmTicket(kt, tkInfo)
}

func convListResPlanDemandItem(items []*cvmapi.CvmCbsPlanQueryItem) []*ptypes.PlanDemandDetail {
	rst := make([]*ptypes.PlanDemandDetail, 0, len(items))
	for _, item := range items {
		rstItem := &ptypes.PlanDemandDetail{
			GetPlanDemandDetailResp: ptypes.GetPlanDemandDetailResp{
				CrpDemandID:        item.DemandId,
				YearMonthWeek:      item.YearMonthWeek,
				ExpectStartDate:    item.ExpectStartDate,
				ExpectEndDate:      item.ExpectEndDate,
				ExpectTime:         item.UseTime,
				BgID:               int64(item.BgId),
				BgName:             item.BgName,
				DeptID:             int64(item.DeptId),
				DeptName:           item.DeptName,
				PlanProductID:      int64(item.PlanProductId),
				PlanProductName:    item.PlanProductName,
				ObsProject:         item.ProjectName,
				RegionName:         item.CityName,
				ZoneName:           item.ZoneName,
				PlanType:           enumor.PlanType(item.InPlan).ToAnotherPlanType(),
				PlanAdvanceWeek:    item.PlanWeek,
				ExpeditedPostponed: item.ExpeditedPostponed,
				CoreTypeID:         item.CoreType,
				CoreType:           item.CoreTypeName,
				DeviceFamily:       item.InstanceFamily,
				DeviceClass:        item.InstanceType,
				DeviceType:         item.InstanceModel,
				OS:                 item.PlanCvmAmount,
				Memory:             item.PlanRamAmount,
				CpuCore:            item.PlanCoreAmount,
				DiskSize:           item.PlanDiskAmount,
				DiskIO:             item.InstanceIO,
				DiskTypeName:       item.DiskTypeName,
				DemandWeek:         item.RequirementWeekType,
				ResPoolType:        item.ResourcePoolType,
				ResPool:            item.ResourcePoolName,
				ResMode:            item.ResourceMode,
				GenerationType:     item.GenerationType,
			},
			Year:             item.Year,
			Month:            item.Month,
			Week:             item.Week,
			TotalOS:          item.PlanCvmAmount,
			AppliedOS:        item.ApplyCvmAmount,
			RemainedOS:       item.RealCvmAmount,
			TotalCpuCore:     item.PlanCoreAmount,
			AppliedCpuCore:   item.ApplyCoreAmount,
			RemainedCpuCore:  item.RealCoreAmount,
			TotalMemory:      item.PlanRamAmount,
			AppliedMemory:    item.ApplyRamAmount,
			RemainedMemory:   item.RealRamAmount,
			TotalDiskSize:    item.PlanDiskAmount,
			AppliedDiskSize:  item.ApplyDiskAmount,
			RemainedDiskSize: item.RealDiskAmount,
		}
		rst = append(rst, rstItem)
	}
	return rst
}

// ListCrpDemands 返回全量数据，listReq 中的分页参数将被忽略
func (c *Controller) ListCrpDemands(kt *kit.Kit, listReq *ptypes.ListResPlanDemandReq) (
	[]*ptypes.PlanDemandDetail, error) {

	params := &cvmapi.CvmCbsPlanQueryParam{
		BgName: []string{cvmapi.CvmCbsPlanQueryBgName}, // 强制仅查询IEG的预测需求
	}
	if len(listReq.CrpDemandIDs) > 0 {
		params.DemandIdList = listReq.CrpDemandIDs
	}
	if len(listReq.DeviceClasses) > 0 {
		params.InstanceType = listReq.DeviceClasses
	}
	if len(listReq.ObsProjects) > 0 {
		params.ProjectName = listReq.ObsProjects
	}
	if len(listReq.RegionNames) > 0 {
		params.CityName = listReq.RegionNames
	}
	if len(listReq.ZoneNames) > 0 {
		params.ZoneName = listReq.ZoneNames
	}
	if len(listReq.OrderIDs) > 0 {
		params.OrderIdList = listReq.OrderIDs
	}
	if listReq.ExpectTimeRange != nil {
		params.UseTime = &cvmapi.UseTime{
			Start: listReq.ExpectTimeRange.Start,
			End:   listReq.ExpectTimeRange.End,
		}
	}
	// 500条一组查询出全部结果
	page := &cvmapi.Page{
		Start: 0,
		Size:  int(core.DefaultMaxPageLimit),
	}

	rst := make([]*ptypes.PlanDemandDetail, 0)
	for {
		resp, err := c.listCrpDemandsPage(kt, params, page)
		if err != nil {
			return nil, err
		}

		rst = append(rst, resp...)

		if len(resp) < page.Size {
			break
		}
		page.Start += page.Size
	}
	return rst, nil
}

// ListCrpDemandsPage 分页查询
func (c *Controller) listCrpDemandsPage(kt *kit.Kit, params *cvmapi.CvmCbsPlanQueryParam, page *cvmapi.Page) (
	[]*ptypes.PlanDemandDetail, error) {

	params.Page = page
	req := &cvmapi.CvmCbsPlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanQueryMethod,
		},
		Params: params,
	}

	resp, err := c.crpCli.QueryCvmCbsPlans(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to list crp demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to list crp demand, code: %d, msg: %s, rid: %s, params: %+v", resp.Error.Code,
			resp.Error.Message,
			kt.Rid, req.Params)
		return nil, fmt.Errorf("failed to list crp demand, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to list crp demand, for result is empty, rid: %s, params: %+v", kt.Rid, req.Params)
		return nil, errors.New("failed to list crp demand, for result is empty")
	}

	return convListResPlanDemandItem(resp.Result.Data), nil
}

func convListDemandChangeLogItem(items []*cvmapi.DemandChangeLogQueryLogItem) []*ptypes.ListDemandChangeLogItem {
	rst := make([]*ptypes.ListDemandChangeLogItem, 0, len(items))
	for _, item := range items {
		rstItem := &ptypes.ListDemandChangeLogItem{
			CrpDemandId:       int64(item.DemandId),
			ExpectTime:        item.UseTime,
			BgName:            item.BgName,
			DeptName:          item.DeptName,
			PlanProductName:   item.PlanProductName,
			ObsProject:        item.ProjectName,
			RegionName:        item.CityName,
			ZoneName:          item.ZoneName,
			DemandWeek:        item.RequirementWeekType,
			ResPoolType:       item.ResourcePoolType,
			DeviceClass:       item.InstanceType,
			DeviceType:        item.InstanceModel,
			ChangeCvmAmount:   item.ChangeCvmAmount,
			AfterCvmAmount:    item.AfterCvmAmount,
			ChangeCoreAmount:  item.ChangeCoreAmount,
			AfterCoreAmount:   item.AfterCoreAmount,
			ChangeRamAmount:   item.ChangeRamAmount,
			AfterRamAmount:    item.AfterRamAmount,
			DiskType:          item.DiskTypeName,
			DiskIo:            item.InstanceIO,
			ChangedDiskAmount: item.ChangedDiskAmount,
			AfterDiskAmount:   item.AfterDiskAmount,
			DemandSource:      item.SourceType,
			CrpSn:             item.OrderId,
			CreateTime:        item.CreateTime,
			Remark:            item.Desc,
			ResPool:           item.ResourcePoolName,
		}
		rst = append(rst, rstItem)
	}
	return rst
}

// ListCrpDemandChangeLog list crp demand change log by demand id, full query
func (c *Controller) ListCrpDemandChangeLog(kt *kit.Kit, crpDemandId int64) (
	[]*ptypes.ListDemandChangeLogItem, error) {

	// 500条一组查询出全部结果
	page := &cvmapi.Page{
		Start: 0,
		Size:  int(core.DefaultMaxPageLimit),
	}

	rst := make([]*ptypes.ListDemandChangeLogItem, 0)
	for {
		resp, err := c.listCrpDemandChangeLogPage(kt, crpDemandId, page)
		if err != nil {
			return nil, err
		}

		rst = append(rst, resp...)

		if len(resp) < page.Size {
			break
		}
		page.Start += page.Size
	}
	return rst, nil
}

// listCrpDemandChangeLogPage list crp demand change log by demand id, page query
func (c *Controller) listCrpDemandChangeLogPage(kt *kit.Kit, crpDemandId int64, page *cvmapi.Page) (
	[]*ptypes.ListDemandChangeLogItem, error) {

	rst := make([]*ptypes.ListDemandChangeLogItem, 0)
	req := &cvmapi.DemandChangeLogQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsDemandChangeLogQueryMethod,
		},
		Params: &cvmapi.DemandChangeLogQueryParam{
			DemandIdList: []int64{crpDemandId},
			Page:         page,
		},
	}
	resp, err := c.crpCli.QueryDemandChangeLog(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to list crp demand change log, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to list crp demand change log, code: %d, msg: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, kt.Rid)
		return nil, fmt.Errorf("failed to list crp demand change log, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to list crp demand change log, for result is empty, rid: %s", kt.Rid)
		return nil, errors.New("failed to list crp demand change log, for result is empty")
	}

	if len(resp.Result.Data) < 1 {
		return rst, nil
	}

	return convListDemandChangeLogItem(resp.Result.Data[0].Info), nil
}

func (c *Controller) checkCrpTicket(kt *kit.Kit, ticket *TicketInfo) error {
	logs.Infof("ready to check crp flow, sn: %s, id: %s", ticket.CrpSn, ticket.ID)

	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderQueryMethod,
		},
		Params: &cvmapi.QueryPlanOrderParam{
			OrderIds: []string{ticket.CrpSn},
		},
	}
	resp, err := c.crpCli.QueryPlanOrder(kt.Ctx, nil, req)
	if err != nil {
		logs.Errorf("failed to query crp plan order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to query crp plan order, code: %d, msg: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, kt.Rid)
		return fmt.Errorf("failed to query crp plan order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to query crp plan order, for result is empty, rid: %s", kt.Rid)
		return errors.New("failed to query crp plan order, for result is empty")
	}

	planItem, ok := resp.Result[ticket.CrpSn]
	if !ok {
		logs.Errorf("query crp plan order return no result by sn: %s, rid: %s", ticket.CrpSn, kt.Rid)
		return fmt.Errorf("query crp plan order return no result by sn: %s", ticket.CrpSn)
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSn:   ticket.ItsmSn,
		ItsmUrl:  ticket.ItsmUrl,
		CrpSn:    ticket.CrpSn,
		CrpUrl:   ticket.CrpUrl,
	}

	switch planItem.Data.BaseInfo.Status {
	case cvmapi.PlanOrderStatusRejected:
		update.Status = enumor.RPTicketStatusRejected
	case cvmapi.PlanOrderStatusApproved:
		return c.finishAuditFlow(kt, ticket)
	default:
		return c.checkTicketTimeout(kt, ticket)
	}

	if err := c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 单据被拒需要释放资源
	if update.Status != enumor.RPTicketStatusRejected {
		return nil
	}
	allCrpDemandIDs := make([]int64, 0)
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

func (c *Controller) checkItsmTicket(kt *kit.Kit, ticket *TicketInfo) error {
	logs.Infof("ready to check itsm flow, sn: %s, id: %s", ticket.ItsmSn, ticket.ID)

	resp, err := c.itsmCli.GetTicketStatus(kt, ticket.ItsmSn)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusRejected,
		ItsmSn:   ticket.ItsmSn,
		ItsmUrl:  ticket.ItsmUrl,
	}

	switch resp.Data.CurrentStatus {
	case string(itsm.StatusFinished), string(itsm.StatusTerminated):
		// rejected
		update.Status = enumor.RPTicketStatusRejected
	case string(itsm.StatusRevoked):
		// revoked
		update.Status = enumor.RPTicketStatusRevoked
	case string(itsm.StatusRunning):
		// check if CRP audit state
		if len(resp.Data.CurrentSteps) == 0 {
			return c.checkTicketTimeout(kt, ticket)
		}

		if resp.Data.CurrentSteps[0].StateId != c.crpAuditNode.ID {
			return c.checkTicketTimeout(kt, ticket)
		}

		// CRP audit state, create CRP ticket
		return c.createCrpTicket(kt, ticket)
	default:
		return c.checkTicketTimeout(kt, ticket)
	}

	if err = c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 单据被拒需要释放资源
	if update.Status != enumor.RPTicketStatusRejected && update.Status != enumor.RPTicketStatusRevoked {
		return nil
	}
	allCrpDemandIDs := make([]int64, 0)
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

func (c *Controller) finishAuditFlow(kt *kit.Kit, ticket *TicketInfo) error {
	itsmStatus, err := c.itsmCli.GetTicketStatus(kt, ticket.ItsmSn)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	// check if CRP audit state
	if len(itsmStatus.Data.CurrentSteps) == 0 {
		return c.checkTicketTimeout(kt, ticket)
	}

	if itsmStatus.Data.CurrentSteps[0].StateId != c.crpAuditNode.ID {
		return c.checkTicketTimeout(kt, ticket)
	}

	approveReq := &itsm.ApproveReq{
		Sn:       ticket.ItsmSn,
		StateID:  int(c.crpAuditNode.ID),
		Approver: c.crpAuditNode.Approver,
		Action:   "true",
		Remark:   "",
	}
	if err := c.itsmCli.Approve(kt, approveReq); err != nil {
		logs.Errorf("request itsm ticket approve failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticket.ID,
		Status:   enumor.RPTicketStatusDone,
		ItsmSn:   ticket.ItsmSn,
		ItsmUrl:  ticket.ItsmUrl,
		CrpSn:    ticket.CrpSn,
		CrpUrl:   ticket.CrpUrl,
	}

	if err := c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// crp单据通过后更新本地数据表
	if err := c.upsertCrpDemand(kt, ticket); err != nil {
		logs.Errorf("failed to upsert crp demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *Controller) checkTicketTimeout(kt *kit.Kit, ticket *TicketInfo) error {
	submitTime, err := time.Parse(constant.TimeStdFormat, ticket.SubmittedAt)
	if err != nil {
		logs.Errorf("failed to parse ticket submit time %s, err: %v, rid: %s", ticket.SubmittedAt, err, kt.Rid)
		return err
	}

	// set timeout as 5 days
	if time.Now().Before(submitTime.AddDate(0, 0, AuditFlowTimeoutDay)) {
		return nil
	}

	return c.updateTicketStatusFailed(kt, ticket, "audit flow timeout")
}

// CreateAuditFlow creates an audit flow for resource plan ticket.
func (c *Controller) CreateAuditFlow(kt *kit.Kit, ticketID string) error {
	ticket, err := c.getTicketInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v", err)
		return err
	}

	sn, err := c.createItsmTicket(kt, ticket)
	if err != nil {
		logs.Errorf("failed to create itsm ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	itsmStatus, err := c.itsmCli.GetTicketStatus(kt, sn)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticketID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSn:   sn,
		ItsmUrl:  itsmStatus.Data.TicketUrl,
	}

	if err = c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *Controller) getTicketInfo(kt *kit.Kit, ticketID string) (*TicketInfo, error) {
	base, err := c.getTicketBaseInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket base info, err: %v", err)
		return nil, err
	}

	var demands rpt.ResPlanDemands
	if err = json.Unmarshal([]byte(base.Demands), &demands); err != nil {
		logs.Errorf("failed to unmarshal demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	status, err := c.getTicketStatusInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket status info, err: %v", err)
		return nil, err
	}

	brief := &TicketInfo{
		ID:               ticketID,
		Type:             base.Type,
		Applicant:        base.Applicant,
		BkBizID:          base.BkBizID,
		BkBizName:        base.BkBizName,
		OpProductID:      base.OpProductID,
		OpProductName:    base.OpProductName,
		PlanProductID:    base.PlanProductID,
		PlanProductName:  base.PlanProductName,
		VirtualDeptID:    base.VirtualDeptID,
		VirtualDeptName:  base.VirtualDeptName,
		DemandClass:      base.DemandClass,
		OriginalCpuCore:  base.OriginalCpuCore,
		OriginalMemory:   base.OriginalMemory,
		OriginalDiskSize: base.OriginalDiskSize,
		UpdatedCpuCore:   base.UpdatedCpuCore,
		UpdatedMemory:    base.UpdatedMemory,
		UpdatedDiskSize:  base.UpdatedDiskSize,
		Demands:          demands,
		SubmittedAt:      base.SubmittedAt,
		Status:           status.Status,
		ItsmSn:           status.ItsmSn,
		ItsmUrl:          status.ItsmUrl,
		CrpSn:            status.CrpSn,
		CrpUrl:           status.CrpUrl,
	}

	return brief, nil
}

func (c *Controller) getTicketBaseInfo(kt *kit.Kit, ticketID string) (*rpt.ResPlanTicketTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := c.dao.ResPlanTicket().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket, but len details != 1")
	}

	return &rst.Details[0], nil
}

func (c *Controller) getTicketStatusInfo(kt *kit.Kit, ticketID string) (*rpts.ResPlanTicketStatusTable, error) {
	// search resource plan ticket table.
	opt := &types.ListOption{
		Filter: tools.EqualExpression("ticket_id", ticketID),
		Page:   core.NewDefaultBasePage(),
	}

	rst, err := c.dao.ResPlanTicketStatus().List(kt, opt)
	if err != nil {
		logs.Errorf("failed to list resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(rst.Details) != 1 {
		logs.Errorf("list resource plan ticket status, but len details != 1, rid: %s", kt.Rid)
		return nil, errors.New("list resource plan ticket status, but len details != 1")
	}

	return &rst.Details[0], nil
}

func (c *Controller) createItsmTicket(kt *kit.Kit, ticket *TicketInfo) (string, error) {
	if ticket == nil {
		return "", errors.New("ticket is nil")
	}

	// TODO：待修改
	contentTemplate := `业务：%s(%d)
预测类型：%s
CPU总核数：%.2f
内存总量(GB)：%.2f
云盘总量(GB)：%.2f
`
	content := fmt.Sprintf(contentTemplate, ticket.BkBizName, ticket.BkBizID, ticket.DemandClass, ticket.UpdatedCpuCore,
		ticket.UpdatedMemory, ticket.UpdatedDiskSize)
	createTicketReq := &itsm.CreateTicketParams{
		ServiceID:      c.itsmFlow.ServiceID,
		Creator:        ticket.Applicant,
		Title:          fmt.Sprintf("%s(业务ID: %d)资源预测申请", ticket.BkBizName, ticket.BkBizID),
		ContentDisplay: content,
		ExtraFields:    map[string]interface{}{"res_plan_url": fmt.Sprintf(c.itsmFlow.RedirectUrlTemplate, ticket.ID)},
	}

	sn, err := c.itsmCli.CreateTicket(kt, createTicketReq)
	if err != nil {
		logs.Errorf("create itsm ticket failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return sn, nil
}

func (c *Controller) updateTicketStatus(kt *kit.Kit, ticket *rpts.ResPlanTicketStatusTable) error {
	expr := tools.EqualExpression("ticket_id", ticket.TicketID)
	if err := c.dao.ResPlanTicketStatus().Update(kt, expr, ticket); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
