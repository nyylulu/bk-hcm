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

package config

import (
	"fmt"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/model/config"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SubnetIf provides management interface for operations of subnet config
type SubnetIf interface {
	// GetSubnet get subnet type config list
	GetSubnet(kt *kit.Kit, cond map[string]interface{}) (*types.GetSubnetResult, error)
	// GetSubnetList get subnet detail config list
	GetSubnetList(kt *kit.Kit, input *types.GetSubnetListParam) (*types.GetSubnetResult, error)
	// CreateSubnet creates subnet type config
	CreateSubnet(kt *kit.Kit, input *types.Subnet) (mapstr.MapStr, error)
	// UpdateSubnet updates subnet type config
	UpdateSubnet(kt *kit.Kit, instId int64, input map[string]interface{}) error
	// UpdateSubnetBatch updates subnet config in batch
	UpdateSubnetBatch(kt *kit.Kit, cond, update map[string]interface{}) error
	// DeleteSubnet deletes subnet type config
	DeleteSubnet(kt *kit.Kit, instId int64) error

	// SyncSubnet sync subnet config from yunti
	SyncSubnet(kt *kit.Kit, param *types.GetSubnetParam) error
}

// NewSubnetOp creates a subnet interface
func NewSubnetOp(thirdCli *thirdparty.Client) SubnetIf {
	return &subnet{
		cvm: thirdCli.CVM,
	}
}

type subnet struct {
	cvm cvmapi.CVMClientInterface
}

// GetSubnet get subnet type config list
func (s *subnet) GetSubnet(kt *kit.Kit, cond map[string]interface{}) (*types.GetSubnetResult, error) {
	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}
	insts, err := config.Operation().Subnet().FindManySubnet(kt.Ctx, page, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetSubnetResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetSubnetList get subnet detail config list
func (s *subnet) GetSubnetList(kt *kit.Kit, input *types.GetSubnetListParam) (*types.GetSubnetResult, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config subnet detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetSubnetResult{}
	if input.Page.EnableCount {
		cnt, err := config.Operation().Subnet().CountSubnet(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get subnet detail count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*types.Subnet, 0)
		return rst, nil
	}

	insts, err := config.Operation().Subnet().FindManySubnet(kt.Ctx, input.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// CreateSubnet creates subnet type config
func (s *subnet) CreateSubnet(kt *kit.Kit, input *types.Subnet) (mapstr.MapStr, error) {
	id, err := config.Operation().Subnet().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create subnet, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Subnet().CreateSubnet(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create subnet, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateSubnet updates subnet type config
func (s *subnet) UpdateSubnet(kt *kit.Kit, instId int64, input map[string]interface{}) error {
	filter := map[string]interface{}{
		"id": instId,
	}

	if err := config.Operation().Subnet().UpdateSubnet(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update subnet, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateSubnetBatch updates subnet config in batch
func (s *subnet) UpdateSubnetBatch(kt *kit.Kit, cond, update map[string]interface{}) error {
	if err := config.Operation().Subnet().UpdateSubnet(kt.Ctx, cond, update); err != nil {
		logs.Errorf("failed to update subnet, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSubnet deletes subnet type config
func (s *subnet) DeleteSubnet(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Subnet().DeleteSubnet(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete subnet, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SyncSubnet sync subnet config from yunti
func (s *subnet) SyncSubnet(kt *kit.Kit, param *types.GetSubnetParam) error {
	req := &cvmapi.SubnetReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmSubnetMethod,
		},
		Params: &cvmapi.SubnetParam{
			DeptId: cvmapi.CvmDeptId,
			Region: param.Region,
			Zone:   param.Zone,
			VpcId:  param.Vpc,
		},
	}

	resp, err := s.cvm.QueryCvmSubnet(kt.Ctx, nil, req)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v", err)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get cvm subnet info, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		return fmt.Errorf("failed to get cvm subnet info, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	for _, subnet := range resp.Result {
		filter := map[string]interface{}{
			"region":      param.Region,
			"zone":        param.Zone,
			"vpc_id":      param.Vpc,
			"subnet_id":   subnet.Id,
			"subnet_name": subnet.Name,
		}
		cnt, err := config.Operation().Subnet().CountSubnet(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to count subnet with filter: %+v, err: %v", filter, err)
			return err
		}
		if cnt <= 0 {
			subnetCfg := &types.Subnet{
				Region:     param.Region,
				Zone:       param.Zone,
				VpcId:      param.Vpc,
				SubnetId:   subnet.Id,
				SubnetName: subnet.Name,
				Enable:     true,
				Comment:    "",
			}
			if _, err := s.CreateSubnet(kt, subnetCfg); err != nil {
				logs.Errorf("failed to create subnet, err: %v", filter, err)
				return err
			}
		}
	}

	return nil
}
