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

// Package generator provides ...
package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/utils"
)

var (
	// cvmApplyNumReg CVM主机申请匹配数量
	cvmApplyNumReg = regexp.MustCompile(`计算最终当前可申领量(\d+)`)
)

// createCVM starts a cvm creating task
func (g *Generator) createCVM(kt *kit.Kit, cvm *types.CVM, order *types.ApplyOrder) (string, error) {
	// construct cvm launch request
	createReq := g.getCreateCvmReq(cvm)

	// 增加日志记录
	jsonReq, err := json.Marshal(createReq)
	if err != nil {
		logs.Warnf("scheduler:logics:generator:create:cvm json marshal failed, err: %+v, rid: %s", err, kt.Rid)
	}
	logs.Infof("scheduler:logics:generator:create:cvm:start, subOrderID: %s, create cvm req: %s, rid: %s",
		order.SubOrderId, string(jsonReq), kt.Rid)

	// call cvm api to launchCvm cvm order
	maxRetry := 3
	resp := new(cvmapi.OrderCreateResp)
	for try := 0; try < maxRetry; try++ {
		// need not wait for the first try
		if try != 0 {
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
		}

		resp, err = g.cvm.CreateCvmOrder(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Warnf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, subOrderID: %s, "+
				"req: %s, err: %v, rid: %s", order.SubOrderId, string(jsonReq), err, kt.Rid)
			continue
		}

		if resp.Error.Code != 0 {
			logs.Warnf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, subOrderID: %s, "+
				"code: %d, msg: %s, crpTraceID: %s, rid: %s", order.SubOrderId, resp.Error.Code, resp.Error.Message,
				resp.TraceId, kt.Rid)
			if isRetry, applyNum := g.needRetryCreateCvm(resp.Error.Code, resp.Error.Message); isRetry {
				if applyNum > 0 {
					createReq.Params.ApplyNum = min(applyNum, createReq.Params.ApplyNum)
				}
				continue
			}
		}

		break
	}

	if err != nil {
		logs.Errorf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, subOrderID: %s, "+
			"req: %s, err: %v, rid: %s", order.SubOrderId, string(jsonReq), err, kt.Rid)
		return "", err
	}

	respStr := ""
	b, err := json.Marshal(resp)
	if err != nil {
		logs.Warnf("scheduler:logics:generator:create:cvm json marshal failed, err: %v, rid: %s", err, kt.Rid)
	}

	respStr = string(b)
	logs.Infof("scheduler:logics:generator:create:cvm:success, subOrderID: %s, create cvm req: %s, resp: %s, rid: %s",
		order.SubOrderId, string(jsonReq), respStr, kt.Rid)

	if resp.Error.Code != 0 {
		return "", fmt.Errorf("cvm order create task failed, code: %d, msg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	if resp.Result.OrderId == "" {
		return "", fmt.Errorf("cvm order create task return empty order id, crpTraceID: %s", resp.TraceId)
	}

	return resp.Result.OrderId, nil
}

