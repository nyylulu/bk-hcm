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

package cvm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	model "hcm/cmd/woa-server/model/cvm"
	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/utils"
)

// CVM create cvm request param
type CVM struct {
	AppId             string            `json:"appId"`
	ApplyType         int64             `json:"applyType"`
	AppModuleId       int64             `json:"appModuleId"`
	Operator          string            `json:"operator"`
	ApplyNumber       uint              `json:"applyNumber"`
	NoteInfo          string            `json:"noteInfo"`
	VPCId             string            `json:"vpcId"`
	SubnetId          string            `json:"subnetId"`
	Area              string            `json:"area"`
	Zone              string            `json:"zone"`
	ImageId           string            `json:"image_id"`
	ImageName         string            `json:"image_name"`
	InstanceType      string            `json:"instanceType"`
	DiskType          string            `json:"disk_type"`
	DiskSize          int64             `json:"disk_size"`
	SecurityGroupId   string            `json:"securityGroupId"`
	SecurityGroupName string            `json:"securityGroupName"`
	SecurityGroupDesc string            `json:"securityGroupDesc"`
	ChargeType        cvmapi.ChargeType `json:"chargeType"`
	ChargeMonths      uint              `json:"chargeMonths"`
	InheritInstanceId string            `json:"inherit_instance_id"`
	BkProductID       int64             `json:"bk_product_id"`
	BkProductName     string            `json:"bk_product_name"`
}

// executeApplyOrder CVM生产-创建单据
func (l *logics) executeApplyOrder(kt *kit.Kit, order *types.ApplyOrder) {
	// 0. update generate record status to running
	if err := l.updateApplyOrder(order, types.ApplyStatusRunning, "handling", "", 0); err != nil {
		logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
			order.OrderId, err, kt.Rid)
		return
	}

	// 1. launch cvm request
	request, err := l.buildCvmReq(kt, order)
	if err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to launch cvm when build cvm request, "+
			"err: %v, order id: %d, rid: %s", err, order.OrderId, kt.Rid)

		// update generate record status to failed
		if err = l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); err != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
				order.OrderId, err, kt.Rid)
			return
		}

		return
	}

	// 2. launch cvm request
	taskId, err := l.createCVM(request)
	if err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to create cvm when launch generate task, "+
			"order id: %d, err: %v, rid: %s", order.OrderId, err, kt.Rid)

		// update generate record status to failed
		if err = l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); err != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
				order.OrderId, err, kt.Rid)
			return
		}

		return
	}

	// 3. update generate record status to running
	if err = l.updateApplyOrder(order, types.ApplyStatusRunning, "handling", taskId, 0); err != nil {
		logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
			order.OrderId, taskId, err, kt.Rid)
		return
	}

	// 4. check cvm task result
	if err = l.checkCVM(taskId); err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to create cvm when check generate task, "+
			"order id: %s, task id: %s, err: %v, rid: %s", order.OrderId, taskId, err, kt.Rid)

		// update generate record status to failed
		if err = l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); err != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
				order.OrderId, taskId, err, kt.Rid)
			return
		}

		return
	}

	// 5. get generated cvm instances
	hosts, err := l.listCVM(taskId)
	if err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to list created cvm, order id: %s, "+
			"task id: %s, err: %v, rid: %s", order.OrderId, taskId, err, kt.Rid)

		// update generate record status to failed
		if err = l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); err != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
				order.OrderId, taskId, err, kt.Rid)
			return
		}

		return
	}

	// 6. save generated cvm instances info
	if err = l.createDeviceInfo(order, hosts, taskId); err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to update generated device, "+
			"order id: %s, taskId: %s, err: %v, rid: %s", order.OrderId, taskId, err, kt.Rid)

		// update generate record status to failed
		if err = l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); err != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
				order.OrderId, taskId, err, kt.Rid)
			return
		}

		return
	}

	// 7. update generate record status to success
	if err = l.updateApplyOrder(order, types.ApplyStatusSuccess, "success", "", uint(len(hosts))); err != nil {
		logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
			order.OrderId, taskId, err, kt.Rid)
		return
	}

	return
}

