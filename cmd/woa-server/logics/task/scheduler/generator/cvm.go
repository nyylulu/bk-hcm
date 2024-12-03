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
	"strings"
	"time"

	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/utils"
)

// createCVM starts a cvm creating task
func (g *Generator) createCVM(cvm *types.CVM) (string, error) {
	// construct cvm launch request
	createReq := &cvmapi.OrderCreateReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmLaunchMethod,
		},
		Params: &cvmapi.OrderCreateParams{
			Zone:          cvm.Zone,
			DeptName:      cvmapi.CvmLaunchDeptName,
			ProductName:   cvm.BkProductName,
			Business1Id:   cvmapi.CvmLaunchBiz1Id,
			Business1Name: cvmapi.CvmLaunchBiz1Name,
			Business2Id:   cvmapi.CvmLaunchBiz2Id,
			Business2Name: cvmapi.CvmLaunchBiz2Name,
			Business3Id:   cvmapi.CvmLaunchBiz3Id,
			Business3Name: cvmapi.CvmLaunchBiz3Name,
			ProjectId:     int(cvm.BkProductID),
			Image: &cvmapi.Image{
				ImageId:   cvm.ImageId,
				ImageName: cvm.ImageName,
			},
			InstanceType: cvm.InstanceType,
			DataDisk:     make([]*cvmapi.DataDisk, 0),
			VpcId:        cvm.VPCId,
			SubnetId:     cvm.SubnetId,
			ApplyNum:     int(cvm.ApplyNumber),
			PassWord:     g.clientConf.CvmOpt.CvmLaunchPassword,
			Security: &cvmapi.Security{
				SecurityGroupId:   cvm.SecurityGroupId,
				SecurityGroupName: cvm.SecurityGroupName,
				SecurityGroupDesc: cvm.SecurityGroupDesc,
			},
			UseTime:           time.Now().Format(constant.DateTimeLayout),
			Memo:              cvm.NoteInfo,
			Operator:          cvm.Operator,
			BakOperator:       cvm.Operator,
			InheritInstanceId: cvm.InheritInstanceId,
		},
	}

	// 计费模式
	if len(cvm.ChargeType) > 0 {
		createReq.Params.ChargeType = cvm.ChargeType
		// 包年包月时才需要设置计费时长
		if createReq.Params.ChargeType == cvmapi.ChargeTypePrePaid {
			createReq.Params.ChargeMonths = cvm.ChargeMonths
		}
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

	jsonReq, _ := json.Marshal(createReq)
	logs.Infof("scheduler:logics:generator:create:cvm:start, create cvm req: %s", string(jsonReq))

	// call cvm api to launchCvm cvm order
	maxRetry := 3
	var err error = nil
	resp := new(cvmapi.OrderCreateResp)
	for try := 0; try < maxRetry; try++ {
		// need not wait for the first try
		if try != 0 {
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
		}

		resp, err = g.cvm.CreateCvmOrder(nil, nil, createReq)
		if err != nil {
			logs.Warnf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, req: %s, err: %v",
				string(jsonReq), err)
			continue
		}

		if resp.Error.Code != 0 {
			logs.Warnf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, code: %d, msg: %s",
				resp.Error.Code, resp.Error.Message)
			if g.needRetryCreateCvm(resp.Error.Code, resp.Error.Message) {
				continue
			}
		}

		break
	}

	if err != nil {
		logs.Errorf("scheduler:logics:generator:create:cvm:failed to create cvm launch order, req: %s, err: %v",
			string(jsonReq), err)
		return "", err
	}

	respStr := ""
	if b, err := json.Marshal(resp); err == nil {
		respStr = string(b)
	}
	logs.Infof("scheduler:logics:generator:create:cvm:success, create cvm req: %s, resp: %s", string(jsonReq), respStr)

	if resp.Error.Code != 0 {
		return "", fmt.Errorf("cvm order create task failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result.OrderId == "" {
		return "", fmt.Errorf("cvm order create task return empty order id")
	}

	return resp.Result.OrderId, nil
}

func (g *Generator) needRetryCreateCvm(code int, msg string) bool {
	// success
	if code == 0 {
		return false
	}

	// no capacity enough
	if code == -20004 && strings.Contains(msg, "无法满足本次需求量") {
		return false
	}

	if code == -20000 && strings.Contains(msg, "无法满足本次需求量") {
		return false
	}

	// sold out
	if code == -20004 && strings.Contains(msg, "已售罄，请更换可用区") {
		return false
	}

	return true
}

