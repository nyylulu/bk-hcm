/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package bandwidthpackage

import (
	csbwpkg "hcm/pkg/api/cloud-server/bandwidth-package"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateRecommendBandPackage 更新推荐带宽包至GlobalConfig,覆盖式更新
func (svc *bandSvc) UpdateRecommendBandPackage(cts *rest.Contexts) (interface{}, error) {
	req := new(csbwpkg.UpdateBandwidthPackageRecommendOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.GlobalConfig, Action: meta.Create}})
	if err != nil {
		logs.Errorf("update global config auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	configID, config, err := svc.getRecommendConfig(cts.Kit)
	if err != nil {
		logs.Errorf("fail to get recommend config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(config) == 0 {
		createReq := &datagconf.BatchCreateReqT[any]{Configs: []cgconf.GlobalConfigT[any]{
			{
				ConfigKey:   constant.GlobalConfigTypeCLBBandwidthPackageRecommend,
				ConfigValue: req.PackageIDs,
				ConfigType:  constant.GlobalConfigTypeCLBBandwidthPackageRecommend,
			},
		}}
		if _, err = svc.client.DataService().Global.GlobalConfig.BatchCreate(cts.Kit, createReq); err != nil {
			logs.Errorf("fail to create recommend config, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	}

	updateReq := &datagconf.BatchUpdateReq{Configs: []cgconf.GlobalConfigT[any]{
		{
			ID:          configID,
			ConfigKey:   constant.GlobalConfigTypeCLBBandwidthPackageRecommend,
			ConfigValue: req.PackageIDs,
			ConfigType:  constant.GlobalConfigTypeCLBBandwidthPackageRecommend,
		},
	}}
	if err = svc.client.DataService().Global.GlobalConfig.BatchUpdate(cts.Kit, updateReq); err != nil {
		logs.Errorf("fail to update recommend config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