func (g *Generator) getCreateCvmReq(cvm *types.CVM) *cvmapi.OrderCreateReq {
	deptName := cvmapi.CvmLaunchDeptName
	if cvm.VirtualDeptName != "" {
		deptName = cvm.VirtualDeptName
	}
	createReq := &cvmapi.OrderCreateReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmLaunchMethod,
		},
		Params: &cvmapi.OrderCreateParams{
			Zone:          cvm.Zone,
			DeptName:      deptName,
			ProductName:   cvm.BkProductName,
			Business1Id:   cvmapi.CvmLaunchBiz1Id,
			Business1Name: cvmapi.CvmLaunchBiz1Name,
			Business2Id:   cvmapi.CvmLaunchBiz2Id,
			Business2Name: cvmapi.CvmLaunchBiz2Name,
			Business3Id:   cvmapi.CvmLaunchBiz3Id,
			Business3Name: cvmapi.CvmLaunchBiz3Name,
			ProjectId:     int(cvm.BkProductID),
			Image:         &cvmapi.Image{ImageId: cvm.ImageId, ImageName: cvm.ImageName},
			InstanceType:  cvm.InstanceType,
			DataDisk:      make([]*cvmapi.DataDisk, 0),
			VpcId:         cvm.VPCId,
			SubnetId:      cvm.SubnetId,
			ApplyNum:      int(cvm.ApplyNumber),
			PassWord:      g.clientConf.CvmOpt.CvmLaunchPassword,
			Security: &cvmapi.Security{
				SecurityGroupId:   cvm.SecurityGroupId,
				SecurityGroupName: cvm.SecurityGroupName,
				SecurityGroupDesc: cvm.SecurityGroupDesc,
			},
			UseTime:           time.Now().Format(constant.DateTimeLayout),
			Memo:              cvm.NoteInfo,
			Operator:          cvm.Operator,
			BakOperator:       cvm.Operator,
			ChargeType:        cvmapi.ChargeTypePrePaid,
			InheritInstanceId: cvm.InheritInstanceId,
		},
	}
	// 计费模式
	if len(cvm.ChargeType) > 0 {
		createReq.Params.ChargeType = cvm.ChargeType
	}
	// 包年包月时才需要设置计费时长
	if createReq.Params.ChargeType == cvmapi.ChargeTypePrePaid && cvm.ChargeMonths > 0 {
		createReq.Params.ChargeMonths = cvm.ChargeMonths
	}
	// set system disk parameters
	itDev := regexp.MustCompile(`^IT3\.|^IT2\.|^I3\.|^IT5\.|^IT5c\.`).FindStringSubmatch(cvm.InstanceType)
	if len(itDev) > 0 {
		createReq.Params.SystemDiskType = cvmapi.CvmLaunchSystemDiskTypeBasic
		createReq.Params.SystemDiskSize = cvmapi.CvmLaunchSystemDiskSizeBasic
	} else {
		createReq.Params.SystemDiskType = cvmapi.CvmLaunchSystemDiskTypePremium
		createReq.Params.SystemDiskSize = cvmapi.CvmLaunchSystemDiskSizePremium
	}
	// set system disk and data disk for special instance type.
	if cvm.InstanceType == "BMGY5.16XLARGE256" {
		createReq.Params.SystemDiskType = cvmapi.CvmLaunchSystemDiskTypeBasic
		createReq.Params.SystemDiskSize = 440
		createReq.Params.DataDisk = append(createReq.Params.DataDisk, &cvmapi.DataDisk{
			DataDiskType: cvmapi.CvmLaunchSystemDiskTypeBasic,
			DataDiskSize: 1320,
		})
	}
	if cvm.DiskSize > 0 {
		createReq.Params.DataDisk = append(createReq.Params.DataDisk, &cvmapi.DataDisk{
			DataDiskType: cvm.DiskType,
			DataDiskSize: int(cvm.DiskSize),
		})
	}
	// set obs project type
	requireType := enumor.RequireType(cvm.ApplyType)
	createReq.Params.ObsProject = string(requireType.ToObsProject())
	if requireType == enumor.RequireTypeGreenChannel {
		createReq.Params.ResourceType = cvmapi.ResourceTypeQuick
	}
	return createReq
}

func (g *Generator) needRetryCreateCvm(code int, msg string) (bool, int) {
	// success
	if code == 0 {
		return false, 0
	}

	// sold out
	if code == -20004 && strings.Contains(msg, "已售罄，请更换可用区") {
		return false, 0
	}

	// no capacity enough
	if code == -20004 || code == -20000 {
		applyNum := g.getCrpCvmRemainNum(msg)
		if applyNum > 0 {
			return true, applyNum
		}
		return false, 0
	}

	return true, 0
}

func (g *Generator) getCrpCvmRemainNum(msg string) int {
	if !strings.Contains(msg, "无法满足本次需求量") {
		return 0
	}

	// 解析CRP的报错消息获取最终可申请的数量
	return g.parseCrpCvmApplyNum(msg)
}

