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

// Package scheduler 调度器
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/dal"
	"hcm/cmd/woa-server/common/language"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/scheduler/dispatcher"
	"hcm/cmd/woa-server/logics/task/scheduler/generator"
	"hcm/cmd/woa-server/logics/task/scheduler/matcher"
	"hcm/cmd/woa-server/logics/task/scheduler/recommender"
	"hcm/cmd/woa-server/logics/task/scheduler/record"
	"hcm/cmd/woa-server/model/task"
	configtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"

	"go.mongodb.org/mongo-driver/mongo"
)

// Interface scheduler interface
type Interface interface {
	// UpdateApplyTicket creates or updates resource apply ticket
	UpdateApplyTicket(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error)
	// GetApplyTicket gets resource apply ticket
	GetApplyTicket(kit *kit.Kit, param *types.GetApplyTicketReq) (*types.GetApplyTicketRst, error)
	// GetApplyAudit gets resource apply ticket audit info
	GetApplyAudit(kit *kit.Kit, param *types.GetApplyAuditReq) (*types.GetApplyAuditRst, error)
	// AuditTicket audit resource apply ticket
	AuditTicket(kit *kit.Kit, param *types.ApplyAuditReq) error
	// AutoAuditTicket system automatic audit resource apply ticket callback
	AutoAuditTicket(kit *kit.Kit, param *types.ApplyAutoAuditReq) (*types.ApplyAutoAuditRst, error)
	// ApproveTicket approve apply ticket callback
	ApproveTicket(kit *kit.Kit, param *types.ApproveApplyReq) error

	// CreateApplyOrder creates resource apply order
	CreateApplyOrder(kit *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error)
	// GetApplyOrder gets resource apply order info
	GetApplyOrder(kit *kit.Kit, param *types.GetApplyParam) (*types.GetApplyOrderRst, error)
	// GetApplyDetail gets resource apply order detail info
	GetApplyDetail(kit *kit.Kit, param *types.GetApplyDetailReq) (*types.GetApplyDetailRst, error)
	// GetApplyGenerate gets resource apply order generate records
	GetApplyGenerate(kit *kit.Kit, param *types.GetApplyGenerateReq) (*types.GetApplyGenerateRst, error)
	// GetApplyInit gets resource apply order init records
	GetApplyInit(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyInitRst, error)
	// GetApplyDiskCheck gets resource apply order disk check records
	GetApplyDiskCheck(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyDiskCheckRst, error)
	// GetApplyDeliver gets resource apply order deliver records
	GetApplyDeliver(kit *kit.Kit, param *types.GetApplyDeliverReq) (*types.GetApplyDeliverRst, error)
	// GetApplyDevice get resource apply delivered devices
	GetApplyDevice(kit *kit.Kit, param *types.GetApplyDeviceReq) (*types.GetApplyDeviceRst, error)
	// ExportDeliverDevice export resource apply delivered devices
	ExportDeliverDevice(kit *kit.Kit, param *types.ExportDeliverDeviceReq) (*types.GetApplyDeviceRst, error)
	// GetMatchDevice get resource apply match devices
	GetMatchDevice(kit *kit.Kit, param *types.GetMatchDeviceReq) (*types.GetMatchDeviceRst, error)
	// MatchDevice execute resource apply match devices
	MatchDevice(kit *kit.Kit, param *types.MatchDeviceReq) error
	// MatchPoolDevice execute resource apply match devices from resource pool
	MatchPoolDevice(kit *kit.Kit, param *types.MatchPoolDeviceReq) error
	// PauseApplyOrder pauses resource apply order
	PauseApplyOrder(kit *kit.Kit, param mapstr.MapStr) error
	// ResumeApplyOrder resumes resource apply order
	ResumeApplyOrder(kit *kit.Kit, param mapstr.MapStr) error
	// StartApplyOrder starts resource apply order
	StartApplyOrder(kit *kit.Kit, param *types.StartApplyOrderReq) error
	// TerminateApplyOrder terminates resource apply order
	TerminateApplyOrder(kit *kit.Kit, param *types.TerminateApplyOrderReq) error
	// ModifyApplyOrder modify resource apply order
	ModifyApplyOrder(kit *kit.Kit, param *types.ModifyApplyReq) error
	// RecommendApplyOrder get resource apply order modification recommendation
	RecommendApplyOrder(kit *kit.Kit, param *types.RecommendApplyReq) (*types.RecommendApplyRst, error)
	// GetApplyModify gets resource apply order modify records
	GetApplyModify(kit *kit.Kit, param *types.GetApplyModifyReq) (*types.GetApplyModifyRst, error)
	// DeliverDevice deliver one device to business
	DeliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error
	// SetDeviceDelivered set device info delivered
	SetDeviceDelivered(info *types.DeviceInfo) error
	// GetGenerateRecords check and update cvm device
	GetGenerateRecords(kt *kit.Kit, orderId string) ([]*types.GenerateRecord, error)
	// AddCvmDevices check and update cvm device
	AddCvmDevices(kit *kit.Kit, taskId string, generateId uint64, order *types.ApplyOrder) error
	// UpdateOrderStatus check generate record by order id
	UpdateOrderStatus(suborderID string) error
	// UpdateHostOperator update operator of host
	UpdateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error
	// ProcessInitStep process init step
	ProcessInitStep(device *types.DeviceInfo) error
	// CheckSopsUpdate check if the sops task is completed and update the initialization status
	CheckSopsUpdate(bkBizID int64, info *types.DeviceInfo, jobUrl string, jobIDStr string) error
	// RunDiskCheck run disk check
	RunDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error)
	// DeliverDevices deliver devices to business
	DeliverDevices(order *types.ApplyOrder, observeDevices []*types.DeviceInfo) error
	// FinalApplyStep after deliver device, update generate record status and order status
	FinalApplyStep(genRecord *types.GenerateRecord, order *types.ApplyOrder) error
	// GetMatcher get matcher
	GetMatcher() *matcher.Matcher
	// GetGenerator get generator
	GetGenerator() *generator.Generator

	// CheckRollingServerHost check rolling server host
	CheckRollingServerHost(kt *kit.Kit, param *types.CheckRollingServerHostReq) (*types.CheckRollingServerHostResp,
		error)
}

// scheduler provides resource apply service
type scheduler struct {
	lang         language.CCLanguageIf
	itsm         itsm.Client
	cc           cmdb.Client
	dispatcher   *dispatcher.Dispatcher
	generator    *generator.Generator
	matcher      *matcher.Matcher
	recommend    *recommender.Recommender
	configLogics config.Logics
}

