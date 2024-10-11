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
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/logics/config"
	model "hcm/cmd/woa-server/model/cvm"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/thirdparty/esb"
	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// Logics provides management interface for operations of model and instance and related resources like association
type Logics interface {
	// CreateApplyOrder creates cvm apply order
	CreateApplyOrder(kt *kit.Kit, param *types.CvmCreateReq) (*types.CvmCreateResult, error)
	// GetApplyOrderById get cvm apply order info
	GetApplyOrderById(kt *kit.Kit, param *types.CvmOrderReq) (*types.CvmOrderResult, error)
	// GetApplyOrder get cvm apply order info
	GetApplyOrder(kt *kit.Kit, param *types.GetApplyParam) (*types.CvmOrderResult, error)
	// GetApplyDevice get cvm apply order launched instances
	GetApplyDevice(kt *kit.Kit, param *types.CvmDeviceReq) (*types.CvmDeviceResult, error)
	// GetCapacity get cvm apply capacity
	GetCapacity(kt *kit.Kit, param *types.CvmCapacityReq) (*types.CvmCapacityResult, error)
}

type logics struct {
	cvm       cvmapi.CVMClientInterface
	cliConf   cc.ClientConfig
	confLogic config.Logics
	esbClient esb.Client
}

// New create a logics manager
func New(thirdCli *thirdparty.Client, cliConf cc.ClientConfig, confLogic config.Logics, esbClient esb.Client) Logics {
	return &logics{
		cvm:       thirdCli.CVM,
		confLogic: confLogic,
		cliConf:   cliConf,
		esbClient: esbClient,
	}
}

// CreateApplyOrder creates cvm apply order(CVM生产-创建单据)
func (l *logics) CreateApplyOrder(kt *kit.Kit, param *types.CvmCreateReq) (*types.CvmCreateResult, error) {
	id, err := model.Operation().ApplyOrder().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create cvm apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	now := time.Now()
	order := &types.ApplyOrder{
		OrderId:     id,
		BkBizId:     param.BkBizId,
		BkModuleId:  param.BkModuleId,
		User:        param.User,
		RequireType: param.RequireType,
		Remark:      param.Remark,
		Spec:        param.Spec,
		Status:      types.ApplyStatusInit,
		Message:     "",
		TaskId:      "",
		TaskLink:    "",
		Total:       param.Replicas,
		SuccessNum:  0,
		FailedNum:   0,
		PendingNum:  param.Replicas,
		CreateAt:    now,
		UpdateAt:    now,
	}

	if err = model.Operation().ApplyOrder().CreateApplyOrder(kt.Ctx, order); err != nil {
		logs.Errorf("failed to create cvm apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	logs.Infof("scheduler:logics:cvm:create:apply:order:init, orderId: %s, param: %+v, rid: %s",
		id, cvt.PtrToVal(param), kt.Rid)

	// execute cvm apply order
	go l.executeApplyOrder(kt, order)

	rst := &types.CvmCreateResult{
		OrderId: id,
	}

	return rst, nil
}

// GetApplyOrderById get cvm apply order info by order id
func (l *logics) GetApplyOrderById(kt *kit.Kit, param *types.CvmOrderReq) (*types.CvmOrderResult, error) {
	filter := map[string]interface{}{
		"order_id": param.OrderId,
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.CvmOrderResult{
		Info: insts,
	}

	return rst, nil
}

// GetApplyOrder get cvm apply order info
func (l *logics) GetApplyOrder(kt *kit.Kit, param *types.GetApplyParam) (*types.CvmOrderResult, error) {
	filter := param.GetFilter()

	rst := &types.CvmOrderResult{}
	if param.Page.EnableCount {
		cnt, err := model.Operation().ApplyOrder().CountApplyOrder(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get apply order count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*types.ApplyOrder, 0)
		return rst, nil
	}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetApplyDevice get cvm apply order launched instances
func (l *logics) GetApplyDevice(kt *kit.Kit, param *types.CvmDeviceReq) (*types.CvmDeviceResult, error) {
	filter := &mapstr.MapStr{
		"order_id": param.OrderId,
	}

	insts, err := model.Operation().CvmInfo().GetCvmInfo(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.CvmDeviceResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetCapacity get cvm apply capacity
func (l *logics) GetCapacity(kt *kit.Kit, param *types.CvmCapacityReq) (*types.CvmCapacityResult, error) {
	req := &cvmapi.CapacityReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmLaunchMethod,
		},
		Params: &cvmapi.CapacityParam{
			DeptId:       cvmapi.CvmDeptId,
			Business3Id:  int(param.BkBizId),
			CloudCampus:  param.Zone,
			InstanceType: param.DeviceType,
			VpcId:        param.VpcId,
			SubnetId:     param.SubnetId,
		},
	}

	// set project name
	req.Params.ProjectName = cvmapi.GetObsProject(param.RequireType)

	resp, err := l.cvm.QueryCvmCapacity(nil, nil, req)
	if err != nil {
		logs.Errorf("scheduler:logics:cvm:capacity:failed, failed to get cvm apply capacity, err: %v, rid: %s", err,
			kt.Rid)
		return nil, err
	}

	// TODO: support return multiple vpc and subnet capacity
	rst := &types.CvmCapacityResult{
		Count: 1,
		Info:  make([]*types.CapacityItem, 0),
	}

	capacityItem := &types.CapacityItem{
		Region:   param.Region,
		Zone:     param.Zone,
		VpcId:    param.VpcId,
		SubnetId: param.SubnetId,
		MaxNum:   resp.Result.MaxNum,
		MaxInfo:  make([]*types.CapacityInfo, 0),
	}

	for _, info := range resp.Result.MaxInfo {
		capacityItem.MaxInfo = append(capacityItem.MaxInfo, &types.CapacityInfo{
			Key:   info.Key,
			Value: info.Value,
		})
	}

	rst.Info = append(rst.Info, capacityItem)
	return rst, nil
}