// parseCrpCvmApplyNum 解析CRP的报错消息获取最终可申请的数量
func (g *Generator) parseCrpCvmApplyNum(msg string) int {
	match := cvmApplyNumReg.FindStringSubmatch(msg)
	if len(match) > 1 {
		applyNum, err := strconv.Atoi(match[1])
		if err == nil && applyNum > 0 {
			return applyNum
		}
	}
	return 0
}

// CheckCVM checks cvm creating task result
func (g *Generator) CheckCVM(kt *kit.Kit, orderId, subOrderID string) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to query cvm order by id %s, err: %v", orderId, err)
		}

		if obj == nil {
			return false, fmt.Errorf("cvm order %s not found", orderId)
		}

		resp, ok := obj.(*cvmapi.OrderQueryResp)
		if !ok {
			return false, fmt.Errorf("object with order id %s is not a cvm order response: %+v", orderId, obj)
		}

		if resp.Error.Code != 0 {
			return false, fmt.Errorf("query cvm order failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		}

		if resp.Result == nil {
			return false, fmt.Errorf("query cvm order failed, for result is null, resp: %+v", resp)
		}

		num := len(resp.Result.Data)
		if num != 1 {
			return false, fmt.Errorf("query cvm order return %d orders with order id: %s", num, orderId)
		}

		// 检查CRP订单是否超出处理时间并记录日志
		g.checkRecordCrpOrderTimeout(kt, subOrderID, resp)

		status := enumor.CrpOrderStatus(resp.Result.Data[0].Status)
		if status != enumor.CrpOrderStatusFinish &&
			status != enumor.CrpOrderStatusReject &&
			status != enumor.CrpOrderStatusFailed {
			return false, fmt.Errorf("cvm order %s handling", orderId)
		}

		if status != enumor.CrpOrderStatusFinish {
			return true, fmt.Errorf("order %s failed, status: %d", resp.Result.Data[0].OrderId, status)
		}

		// crp侧订单完成时，不一定代表cvm生产成功，这里需要做处理，如果没有成功创建的实例，则也认为创建失败
		if resp.Result.Data[0].SucInstanceCount <= 0 {
			return true, fmt.Errorf("CRP申领失败，详情可咨询2000(TEG技术支持)，CRP申请单链接: %s, status: %d, "+
				"sucInstanceCount: %d", cvmapi.CvmOrderLinkPrefix+resp.Result.Data[0].OrderId, status,
				resp.Result.Data[0].SucInstanceCount)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := cvmapi.NewOrderQueryReq(&cvmapi.OrderQueryParam{OrderId: []string{orderId}})
		resp, err := g.cvm.QueryCvmOrders(nil, nil, req)
		if err != nil {
			return nil, err
		}

		// call cvm api to query cvm order status
		return resp, nil
	}

	// TODO: get retry strategy from config
	_, err := utils.Retry(doFunc, checkFunc, uint64(7*types.OneDayDuration.Seconds()), 60)
	return err
}

// listCVM lists created cvm by order id
func (g *Generator) listCVM(orderId string) ([]*cvmapi.InstanceItem, error) {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, err
		}
		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := &cvmapi.InstanceQueryReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.CvmInstanceStatusMethod,
			},
			Params: &cvmapi.InstanceQueryParam{
				OrderId: []string{orderId},
			},
		}
		return g.cvm.QueryCvmInstances(nil, nil, req)
	}

	// TODO: get retry strategy from config
	obj, err := utils.Retry(doFunc, checkFunc, 120, 5)

	if err != nil {
		return nil, err
	}
	resp, ok := obj.(*cvmapi.InstanceQueryResp)
	if !ok {
		return nil, fmt.Errorf("object with order id %s is not a cvm instance response: %+v", orderId, obj)
	}

	logs.Infof("get cvm instance resp: %+v", resp)

	if resp.Error.Code != 0 {
		return nil, fmt.Errorf("list cvm instance failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		return nil, errors.New("list cvm instance failed, for result is null")
	}

	return resp.Result.Data, nil
}

