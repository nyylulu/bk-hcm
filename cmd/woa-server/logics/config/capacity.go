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

// Package config config
package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	arrayutil "hcm/pkg/tools/util"
)

// CapacityIf provides management interface for operations of resource apply capacity
type CapacityIf interface {
	// GetCapacity gets resource apply capacity info
	GetCapacity(kt *kit.Kit, input *types.GetCapacityParam) (*types.GetCapacityRst, error)
	// UpdateCapacity updates resource apply capacity info
	UpdateCapacity(kt *kit.Kit, input *types.UpdateCapacityParam) error
}

// NewCapacityOp creates a capacity interface
func NewCapacityOp(thirdCli *thirdparty.Client) CapacityIf {
	return &capacity{
		cvm: thirdCli.OldCVM,
	}
}

type capacity struct {
	cvm cvmapi.CVMClientInterface
}

// GetCapacity gets resource apply capacity info
func (c *capacity) GetCapacity(kt *kit.Kit, input *types.GetCapacityParam) (*types.GetCapacityRst, error) {
	// 1. query subnet from db
	filter := map[string]interface{}{
		"region": input.Region,
	}
	if input.Zone != "" && input.Zone != cvmapi.CvmSeparateCampus {
		filter["zone"] = input.Zone
	}
	vpcID := input.Vpc
	if vpcID == "" {
		dftVpc, err := GetDftCvmVpc(input.Region)
		if err != nil {
			return nil, err
		}
		vpcID = dftVpc
	}
	filter["vpc_id"] = vpcID

	if input.Subnet != "" {
		filter["subnet_id"] = input.Subnet
	} else {
		if IsDftCvmVpc(vpcID) {
			// filter subnet with name prefix cvm_use_
			filter["subnet_name"] = mapstr.MapStr{
				pkg.BKDBLIKE: "^cvm_use_",
			}
		}
	}

	// get subnet with enable flag only
	filter["enable"] = true

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	subnetList, err := config.Operation().Subnet().FindManySubnet(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to find subnet with filter: %+v, err: %v, rid: %s", filter, err, kt.Rid)
		return nil, err
	}

	zoneToVpc := make(map[string][]string)
	vpcToSubnet := make(map[string][]string)

	for _, subnetItem := range subnetList {
		zoneToVpc[subnetItem.Zone] = append(zoneToVpc[subnetItem.Zone], subnetItem.VpcId)
		vpcToSubnet[subnetItem.VpcId] = append(vpcToSubnet[subnetItem.VpcId], subnetItem.SubnetId)
	}

	// 2. query apply capacity
	zoneToCapacity := make(map[string]*types.CapacityInfo)
	for zoneID, vpcList := range zoneToVpc {
		vpcUniq := arrayutil.StrArrayUnique(vpcList)
		capa := c.getZoneCapacity(kt, input, zoneID, vpcUniq, vpcToSubnet, input.IgnorePrediction)
		if capa != nil {
			zoneToCapacity[zoneID] = capa
		}
	}

	rst := &types.GetCapacityRst{}
	for _, capInfo := range zoneToCapacity {
		rst.Info = append(rst.Info, capInfo)
	}
	rst.Count = int64(len(rst.Info))

	// 为方便排查问题，增加日志记录
	jsonRst, err := json.Marshal(rst)
	if err != nil {
		logs.Errorf("cvm apply order get capacity failed to marshal capacityRst, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	logs.Infof("cvm apply order get capacity, input: %+v, zoneInfo: %s, rid: %s",
		cvt.PtrToVal(input), string(jsonRst), kt.Rid)

	return rst, nil
}

// UpdateCapacity updates resource apply capacity info
func (c *capacity) UpdateCapacity(kt *kit.Kit, input *types.UpdateCapacityParam) error {
	// 1. get capacity
	param := &types.GetCapacityParam{
		RequireType: input.RequireType,
		DeviceType:  input.DeviceType,
		Region:      input.Region,
		Zone:        input.Zone,
	}

	rst, err := c.GetCapacity(kt, param)
	if err != nil {
		logs.Errorf("failed to get capacity, err: %v, %s", err, kt.Rid)
		return err
	}

	count := len(rst.Info)
	if count != 1 {
		logs.Errorf("get invalid capacity info num %d not equal 1, input: %+v, rid: %s",
			count, cvt.PtrToVal(input), kt.Rid)
		return fmt.Errorf("get invalid capacity info num %d not equal 1", count)
	}

	if rst.Info[0] == nil {
		logs.Errorf("get invalid null capacity info, rid: %s", kt.Rid)
		return errors.New("get invalid null capacity info")
	}

	maxNum := rst.Info[0].MaxNum

	// 2. calculate capacity flag
	flag := c.getCapacityFlag(int(maxNum))

	// 3. update capacity info in db
	filter := map[string]interface{}{
		"require_type": input.RequireType,
		"region":       input.Region,
		"zone":         input.Zone,
		"device_type":  input.DeviceType,
	}

	update := map[string]interface{}{
		"capacity_flag": flag,
	}

	if err = config.Operation().CvmDevice().UpdateDevice(kt.Ctx, filter, update); err != nil {
		logs.Errorf("failed to update capacity info in db, err: %v, flag: %d, input: %+v, rid: %s",
			err, flag, cvt.PtrToVal(input), kt.Rid)
		return err
	}
	// 记录日志方便排查问题
	logs.Errorf("update device capacity success, maxNum: %d, flag: %d, input: %+v, crpResp: %+v, rid: %s",
		maxNum, flag, cvt.PtrToVal(input), cvt.PtrToSlice(rst.Info), kt.Rid)

	return nil
}

func (c *capacity) getZoneCapacity(kt *kit.Kit, input *types.GetCapacityParam, zone string, vpcList []string,
	vpcToSubnet map[string][]string, ignorePrediction bool) *types.CapacityInfo {

	// 1. query cvm capacity
	if len(vpcList) == 0 {
		capacityInfo := &types.CapacityInfo{
			Region:  input.Region,
			Zone:    zone,
			Vpc:     "",
			Subnet:  "",
			MaxNum:  0,
			MaxInfo: make([]*types.CapacityMaxInfo, 0),
		}
		return capacityInfo
	}

	req := c.createCapacityReq(input, zone, vpcList, vpcToSubnet)
	resp, err := c.cvm.QueryCvmCapacity(nil, nil, req)
	if err != nil {
		logs.ErrorJson("failed to get cvm apply capacity, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get cvm apply capacity, code: %d, msg: %s, crpTraceID: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId, kt.Rid)
		return nil
	}

	if resp.Result == nil {
		logs.Errorf("failed to get cvm apply capacity, for result is nil, crpTraceID: %s, rid: %s",
			resp.TraceId, kt.Rid)
		return nil
	}

	capacityItem := &types.CapacityInfo{
		Region:  input.Region,
		Zone:    zone,
		MaxNum:  int64(resp.Result.MaxNum),
		MaxInfo: make([]*types.CapacityMaxInfo, 0),
	}

	for _, info := range resp.Result.MaxInfo {
		capacityItem.MaxInfo = append(capacityItem.MaxInfo, &types.CapacityMaxInfo{
			Key:   c.translateCapacityKey(info.Key),
			Value: int64(info.Value),
		})
	}

	// 2. query all subnet info for left ip number
	subnetToLeftIp := make(map[string]*cvmapi.SubnetInfo)
	for _, vpcItem := range vpcList {
		subnetList, err := c.querySubnet(kt, input.Region, zone, vpcItem)
		if err != nil {
			logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
			return nil
		}
		for _, subnetItem := range subnetList {
			subnetToLeftIp[subnetItem.Id] = subnetItem
		}
	}

	// 3. sum up total left ip number
	totalLeftIp := c.sumLeftIp(subnetToLeftIp, vpcList, vpcToSubnet)

	// 4. update max info
	c.updateCapacityMaxInfo(capacityItem, totalLeftIp, ignorePrediction)

	jsonReq, err := json.Marshal(req)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityReq, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	// 需要记录crp返回的所有结果包括traceid
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityResp, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	jsonCapacityItem, err := json.Marshal(capacityItem)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityItem, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	logs.Infof("get zone capacity info, input: %+v, zone: %s, capacityReq: %s, capacityResp: %s, capacityItem: %s, "+
		"vpcList: %v, rid: %s", cvt.PtrToVal(input), zone, string(jsonReq), string(jsonResp), jsonCapacityItem,
		vpcList, kt.Rid)

	return capacityItem
}

func (c *capacity) createCapacityReq(input *types.GetCapacityParam, zone string, vpcList []string,
	vpcToSubnet map[string][]string) *cvmapi.CapacityReq {

	projectName := input.RequireType.ToObsProject()

	req := &cvmapi.CapacityReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCapacityMethod,
		},
		Params: &cvmapi.CapacityParam{
			DeptId:       cvmapi.CvmDeptId,
			Business3Id:  cvmapi.CvmLaunchBiz3Id,
			CloudCampus:  zone,
			InstanceType: input.DeviceType,
			VpcId:        vpcList[0],
			SubnetId:     vpcToSubnet[vpcList[0]][0],
			ProjectName:  string(projectName),
		},
	}
	// 计费模式,默认包年包月
	if len(input.ChargeType) > 0 {
		req.Params.ChargeType = input.ChargeType
	}

	return req
}

