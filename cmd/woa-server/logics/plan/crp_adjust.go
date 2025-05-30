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

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"
)

// AdjustAbleRemainObj adjust able resource plan remained avail cpu core.
type AdjustAbleRemainObj struct {
	OriginDemand *cvmapi.CvmCbsPlanQueryItem
	AdjustType   enumor.CrpAdjustType
	// expectTime 当 adjustType 为 CrpAdjustTypeDelay 时，记录最新的期望交付时间
	ExpectTime  string
	WillConsume int64
}

// CrpAdjustTicketCreator crp adjust ticket creator
// CRP预测修改请求分为两部分：对CRP原有预测的调减，对用户更新内容的追加
type CrpAdjustTicketCreator struct {
	planLogics Logics
	crpCli     cvmapi.CVMClientInterface

	// adjCRPDemandsRst 记录对CRP中可修改的原有预测，将要产生调减的内容
	adjCRPDemandsRst map[string]*AdjustAbleRemainObj
	// appendUpdateDemand 记录将要追加的用户更新的预测需求内容
	appendUpdateDemand []*cvmapi.AdjustUpdatedData
	// adjustAbleDemands 记录每个本地预测需求对应的，CRP中可修改的原有预测
	adjustAbleDemands map[string][]*cvmapi.CvmCbsPlanQueryItem
}

// NewCrpAdjustTicketCreator new CrpAdjustTicketCreator
func NewCrpAdjustTicketCreator(planLogics Logics, crpCli cvmapi.CVMClientInterface) *CrpAdjustTicketCreator {
	return &CrpAdjustTicketCreator{
		planLogics:         planLogics,
		crpCli:             crpCli,
		adjCRPDemandsRst:   make(map[string]*AdjustAbleRemainObj),
		appendUpdateDemand: make([]*cvmapi.AdjustUpdatedData, 0),
		adjustAbleDemands:  make(map[string][]*cvmapi.CvmCbsPlanQueryItem),
	}
}