// buildCvmReq construct a cvm creating task request
func (g *Generator) buildCvmReq(kt *kit.Kit, order *types.ApplyOrder, zone string, replicas uint,
	excludeSubnetIDMap map[string]struct{}) (*types.CVM, error) {

	// TODO: get parameters from config
	// construct cvm launch req
	req := &types.CVM{
		AppId:             "931",
		ApplyType:         int64(order.RequireType),
		AppModuleId:       51524,
		Operator:          "dommyzhang",
		ApplyNumber:       replicas,
		NoteInfo:          order.Remark,
		Area:              order.Spec.Region,
		Zone:              zone,
		InstanceType:      order.Spec.DeviceType,
		DiskType:          order.Spec.DiskType,
		DiskSize:          order.Spec.DiskSize,
		ChargeType:        order.Spec.ChargeType,
		ChargeMonths:      order.Spec.ChargeMonths,
		InheritInstanceId: order.Spec.InheritInstanceId,
	}
	// set disk type default value
	if len(req.DiskType) == 0 {
		req.DiskType = cvmapi.CvmLaunchSystemDiskTypePremium
	}
	// vpc and subnet
	if order.Spec.Vpc != "" {
		req.VPCId = order.Spec.Vpc
	} else {
		vpc, err := g.configLogics.Vpc().GetRegionDftVpc(kt, order.Spec.Region)
		if err != nil {
			logs.Errorf("failed to get region default vpc, err: %v, subOrderID: %s, region: %s, rid: %s", err,
				order.SubOrderId, order.Spec.Region, kt.Rid)
			return nil, err
		}
		req.VPCId = vpc
	}
	if order.Spec.Subnet != "" {
		req.SubnetId = order.Spec.Subnet
	} else {
		subnetList, err := g.getCvmSubnet(kt, zone, req.VPCId, order)
		if err != nil {
			logs.Errorf("failed to get available subnet, subOrderID: %s, err: %v, region: %s, zone: %s, vpcID: %s, "+
				"rid: %s", order.SubOrderId, err, order.Spec.Region, zone, req.VPCId, kt.Rid)
			return nil, err
		}
		sort.Sort(sort.Reverse(subnetList))
		subnetID := ""
		applyNum := uint(0)
		for _, subnet := range subnetList {
			if _, ok := excludeSubnetIDMap[subnet.Id]; ok {
				logs.Warnf("exclude subnet id: %s, subOrderID: %s, rid: %s", subnet.Id, order.SubOrderId, kt.Rid)
				continue
			}

			capacity, err := g.getCapacity(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region, zone,
				req.VPCId, subnet.Id, order.Spec.ChargeType)
			if err != nil {
				logs.Errorf("failed to get capacity with subnet %s, subOrderID: %s, subnetNum: %d, zone: %s, "+
					"reqVpcID: %s, err: %v, rid: %s", subnet.Id, order.SubOrderId, len(subnetList), zone,
					req.VPCId, err, kt.Rid)
				continue
			}
			maxNum, ok := capacity[zone]
			if !ok {
				logs.Warnf("get no capacity with zone %s and subnet %s, subOrderID: %s, rid: %s",
					zone, subnet.Id, order.SubOrderId, kt.Rid)
				continue
			}
			if maxNum > 0 {
				subnetID = subnet.Id
				applyNum = uint(maxNum)
				break
			}
			// 记录日志，方便排查线上资源申请问题
			logs.Errorf("buildCvmReq:get no available capacity info, subOrderID: %s, subnetNum: %d, zone: %s, "+
				"reqVpcID: %s, subnet: %+v, orderInfo: %+v, capacity: %+v, rid: %s", order.SubOrderId, len(subnetList),
				zone, req.VPCId, cvt.PtrToVal(subnet), cvt.PtrToVal(order), capacity, kt.Rid)
		}

		if subnetID == "" || applyNum <= 0 {
			// get capacity detail as component of error message
			capInfo, _ := g.getCapacityDetail(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region, zone,
				req.VPCId, "", order.Spec.ChargeType)
			capInfoStr, err := json.Marshal(capInfo)
			if err != nil {
				logs.Warnf("buildCvmReq:get empty subnet info json marshal failed, err: %+v, rid: %s", err, kt.Rid)
			}
			// 记录日志，方便排查线上资源申请问题
			logs.Errorf("buildCvmReq:get empty subnet info failed, subOrderID: %s, subnetNum: %d, applyNum: %d, "+
				"zone: %s, reqVpcID: %s, subnetID: %s, orderInfo: %+v, capInfoStr: %s, rid: %s", order.SubOrderId,
				len(subnetList), applyNum, zone, req.VPCId, subnetID, cvt.PtrToVal(order), capInfoStr, kt.Rid)
			return nil, fmt.Errorf("no capacity: %s", capInfoStr)
		}
		req.SubnetId = subnetID
		if applyNum < replicas {
			// set apply number to min(replicas, leftIp)
			req.ApplyNumber = applyNum
		}
		// 记录日志，方便排查线上资源申请问题
		subnetListRemain, err := json.Marshal(subnetList)
		if err != nil {
			logs.Warnf("buildCvmReq:get subnet info json marshal failed, err: %+v, rid: %s", err, kt.Rid)
		}
		logs.Infof("buildCvmReq:get subnet info success, subOrderID: %s, subnetNum: %d, applyNum: %d, replicas: %d, "+
			"zone: %s, reqVpcID: %s, subnetID: %s, orderInfo: %+v, subnetList: %s, req: %+v, rid: %s",
			order.SubOrderId, len(subnetList), applyNum, replicas, zone, req.VPCId, subnetID, cvt.PtrToVal(order),
			subnetListRemain, cvt.PtrToVal(req), kt.Rid)
	}

	// image
	req.ImageId = order.Spec.ImageId
	req.ImageName = order.Spec.Image
	// security group
	sg, err := g.configLogics.Sg().GetRegionDftSg(kt, order.Spec.Region)
	if err != nil {
		logs.Errorf("failed to get region default sg, err: %v, subOrderID: %s, region: %s, rid: %s", err,
			order.SubOrderId, order.Spec.Region, kt.Rid)
		return nil, err
	}
	req.SecurityGroupId = sg.SgID
	req.SecurityGroupName = sg.SgName
	req.SecurityGroupDesc = sg.SgDesc

	productInfo, err := g.getProductInfo(kt, order)
	if err != nil {
		logs.Errorf("get product message failed, err: %v, order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}

	req.BkProductID = productInfo.BkProductID
	req.BkProductName = productInfo.BkProductName
	req.VirtualDeptID = productInfo.VirtualDeptID
	req.VirtualDeptName = productInfo.VirtualDeptName

	return req, nil
}

func (g *Generator) getProductInfo(kt *kit.Kit, order *types.ApplyOrder) (cmdb.CompanyCmdbInfo, error) {
	if order.RequireType.IsUseManageBizPlan() {
		return cmdb.CompanyCmdbInfo{
			BkProductID: cvmapi.CvmLaunchProjectId, BkProductName: cvmapi.CvmLaunchProductName,
			VirtualDeptID: cvmapi.CvmDeptId, VirtualDeptName: cvmapi.CvmLaunchDeptName}, nil
	}

	param := &cmdb.SearchBizCompanyCmdbInfoParams{BizIDs: []int64{order.BkBizId}}
	resp, err := g.cc.SearchBizCompanyCmdbInfo(kt, param)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, param: %+v, rid: %s", err, *param, kt.Rid)
		return cmdb.CompanyCmdbInfo{}, err
	}
	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return cmdb.CompanyCmdbInfo{}, errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	return (*resp)[0], nil
}

// AvailSubnetList available subnet list
type AvailSubnetList []*cvmapi.SubnetInfo

// Len available subnet list length
func (l AvailSubnetList) Len() int {
	return len(l)
}

// Less compare two host priority
func (l AvailSubnetList) Less(i, j int) bool {
	if l[i].LeftIpNum == l[j].LeftIpNum {
		return l[i].Id < l[j].Id
	}
	return l[i].LeftIpNum < l[j].LeftIpNum
}

// Swap swap two subnet
func (l AvailSubnetList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (g *Generator) getCvmSubnet(kt *kit.Kit, zone, vpc string, order *types.ApplyOrder) (AvailSubnetList, error) {
	subnetList := AvailSubnetList{}
	subnetReq := cvmapi.SubnetRealParam{
		Region:      order.Spec.Region,
		CloudCampus: zone,
		VpcId:       vpc,
	}
	resp, err := g.cvm.QueryRealCvmSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, subOrderID: %s, err: %v, region: %s, zone: %s, vpc: %s, "+
			"crpResp: %+v, rid: %s", order.SubOrderId, err, order.Spec.Region, zone, vpc, cvt.PtrToVal(resp), kt.Rid)
		return subnetList, err
	}
	// 记录crp返回的剩余IP日志
	crpRemainIPJSON, err := json.Marshal(resp.Result)
	if err != nil {
		logs.Warnf("get crp cvm subnet remainIP num json marshal failed, err: %+v, rid: %s", err, kt.Rid)
	}
	logs.Infof("get crp cvm subnet remainIP num success, subOrderID: %s, region: %s, zone: %s, vpc: %s, "+
		"crpRemainIP: %s, crpTraceID: %s, rid: %s", order.SubOrderId, order.Spec.Region, zone, vpc,
		crpRemainIPJSON, resp.TraceId, kt.Rid)

	cond := map[string]interface{}{
		"region": order.Spec.Region,
		"zone":   zone,
		"vpc_id": vpc,
	}
	// get subnet with enable flag only
	cond["enable"] = true
	cfgSubnets, err := g.configLogics.Subnet().GetSubnet(kt, cond)
	if err != nil {
		logs.Errorf("failed to get config cvm subnet info, subOrderID: %s, err: %v, region: %s, zone: %s, vpc: %s, "+
			"rid: %s", order.SubOrderId, err, order.Spec.Region, zone, vpc, kt.Rid)
		return subnetList, err
	}
	mapIdTosubnet := make(map[string]*cfgtypes.Subnet)
	for _, subnet := range cfgSubnets.Info {
		mapIdTosubnet[subnet.SubnetId] = subnet
	}

	for _, subnet := range resp.Result {
		// subnet is not effective
		if _, ok := mapIdTosubnet[subnet.Id]; !ok {
			continue
		}
		// use subnet name with prefix "cvm_use" only
		if !strings.HasPrefix(subnet.Name, "cvm_use") {
			continue
		}
		// return subnet with positive left ip
		if subnet.LeftIpNum > 0 {
			subnetList = append(subnetList, subnet)
		}
	}

	if subnetList.Len() == 0 {
		return subnetList, fmt.Errorf("found no available subnet with region:%s, zone:%s, vpc:%s, crpTraceID: %s",
			order.Spec.Region, zone, vpc, resp.TraceId)
	}

	return subnetList, nil
}

func (g *Generator) checkRecordCrpOrderTimeout(kt *kit.Kit, subOrderID string, crpResp *cvmapi.OrderQueryResp) {
	if crpResp == nil || crpResp.Result == nil || crpResp.Error.Code != 0 || len(crpResp.Result.Data) != 1 {
		return
	}

	createTime, err := time.Parse(constant.DateTimeLayout, crpResp.Result.Data[0].CreateTime)
	if err == nil && createTime.Add(types.OneDayDuration).Before(time.Now()) {
		logs.Warnf("%s: query crp cvm apply order timeout, subOrderID: %s, crpOrderID: %s, crpTraceID: %s, rid: %s",
			constant.CvmApplyOrderCrpProductTimeout, subOrderID, crpResp.Result.Data[0].OrderId,
			crpResp.TraceId, kt.Rid)
	}
}
