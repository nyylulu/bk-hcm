/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package ziyan

import (
	"fmt"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"

	"github.com/tidwall/gjson"
)

// SgIf provides management interface for operations of security group config
type SgIf interface {
	// GetRegionDftSg get the default security group of a region.
	GetRegionDftSg(kt *kit.Kit, region string) (*types.DftSecurityGroup, error)
	// UpsertRegionDftSg upsert the default security group of region.
	UpsertRegionDftSg(kt *kit.Kit, input []types.RegionDftSg) error
	// GetAllDftSg get all default security group.
	GetAllDftSg(kt *kit.Kit) (map[string]string, error)
}

// NewSgOp creates a security group interface
func NewSgOp(client *client.ClientSet) SgIf {
	return &sg{client: client}
}

type sg struct {
	client *client.ClientSet
}

var regionToSecGroup = map[string]*types.DftSecurityGroup{
	"ap-guangzhou": {
		SgID:   "sg-ka67ywe9",
		SgName: "云梯默认安全组",
		SgDesc: "腾讯自研上云-默认安全组",
	},
	"ap-tianjin": {
		SgID:   "sg-c28492qp",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-shanghai": {
		SgID:   "sg-ibqae0te",
		SgName: "云梯默认安全组",
		SgDesc: "腾讯自研上云-默认安全组",
	},
	"eu-frankfurt": {
		SgID:   "sg-cet13de0",
		SgName: "云梯默认安全组",
		SgDesc: "云梯默认安全组",
	},
	"ap-singapore": {
		SgID:   "sg-hjtqedoe",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-tokyo": {
		SgID:   "sg-o1lfldnk",
		SgName: "云梯默认安全组",
		SgDesc: "云梯默认安全组",
	},
	"ap-seoul": {
		SgID:   "sg-i7h8hv5r",
		SgName: "云梯默认安全组",
		SgDesc: "云梯默认安全组",
	},
	"ap-hongkong": {
		SgID:   "sg-59kfufmn",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"na-toronto": {
		SgID:   "sg-7l82d7km",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-xian-ec": {
		SgID:   "sg-o4bmz4kg",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-nanjing": {
		SgID:   "sg-dybs7i3y",
		SgName: "云梯默认安全组",
		SgDesc: "腾讯自研上云-默认安全组",
	},
	"ap-chongqing": {
		SgID:   "sg-l5usnzxw",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-shenzhen": {
		SgID:   "sg-qkfewp0u",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"na-siliconvalley": {
		SgID:   "sg-q7usygae",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-hangzhou-ec": {
		SgID:   "sg-4ezyvbvl",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-fuzhou-ec": {
		SgID:   "sg-leqa6w29",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-wuhan-ec": {
		SgID:   "sg-p5ld4xyq",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-beijing": {
		SgID:   "sg-rjwj7cnt",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-jinan-ec": {
		SgID:   "sg-eag5dvzm",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-chengdu": {
		SgID:   "sg-g504fnlx",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-zhengzhou-ec": {
		SgID:   "sg-mdzp3pem",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-shenyang-ec": {
		SgID:   "sg-jvdlgqyx",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-changsha-ec": {
		SgID:   "sg-fohw41u4",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-hefei-ec": {
		SgID:   "sg-qjn542yi",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-shijiazhuang-ec": {
		SgID:   "sg-5qwjawx2",
		SgName: "云梯默认安全组",
		SgDesc: "",
	},
	"ap-qingyuan": {SgID: "sg-rzheledx"},
	"ap-bangkok":  {SgID: "sg-m33on5qq"},
	"ap-ashburn":  {SgID: "sg-osi7m525"},
	"sa-saopaulo": {SgID: "sg-9sfhy229"},
}

// GetAllDftSg get all region default security group
func (s *sg) GetAllDftSg(kt *kit.Kit) (map[string]string, error) {

	page := core.NewDefaultBasePage()
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultSecGroup),
		),
		Page: page,
	}
	var allConfigs = map[string]string{}
	for {
		list, err := s.client.DataService().Global.GlobalConfig.List(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list all default security group, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for i := range list.Details {
			cfgItem := list.Details[i]
			region := cfgItem.ConfigKey
			sgCloudID := gjson.Get(string(cfgItem.ConfigValue), "security_group_id").String()
			if sgCloudID == "" {
				return nil, fmt.Errorf("default security group data broken for region: %s, raw value: %s, rid: %s",
					cfgItem.ConfigValue, region, kt.Rid)
			}
			allConfigs[cfgItem.ConfigKey] = sgCloudID
		}
		if len(list.Details) < int(page.Limit) {
			break
		}
		page.Start += uint32(page.Limit)
	}
	if len(allConfigs) == 0 {
		for region, dftSg := range regionToSecGroup {
			allConfigs[region] = dftSg.SgID
		}
	}
	return allConfigs, nil
}