// createCVM starts a cvm creating task(CVM生产-创建单据)
func (l *logics) createCVM(cvm *CVM) (string, error) {
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
			// Business3Id:   cvmapi.CvmLaunchBiz3Id,
			// Business3Name: cvmapi.CvmLaunchBiz3Name,
			Business3Id:   662584,
			Business3Name: "CC_SA云化池",
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
			PassWord:     l.cliConf.CvmOpt.CvmLaunchPassword,
			Security: &cvmapi.Security{
				SecurityGroupId:   cvm.SecurityGroupId,
				SecurityGroupName: cvm.SecurityGroupName,
				SecurityGroupDesc: cvm.SecurityGroupDesc,
			},
			UseTime:           time.Now().Format(constant.DateTimeLayout),
			Memo:              cvm.NoteInfo,
			Operator:          cvm.Operator,
			BakOperator:       cvm.Operator,
			ChargeType:        cvm.ChargeType,        // 计费模式
			InheritInstanceId: cvm.InheritInstanceId, // 被继承云主机的实例id
		},
	}

	// 计费模式-包年包月时才需要设置计费时长
	if cvm.ChargeType == cvmapi.ChargeTypePrePaid {
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

	if cvm.DiskSize > 0 {
		createReq.Params.DataDisk = append(createReq.Params.DataDisk, &cvmapi.DataDisk{
			DataDiskType: cvm.DiskType,
			DataDiskSize: int(cvm.DiskSize),
		})
	}

	// set obs project type
	createReq.Params.ObsProject = cvmapi.GetObsProject(cvm.ApplyType)

	// call cvm api to launchCvm cvm order
	createResp, err := l.cvm.CreateCvmOrder(nil, nil, createReq)
	if err != nil {
		logs.Errorf("scheduler:logics:create:cvm:failed, err: %v, req: %+v", err, createReq)
		return "", err
	}

	// 记录cvm创建成功的日志，方便排查问题
	logs.Infof("scheduler:logics:create:cvm:success, req: %+v, resp: %+v",
		cvt.PtrToVal(createReq), cvt.PtrToVal(createResp))

	if createResp.Error.Code != 0 {
		return "", fmt.Errorf("cvm order create task failed, code: %d, msg: %s", createResp.Error.Code,
			createResp.Error.Message)
	}

	if createResp.Result.OrderId == "" {
		return "", fmt.Errorf("cvm order create task return empty order id")
	}

	return createResp.Result.OrderId, nil
}

// checkCVM checks cvm creating task result
func (l *logics) checkCVM(orderId string) error {
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

		if resp.Result.Data[0].Status != cvmapi.OrderStatusFinish &&
			resp.Result.Data[0].Status != cvmapi.OrderStatusReject &&
			resp.Result.Data[0].Status != cvmapi.OrderStatusFailed {
			return false, fmt.Errorf("cvm order %s handling", orderId)
		}

		if resp.Result.Data[0].Status != cvmapi.OrderStatusFinish {
			return true, fmt.Errorf("order %s failed, status: %d", resp.Result.Data[0].OrderId,
				resp.Result.Data[0].Status)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := &cvmapi.OrderQueryReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.CvmOrderStatusMethod,
			},
			Params: &cvmapi.OrderQueryParam{
				OrderId: []string{orderId},
			},
		}

		// call cvm api to query cvm order status
		return l.cvm.QueryCvmOrders(nil, nil, req)
	}

	// TODO: get retry strategy from config
	_, err := utils.Retry(doFunc, checkFunc, 86400, 60)
	return err
}

