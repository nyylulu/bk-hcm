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
	"strings"

	mtypes "hcm/cmd/woa-server/types/meta"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	rpcd "hcm/pkg/dal/table/resource-plan/res-plan-crp-demand"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"

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
			CityName:        demand.Updated.RegionName,
			ZoneName:        demand.Updated.ZoneName,
			CoreTypeName:    demand.Updated.Cvm.CoreType,
			InstanceModel:   demand.Updated.Cvm.DeviceType,
			CvmAmount:       demand.Updated.Cvm.Os.InexactFloat64(),
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
	adjustReq, err := c.constructCrpAdjustReq(kt, ticket)
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
		logs.Errorf("failed to adjust cvm & cbs plan order, code: %d, msg: %s, req: %v, crp_trace: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, *adjustReq, resp.TraceId, kt.Rid)
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to adjust cvm & cbs plan order, for return empty order id, rid: %s", kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// constructCrpAdjustReq construct cvm cbs plan adjust request.
func (c *Controller) constructCrpAdjustReq(kt *kit.Kit, ticket *TicketInfo) (*cvmapi.CvmCbsPlanAdjustReq, error) {
	adjustReq := &cvmapi.CvmCbsPlanAdjustReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAdjustMethod,
		},
		Params: &cvmapi.CvmCbsPlanAdjustParam{
			BaseInfo: &cvmapi.AdjustBaseInfo{
				DeptId:          int(ticket.VirtualDeptID),
				DeptName:        ticket.VirtualDeptName,
				PlanProductName: ticket.PlanProductName,
				Desc:            "",
			},
			SrcData:     make([]*cvmapi.AdjustSrcData, 0),
			UpdatedData: make([]*cvmapi.AdjustUpdatedData, 0),
			UserName:    ticket.Applicant,
		},
	}

	switch ticket.DemandClass {
	case enumor.DemandClassCVM:
		adjustReq.Params.BaseInfo.Desc = cvmapi.CvmCbsPlanDefaultCvmDesc
	case enumor.DemandClassCA:
		adjustReq.Params.BaseInfo.Desc = cvmapi.CvmCbsPlanDefaultCADesc
	default:
		logs.Warnf("failed to construct adjust desc, unsupported demand class: %s, rid: %s", ticket.DemandClass, kt.Rid)
	}

	adjCRPDemandsRst := make(map[string]*AdjustAbleRemainObj)
	for _, demand := range ticket.Demands {
		if demand.Original == nil {
			logs.Errorf("failed to construct adjust request, demand original is nil, rid: %s", kt.Rid)
			return nil, errors.New("demand original is nil")
		}

		// query crp for a set of res plan demands that can be adjusted.
		adjustAbleReq := &ptypes.AdjustAbleDemandsReq{
			RegionName:      demand.Original.RegionName,
			DeviceFamily:    demand.Original.Cvm.DeviceFamily,
			DeviceType:      demand.Original.Cvm.DeviceType,
			ExpectTime:      demand.Original.ExpectTime,
			PlanProductName: ticket.PlanProductName,
			ObsProject:      demand.Original.ObsProject,
			DiskType:        demand.Original.Cbs.DiskType,
			ResMode:         demand.Original.Cvm.ResMode,
		}
		adjustAbleDemands, err := c.QueryAdjustAbleDemands(kt, adjustAbleReq)
		if err != nil {
			logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, *adjustReq, kt.Rid)
			return nil, err
		}

		// 预处理CRP中预测的调整结果，以及给出调整后追加的更新请求
		updatedItem, err := c.constructAdjustDemandDetails(kt, ticket, demand, adjustAbleDemands, adjCRPDemandsRst)
		if err != nil {
			logs.Errorf("failed to construct adjust demand details, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if updatedItem != nil {
			adjustReq.Params.UpdatedData = append(adjustReq.Params.UpdatedData, updatedItem)
		}
	}

	for _, adjustObj := range adjCRPDemandsRst {
		srcItem, updatedItem, err := c.constructAdjustUpdatedData(kt, adjustObj)
		if err != nil {
			logs.Errorf("failed to construct adjust updated data, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		adjustReq.Params.SrcData = append(adjustReq.Params.SrcData, srcItem)
		adjustReq.Params.UpdatedData = append(adjustReq.Params.UpdatedData, updatedItem)
	}

	return adjustReq, nil
}

// constructAdjustDemandDetails 构造调整预测请求参数
// 因crp不支持部分调整，这里通过将crp中的预测（可能有多条）调减调整的原始量，再在crp中追加一条调整的目标量，来间接实现部分调整。
// 对crp数据的调减记录在入参 adjCRPDemandsRst map中，上层调用时需注意入参 adjCRPDemandsRst 会被本方法修改
func (c *Controller) constructAdjustDemandDetails(kt *kit.Kit, ticket *TicketInfo, demand rpt.ResPlanDemand,
	adjustAbleCrpDemands []*cvmapi.CvmCbsPlanQueryItem, adjCRPDemandsRst map[string]*AdjustAbleRemainObj) (
	*cvmapi.AdjustUpdatedData, error) {

	var adjustType enumor.CrpAdjustType
	switch ticket.Type {
	case enumor.RPTicketTypeAdjust:
		adjustType = enumor.CrpAdjustTypeUpdate
		if demand.Updated.ExpectTime != demand.Original.ExpectTime {
			adjustType = enumor.CrpAdjustTypeDelay
		}
	case enumor.RPTicketTypeDelete:
		// 我们的删除对crp来说是部分调减
		adjustType = enumor.CrpAdjustTypeUpdate
	default:
		logs.Errorf("unsupported ticket type: %s， rid: %s", ticket.Type, kt.Rid)
		return nil, errors.New("unsupported ticket type")
	}

	// 预处理，计算crp中的预测调减结果
	adjCRPDemandsRst, err := c.prePrepareAdjustAbleData(kt, adjustType, adjustAbleCrpDemands, adjCRPDemandsRst, demand)
	if err != nil {
		logs.Errorf("failed to pre prepare adjust able data, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 加急延期、删除时不追加预测
	if adjustType == enumor.CrpAdjustTypeDelay || ticket.Type == enumor.RPTicketTypeDelete {
		return nil, nil
	}

	// 追加一条等于调整目标量的预测
	addItem, err := c.constructAdjustAppendData(kt, ticket.PlanProductID, ticket.PlanProductName, demand)
	if err != nil {
		logs.Errorf("failed to construct adjust append data, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return addItem, nil
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

// prePrepareAdjustAbleData 预处理可调减的数据.
// 适用于多个子订单会重复调整到同一条预测数据的场景，先将所有可能的影响汇总.
// 再通过 constructAdjustUpdatedData 生成最终的调整请求
func (c *Controller) prePrepareAdjustAbleData(kt *kit.Kit, adjustType enumor.CrpAdjustType,
	adjustAbleCrpDemands []*cvmapi.CvmCbsPlanQueryItem, adjustDemandsRemainAvail map[string]*AdjustAbleRemainObj,
	demand rpt.ResPlanDemand) (map[string]*AdjustAbleRemainObj, error) {

	// 查询原始预测的预测内外情况
	demandDetail, err := c.GetResPlanDemandDetail(kt, demand.Original.DemandID, []int64{})
	if err != nil {
		logs.Errorf("failed to get res plan demand detail, err: %v, demand_id: %s, rid: %s", err,
			demand.Original.DemandID, kt.Rid)
		return nil, err
	}

	// 遍历可用于调减的crp预测，凑齐调减总量
	needCpuCores := demand.Original.Cvm.CpuCore
	for _, adjustAbleD := range adjustAbleCrpDemands {
		if needCpuCores <= 0 {
			break
		}

		// 预测内外需一致
		if demandDetail.PlanType.GetCode() != enumor.PlanType(adjustAbleD.InPlan).GetCode() {
			continue
		}

		remainedCpuCores := adjustAbleD.RealCoreAmount
		if _, ok := adjustDemandsRemainAvail[adjustAbleD.SliceId]; ok {
			remainedCpuCores -= adjustDemandsRemainAvail[adjustAbleD.SliceId].WillConsume
		}

		canConsume := min(needCpuCores, remainedCpuCores)
		// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
		if canConsume <= 0 || adjustAbleD.CvmAmount == 0 {
			continue
		}

		if _, ok := adjustDemandsRemainAvail[adjustAbleD.SliceId]; !ok {
			adjustDemandsRemainAvail[adjustAbleD.SliceId] = &AdjustAbleRemainObj{
				OriginDemand: adjustAbleD.Clone(),
				AdjustType:   adjustType,
			}
			// 取消类型的调整单据，没有updated字段
			if demand.Updated != nil {
				adjustDemandsRemainAvail[adjustAbleD.SliceId].ExpectTime = demand.Updated.ExpectTime
			}
		}
		adjustAbleRemain := adjustDemandsRemainAvail[adjustAbleD.SliceId]

		if adjustType != adjustAbleRemain.AdjustType {
			logs.Warnf("adjust type is not eq, need adjust: %+v, adjust type: %s, slice: %s, rid: %s",
				converter.PtrToVal(demand.Original), adjustType, adjustAbleD.SliceId, kt.Rid)
			continue
		}

		if adjustType == enumor.CrpAdjustTypeDelay {
			if demand.Updated == nil {
				logs.Warnf("adjust type is delay, but updated is nil, need adjust: %+v, slice: %s, rid: %s",
					converter.PtrToVal(demand.Original), adjustAbleD.SliceId, kt.Rid)
				continue
			}
			if demand.Updated.ExpectTime != adjustAbleRemain.ExpectTime {
				logs.Warnf("adjust type is delay, but expect time is not eq, need adjust: %+v, slice: %s, rid: %s",
					converter.PtrToVal(demand.Original), adjustAbleD.SliceId, kt.Rid)
				continue
			}
		}

		adjustAbleRemain.WillConsume += canConsume
		needCpuCores -= canConsume
	}

	if needCpuCores > 0 {
		logs.Errorf("crp demand remained is not enough to deduction, adjust: %+v, need cpu cores: %d, rid: %s",
			demand.Original.Cvm, needCpuCores, kt.Rid)
		return nil, fmt.Errorf("crp demand remained is not enough to deduction, adjust cores: %d, need cores: %d",
			demand.Original.Cvm.CpuCore, needCpuCores)
	}

	return adjustDemandsRemainAvail, nil
}

// constructAdjustUpdatedData construct adjust updated data.
// if adjust type is update, updated item are normal.
// if adjust type is delay, updated will fill parameter TimeAdjustCvmAmount with OriginOs.
// if adjust type is cancel, error.
func (c *Controller) constructAdjustUpdatedData(kt *kit.Kit, adjustObj *AdjustAbleRemainObj) (
	*cvmapi.AdjustSrcData, *cvmapi.AdjustUpdatedData, error) {

	adjustAbleD := adjustObj.OriginDemand
	willConsume := adjustObj.WillConsume

	deviceCore := float64(adjustAbleD.CoreAmount) / adjustAbleD.CvmAmount
	// 和CRP确认保留2位小数可以，但是肯定会存在误差
	willChangeCvm, err := math.RoundToDecimalPlaces(float64(willConsume)/deviceCore, 2)
	if err != nil {
		logs.Errorf("failed to round change cvm to 2 decimal places, err: %v, crp_demand:%s, change cvm: %f, "+
			"rid: %s", err, adjustAbleD.DemandId, float64(willConsume)/deviceCore, kt.Rid)
		return nil, nil, err
	}

	srcItem := &cvmapi.AdjustSrcData{
		AdjustType:          string(adjustObj.AdjustType),
		CvmCbsPlanQueryItem: adjustAbleD.Clone(),
	}
	updatedItem := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(adjustObj.AdjustType),
		CvmCbsPlanQueryItem: adjustAbleD.Clone(),
	}

	switch adjustObj.AdjustType {
	case enumor.CrpAdjustTypeUpdate:
		// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
		updatedItem.CvmAmount = max(updatedItem.CvmAmount-willChangeCvm, 0)
		updatedItem.CoreAmount -= willConsume
	case enumor.CrpAdjustTypeDelay:
		updatedItem.UseTime = adjustObj.ExpectTime
		updatedItem.TimeAdjustCvmAmount = willChangeCvm
	default:
		logs.Errorf("failed to construct adjust updated data. unsupported adjust type: %s, rid: %s",
			adjustObj.AdjustType, kt.Rid)
		return nil, nil, fmt.Errorf("unsupported adjust type: %s", adjustObj.AdjustType)
	}

	return srcItem, updatedItem, nil
}

// constructAdjustAppendData construct adjust append data.
func (c *Controller) constructAdjustAppendData(kt *kit.Kit, planProductID int64, planProductName string,
	demand rpt.ResPlanDemand) (
	*cvmapi.AdjustUpdatedData, error) {

	demandItem := &cvmapi.CvmCbsPlanQueryItem{
		SliceId:         demand.Original.DemandID,
		ProjectName:     string(demand.Updated.ObsProject),
		PlanProductId:   int(planProductID),
		PlanProductName: planProductName,
		InstanceType:    demand.Updated.Cvm.DeviceClass,
		CityId:          0,
		CityName:        demand.Updated.RegionName,
		ZoneId:          0,
		ZoneName:        demand.Updated.ZoneName,
		InstanceModel:   demand.Updated.Cvm.DeviceType,
		UseTime:         demand.Updated.ExpectTime,
		CvmAmount:       demand.Updated.Cvm.Os.InexactFloat64(),
		InstanceIO:      int(demand.Updated.Cbs.DiskIo),
		DiskType:        0,
		DiskTypeName:    demand.Updated.Cbs.DiskTypeName,
		AllDiskAmount:   demand.Updated.Cbs.DiskSize,
	}

	updatedData := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(enumor.CrpAdjustTypeUpdate),
		CvmCbsPlanQueryItem: demandItem,
	}

	return updatedData, nil
}

// upsertCrpDemandBizRel upsert crp demand biz rel.
func (c *Controller) upsertCrpDemandBizRel(kt *kit.Kit, crpDemandIDs []int64, demandClass enumor.DemandClass,
	bizOrgRel mtypes.BizOrgRel, reviser string) error {

	if len(crpDemandIDs) == 0 {
		logs.Errorf("crp demand ids is empty, rid: %s", kt.Rid)
		return errors.New("crp demand ids is empty")
	}

	crpDemandFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "crp_demand_id", Op: filter.In.Factory(), Value: crpDemandIDs},
			filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizOrgRel.BkBizID},
		},
	}

	listOpt := &types.ListOption{
		Filter: crpDemandFilter,
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
		err = c.dao.ResPlanCrpDemand().Update(kt, crpDemandFilter, update)
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