// checkCVM checks cvm creating task result
func (g *Generator) checkCVM(orderId string) error {
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

		status := enumor.CrpOrderStatus(resp.Result.Data[0].Status)
		if status != enumor.CrpOrderStatusFinish &&
			status != enumor.CrpOrderStatusReject &&
			status != enumor.CrpOrderStatusFailed {
			return false, fmt.Errorf("cvm order %s handling", orderId)
		}

		if status != enumor.CrpOrderStatusFinish {
			return true, fmt.Errorf("order %s failed, status: %d", resp.Result.Data[0].OrderId,
				resp.Result.Data[0].Status)
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
	_, err := utils.Retry(doFunc, checkFunc, 86400, 60)
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
func (g *Generator) buildCvmReq(kt *kit.Kit, order *types.ApplyOrder, zone string, replicas uint) (*types.CVM, error) {
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
		vpc, err := g.getCvmVpc(order.Spec.Region)
		if err != nil {
			return nil, err
		}
		req.VPCId = vpc
	}
	if order.Spec.Subnet != "" {
		req.SubnetId = order.Spec.Subnet
	} else {
		subnetList, err := g.getCvmSubnet(kt, order.Spec.Region, zone, req.VPCId)
		if err != nil {
			logs.Errorf("failed to get available subnet, err: %v", err)
			return nil, err
		}
		sort.Sort(sort.Reverse(subnetList))
		subnetID := ""
		applyNum := uint(0)
		ignorePrediction := false
		if order.RequireType == enumor.RequireTypeGreenChannel {
			ignorePrediction = true
		}

		for _, subnet := range subnetList {
			capacity, err := g.getCapacity(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region, zone,
				req.VPCId, subnet.Id, order.Spec.ChargeType, ignorePrediction)
			if err != nil {
				logs.Errorf("failed to get capacity with subnet %s, subnetNum: %d, zone: %s, reqVpcID: %s, err: %v",
					subnet.Id, len(subnetList), zone, req.VPCId, err)
				continue
			}
			maxNum, ok := capacity[zone]
			if !ok {
				logs.Warnf("get no capacity with zone %s and subnet %s, err: %v", zone, subnet.Id)
				continue
			}
			if maxNum > 0 {
				subnetID = subnet.Id
				applyNum = uint(maxNum)
				break
			}
			// 记录日志，方便排查线上资源申请问题
			logs.Errorf("buildCvmReq:get no available capacity info, subnetNum: %d, zone: %s, reqVpcID: %s, "+
				"subnet: %+v, orderInfo: %+v, capacity: %+v",
				len(subnetList), zone, req.VPCId, cvt.PtrToVal(subnet), cvt.PtrToVal(order), capacity)
		}

		if subnetID == "" || applyNum <= 0 {
			// get capacity detail as component of error message
			capInfo, _ := g.getCapacityDetail(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region, zone,
				req.VPCId, "", order.Spec.ChargeType)
			capInfoStr, _ := json.Marshal(capInfo)
			// 记录日志，方便排查线上资源申请问题
			logs.Errorf("buildCvmReq:get empty subnet info failed, subnetNum: %d, applyNum: %d, zone: %s, "+
				"reqVpcID: %s, subnetID: %s, orderInfo: %+v, capInfoStr: %s",
				len(subnetList), applyNum, zone, req.VPCId, subnetID, cvt.PtrToVal(order), capInfoStr)
			return nil, fmt.Errorf("no capacity: %s", capInfoStr)
		}
		req.SubnetId = subnetID
		if applyNum < replicas {
			// set apply number to min(replicas, leftIp)
			req.ApplyNumber = applyNum
		}
		// 记录日志，方便排查线上资源申请问题
		logs.Infof("buildCvmReq:get subnet info success, subnetNum: %d, applyNum: %d, replicas: %d, zone: %s, "+
			"reqVpcID: %s, subnetID: %s, orderInfo: %+v, subnetList: %+v, req: %+v", len(subnetList), applyNum,
			replicas, zone, req.VPCId, subnetID, cvt.PtrToVal(order), subnetList, cvt.PtrToVal(req))
	}

	// image
	req.ImageId = order.Spec.ImageId
	req.ImageName = order.Spec.Image
	// security group
	sg, err := g.getCvmDftSecGroup(order.Spec.Region)
	if err != nil {
		return nil, err
	}
	req.SecurityGroupId = sg.SecurityGroupId
	req.SecurityGroupName = sg.SecurityGroupName
	req.SecurityGroupDesc = sg.SecurityGroupDesc

	productID, productName, err := g.getProductMsg(kt, order)
	if err != nil {
		logs.Errorf("get product message failed, err: %v, order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}

	req.BkProductID = productID
	req.BkProductName = productName

	return req, nil
}

func (g *Generator) getProductMsg(kt *kit.Kit, order *types.ApplyOrder) (int64, string, error) {
	if order.RequireType == enumor.RequireTypeRollServer {
		return cvmapi.CvmLaunchProjectId, cvmapi.CvmLaunchProductName, nil
	}

	param := &cmdb.SearchBizBelongingParams{BizIDs: []int64{order.BkBizId}}
	resp, err := g.cc.SearchBizBelonging(kt, param)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, param: %+v, rid: %s", err, *param, kt.Rid)
		return 0, "", err
	}
	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return 0, "", errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	bizBelong := (*resp)[0]
	return bizBelong.OpProductID, bizBelong.OpProductName, nil
}

var regionToVpc = map[string]string{
	"ap-guangzhou":     "vpc-03nkx9tv",
	"ap-tianjin":       "vpc-1yoew5gc",
	"ap-shanghai":      "vpc-2x7lhtse",
	"eu-frankfurt":     "vpc-38klpz7z",
	"ap-singapore":     "vpc-706wf55j",
	"ap-tokyo":         "vpc-8iple1iq",
	"ap-seoul":         "vpc-99wg8fre",
	"ap-hongkong":      "vpc-b5okec48",
	"na-toronto":       "vpc-drefwt2v",
	"ap-xian-ec":       "vpc-efw4kf6r",
	"ap-nanjing":       "vpc-fb7sybzv",
	"ap-chongqing":     "vpc-gelpqsur",
	"ap-shenzhen":      "vpc-kwgem8tj",
	"na-siliconvalley": "vpc-n040n5bl",
	"ap-hangzhou-ec":   "vpc-puhasca0",
	"ap-fuzhou-ec":     "vpc-hdxonj2q",
	"ap-wuhan-ec":      "vpc-867lsj6w",
	"ap-beijing":       "vpc-bhb0y6g8",
}

func (g *Generator) getCvmVpc(region string) (string, error) {
	vpc, ok := regionToVpc[region]
	if !ok {
		return "", fmt.Errorf("found no vpc with region %s", region)
	}

	return vpc, nil
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

func (g *Generator) getCvmSubnet(kt *kit.Kit, region, zone, vpc string) (AvailSubnetList, error) {
	subnetList := AvailSubnetList{}
	req := &cvmapi.SubnetReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmSubnetMethod,
		},
		Params: &cvmapi.SubnetParam{
			DeptId: cvmapi.CvmDeptId,
			Region: region,
			Zone:   zone,
			VpcId:  vpc,
		},
	}

	resp, err := g.cvm.QueryCvmSubnet(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v", err)
		return subnetList, err
	}

	cond := map[string]interface{}{
		"region": region,
		"zone":   zone,
		"vpc_id": vpc,
	}
	// get subnet with enable flag only
	cond["enable"] = true
	cfgSubnets, err := g.configLogics.Subnet().GetSubnet(kt, cond)
	if err != nil {
		logs.Errorf("failed to get config cvm subnet info, err: %v", err)
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
		return subnetList, fmt.Errorf("found no available subnet with region %s, zone %s, vpc %s", region, zone, vpc)
	}

	return subnetList, nil
}

// SecGroup network security group
type SecGroup struct {
	SecurityGroupId   string `json:"securityGroupId"`
	SecurityGroupName string `json:"securityGroupName"`
	SecurityGroupDesc string `json:"securityGroupDesc"`
}

var regionToSecGroup = map[string]*SecGroup{
	"ap-guangzhou": {
		SecurityGroupId:   "sg-ka67ywe9",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"ap-tianjin": {
		SecurityGroupId:   "sg-c28492qp",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shanghai": {
		SecurityGroupId:   "sg-ibqae0te",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"eu-frankfurt": {
		SecurityGroupId:   "sg-cet13de0",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-singapore": {
		SecurityGroupId:   "sg-hjtqedoe",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-tokyo": {
		SecurityGroupId:   "sg-o1lfldnk",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-seoul": {
		SecurityGroupId:   "sg-i7h8hv5r",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-hongkong": {
		SecurityGroupId:   "sg-59kfufmn",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"na-toronto": {
		SecurityGroupId:   "sg-7l82d7km",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-xian-ec": {
		SecurityGroupId:   "sg-o4bmz4kg",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-nanjing": {
		SecurityGroupId:   "sg-dybs7i3y",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"ap-chongqing": {
		SecurityGroupId:   "sg-l5usnzxw",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shenzhen": {
		SecurityGroupId:   "sg-qkfewp0u",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"na-siliconvalley": {
		SecurityGroupId:   "sg-q7usygae",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-hangzhou-ec": {
		SecurityGroupId:   "sg-4ezyvbvl",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-fuzhou-ec": {
		SecurityGroupId:   "sg-leqa6w29",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-wuhan-ec": {
		SecurityGroupId:   "sg-p5ld4xyq",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-beijing": {
		SecurityGroupId:   "sg-rjwj7cnt",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
}

func (g *Generator) getCvmDftSecGroup(region string) (*SecGroup, error) {
	sg, ok := regionToSecGroup[region]
	if !ok {
		return nil, fmt.Errorf("found no security group with region %s", region)
	}

	return sg, nil
}
