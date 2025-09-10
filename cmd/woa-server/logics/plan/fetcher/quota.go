/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

package fetcher

import (
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	tablegconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"
)

// GetPlanTransferQuotaConfigs 获取预测转移额度配置
func (f *ResPlanFetcher) GetPlanTransferQuotaConfigs(kt *kit.Kit) (ptypes.TransferQuotaConfig, error) {
	dbConfigs, err := f.GetConfigsFromData(kt, constant.ResourcePlanTransferKey)
	if err != nil {
		logs.Errorf("failed to get plan transfer quota configs from db, err: %v, rid: %s", err, kt.Rid)
		return ptypes.TransferQuotaConfig{}, err
	}

	configMap := cvt.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, interface{}) {
		return t.ConfigKey, t.ConfigValue
	})

	config := ptypes.TransferQuotaConfig{
		Quota:      f.resPlanCfg.RefreshTransferQuota.Quota,      // 预测转移额度
		AuditQuota: f.resPlanCfg.RefreshTransferQuota.AuditQuota, // 预测转移审批额度
	}
	if v, ok := configMap[constant.TransferQuotaKey]; ok {
		config.Quota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert biz quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return ptypes.TransferQuotaConfig{}, err
		}
	}

	if v, ok := configMap[constant.TransferAuditQuotaKey]; ok {
		config.AuditQuota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert ieg quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return ptypes.TransferQuotaConfig{}, err
		}
	}

	return config, nil
}

// GetConfigsFromData 从数据层global_config获取配置
func (f *ResPlanFetcher) GetConfigsFromData(kt *kit.Kit, configType string) ([]tablegconf.GlobalConfigTable, error) {
	filter := tools.ExpressionAnd(tools.RuleEqual("config_type", configType))

	dataReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	dataResp, err := f.client.DataService().Global.GlobalConfig.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return nil, err
	}

	return dataResp.Details, nil
}