// New creates a scheduler
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client, informerIf informer.Interface,
	clientConf cc.ClientConfig) (*scheduler, error) {

	// new recommend module
	recommend, err := recommender.New(ctx, thirdCli)
	if err != nil {
		return nil, err
	}

	// new matcher
	match, err := matcher.New(ctx, thirdCli, esbCli, clientConf, informerIf)
	if err != nil {
		return nil, err
	}

	// new generator
	generate, err := generator.New(ctx, thirdCli, esbCli, clientConf)
	if err != nil {
		return nil, err
	}

	// new dispatcher
	dispatch, err := dispatcher.New(ctx, informerIf)
	if err != nil {
		return nil, err
	}
	dispatch.SetGenerator(generate)

	scheduler := &scheduler{
		lang:         language.NewFromCtx(language.EmptyLanguageSetting),
		itsm:         thirdCli.ITSM,
		cc:           esbCli.Cmdb(),
		dispatcher:   dispatch,
		generator:    generate,
		matcher:      match,
		recommend:    recommend,
		configLogics: config.New(thirdCli),
	}

	return scheduler, nil
}

// GetDispatcher get dispatcher
func (s *scheduler) GetDispatcher() *dispatcher.Dispatcher {
	return s.dispatcher
}

// GetGenerator get generator
func (s *scheduler) GetGenerator() *generator.Generator { return s.generator }

// UpdateApplyTicket creates or updates resource apply ticket
func (s *scheduler) UpdateApplyTicket(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error) {
	if param.OrderId <= 0 {
		return s.createApplyTicket(kt, param, types.TicketStageUncommit)
	}

	return s.updateApplyTicket(kt, param, types.TicketStageUncommit)
}

func (s *scheduler) createApplyTicket(kt *kit.Kit, param *types.ApplyReq,
	stage types.TicketStage) (*types.CreateApplyOrderResult, error) {

	orderId, err := model.Operation().ApplyOrder().NextSequence(kt.Ctx)
	if err != nil {
		return nil, errf.Newf(common.CCErrObjectDBOpErrno, err.Error())
	}

	now := time.Now()
	ticket := &types.ApplyTicket{
		OrderId:      orderId,
		Stage:        stage,
		BkBizId:      param.BkBizId,
		User:         param.User,
		Follower:     param.Follower,
		EnableNotice: param.EnableNotice,
		RequireType:  param.RequireType,
		ExpectTime:   param.ExpectTime,
		Remark:       param.Remark,
		Suborders:    param.Suborders,
		CreateAt:     now,
		UpdateAt:     now,
	}

	logs.V(9).Infof("ticket data: %+v", ticket)

	if err := model.Operation().ApplyTicket().CreateApplyTicket(kt.Ctx, ticket); err != nil {
		logs.Errorf("failed to create apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.CreateApplyOrderResult{
		OrderId: orderId,
	}

	return rst, nil
}

func (s *scheduler) updateApplyTicket(kt *kit.Kit, param *types.ApplyReq,
	stage types.TicketStage) (*types.CreateApplyOrderResult, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	origin, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if origin.Stage != types.TicketStageUncommit {
		logs.Errorf("failed to update apply ticket, for invalid stage: %s != %s, rid: %s", origin.Stage,
			types.TicketStageUncommit, kt.Rid)
		return nil, fmt.Errorf("invalid ticket stage:%s != %s", origin.Stage, types.TicketStageUncommit)
	}

	update := mapstr.MapStr{
		"order_id":      param.OrderId,
		"stage":         stage,
		"bk_biz_id":     param.BkBizId,
		"bk_username":   param.User,
		"follower":      param.Follower,
		"enable_notice": param.EnableNotice,
		"require_type":  param.RequireType,
		"expect_time":   param.ExpectTime,
		"remark":        param.Remark,
		"suborders":     param.Suborders,
		"update_at":     time.Now(),
	}

	if err := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, update); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.CreateApplyOrderResult{
		OrderId: param.OrderId,
	}

	return rst, nil
}

// GetApplyTicket gets resource apply ticket
func (s *scheduler) GetApplyTicket(kit *kit.Kit, param *types.GetApplyTicketReq) (
	*types.GetApplyTicketRst, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}
	// 业务下查询时，只查询传入业务对应的单据
	if param.BkBizID > 0 && param.BkBizID != constant.UnassignedBiz {
		filter["bk_biz_id"] = param.BkBizID
	}

	inst, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetApplyTicketRst{
		ApplyTicket: inst,
	}

	return rst, nil
}