func (c *capacity) querySubnet(kt *kit.Kit, region, zone, vpc string) ([]*cvmapi.SubnetInfo, error) {
	subnetReq := cvmapi.SubnetRealParam{
		Region:      region,
		CloudCampus: zone,
		VpcId:       vpc,
	}
	resp, err := c.cvm.QueryRealCvmSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Result, nil
}

func (c *capacity) sumLeftIp(subnetToLeftIp map[string]*cvmapi.SubnetInfo, vpcList []string,
	vpcToSubnet map[string][]string) int64 {

	subnetIdList := make([]string, 0)
	for _, vpc := range vpcList {
		subnetIdList = append(subnetIdList, vpcToSubnet[vpc]...)
	}

	subnetIdList = arrayutil.StrArrayUnique(subnetIdList)

	total := 0
	for _, subnetId := range subnetIdList {
		if subnetToLeftIp[subnetId] != nil {
			total = total + subnetToLeftIp[subnetId].LeftIpNum
		}
	}

	return int64(total)
}

func (c *capacity) updateCapacityMaxInfo(capacity *types.CapacityInfo, leftIp int64, ignorePrediction bool) {
	maxNum := leftIp
	for _, maxInfo := range capacity.MaxInfo {
		key := maxInfo.Key
		if key == hcmKeyIPCap {
			maxInfo.Value = leftIp
		}

		// 所有key的最小值，为可申请的最大值；当忽略预测时，只需要关心所选VPC子网可用IP数和云梯系统单次最大申请量。
		if maxInfo.Value < maxNum && (!ignorePrediction || key == hcmKeyIPCap || key == crpKeyApplyLimit) {
			maxNum = maxInfo.Value
		}
	}

	capacity.MaxNum = maxNum
}

