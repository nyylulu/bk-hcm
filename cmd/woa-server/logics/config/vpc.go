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

	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"go.mongodb.org/mongo-driver/mongo"
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
	// GetRegionDftVpc gets the default vpc of a region.
	GetRegionDftVpc(kt *kit.Kit, region string) (string, error)
	// IsRegionDftVpc check if given vpc is the default vpc of a region.
	IsRegionDftVpc(kt *kit.Kit, vpc string) (bool, error)
	// UpsertRegionDftVpc upsert the default vpc of region.
	UpsertRegionDftVpc(kt *kit.Kit, input []types.RegionDftVpc) error
}

// NewVpcOp creates a vpc interface
func NewVpcOp(client *client.ClientSet, thirdCli *thirdparty.Client) VpcIf {
	return &vpc{
		cvm:    thirdCli.OldCVM,
		client: client,
	}
}

type vpc struct {
	cvm    cvmapi.CVMClientInterface
	client *client.ClientSet
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
		logs.Errorf("failed to create vpc, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().Vpc().CreateVpc(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create vpc, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("failed to update vpc, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("failed to delete vpc, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("failed to get cvm vpc info, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get cvm vpc info, code: %d, msg: %s, region: %s, crpTraceID: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, param.Region, resp.TraceId, kt.Rid)
		return fmt.Errorf("failed to get cvm vpc info, code: %d, msg: %s, crpTraceID: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId)
	}

	for _, vpcItem := range resp.Result {
		filter := map[string]interface{}{
			"region":   param.Region,
			"vpc_id":   vpcItem.Id,
			"vpc_name": vpcItem.Name,
		}
		count, err := config.Operation().Vpc().CountVpc(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to count vpc with filter: %+v, err: %v, rid: %s", filter, err, kt.Rid)
			return err
		}
		// 按云端返回的VPCID、VPC名称能查到数据，说明已同步，用于多次执行同步的场景
		if count > 0 {
			continue
		}

		listFilter := &mapstr.MapStr{
			"region": param.Region,
			"vpc_id": vpcItem.Id,
		}
		vpcList, err := config.Operation().Vpc().FindManyVpc(kt.Ctx, listFilter)
		if err != nil {
			logs.Errorf("failed to list vpc with filter: %+v, err: %v, rid: %s", filter, err, kt.Rid)
			return err
		}

		txnErr := dal.RunTransaction(kit.New(), func(sc mongo.SessionContext) error {
			kt.Ctx = sc
			// 清理旧的VPC
			for _, vpcInfo := range vpcList {
				err = v.DeleteVpc(kt, vpcInfo.BkInstId)
				if err != nil {
					logs.Errorf("failed to delete vpc, err: %v, vpcInstID: %d, region: %s, vpcItem: %+v, rid: %s",
						err, vpcInfo.BkInstId, param.Region, cvt.PtrToVal(vpcItem), kt.Rid)
					return err
				}
			}

			// 插入新的VPC
			vpcCfg := &types.Vpc{
				Region:  param.Region,
				VpcId:   vpcItem.Id,
				VpcName: vpcItem.Name,
			}
			if _, err = v.CreateVpc(kt, vpcCfg); err != nil {
				logs.Errorf("failed to create vpc, err: %v, rid: %s", filter, err, kt.Rid)
				return err
			}
			return nil
		})
		if txnErr != nil {
			logs.Errorf("failed to create vpc with transation, err: %v, rid: %s", filter, txnErr, kt.Rid)
			return txnErr
		}
	}
	return nil
}

// 请勿继续添加内容，应该通过/config/region/default_vpc/upsert接口添加到db
var regionToVpc = map[string]string{
	"ap-guangzhou":       "vpc-03nkx9tv",
	"ap-tianjin":         "vpc-1yoew5gc",
	"ap-shanghai":        "vpc-2x7lhtse",
	"eu-frankfurt":       "vpc-38klpz7z",
	"ap-singapore":       "vpc-706wf55j",
	"ap-tokyo":           "vpc-8iple1iq",
	"ap-seoul":           "vpc-99wg8fre",
	"ap-hongkong":        "vpc-b5okec48",
	"na-toronto":         "vpc-drefwt2v",
	"ap-xian-ec":         "vpc-efw4kf6r",
	"ap-nanjing":         "vpc-fb7sybzv",
	"ap-chongqing":       "vpc-gelpqsur",
	"ap-shenzhen":        "vpc-kwgem8tj",
	"na-siliconvalley":   "vpc-n040n5bl",
	"ap-hangzhou-ec":     "vpc-puhasca0",
	"ap-fuzhou-ec":       "vpc-hdxonj2q",
	"ap-wuhan-ec":        "vpc-867lsj6w",
	"ap-beijing":         "vpc-bhb0y6g8",
	"ap-jinan-ec":        "vpc-kgepmcdd",
	"ap-chengdu":         "vpc-r1wicnlq",
	"ap-zhengzhou-ec":    "vpc-54mjeaf8",
	"ap-shenyang-ec":     "vpc-rea7a2kc",
	"ap-changsha-ec":     "vpc-erdqk82h",
	"ap-hefei-ec":        "vpc-e0a5jxa7",
	"ap-shijiazhuang-ec": "vpc-6b3vbija",
}

// GetRegionDftVpc gets the default vpc of a region.
func (v *vpc) GetRegionDftVpc(kt *kit.Kit, region string) (string, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleEqual("config_key", region),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return "", err
	}
	if len(list.Details) == 0 {
		// 兜底兼容逻辑，防止部署时还没添加默认值
		vpcVal, ok := regionToVpc[region]
		if !ok {
			return "", fmt.Errorf("found no default vpc with region: %s", region)
		}
		return vpcVal, nil
	}
	result := new(types.DftVpc)
	if err = json.UnmarshalFromString(string(list.Details[0].ConfigValue), &result); err != nil {
		logs.Errorf("failed to unmarshal vpc, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return "", err
	}

	return result.VpcID, nil
}

// IsRegionDftVpc check if given vpc is the default vpc of a region.
func (v *vpc) IsRegionDftVpc(kt *kit.Kit, vpc string) (bool, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleJSONEqual("config_value.vpc_id", vpc),
		),
		Page: &core.BasePage{Count: true},
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, vpc: %s, rid: %s", err, vpc, kt.Rid)
		return false, err
	}
	if list.Count > 0 {
		return true, nil
	}

	// 兜底兼容逻辑，防止部署时还没添加默认值
	for _, val := range regionToVpc {
		if vpc == val {
			return true, nil
		}
	}

	return false, nil
}