// GetApplyAudit gets resource apply ticket audit info
func (s *scheduler) GetApplyAudit(kit *kit.Kit, param *types.GetApplyAuditReq) (
	*types.GetApplyAuditRst, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}
	// 业务下查询时，只查询传入业务对应的单据
	if param.BkBizID > 0 && param.BkBizID != constant.UnassignedBiz {
		filter["bk_biz_id"] = param.BkBizID
	}

	inst, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if inst.ItsmTicketId == "" {
		logs.Errorf("failed to get apply ticket audit info, for itsm ticket sn is empty, rid: %s", kit.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, for itsm ticket sn is empty")
	}

	statusResp, err := s.itsm.GetTicketStatus(kit, inst.ItsmTicketId)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if statusResp.Code != 0 {
		logs.Errorf("failed to get apply ticket audit info, code: %d, msg: %s, rid: %s", statusResp.Code,
			statusResp.ErrMsg, kit.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, code: %d, msg: %s", statusResp.Code,
			statusResp.ErrMsg)
	}

	status := statusResp.Data.CurrentStatus
	link := statusResp.Data.TicketUrl

	logResp, err := s.itsm.GetTicketLog(kit, inst.ItsmTicketId)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if logResp.Code != 0 {
		logs.Errorf("failed to get apply ticket audit info, code: %d, msg: %s, rid: %s", logResp.Code, logResp.ErrMsg,
			kit.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, code: %d, msg: %s", logResp.Code, logResp.ErrMsg)
	}

	rst := &types.GetApplyAuditRst{
		ApplyAudit: &types.ApplyAudit{
			OrderId:        param.OrderId,
			ItsmTicketId:   inst.ItsmTicketId,
			ItsmTicketLink: link,
			Status:         status,
			CurrentSteps:   make([]*types.ApplyAuditStep, 0),
			Logs:           make([]*types.ApplyAuditLog, 0),
		},
	}

	for _, step := range statusResp.Data.CurrentSteps {
		rst.CurrentSteps = append(rst.CurrentSteps, &types.ApplyAuditStep{
			Name:       step.Name,
			Processors: step.Processors,
			StateId:    step.StateId,
		})
	}
	for _, log := range logResp.Data.Logs {
		rst.Logs = append(rst.Logs, &types.ApplyAuditLog{
			Operator:  log.Operator,
			OperateAt: log.OperateAt,
			Message:   log.Message,
			Source:    log.Source,
		})
	}

	return rst, nil
}

// AuditTicket audit resource apply ticket
func (s *scheduler) AuditTicket(kit *kit.Kit, param *types.ApplyAuditReq) error {
	req := &itsm.OperateNodeReq{
		Sn:         param.ItsmTicketId,
		StateId:    param.StateId,
		Operator:   param.Operator,
		ActionType: itsm.ActionTypeTransition,
		Fields:     make([]*itsm.TicketField, 0),
	}

	keys, ok := itsm.MapStateKey[param.StateId]
	if !ok {
		logs.Errorf("failed to audit apply ticket, invalid state id: %d, rid: %s", param.StateId, kit.Rid)
		return fmt.Errorf("failed to audit apply ticket, invalid state id: %d", param.StateId)
	}

	req.Fields = append(req.Fields, &itsm.TicketField{
		Key:   keys[0],
		Value: strconv.FormatBool(param.Approval),
	})
	req.Fields = append(req.Fields, &itsm.TicketField{
		Key:   keys[1],
		Value: param.Remark,
	})

	resp, err := s.itsm.OperateNode(kit, req)
	if err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	if resp.Code != 0 {
		logs.Errorf("failed to audit apply ticket, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kit.Rid)
		return fmt.Errorf("failed to audit apply ticket, order id: %d, sn: %s, code: %d, msg: %s", param.OrderId,
			param.ItsmTicketId, resp.Code, resp.ErrMsg)
	}

	return nil
}

// AutoAuditTicket system automatic audit resource apply ticket callback
func (s *scheduler) AutoAuditTicket(kit *kit.Kit, param *types.ApplyAutoAuditReq) (*types.ApplyAutoAuditRst, error) {
	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	order, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to auto audit order %d, err: %v, rid: %s", param.OrderId, err, kit.Rid)
		return nil, fmt.Errorf("failed to auto audit order %d, err: %v", param.OrderId, err)
	}

	if order.Stage != types.TicketStageAudit {
		logs.Errorf("failed to auto audit order %d, for invalid stage %s != AUDIT, rid: %s", param.OrderId, order.Stage,
			kit.Rid)
		return nil, fmt.Errorf("order %d is not at AUDIT stage", param.OrderId)
	}

	rst := &types.ApplyAutoAuditRst{
		Operator: "icr",
		Approval: 1,
		Remark:   "approved",
	}

	total := uint(0)
	for _, suborder := range order.Suborders {
		total += suborder.Replicas
		// 所有物理机资源申请（除故障替换外），都需要人工审核
		if order.RequireType != 4 {
			if suborder.ResourceType == types.ResourceTypePm {
				logs.Errorf("failed to auto audit order %d, for resource type include %s, rid: %s", param.OrderId,
					types.ResourceTypePm, kit.Rid)
				rst.Approval = 0
				rst.Remark = fmt.Sprintf("order %d resource type include %s", param.OrderId, types.ResourceTypePm)
				return rst, nil
			}
		}
	}

	auditThreshold := uint(50)
	if total > auditThreshold {
		logs.Errorf("failed to auto audit order %d, for apply number exceeds %d, rid: %s", param.OrderId,
			auditThreshold, kit.Rid)
		rst.Approval = 0
		rst.Remark = fmt.Sprintf("order %d apply number %d exceed auto audit threshold %d", param.OrderId, total,
			auditThreshold)
	}

	return rst, nil
}

// ApproveTicket approve or reject resource apply ticket
func (s *scheduler) ApproveTicket(kt *kit.Kit, param *types.ApproveApplyReq) error {
	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	stage := types.TicketStageTerminate
	if param.Approval {
		stage = types.TicketStageRunning
	}
	update := mapstr.MapStr{
		"stage":     stage,
		"update_at": time.Now(),
	}

	err := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		if err := model.Operation().ApplyTicket().UpdateApplyTicket(sc, &filter, update); err != nil {
			logs.Errorf("failed to update apply ticket, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
			return err
		}

		sessionKit := &kit.Kit{Ctx: sc, Rid: kt.Rid}
		if param.Approval {
			if err := s.createSubOrders(sessionKit, param.OrderId); err != nil {
				logs.Errorf("failed to create subOrders, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
				return err
			}
		}

		return nil
	})

	if err != nil {
		logs.Errorf("failed to approve apply ticket %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
		return err
	}
	return nil
}

func (s *scheduler) createSubOrders(kt *kit.Kit, orderId uint64) error {
	filter := mapstr.MapStr{
		"order_id": orderId,
	}

	ticket, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket by filter: %+v, err: %v, rid: %s", filter, err, kt.Rid)
		return err
	}

	now := time.Now()
	for index, suborder := range ticket.Suborders {
		// TODO: delete debug log
		logs.V(5).Infof("suborder data: %+v", suborder)

		subOrder := &types.ApplyOrder{
			OrderId:           orderId,
			SubOrderId:        fmt.Sprintf("%d-%d", orderId, index+1),
			BkBizId:           ticket.BkBizId,
			User:              ticket.User,
			Follower:          ticket.Follower,
			Auditor:           "",
			RequireType:       ticket.RequireType,
			ExpectTime:        ticket.ExpectTime,
			ResourceType:      suborder.ResourceType,
			Spec:              suborder.Spec,
			AntiAffinityLevel: suborder.AntiAffinityLevel,
			EnableDiskCheck:   suborder.EnableDiskCheck,
			Description:       ticket.Remark,
			Remark:            suborder.Remark,
			Stage:             types.TicketStageRunning,
			Status:            types.ApplyStatusWaitForMatch,
			Total:             suborder.Replicas,
			PendingNum:        suborder.Replicas,
			SuccessNum:        0,
			RetryTime:         0,
			ModifyTime:        0,
			CreateAt:          now,
			UpdateAt:          now,
		}
		logs.V(4).Infof("suborder data: %+v", subOrder)

		if err := model.Operation().ApplyOrder().CreateApplyOrder(kt.Ctx, subOrder); err != nil {
			logs.Errorf("failed to create apply order, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		// init all step record
		if err := s.initAllSteps(kt, subOrder.SubOrderId, subOrder.Total, subOrder.EnableDiskCheck); err != nil {
			logs.Errorf("failed to init apply step record, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// initAllSteps init apply order all steps
func (s *scheduler) initAllSteps(kt *kit.Kit, suborderId string, total uint,
	enableDiskCheck bool) error {
	// init commit step
	stepID := 1
	if err := record.CreateCommitStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create commit step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	// init generate step
	stepID++
	if err := record.CreateGenerateStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create generate step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	// init init step
	stepID++
	if err := record.CreateInitStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create init step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	if enableDiskCheck {
		// init disk check step
		stepID++
		if err := record.CreateDiskCheckStep(kt.Ctx, suborderId, total, stepID); err != nil {
			logs.Errorf("order %s failed to create disk check step, err: %v, rid: %s", suborderId, err, kt.Rid)
			return err
		}
	}

	// init deliver step
	stepID++
	if err := record.CreateDeliverStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create deliver step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	return nil
}

// CreateApplyOrder creates resource apply order
func (s *scheduler) CreateApplyOrder(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error) {
	rst := new(types.CreateApplyOrderResult)
	var err error = nil

	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		sessionKit := &kit.Kit{Ctx: sc, Rid: kt.Rid}
		if param.OrderId <= 0 {
			rst, err = s.createApplyTicket(sessionKit, param, types.TicketStageAudit)
		} else {
			rst, err = s.updateApplyTicket(sessionKit, param, types.TicketStageAudit)
		}
		if err != nil {
			logs.Errorf("failed to create apply order, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
			return err
		}

		resp, err := s.itsm.CreateApplyTicket(sessionKit, param.User, rst.OrderId, param.BkBizId)
		if err != nil {
			logs.Errorf("failed to create apply order, for create itsm ticket err: %v, rid: %s, orderId: %d, BkBIzId: %d",
				err, kt.Rid, rst.OrderId, param.BkBizId)
			return err
		}

		if resp.Code != 0 {
			logs.Errorf("failed to create apply order, for create itsm ticket err, code: %d, msg: %s, rid: %s, orderId: %d,"+
				" BkBIzId: %d", resp.Code, resp.ErrMsg, kt.Rid, rst.OrderId, param.BkBizId)
			return err
		}

		if err := s.setTicketId(sessionKit, rst.OrderId, resp.Data.Sn); err != nil {
			logs.Errorf("failed to create apply order, for set ticket id err: %v, rid: %s, orderId: %d, sn: %s",
				err, kt.Rid, rst.OrderId, resp.Data.Sn)
			return err
		}
		return nil
	})

	return rst, txnErr
}

func (s *scheduler) setTicketId(kt *kit.Kit, orderId uint64, itsmTicketId string) error {
	filter := mapstr.MapStr{
		"order_id": orderId,
	}

	doc := mapstr.MapStr{
		"itsm_ticket_id": itsmTicketId,
		"update_at":      time.Now(),
	}

	if err := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, doc); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetApplyOrder gets resource apply order info
func (s *scheduler) GetApplyOrder(kit *kit.Kit, param *types.GetApplyParam) (*types.GetApplyOrderRst, error) {
	orderFilter := param.GetFilter(false)
	ticketFilter := param.GetFilter(true)

	cntTicket, err := model.Operation().ApplyTicket().CountApplyTicket(kit.Ctx, ticketFilter)
	if err != nil {
		return nil, err
	}

	cntOrder, err := model.Operation().ApplyOrder().CountApplyOrder(kit.Ctx, orderFilter)
	if err != nil {
		return nil, err
	}

	page := metadata.BasePage{
		Sort:  "-create_at",
		Limit: common.BKNoLimit,
		Start: 0,
	}

	tickets, err := model.Operation().ApplyTicket().FindManyApplyTicket(kit.Ctx, page, ticketFilter)
	if err != nil {
		logs.Errorf("get apply ticket failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, orderFilter)
	if err != nil {
		logs.Errorf("get apply order failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	cnt := cntTicket + cntOrder
	mergedOrders := s.mergeApplyTicketOrder(tickets, orders)

	begin := int(math.Max(0, float64(param.Page.Start)))
	end := int(cnt)
	if param.Page.Limit > 0 {
		end = int(math.Min(float64(begin+param.Page.Limit), float64(cnt)))
	}

	rst := &types.GetApplyOrderRst{
		Count: int64(cnt),
		Info:  mergedOrders[begin:end],
	}

	return rst, nil
}

func (s *scheduler) mergeApplyTicketOrder(tickets []*types.ApplyTicket,
	orders []*types.ApplyOrder) []*types.UnifyOrder {

	mergeOrders := types.UnifyOrderList{}

	unifyTickets := s.ticketToUnifyOrder(tickets)
	unifyOrders := s.orderToUnifyOrder(orders)

	mergeOrders = append(mergeOrders, unifyTickets...)
	mergeOrders = append(mergeOrders, unifyOrders...)

	sort.Sort(sort.Reverse(mergeOrders))
	return mergeOrders
}

func (s *scheduler) ticketToUnifyOrder(tickets []*types.ApplyTicket) []*types.UnifyOrder {
	unifyOrders := make([]*types.UnifyOrder, 0)

	for _, ticket := range tickets {
		total := uint(0)
		for _, suborder := range ticket.Suborders {
			total += suborder.Replicas
		}
		order := &types.UnifyOrder{
			OrderId:     ticket.OrderId,
			BkBizId:     ticket.BkBizId,
			User:        ticket.User,
			RequireType: ticket.RequireType,
			ExpectTime:  ticket.ExpectTime,
			Description: ticket.Remark,
			Stage:       ticket.Stage,
			Total:       total,
			CreateAt:    ticket.CreateAt,
			UpdateAt:    ticket.UpdateAt,
		}

		unifyOrders = append(unifyOrders, order)
	}

	return unifyOrders
}

func (s *scheduler) orderToUnifyOrder(orders []*types.ApplyOrder) []*types.UnifyOrder {
	unifyOrders := make([]*types.UnifyOrder, 0)

	for _, order := range orders {
		order := &types.UnifyOrder{
			OrderId:           order.OrderId,
			SubOrderId:        order.SubOrderId,
			BkBizId:           order.BkBizId,
			User:              order.User,
			RequireType:       order.RequireType,
			ResourceType:      order.ResourceType,
			ExpectTime:        order.ExpectTime,
			Description:       order.Description,
			Remark:            order.Remark,
			Spec:              order.Spec,
			AntiAffinityLevel: order.AntiAffinityLevel,
			EnableDiskCheck:   order.EnableDiskCheck,
			Stage:             order.Stage,
			Status:            order.Status,
			Total:             order.Total,
			SuccessNum:        order.SuccessNum,
			PendingNum:        order.PendingNum,
			ModifyTime:        order.ModifyTime,
			CreateAt:          order.CreateAt,
			UpdateAt:          order.UpdateAt,
		}
		unifyOrders = append(unifyOrders, order)
	}

	return unifyOrders
}

// GetApplyDetail gets resource apply order detail info
func (s *scheduler) GetApplyDetail(kit *kit.Kit, param *types.GetApplyDetailReq) (*types.GetApplyDetailRst, error) {
	filter := &mapstr.MapStr{
		"suborder_id": param.SuborderId,
	}

	insts, err := model.Operation().ApplyStep().FindManyApplyStep(kit.Ctx, filter)
	if err != nil {
		logs.Errorf("get apply order detail info failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetApplyDetailRst{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyGenerate gets resource apply order generate records
func (s *scheduler) GetApplyGenerate(kit *kit.Kit, param *types.GetApplyGenerateReq) (*types.GetApplyGenerateRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order generate record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().GenerateRecord().CountGenerateRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().GenerateRecord().FindManyGenerateRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyGenerateRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyInit gets resource apply order init records
func (s *scheduler) GetApplyInit(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyInitRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order init record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().InitRecord().CountInitRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().InitRecord().FindManyInitRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyInitRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDiskCheck gets resource apply order disk check records
func (s *scheduler) GetApplyDiskCheck(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyDiskCheckRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order disk check record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().DiskCheckRecord().CountDiskCheckRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().DiskCheckRecord().FindManyDiskCheckRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDiskCheckRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDeliver gets resource apply order deliver records
func (s *scheduler) GetApplyDeliver(kit *kit.Kit, param *types.GetApplyDeliverReq) (*types.GetApplyDeliverRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order deliver record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().DeliverRecord().CountDeliverRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().DeliverRecord().FindManyDeliverRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDeliverRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDevice get resource apply delivered devices
func (s *scheduler) GetApplyDevice(kit *kit.Kit, param *types.GetApplyDeviceReq) (*types.GetApplyDeviceRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get apply order device info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	// return delivered device only
	filter["is_delivered"] = true

	count, err := model.Operation().DeviceInfo().CountDeviceInfo(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().DeviceInfo().FindManyDeviceInfo(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDeviceRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// ExportDeliverDevice export resource apply delivered devices
func (s *scheduler) ExportDeliverDevice(kit *kit.Kit, param *types.ExportDeliverDeviceReq) (*types.GetApplyDeviceRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	// return delivered device only
	filter["is_delivered"] = true

	count, err := model.Operation().DeviceInfo().CountDeviceInfo(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}

	insts, err := model.Operation().DeviceInfo().FindManyDeviceInfo(kit.Ctx, page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDeviceRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetMatchDevice get resource apply match devices
func (s *scheduler) GetMatchDevice(kit *kit.Kit, param *types.GetMatchDeviceReq) (*types.GetMatchDeviceRst, error) {
	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules:     make([]querybuilder.Rule, 0),
	}
	if len(param.Ips) != 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_host_innerip",
			Operator: querybuilder.OperatorIn,
			Value:    param.Ips,
		})
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_cloud_id",
			Operator: querybuilder.OperatorEqual,
			Value:    0,
		})
	}
	if param.Spec != nil {
		if param.ResourceType != types.ResourceTypeCvm {
			if len(param.Spec.Region) != 0 {
				rule.Rules = append(rule.Rules, querybuilder.AtomRule{
					Field:    "bk_zone_name",
					Operator: querybuilder.OperatorIn,
					Value:    param.Spec.Region,
				})
			}
			if len(param.Spec.Zone) != 0 {
				rule.Rules = append(rule.Rules, querybuilder.AtomRule{
					Field:    "sub_zone",
					Operator: querybuilder.OperatorIn,
					Value:    param.Spec.Zone,
				})
			}
		} else {
			if len(param.Spec.Zone) != 0 {
				filter := mapstr.MapStr{}
				filter["zone"] = mapstr.MapStr{
					common.BKDBIN: param.Spec.Zone,
				}
				if len(param.Spec.Region) != 0 {
					filter["region"] = mapstr.MapStr{
						common.BKDBIN: param.Spec.Region,
					}
				}
				zones, err := model.Operation().Zone().FindManyZone(context.Background(), &filter)
				if err != nil {
					return nil, err
				}
				cmdbZoneNames := make([]string, 0)
				for _, zone := range zones {
					cmdbZoneNames = append(cmdbZoneNames, zone.CmdbZoneName)
				}
				cmdbZoneNames = util.StrArrayUnique(cmdbZoneNames)
				rule.Rules = append(rule.Rules, querybuilder.AtomRule{
					Field:    "sub_zone",
					Operator: querybuilder.OperatorIn,
					Value:    cmdbZoneNames,
				})
			} else if len(param.Spec.Region) != 0 {
				filter := mapstr.MapStr{}
				filter["region"] = mapstr.MapStr{
					common.BKDBIN: param.Spec.Region,
				}
				zones, err := model.Operation().Zone().FindManyZone(context.Background(), &filter)
				if err != nil {
					return nil, err
				}
				cmdbRegionNames := make([]string, 0)
				for _, zone := range zones {
					cmdbRegionNames = append(cmdbRegionNames, zone.CmdbRegionName)
				}
				cmdbRegionNames = util.StrArrayUnique(cmdbRegionNames)
				rule.Rules = append(rule.Rules, querybuilder.AtomRule{
					Field:    "bk_zone_name",
					Operator: querybuilder.OperatorIn,
					Value:    cmdbRegionNames,
				})
			}
		}
		if len(param.Spec.DeviceType) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "svr_device_class",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.DeviceType,
			})
		}
		if len(param.Spec.OsType) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "bk_os_name",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.OsType,
			})
		}
		if len(param.Spec.RaidType) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "raid_name",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.RaidType,
			})
		}
		if len(param.Spec.Isp) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "bk_ip_oper_name",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.Isp,
			})
		}
	}
	req := &cmdb.ListBizHostParams{
		BizID:       931,
		BkModuleIDs: []int64{239149},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			// 外网运营商
			"bk_ip_oper_name",
			// 机型
			"svr_device_class",
			"bk_os_name",
			// 地域
			"bk_zone_name",
			// 可用区
			"sub_zone",
			"module_name",
			// 机架号，string类型
			"rack_id",
			"idc_unit_name",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: common.BKMaxInstanceLimit,
		},
	}
	if len(rule.Rules) > 0 {
		req.HostPropertyFilter = &cmdb.QueryFilter{
			Rule: rule,
		}
	}

	resp, err := s.cc.ListBizHost(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// TODO: filter and sort devices
	rst := &types.GetMatchDeviceRst{
		Count: 0,
		Info:  make([]*types.MatchDevice, 0),
	}
	tagNum := int64(0)
	for _, host := range resp.Info {
		rackId, err := strconv.Atoi(host.RackId)
		if err != nil {
			logs.Warnf("failed to convert host %d rack_id %s to int", host.BkHostID, host.RackId)
			rackId = 0
		}
		tag := false
		if tagNum < param.PendingNum {
			tag = true
			tagNum++
		}
		device := &types.MatchDevice{
			BkHostId:     host.BkHostID,
			AssetId:      host.BkAssetID,
			Ip:           host.GetUniqIp(),
			OuterIp:      host.BkHostOuterIP,
			Isp:          host.BkIpOerName,
			DeviceType:   host.SvrDeviceClass,
			OsType:       host.BkOSName,
			Region:       host.BkZoneName,
			Zone:         host.SubZone,
			Module:       host.ModuleName,
			Equipment:    int64(rackId),
			IdcUnit:      host.IdcUnitName,
			IdcLogicArea: host.LogicDomain,
			RaidType:     host.RaidName,
			InputTime:    host.SvrInputTime,
			MatchScore:   1.0,
			MatchTag:     tag,
		}

		rst.Info = append(rst.Info, device)
	}
	rst.Count = int64(len(rst.Info))

	return rst, nil
}

// MatchDevice execute resource apply match devices
func (s *scheduler) MatchDevice(kit *kit.Kit, param *types.MatchDeviceReq) error {
	if err := s.generator.MatchCVM(param); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	return nil
}

// MatchPoolDevice execute resource apply match devices from resource pool
func (s *scheduler) MatchPoolDevice(kit *kit.Kit, param *types.MatchPoolDeviceReq) error {
	go s.generator.MatchPoolDevice(param)

	return nil
}

// PauseApplyOrder pauses resource apply order
func (s *scheduler) PauseApplyOrder(kit *kit.Kit, param mapstr.MapStr) error {
	// TODO
	return nil
}

// ResumeApplyOrder resumes resource apply order
func (s *scheduler) ResumeApplyOrder(kit *kit.Kit, param mapstr.MapStr) error {
	// TODO
	return nil
}

// StartApplyOrder starts resource apply order
func (s *scheduler) StartApplyOrder(kit *kit.Kit, param *types.StartApplyOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("found no apply order to start, orderNum: %d, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no apply order to start")
	}

	// check status
	for _, order := range insts {
		// cannot start apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot terminate order %s, for its stage %s != %s ", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}
	}

	// set order status wait
	if err := s.startOrder(insts); err != nil {
		logs.Errorf("failed to start apply order, err: %v", err)
		return fmt.Errorf("failed to start apply order, err: %v", err)
	}

	return nil
}

func (s *scheduler) startOrder(orders []*types.ApplyOrder) error {
	now := time.Now()
	for _, order := range orders {
		// cannot start apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot start order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot start order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SubOrderId,
		}

		update := &mapstr.MapStr{
			"stage":      types.TicketStageRunning,
			"status":     types.ApplyStatusWaitForMatch,
			"retry_time": 0,
			"update_at":  now,
		}

		if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s running, err: %v", order.SubOrderId, err)
			return fmt.Errorf("failed to set order %s running, err: %v", order.SubOrderId, err)
		}
	}

	return nil
}

// TerminateApplyOrder terminates resource apply order
func (s *scheduler) TerminateApplyOrder(kit *kit.Kit, param *types.TerminateApplyOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("found no apply order to terminate, rid: %s", kit.Rid)
		return fmt.Errorf("found no apply order to terminate")
	}

	// check status
	for _, order := range insts {
		// cannot terminate apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}
	}

	// set order status terminate
	if err := s.terminateOrder(insts); err != nil {
		logs.Errorf("failed to terminate apply order, err: %v", err)
		return fmt.Errorf("failed to terminate apply order, err: %v", err)
	}

	return nil
}

