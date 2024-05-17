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

	"hcm/cmd/woa-server/common/blog"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/model/config"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
)

// VpcIf provides management interface for operations of vpc config
type VpcIf interface {
	// GetVpc get vpc type config list
	GetVpc(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetVpcResult, error)
	// GetVpcList get vpc id list
	GetVpcList(kt *kit.Kit, cond map[string]interface{}) (*types.GetVpcListRst, error)
	// CreateVpc creates vpc type config
	CreateVpc(kt *kit.Kit, input *types.Vpc) (mapstr.MapStr, error)
	// UpdateVpc updates vpc type config
	UpdateVpc(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteVpc deletes vpc type config
	DeleteVpc(kt *kit.Kit, instId int64) error
	// SyncVpc sync vpc config from yunti
	SyncVpc(kt *kit.Kit, param *types.GetVpcParam) error
}

// NewVpcOp creates a vpc interface
func NewVpcOp(thirdCli *thirdparty.Client) VpcIf {
	return &vpc{
		cvm: thirdCli.CVM,
	}
}

type vpc struct {
	cvm cvmapi.CVMClientInterface
}

// GetVpc get vpc type config list
func (v *vpc) GetVpc(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetVpcResult, error) {
	insts, err := config.Operation().Vpc().FindManyVpc(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetVpcResult{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetVpcList get vpc id list
func (v *vpc) GetVpcList(kt *kit.Kit, cond map[string]interface{}) (*types.GetVpcListRst, error) {
	insts, err := config.Operation().Vpc().FindManyVpcId(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetVpcListRst{
		Info: insts,
	}

	return rst, nil
}

// CreateVpc creates vpc type config
func (v *vpc) CreateVpc(kt *kit.Kit, input *types.Vpc) (mapstr.MapStr, error) {
	id, err := config.Operation().Vpc().NextSequence(kt.Ctx)
	if err != nil {
		blog.Errorf("failed to create vpc, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Vpc().CreateVpc(kt.Ctx, input); err != nil {
		blog.Errorf("failed to create vpc, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateVpc updates vpc type config
func (v *vpc) UpdateVpc(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Vpc().UpdateVpc(kt.Ctx, filter, input); err != nil {
		blog.Errorf("failed to update vpc, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteVpc deletes vpc type config
func (v *vpc) DeleteVpc(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().Vpc().DeleteVpc(kt.Ctx, filter); err != nil {
		blog.Errorf("failed to delete vpc, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SyncVpc sync vpc config from yunti
func (v *vpc) SyncVpc(kt *kit.Kit, param *types.GetVpcParam) error {
	req := &cvmapi.VpcReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmVpcMethod,
		},
		Params: &cvmapi.VpcParam{
			DeptId: cvmapi.CvmDeptId,
			Region: param.Region,
		},
	}

	resp, err := v.cvm.QueryCvmVpc(kt.Ctx, nil, req)
	if err != nil {
		blog.Errorf("failed to get cvm vpc info, err: %v", err)
		return err
	}

	if resp.Error.Code != 0 {
		blog.Errorf("failed to get cvm vpc info, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		return fmt.Errorf("failed to get cvm vpc info, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	for _, vpc := range resp.Result {
		filter := map[string]interface{}{
			"region":   param.Region,
			"vpc_id":   vpc.Id,
			"vpc_name": vpc.Name,
		}
		cnt, err := config.Operation().Vpc().CountVpc(kt.Ctx, filter)
		if err != nil {
			blog.Errorf("failed to count vpc with filter: %+v, err: %v", filter, err)
			return err
		}
		if cnt <= 0 {
			vpcCfg := &types.Vpc{
				Region:  param.Region,
				VpcId:   vpc.Id,
				VpcName: vpc.Name,
			}
			if _, err := v.CreateVpc(kt, vpcCfg); err != nil {
				blog.Errorf("failed to create vpc, err: %v", filter, err)
				return err
			}
		}
	}

	return nil
}