// GetRegionDftSg get the default security group of a region.
func (s *sg) GetRegionDftSg(kt *kit.Kit, region string) (*types.DftSecurityGroup, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultSecGroup),
			tools.RuleEqual("config_key", region),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := s.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default security group, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return nil, err
	}
	if len(list.Details) == 0 {
		// 兜底兼容逻辑，防止部署时还没添加默认值
		securityGroup, ok := regionToSecGroup[region]
		if !ok {
			return nil, fmt.Errorf("found no default security group with region: %s", region)
		}
		return securityGroup, nil
	}

	result := new(types.DftSecurityGroup)
	if err = json.UnmarshalFromString(string(list.Details[0].ConfigValue), &result); err != nil {
		logs.Errorf("failed to unmarshal security group, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return nil, err
	}

	return result, nil
}

// UpsertRegionDftSg upsert the default security group of region.
func (s *sg) UpsertRegionDftSg(kt *kit.Kit, input []types.RegionDftSg) error {
	if len(input) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("input length must be less than %d", constant.BatchOperationMaxLimit)
	}

	regions := make([]string, 0, len(input))
	regionSgMap := make(map[string]types.DftSecurityGroup, len(input))
	for _, regionDftSg := range input {
		if err := regionDftSg.Validate(); err != nil {
			return err
		}

		if _, ok := regionSgMap[regionDftSg.Region]; ok {
			return fmt.Errorf("found duplicate region: %s", regionDftSg.Region)
		}

		regions = append(regions, regionDftSg.Region)
		regionSgMap[regionDftSg.Region] = regionDftSg.DftSecurityGroup
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultSecGroup),
			tools.RuleJsonIn("config_key", regions),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := s.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default security group, err: %v, region: %v, rid: %s", err, regions, kt.Rid)
		return err
	}
	existRegionDftSg := make(map[string]cgconf.GlobalConfig, len(list.Details))
	for _, detail := range list.Details {
		existRegionDftSg[detail.ConfigKey] = cgconf.GlobalConfig{
			ID:          detail.ID,
			ConfigKey:   detail.ConfigKey,
			ConfigValue: detail.ConfigValue,
			ConfigType:  detail.ConfigType,
			Memo:        detail.Memo,
		}
	}

	update := make([]cgconf.GlobalConfig, 0)
	create := make([]cgconf.GlobalConfigT[any], 0)
	for regionKey, sgVal := range regionSgMap {
		if detail, ok := existRegionDftSg[regionKey]; ok {
			detail.ConfigValue = sgVal
			update = append(update, detail)
			continue
		}
		create = append(create, cgconf.GlobalConfigT[any]{
			ConfigKey:   regionKey,
			ConfigValue: sgVal,
			ConfigType:  constant.GlobalConfigTypeRegionDefaultSecGroup,
		})
	}

	if len(update) != 0 {
		updateReq := &datagconf.BatchUpdateReq{Configs: update}
		if err = s.client.DataService().Global.GlobalConfig.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("failed to update region default security group, err: %v, data: %v, rid: %s", err, update,
				kt.Rid)
			return err
		}
	}

	if len(create) != 0 {
		createReq := &datagconf.BatchCreateReqT[any]{Configs: create}
		if _, err = s.client.DataService().Global.GlobalConfig.BatchCreate(kt, createReq); err != nil {
			logs.Errorf("failed to create region default security group, err: %v, data: %v, rid: %s", err, create,
				kt.Rid)
			return err
		}
	}

	return nil
}