func (s *scheduler) terminateOrder(orders []*types.ApplyOrder) error {
	now := time.Now()
	for _, order := range orders {
		// cannot terminate apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Status,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SubOrderId,
		}

		update := &mapstr.MapStr{
			"stage":     types.TicketStageTerminate,
			"update_at": now,
		}

		if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
			return fmt.Errorf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
		}
	}

	return nil
}

// ModifyApplyOrder modify resource apply order
func (s *scheduler) ModifyApplyOrder(kt *kit.Kit, param *types.ModifyApplyReq) error {
	filter := &mapstr.MapStr{
		"suborder_id": mapstr.MapStr{
			common.BKDBEQ: param.SuborderID,
		},
	}

	order, err := model.Operation().ApplyOrder().GetApplyOrder(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// cannot modify apply order if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
		return fmt.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// validate modification
	if err := s.validateModification(kt, order, param); err != nil {
		logs.Errorf("modification is invalid, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// modify apply order
	if err := s.modifyOrder(order, param); err != nil {
		logs.Errorf("failed to modify apply order, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("failed to modify apply order, err: %v", err)
	}

	// create apply order modify record
	if err := s.createModifyRecord(kt, order, param); err != nil {
		logs.Errorf("failed to create apply order modify record, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) validateModification(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	// validate replicas
	if param.Replicas > order.Total {
		logs.Errorf("modified replicas cannot exceeds origin value %d", order.Total)
		return fmt.Errorf("modified replicas cannot exceeds origin value %d", order.Total)
	}
	if param.Replicas <= 0 {
		logs.Errorf("modified replicas should be positive integer")
		return errors.New("modified replicas should be positive integer")
	}
	if param.Replicas < order.SuccessNum {
		logs.Errorf("modified replicas cannot be less than successfully delivered amount %d", order.SuccessNum)
		return fmt.Errorf("modified replicas cannot be less than successfully delivered amount %d", order.SuccessNum)
	}

	// validate device type
	if err := s.validateModifyDeviceType(kt, order, param); err != nil {
		logs.Errorf("failed to validate modify device type, err: %v", err)
		return err
	}

	// validate zone
	if err := s.validateModifyZone(order, param); err != nil {
		logs.Errorf("failed to validate modify zone, err: %v", err)
		return err
	}

	return nil
}

func (s *scheduler) validateModifyDeviceType(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	originDeviceGroup, err := s.getDeviceGroup(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region,
		order.Spec.Zone)
	if err != nil {
		logs.Errorf("failed to get device group, err: %v", err)
		return err
	}

	modifiedDeviceGroup, err := s.getDeviceGroup(kt, order.RequireType, param.Spec.DeviceType, param.Spec.Region,
		param.Spec.Zone)
	if err != nil {
		logs.Errorf("failed to get device group, err: %v", err)
		return err
	}

	// modification is valid if found no device config
	if originDeviceGroup == "" {
		return nil
	}

	if originDeviceGroup != modifiedDeviceGroup {
		logs.Errorf("modify device type is invalid, for its device group changed")
		return errors.New("modify device type is invalid, for its device group changed")
	}

	return nil
}

func (s *scheduler) getDeviceGroup(kt *kit.Kit, requireType int64, deviceType, region, zone string) (string, error) {
	rules := []querybuilder.Rule{
		querybuilder.AtomRule{
			Field:    "device_type",
			Operator: querybuilder.OperatorEqual,
			Value:    deviceType,
		},
		querybuilder.AtomRule{
			Field:    "require_type",
			Operator: querybuilder.OperatorEqual,
			Value:    requireType,
		},
		querybuilder.AtomRule{
			Field:    "region",
			Operator: querybuilder.OperatorEqual,
			Value:    region,
		},
	}

	if zone != "" && zone != cvmapi.CvmSeparateCampus {
		rules = append(rules, querybuilder.AtomRule{
			Field:    "zone",
			Operator: querybuilder.OperatorEqual,
			Value:    zone,
		})
	}

	req := &configtypes.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules:     rules,
			},
		},
		Page: metadata.BasePage{
			Limit: 1,
			Start: 0,
		},
	}

	deviceInfo, err := s.configLogics.Device().GetDevice(kt, req)
	if err != nil {
		logs.Errorf("failed to get device info, err: %v", err)
		return "", err
	}

	num := len(deviceInfo.Info)
	if num == 0 {
		// return empty when found no device config
		return "", nil
	} else if num != 1 {
		logs.Errorf("failed to get device info, for len %d != 1", num)
		return "", fmt.Errorf("failed to get device info, for len %d != 1", num)
	}

	deviceGroup, ok := deviceInfo.Info[0].Label["device_group"]
	if !ok {
		return "", errors.New("get invalid empty device group")
	}

	ret, ok := deviceGroup.(string)
	if !ok {
		return "", errors.New("get invalid non-string device group")
	}

	return ret, nil
}

func (s *scheduler) validateModifyZone(order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	if param.Spec.Region != order.Spec.Region {
		logs.Errorf("region cannot be modified")
		return errors.New("region cannot be modified")
	}

	return nil
}

func (s *scheduler) modifyOrder(order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	now := time.Now()
	// cannot modify apply order if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend)
		return fmt.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend)
	}

	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}

	update := &mapstr.MapStr{
		"spec.region":       param.Spec.Region,
		"spec.zone":         param.Spec.Zone,
		"spec.device_type":  param.Spec.DeviceType,
		"spec.image_id":     param.Spec.ImageId,
		"spec.disk_size":    param.Spec.DiskSize,
		"spec.disk_type":    param.Spec.DiskType,
		"spec.network_type": param.Spec.NetworkType,
		"spec.vpc":          param.Spec.Vpc,
		"spec.subnet":       param.Spec.Subnet,
		"stage":             types.TicketStageRunning,
		"status":            types.ApplyStatusWaitForMatch,
		"total_num":         param.Replicas,
		"pending_num":       param.Replicas - order.SuccessNum,
		"retry_time":        0,
		"modify_time":       order.ModifyTime + 1,
		"update_at":         now,
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
		logs.Warnf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
		return fmt.Errorf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
	}

	return nil
}