// CreateCRPTicket create crp adjust ticket
func (c *CrpAdjustTicketCreator) CreateCRPTicket(kt *kit.Kit, ticket *TicketInfo) (string, error) {
	// 1. 生成调整请求
	srcData, updateData, err := c.constructCrpAdjustReqParams(kt, ticket)
	if err != nil {
		logs.Errorf("failed to construct adjust crp plan order request params, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

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
			SrcData:     srcData,
			UpdatedData: updateData,
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

	// 2. 发起提单
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
func (c *CrpAdjustTicketCreator) constructCrpAdjustReqParams(kt *kit.Kit, ticket *TicketInfo) ([]*cvmapi.AdjustSrcData,
	[]*cvmapi.AdjustUpdatedData, error) {

	// 1. 通过通配的方式，查询可修改的CRP原有预测
	err := c.getAllCRPAdjustAbleDemands(kt, ticket.Demands, ticket.PlanProductName)
	if err != nil {
		logs.Errorf("failed to get crp adjust able demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// 2. 构造调整请求的详细内容。从可修改的CRP预测中进行调减，以及从预测需求内容中进行追加
	for _, demand := range ticket.Demands {
		if err := c.constructAdjustDemandDetails(kt, ticket, demand); err != nil {
			logs.Errorf("failed to construct adjust demand details, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
	}

	// 3. 根据暂存的修改信息生成最终的变更请求内容
	srcData := make([]*cvmapi.AdjustSrcData, 0)
	updatedData := make([]*cvmapi.AdjustUpdatedData, 0)
	// 3.1. 构造CRP原始预测调减的请求数据
	for _, adjustObj := range c.adjCRPDemandsRst {
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

		// if adjust type is update, updated item are normal.
		// if adjust type is delay, updated will fill parameter TimeAdjustCvmAmount with OriginOs.
		// if adjust type is cancel, error.
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

		srcData = append(srcData, srcItem)
		updatedData = append(updatedData, updatedItem)
	}

	// 3.2. 构造追加的调整量
	for _, appendItem := range c.appendUpdateDemand {
		updatedData = append(updatedData, appendItem)
	}

	return srcData, updatedData, nil
}

// getCRPAdjustAbleDemandsForAdjustDemands get all adjustable demands from CRP based on the request's adjust demands.
func (c *CrpAdjustTicketCreator) getAllCRPAdjustAbleDemands(kt *kit.Kit, demands rpt.ResPlanDemands,
	planProductName string) error {

	for _, demand := range demands {
		if demand.Original == nil {
			logs.Errorf("failed to construct adjust request, demand original is nil, rid: %s", kt.Rid)
			return errors.New("demand original is nil")
		}

		// query crp for a set of res plan demands that can be adjusted.
		adjustAbleReq := &ptypes.AdjustAbleDemandsReq{
			RegionName:      demand.Original.RegionName,
			DeviceFamily:    demand.Original.Cvm.DeviceFamily,
			DeviceType:      demand.Original.Cvm.DeviceType,
			ExpectTime:      demand.Original.ExpectTime,
			PlanProductName: planProductName,
			ObsProject:      demand.Original.ObsProject,
			DiskType:        demand.Original.Cbs.DiskType,
			ResMode:         demand.Original.Cvm.ResMode,
		}
		adjustAbleDemands, err := c.queryAdjustAbleDemands(kt, adjustAbleReq)
		if err != nil {
			logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, adjustAbleReq,
				kt.Rid)
			return err
		}

		c.adjustAbleDemands[demand.Original.DemandID] = adjustAbleDemands
	}

	return nil
}

// queryAdjustAbleDemands query demands that can be adjusted.
func (c *CrpAdjustTicketCreator) queryAdjustAbleDemands(kt *kit.Kit, req *ptypes.AdjustAbleDemandsReq) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// init request parameter.
	queryReq := &cvmapi.CvmCbsAdjustAblePlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsAdjustAblePlanQueryMethod,
		},
		Params: convAdjustAbleQueryParam(req),
	}

	rst, err := c.crpCli.QueryAdjustAbleDemand(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		logs.Errorf("failed to query adjust able demands, err: %s, req: %+v, rid: %s", err, *queryReq.Params,
			kt.Rid)
		return nil, err
	}

	if rst.Error.Code != 0 {
		logs.Errorf("failed to query adjust able demands, err: %s, crp_trace: %s, rid: %s", rst.Error.Message,
			rst.TraceId, kt.Rid)
		return nil, errors.New(rst.Error.Message)
	}

	if rst.Result == nil || len(rst.Result.Data) == 0 {
		logs.Errorf("failed to query adjust able demands, return is empty, crp_trace: %s, rid: %s",
			rst.TraceId, kt.Rid)
		return nil, errors.New("no demands can be adjusted in CRP")
	}

	return rst.Result.Data, nil
}

// constructAdjustDemandDetails 构造调整预测请求参数
// 因crp不支持部分调整，这里通过将crp中的预测（可能有多条）调减调整的原始量，再在crp中追加一条调整的目标量，来间接实现部分调整。
// 在 prePrepareAdjustAbleData 中计算并记录crp中原始预测的调减量，将追加的目标量暂存在 appendUpdateDemand 中
func (c *CrpAdjustTicketCreator) constructAdjustDemandDetails(kt *kit.Kit, ticket *TicketInfo,
	demand rpt.ResPlanDemand) error {

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
		return errors.New("unsupported ticket type")
	}

	// 预处理，计算crp中的预测调减结果
	err := c.prePrepareAdjustAbleData(kt, adjustType, demand)
	if err != nil {
		logs.Errorf("failed to pre prepare adjust able data, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 加急延期、删除时不追加预测
	if adjustType == enumor.CrpAdjustTypeDelay || ticket.Type == enumor.RPTicketTypeDelete {
		return nil
	}

	// 追加一条等于调整目标量的预测
	addItem, err := c.constructAdjustAppendData(kt, ticket.PlanProductID, ticket.PlanProductName, demand)
	if err != nil {
		logs.Errorf("failed to construct adjust append data, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	c.appendUpdateDemand = append(c.appendUpdateDemand, addItem)

	return nil
}

// prePrepareAdjustAbleData 预处理可调减的数据.
// 适用于多个子订单会重复调整到同一条预测数据的场景，先将所有可能的影响汇总到 adjCRPDemandsRst 中
func (c *CrpAdjustTicketCreator) prePrepareAdjustAbleData(kt *kit.Kit, adjustType enumor.CrpAdjustType,
	demand rpt.ResPlanDemand) error {

	adjustAbleDemands, ok := c.adjustAbleDemands[demand.Original.DemandID]
	if !ok {
		logs.Errorf("failed to get adjust able demands, demand id: %s, rid: %s", demand.Original.DemandID, kt.Rid)
		return errors.New("no adjust able demands")
	}

	// 查询原始预测的预测内外情况
	demandDetail, err := c.planLogics.GetResPlanDemandDetail(kt, demand.Original.DemandID, []int64{})
	if err != nil {
		logs.Errorf("failed to get res plan demand detail, err: %v, demand_id: %s, rid: %s", err,
			demand.Original.DemandID, kt.Rid)
		return err
	}

	// 遍历可用于调减的crp预测，凑齐调减总量
	needCpuCores := demand.Original.Cvm.CpuCore
	for _, adjustAbleD := range adjustAbleDemands {
		if needCpuCores <= 0 {
			break
		}

		// 取消类型的调整单据，不需要给expectTime
		var updateExpectTime string
		if demand.Updated != nil {
			updateExpectTime = demand.Updated.ExpectTime
		}

		canConsume, err := c.calcAdjustAbleDCanConsumeCPU(kt, adjustAbleD, adjustType, needCpuCores, demandDetail,
			updateExpectTime)
		if err != nil {
			logs.Warnf("adjust type is delay, but updated is nil, need adjust: %+v, slice: %s, rid: %s",
				converter.PtrToVal(demand.Original), adjustAbleD.SliceId, kt.Rid)
			continue
		}

		needCpuCores -= canConsume
	}

	if needCpuCores > 0 {
		logs.Errorf("crp demand remained is not enough to deduction, adjust: %+v, need cpu cores: %d, rid: %s",
			demand.Original.Cvm, needCpuCores, kt.Rid)
		return fmt.Errorf("crp demand remained is not enough to deduction, adjust cores: %d, need cores: %d",
			demand.Original.Cvm.CpuCore, needCpuCores)
	}

	return nil
}

// calcAdjustAbleDCanConsumeCPU 计算单个可调整的crp预测，有多少CPU可以被用在demand的调整中
// 并将结果暂存在 adjCRPDemandsRst 中
func (c *CrpAdjustTicketCreator) calcAdjustAbleDCanConsumeCPU(kt *kit.Kit, adjustAbleD *cvmapi.CvmCbsPlanQueryItem,
	adjustType enumor.CrpAdjustType, needCpuCores int64, demandDetail *ptypes.GetPlanDemandDetailResp,
	updateExpectTime string) (int64, error) {

	var canConsume int64
	// 磁盘类型需一致，未知的磁盘类型（包括空值）除外
	if adjustAbleD.DiskType.Name() != demandDetail.DiskTypeName {
		// 加急延期场景CRP会根据原预测的磁盘类型创建新的预测，此时需保证磁盘类型完全一致，不能为空
		if adjustType == enumor.CrpAdjustTypeDelay || adjustAbleD.DiskType.Name() != "" {
			return canConsume, nil
		}
	}

	// 预测内外需一致
	if demandDetail.PlanType.GetCode() != enumor.PlanType(adjustAbleD.InPlan).GetCode() {
		return canConsume, nil
	}

	remainedCpuCores := adjustAbleD.RealCoreAmount
	if _, ok := c.adjCRPDemandsRst[adjustAbleD.SliceId]; ok {
		remainedCpuCores -= c.adjCRPDemandsRst[adjustAbleD.SliceId].WillConsume
	}

	canConsume = min(needCpuCores, remainedCpuCores)
	// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
	if canConsume <= 0 || adjustAbleD.CvmAmount == 0 {
		return 0, nil
	}

	if _, ok := c.adjCRPDemandsRst[adjustAbleD.SliceId]; !ok {
		c.adjCRPDemandsRst[adjustAbleD.SliceId] = &AdjustAbleRemainObj{
			OriginDemand: adjustAbleD.Clone(),
			AdjustType:   adjustType,
		}
		// 取消类型的调整单据，不需要expectTime
		if updateExpectTime != "" {
			c.adjCRPDemandsRst[adjustAbleD.SliceId].ExpectTime = updateExpectTime
		}
	}
	adjustAbleRemain := c.adjCRPDemandsRst[adjustAbleD.SliceId]

	if adjustType != adjustAbleRemain.AdjustType {
		logs.Warnf("adjust type is not eq, adjust type: %s, slice: %s, adjust demand: %+v, rid: %s", adjustType,
			adjustAbleD.SliceId, demandDetail, kt.Rid)
		return canConsume, errors.New("adjust type is not eq")
	}

	if adjustType == enumor.CrpAdjustTypeDelay {
		if updateExpectTime == "" {
			logs.Warnf("adjust type is delay, but updated expect time is nil, slice: %s, adjust demand: %+v, rid: %s",
				adjustAbleD.SliceId, demandDetail, kt.Rid)
			return canConsume, errors.New("adjust type is delay, but updated expect time is nil")
		}
		// 如果多条通配的预测一起延期，延期的期望时间必须一致。否则不能在同一个CRP demand中操作调整
		if updateExpectTime != adjustAbleRemain.ExpectTime {
			logs.Warnf("adjust type is delay, but expect time is not eq, slice: %s, adjust demand: %+v, rid: %s",
				adjustAbleD.SliceId, demandDetail, kt.Rid)
			return canConsume, errors.New("adjust type is delay, but expect time is not eq")
		}
	}

	adjustAbleRemain.WillConsume += canConsume

	return canConsume, nil
}

// constructAdjustAppendData construct adjust append data.
func (c *CrpAdjustTicketCreator) constructAdjustAppendData(kt *kit.Kit, planProductID int64, planProductName string,
	demand rpt.ResPlanDemand) (
	*cvmapi.AdjustUpdatedData, error) {

	// 根据HCM的diskType，生成CRP的diskType和diskName
	diskTypeName := demand.Updated.Cbs.DiskType.Name()
	diskType, err := enumor.GetCRPDiskTypeFromCRPName(diskTypeName)
	if err != nil {
		return nil, err
	}

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
		DiskType:        diskType,
		DiskTypeName:    diskTypeName,
		AllDiskAmount:   demand.Updated.Cbs.DiskSize,
	}

	updatedData := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(enumor.CrpAdjustTypeUpdate),
		CvmCbsPlanQueryItem: demandItem,
	}

	return updatedData, nil
}