func (c *capacity) getCapacityFlag(num int) int {
	flag := types.CapLevelEmpty
	if num <= 10 {
		flag = types.CapLevelLow
	} else if num <= 50 {
		flag = types.CapLevelMedium
	} else {
		flag = types.CapLevelHigh
	}

	return flag
}

const (
	crpKeyCBSCap        = "云后端CBS容量计算可申领量"
	crpKeyCVMCap        = "云后端CVM容量计算可申领量"
	crpKeyIPCap         = "所选VPC子网可用IP数"
	crpKeyPredictionCap = "未执行需求预测的可申领量"
	crpKeyApplyLimit    = "云梯系统单次提单最大量"

	hcmKeyCBSCap        = "云后端CBS库存可申请量"
	hcmKeyCVMCap        = "云后端CVM库存可申请量"
	hcmKeyIPCap         = "所选VPC子网可用IP数"
	hcmKeyPredictionCap = "未执行需求预测的可申请量"
	hcmKeyApplyLimit    = "云梯系统单次最大申请量"
)

// translateCapacityKey translate yunti capacity info key to cr capacity info key
func (c *capacity) translateCapacityKey(key string) string {
	switch key {
	case crpKeyCBSCap:
		return hcmKeyCBSCap
	case crpKeyCVMCap:
		return hcmKeyCVMCap
	case crpKeyIPCap:
		return hcmKeyIPCap
	case crpKeyPredictionCap:
		return hcmKeyPredictionCap
	case crpKeyApplyLimit:
		return hcmKeyApplyLimit
	default:
		return key
	}
}