func (s *scheduler) createModifyRecord(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	id, err := dao.Set().ModifyRecord().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to get modify record next sequence id, order id: %s, err: %v", err)
		return errf.Newf(common.CCErrObjectDBOpErrno, err.Error())
	}

	record := &table.ModifyRecord{
		ID:         id,
		SuborderID: order.SubOrderId,
		User:       kt.User,
		Details: &table.ModifyDetail{
			PreData: &table.ModifyData{
				Replicas:    order.Total,
				Region:      order.Spec.Region,
				Zone:        order.Spec.Zone,
				DeviceType:  order.Spec.DeviceType,
				ImageId:     order.Spec.ImageId,
				DiskSize:    order.Spec.DiskSize,
				DiskType:    order.Spec.DiskType,
				NetworkType: order.Spec.NetworkType,
				Vpc:         order.Spec.Vpc,
				Subnet:      order.Spec.Subnet,
			},
			CurData: &table.ModifyData{
				Replicas:    param.Replicas,
				Region:      param.Spec.Region,
				Zone:        param.Spec.Zone,
				DeviceType:  param.Spec.DeviceType,
				ImageId:     param.Spec.ImageId,
				DiskSize:    param.Spec.DiskSize,
				DiskType:    param.Spec.DiskType,
				NetworkType: param.Spec.NetworkType,
				Vpc:         param.Spec.Vpc,
				Subnet:      param.Spec.Subnet,
			},
		},
		CreateAt: time.Now(),
	}

	if err := dao.Set().ModifyRecord().CreateModifyRecord(kt.Ctx, record); err != nil {
		logs.Errorf("failed to create modify record, err: %v", err)
		return err
	}

	return nil
}