// listCVM lists created cvm by order id
func (l *logics) listCVM(orderId string) ([]*cvmapi.InstanceItem, error) {
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
		return l.cvm.QueryCvmInstances(nil, nil, req)
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

	logs.Infof("scheduler:logics:cvm:list:success, orderId: %s, get cvm instance resp: %+v", orderId, resp)

	if resp.Error.Code != 0 {
		return nil, fmt.Errorf("list cvm instance failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		return nil, errors.New("list cvm instance failed, for result is null")
	}

	return resp.Result.Data, nil
}

// createDeviceInfo update generate record
func (l *logics) createDeviceInfo(order *types.ApplyOrder, items []*cvmapi.InstanceItem, taskId string) error {
	// 1. save device info to db
	now := time.Now()
	for _, item := range items {
		device := &types.CvmInfo{
			OrderId:   order.OrderId,
			CvmTaskId: taskId,
			CvmInstId: item.InstanceId,
			Ip:        item.LanIp,
			AssetId:   item.AssetId,
			UpdateAt:  now,
		}
		if err := model.Operation().CvmInfo().CreateCvmInfo(context.Background(), device); err != nil {
			logs.Errorf("failed to save device info to db, order id: %d, err: %v", order.OrderId, err)
			return err
		}
	}

	return nil
}

// updateApplyOrder updates generate record
func (l *logics) updateApplyOrder(order *types.ApplyOrder, status types.ApplyStatus, msg, vmTaskId string,
	succNum uint) error {

	filter := &mapstr.MapStr{
		"order_id": order.OrderId,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"update_at": now,
	}

	if len(msg) != 0 {
		doc["message"] = msg
	}

	if len(vmTaskId) != 0 {
		link := cvmapi.CvmOrderLinkPrefix + vmTaskId
		doc["task_id"] = vmTaskId
		doc["task_link"] = link
	}

	if status == types.ApplyStatusSuccess || status == types.ApplyStatusFailed {
		doc["success_num"] = succNum
		doc["pending_num"] = 0
		doc["failed_num"] = order.Total - succNum
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update apply order, order id: %d, update: %+v, err: %v", order.OrderId, doc, err)
		return err
	}

	return nil
}

// buildCvmReq construct a cvm creating task request(CVM生产-创建单据)
func (l *logics) buildCvmReq(kt *kit.Kit, order *types.ApplyOrder) (*CVM, error) {
	kt = kt.NewSubKit()
	kt.Ctx = context.Background()
	// 记录cvm请求日志，方便排查问题
	logs.Infof("scheduler:logics:build:cvm:request:start, order: %+v, rid: %s", cvt.PtrToVal(order), kt.Rid)
	// TODO: get parameters from config
	// construct cvm launch req
	req := &CVM{
		AppId:             "931",
		ApplyType:         order.RequireType,
		AppModuleId:       51524,
		Operator:          order.User,
		ApplyNumber:       order.Total,
		NoteInfo:          order.Remark,
		Area:              order.Spec.Region,
		Zone:              order.Spec.Zone,
		InstanceType:      order.Spec.DeviceType,
		DiskType:          order.Spec.DiskType,
		DiskSize:          order.Spec.DiskSize,
		ChargeType:        order.Spec.ChargeType, // 计费模式
		InheritInstanceId: order.Spec.InheritInstanceId,
	}

	// 计费模式-包年包月时才需要设置计费时长
	if req.ChargeType == cvmapi.ChargeTypePrePaid {
		req.ChargeMonths = order.Spec.ChargeMonths
	}

	// set disk type default value
	if len(req.DiskType) == 0 {
		req.DiskType = cvmapi.CvmLaunchSystemDiskTypePremium
	}
	// vpc and subnet
	if order.Spec.Vpc != "" {
		req.VPCId = order.Spec.Vpc
	} else {
		vpc, err := l.getCvmVpc(order.Spec.Region)
		if err != nil {
			logs.Errorf("scheduler:logics:build:cvm:request:failed, build cvm req get cvm vpc failed, err: %v, "+
				"order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
			return nil, err
		}
		req.VPCId = vpc
	}
	if order.Spec.Subnet != "" {
		req.SubnetId = order.Spec.Subnet
	} else {
		subnet, leftIp, err := l.getCvmSubnet(kt, order.Spec.Region, order.Spec.Zone, req.VPCId)
		if err != nil {
			logs.Errorf("scheduler:logics:build:cvm:request:failed, build cvm req get subnet failed, err: %v, "+
				"req: %+v, order: %+v, rid: %s", err, req, cvt.PtrToVal(order), kt.Rid)
			return nil, err
		}
		req.SubnetId = subnet
		if leftIp < order.Total {
			// set apply number to min(replicas, leftIp)
			req.ApplyNumber = leftIp
		}
	}
	// image
	req.ImageId = order.Spec.ImageId
	// security group
	sg, err := l.getCvmDftSecGroup(order.Spec.Region)
	if err != nil {
		logs.Errorf("scheduler:logics:build:cvm:request:failed, build cvm req get cvm drg secGroup failed, "+
			"err: %v, order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}
	req.SecurityGroupId = sg.SecurityGroupId
	req.SecurityGroupName = sg.SecurityGroupName
	req.SecurityGroupDesc = sg.SecurityGroupDesc

	productID, productName, err := l.getProductMsg(kt, order)
	if err != nil {
		logs.Errorf("get product message failed, err: %v, order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}

	req.BkProductID = productID
	req.BkProductName = productName

	logs.Infof("scheduler:logics:build:cvm:request:end, order: %+v, req: %+v, rid: %s",
		cvt.PtrToVal(order), cvt.PtrToVal(req), kt.Rid)
	return req, nil
}

func (l *logics) getProductMsg(kt *kit.Kit, order *types.ApplyOrder) (int64, string, error) {
	if types.RequireType(order.RequireType) == types.RollingServer {
		return cvmapi.CvmLaunchProjectId, cvmapi.CvmLaunchProductName, nil
	}

	param := &cmdb.SearchBizBelongingParams{BizIDs: []int64{order.BkBizId}}
	resp, err := l.esbClient.Cmdb().SearchBizBelonging(kt, param)
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

func (l *logics) getCvmVpc(region string) (string, error) {
	vpc, ok := regionToVpc[region]
	if !ok {
		return "", fmt.Errorf("found no vpc with region %s", region)
	}

	return vpc, nil
}

func (l *logics) getCvmSubnet(kt *kit.Kit, region, zone, vpc string) (string, uint, error) {
	req := &cvmapi.SubnetReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmSubnetMethod,
		},
		Params: &cvmapi.SubnetParam{
			DeptId: cvmapi.CvmDeptId,
			Region: region,
			VpcId:  vpc,
		},
	}
	// 园区-分区Campus
	if len(zone) > 0 && zone != cvmapi.CvmSeparateCampus {
		req.Params.Zone = zone
	}

	resp, err := l.cvm.QueryCvmSubnet(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v, region: %s, zone: %s, vpc: %s, rid: %s",
			err, region, zone, vpc, kt.Rid)
		return "", 0, err
	}

	cond := map[string]interface{}{
		"region": region,
		"vpc_id": vpc,
		// get subnet with enable flag only
		"enable": true,
	}
	// 园区-分区Campus
	if len(zone) > 0 && zone != cvmapi.CvmSeparateCampus {
		cond["zone"] = zone
	}
	cfgSubnets, err := l.confLogic.Subnet().GetSubnet(kt, cond)
	if err != nil {
		logs.Errorf("failed to get config cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return "", 0, err
	}
	mapIdTosubnet := make(map[string]*cfgtypes.Subnet)
	for _, subnet := range cfgSubnets.Info {
		mapIdTosubnet[subnet.SubnetId] = subnet
	}

	subnetId := ""
	leftIp := uint(0)
	for _, subnet := range resp.Result {
		// subnet is not effective
		if _, ok := mapIdTosubnet[subnet.Id]; !ok {
			continue
		}
		// use subnet name with prefix "cvm_use" only
		if !strings.HasPrefix(subnet.Name, "cvm_use") {
			continue
		}
		if subnet.LeftIpNum >= int(leftIp) {
			subnetId = subnet.Id
			leftIp = uint(subnet.LeftIpNum)
		}
	}

	if subnetId == "" {
		logs.Errorf("getCvmSubnet found no subnet with region: %s, zone: %s, vpc: %s", region, zone, vpc,
			cvt.PtrToSlice(cfgSubnets.Info), cvt.PtrToSlice(resp.Result))
		return "", 0, fmt.Errorf("found no subnet with region: %s, zone: %s, vpc: %s", region, zone, vpc)
	}

	if leftIp <= 0 {
		logs.Errorf("getCvmSubnet found no subnet with left ip > 0, region: %s, zone: %s, vpc: %s", region, zone, vpc,
			cvt.PtrToSlice(cfgSubnets.Info), cvt.PtrToSlice(resp.Result))
		return subnetId, leftIp, fmt.Errorf("found no subnet with left ip > 0, region: %s, zone: %s, vpc: %s", region,
			zone, vpc)
	}

	return subnetId, leftIp, nil
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

func (l *logics) getCvmDftSecGroup(region string) (*SecGroup, error) {
	sg, ok := regionToSecGroup[region]
	if !ok {
		return nil, fmt.Errorf("found no security group with region %s", region)
	}

	return sg, nil
}
