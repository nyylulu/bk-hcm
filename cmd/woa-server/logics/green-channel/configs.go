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

// Package greenchannel ...
package greenchannel

import (
	gctypes "hcm/cmd/woa-server/types/green-channel"
	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	tablegconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"
)

// defaultConfigs ...
var defaultConfigs = gctypes.Config{
	BizQuota:       500,
	IEGQuota:       50000,
	AuditThreshold: 100,
}

// GetConfigs ...
func (l *logics) GetConfigs(kt *kit.Kit) (gctypes.Config, error) {
	dbConfigs, err := l.getConfigsFromData(kt)
	if err != nil {
		logs.Errorf("failed to get configs from db, err: %v, rid: %s", err, kt.Rid)
		return gctypes.Config{}, err
	}

	configMap := converter.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, interface{}) {
		return t.ConfigKey, t.ConfigValue
	})

	config := defaultConfigs
	if v, ok := configMap[constant.ConfigKeyGCBizQuota]; ok {
		config.BizQuota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert biz quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return gctypes.Config{}, err
		}
	}

	if v, ok := configMap[constant.ConfigTypeGCIEGQuota]; ok {
		config.IEGQuota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert ieg quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return gctypes.Config{}, err
		}
	}

	if v, ok := configMap[constant.ConfigTypeGCAuditThreshold]; ok {
		config.AuditThreshold, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert audit threshold, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return gctypes.Config{}, err
		}
	}

	return config, nil
}

// getConfigMap ...
func (l *logics) getConfigsFromData(kt *kit.Kit) ([]tablegconf.GlobalConfigTable, error) {
	filter := tools.ExpressionAnd(tools.RuleEqual("config_type", constant.ConfigTypeGreenChannel))

	dataReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	dataResp, err := l.client.DataService().Global.GlobalConfig.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return nil, err
	}

	return dataResp.Details, nil
}

// UpdateConfigs ...
func (l *logics) UpdateConfigs(kt *kit.Kit, req *gctypes.UpdateConfigsReq) error {
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate request, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	dbConfigs, err := l.getConfigsFromData(kt)
	if err != nil {
		logs.Errorf("failed to get configs from db, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	configIDMap := converter.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, string) {
		return t.ConfigKey, t.ID
	})

	dataReq := &datagconf.BatchUpdateReq{
		Configs: make([]cgconf.GlobalConfig, 0),
	}

	if req.BizQuota != nil {
		dataReq.Configs = append(dataReq.Configs, cgconf.GlobalConfig{
			ID:          configIDMap[constant.ConfigKeyGCBizQuota],
			ConfigValue: req.BizQuota,
		})
	}

	if req.IEGQuota != nil {
		dataReq.Configs = append(dataReq.Configs, cgconf.GlobalConfig{
			ID:          configIDMap[constant.ConfigTypeGCIEGQuota],
			ConfigValue: req.IEGQuota,
		})
	}

	if req.AuditThreshold != nil {
		dataReq.Configs = append(dataReq.Configs, cgconf.GlobalConfig{
			ID:          configIDMap[constant.ConfigTypeGCAuditThreshold],
			ConfigValue: req.AuditThreshold,
		})
	}

	if len(dataReq.Configs) == 0 {
		return nil
	}

	if err := l.client.DataService().Global.GlobalConfig.BatchUpdate(kt, dataReq); err != nil {
		logs.Errorf("failed to update global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return err
	}

	return nil
}