// RecommendApplyOrder get resource apply order modification recommendation
func (s *scheduler) RecommendApplyOrder(kt *kit.Kit, param *types.RecommendApplyReq) (*types.RecommendApplyRst,
	error) {

	rst, err := s.recommend.GetApplyRecommendation(param.SuborderID)
	if err != nil {
		logs.Errorf("failed to get apply recommendation, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("failed to get apply recommendation, err: %v", err)
	}

	return rst, nil
}

// GetApplyModify gets resource apply order modify records
func (s *scheduler) GetApplyModify(kt *kit.Kit, param *types.GetApplyModifyReq) (*types.GetApplyModifyRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get apply order modify record, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetApplyModifyRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().ModifyRecord().CountModifyRecord(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get apply order modify record count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.ModifyRecord, 0)
		return rst, nil
	}

	insts, err := dao.Set().ModifyRecord().FindManyModifyRecord(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// ProcessInitStep processes orders with init step
func (s *scheduler) ProcessInitStep(device *types.DeviceInfo) error {
	return s.matcher.ProcessInitStep(device)
}

// CheckSopsUpdate checks sops task and update status, return err if sops task failed or update failed
func (s *scheduler) CheckSopsUpdate(bkBizID int64, info *types.DeviceInfo, jobUrl string, jobIDStr string) error {
	return s.matcher.CheckSopsUpdate(bkBizID, info, jobUrl, jobIDStr)
}

// RunDiskCheck runs disk check
func (s *scheduler) RunDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
	return s.matcher.RunDiskCheck(order, devices)
}

// DeliverDevices delivers devices to order biz
func (s *scheduler) DeliverDevices(order *types.ApplyOrder, observeDevices []*types.DeviceInfo) error {
	return s.matcher.DeliverDevices(order, observeDevices)
}

// FinalApplyStep checks whether the record is updated
func (s *scheduler) FinalApplyStep(genRecord *types.GenerateRecord, order *types.ApplyOrder) error {
	return s.matcher.FinalApplyStep(genRecord, order)
}

// GetGenerateRecords get generate record by order id
func (s *scheduler) GetGenerateRecords(kt *kit.Kit, subOrderId string) ([]*types.GenerateRecord, error) {
	recordInfo, err := s.matcher.GetOrderGenRecords(subOrderId)
	if err != nil {
		logs.Errorf("failed to get generate generateRecord by subOrderId, subOrderId: %s, err: %v, rid: %s", subOrderId,
			err, kt.Rid)
		return nil, err
	}
	return recordInfo, nil
}

// AddCvmDevices check and add cvm device
func (s *scheduler) AddCvmDevices(kt *kit.Kit, taskId string, generateId uint64,
	order *types.ApplyOrder) error {

	_, err := s.generator.AddCvmDevices(kt, taskId, generateId, order)
	if err != nil {
		logs.Errorf("failed to check and update cvm device, orderId: %s, err: %v, rid: %s", err, order.SubOrderId,
			kt.Rid)
		return err
	}
	return nil
}

// UpdateOrderStatus check generate record by order id
func (s *scheduler) UpdateOrderStatus(suborderID string) error {
	return s.generator.UpdateOrderStatus(suborderID)
}

// GetMatcher get matcher
func (s *scheduler) GetMatcher() *matcher.Matcher {
	return s.matcher
}

// DeliverDevice deliver device
func (s *scheduler) DeliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error {
	return s.matcher.DeliverDevice(info, order)
}

// UpdateHostOperator update host operator
func (s *scheduler) UpdateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error {
	return s.matcher.UpdateHostOperator(info, hostId, operator)
}

// SetDeviceDelivered set device delivered
func (s *scheduler) SetDeviceDelivered(info *types.DeviceInfo) error {
	return s.matcher.SetDeviceDelivered(info)
}

// CheckRollingServerHost check rolling server host
func (s *scheduler) CheckRollingServerHost(kt *kit.Kit, param *types.CheckRollingServerHostReq) (
	*types.CheckRollingServerHostResp, error) {

	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionOr,
		Rules: []querybuilder.Rule{
			querybuilder.AtomRule{
				Field:    "bk_asset_id",
				Operator: querybuilder.OperatorEqual,
				Value:    param.AssetID,
			},
		},
	}
	fields := []string{"bk_svr_device_cls_name", "instance_charge_type", "billing_start_time", "billing_expire_time",
		"bk_cloud_inst_id"}
	page := cmdb.BasePage{Start: 0, Limit: 1}

	var info []*cmdb.Host
	if param.BizID != 0 {
		req := &cmdb.ListBizHostParams{
			BizID:              param.BizID,
			HostPropertyFilter: &cmdb.QueryFilter{Rule: rule},
			Fields:             fields,
			Page:               page,
		}
		resp, err := s.cc.ListBizHost(kt, req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		hosts := make([]*cmdb.Host, 0)
		for _, host := range resp.Info {
			hosts = append(hosts, &host)
		}
		info = hosts

	} else {
		req := &cmdb.ListHostReq{
			HostPropertyFilter: &cmdb.QueryFilter{Rule: rule},
			Fields:             fields,
			Page:               page,
		}
		resp, err := s.cc.ListHost(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		info = resp.Data.Info
	}

	if len(info) != 1 {
		logs.Errorf("host is invalid, count: %d, param: %+v, rid: %s", len(info), param, kt.Rid)
		return nil, errors.New("host is invalid")
	}

	host := info[0]
	chargeMonths := calculateMonths(time.Now(), host.BillingExpireTime)

	// 兜底逻辑，如果当前时间加申请的月份数时间还是小于原来的套餐时间，那么就加上一个月
	if time.Now().AddDate(0, chargeMonths, 0).Before(host.BillingExpireTime) {
		chargeMonths++
	}

	return &types.CheckRollingServerHostResp{
		DeviceType:           host.SvrDeviceClassName,
		InstanceChargeType:   host.InstanceChargeType,
		BillingStartTime:     host.BillingStartTime,
		OldBillingExpireTime: host.BillingExpireTime,
		NewBillingExpireTime: time.Now().AddDate(0, chargeMonths, 0),
		ChargeMonths:         chargeMonths,
		CloudInstID:          host.BkCloudInstID,
	}, nil
}

func calculateMonths(startTime, endTime time.Time) int {
	// 计算年份差和月份差
	yearDiff := endTime.Year() - startTime.Year()
	monthDiff := endTime.Month() - startTime.Month()

	// 总月数 = 年份差 * 12 + 月份差
	totalMonths := yearDiff*12 + int(monthDiff)

	// 如果结束时间的日大于开始时间的日，则添加一个月
	if endTime.Day() > startTime.Day() {
		totalMonths++
	}

	return totalMonths
}