// UpsertRegionDftVpc upsert the default vpc of region.
func (v *vpc) UpsertRegionDftVpc(kt *kit.Kit, input []types.RegionDftVpc) error {
	if len(input) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("input length must be less than %d", constant.BatchOperationMaxLimit)
	}

	regions := make([]string, 0, len(input))
	regionVpcMap := make(map[string]types.DftVpc, len(input))
	for _, regionDftVpc := range input {
		if err := regionDftVpc.Validate(); err != nil {
			return err
		}

		if _, ok := regionVpcMap[regionDftVpc.Region]; ok {
			return fmt.Errorf("found duplicate region: %s", regionDftVpc.Region)
		}

		regions = append(regions, regionDftVpc.Region)
		regionVpcMap[regionDftVpc.Region] = regionDftVpc.DftVpc
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleJsonIn("config_key", regions),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, region: %v, rid: %s", err, regions, kt.Rid)
		return err
	}
	existRegionDftVpc := make(map[string]cgconf.GlobalConfig, len(list.Details))
	for _, detail := range list.Details {
		existRegionDftVpc[detail.ConfigKey] = cgconf.GlobalConfig{
			ID:          detail.ID,
			ConfigKey:   detail.ConfigKey,
			ConfigValue: detail.ConfigValue,
			ConfigType:  detail.ConfigType,
			Memo:        detail.Memo,
		}
	}

	update := make([]cgconf.GlobalConfig, 0)
	create := make([]cgconf.GlobalConfigT[any], 0)
	for regionKey, vpcVal := range regionVpcMap {
		if detail, ok := existRegionDftVpc[regionKey]; ok {
			detail.ConfigValue = vpcVal
			update = append(update, detail)
			continue
		}
		create = append(create, cgconf.GlobalConfigT[any]{
			ConfigKey:   regionKey,
			ConfigValue: vpcVal,
			ConfigType:  constant.GlobalConfigTypeRegionDefaultVpc,
		})
	}

	if len(update) != 0 {
		updateReq := &datagconf.BatchUpdateReq{Configs: update}
		if err = v.client.DataService().Global.GlobalConfig.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("failed to update region default vpc, err: %v, data: %v, rid: %s", err, update, kt.Rid)
			return err
		}
	}

	if len(create) != 0 {
		createReq := &datagconf.BatchCreateReqT[any]{Configs: create}
		if _, err = v.client.DataService().Global.GlobalConfig.BatchCreate(kt, createReq); err != nil {
			logs.Errorf("failed to create region default vpc, err: %v, data: %v, rid: %s", err, create, kt.Rid)
			return err
		}
	}

	return nil
}
