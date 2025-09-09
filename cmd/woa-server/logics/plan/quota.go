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

// Package plan ...
package plan

import (
	"errors"

	plantypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tablegconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/util"
)

// GetPlanTransferQuotaConfigs 获取预测转移额度配置
func (c *Controller) GetPlanTransferQuotaConfigs(kt *kit.Kit) (plantypes.TransferQuotaConfig, error) {
	dbConfigs, err := c.getConfigsFromData(kt, constant.ResourcePlanTransferKey)
	if err != nil {
		logs.Errorf("failed to get plan transfer quota configs from db, err: %v, rid: %s", err, kt.Rid)
		return plantypes.TransferQuotaConfig{}, err
	}

	configMap := cvt.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, interface{}) {
		return t.ConfigKey, t.ConfigValue
	})

	config := plantypes.TransferQuotaConfig{
		Quota:      c.resPlanCfg.RefreshTransferQuota.Quota,      // 预测转移额度
		AuditQuota: c.resPlanCfg.RefreshTransferQuota.AuditQuota, // 预测转移审批额度
	}
	if v, ok := configMap[constant.TransferQuotaKey]; ok {
		config.Quota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert biz quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return plantypes.TransferQuotaConfig{}, err
		}
	}

	if v, ok := configMap[constant.TransferAuditQuotaKey]; ok {
		config.AuditQuota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert ieg quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return plantypes.TransferQuotaConfig{}, err
		}
	}

	return config, nil
}

// getConfigMap ...
func (c *Controller) getConfigsFromData(kt *kit.Kit, configType string) ([]tablegconf.GlobalConfigTable, error) {
	filter := tools.ExpressionAnd(tools.RuleEqual("config_type", configType))

	dataReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	dataResp, err := c.client.DataService().Global.GlobalConfig.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return nil, err
	}

	return dataResp.Details, nil
}

// UpdatePlanTransferQuotaConfigs 更新预测转移额度配置
func (c *Controller) UpdatePlanTransferQuotaConfigs(kt *kit.Kit,
	req *plantypes.UpdatePlanTransferQuotaConfigsReq) error {

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate request, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}

	dbConfigs, err := c.getConfigsFromData(kt, constant.ResourcePlanTransferKey)
	if err != nil {
		logs.Errorf("failed to get resplan transfer configs from db, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	configIDMap := cvt.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, string) {
		return t.ConfigKey, t.ID
	})

	dataReq := &datagconf.BatchUpdateReq{
		Configs: make([]cgconf.GlobalConfig, 0),
	}

	if req.Quota != nil {
		if configID, ok := configIDMap[constant.TransferQuotaKey]; ok {
			dataReq.Configs = append(dataReq.Configs, cgconf.GlobalConfig{
				ID:          configID,
				ConfigValue: req.Quota,
			})
		}
	}

	if req.AuditQuota != nil {
		if configID, ok := configIDMap[constant.TransferAuditQuotaKey]; ok {
			dataReq.Configs = append(dataReq.Configs, cgconf.GlobalConfig{
				ID:          configID,
				ConfigValue: req.AuditQuota,
			})
		}
	}

	if len(dataReq.Configs) == 0 {
		return errors.New("no global config to update")
	}

	if err = c.client.DataService().Global.GlobalConfig.BatchUpdate(kt, dataReq); err != nil {
		logs.Errorf("failed to update resplan transfer global config, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), kt.Rid)
		return err
	}

	return nil
}

// QueryCrpDemandsQuota 查询crp预测额度
func (c *Controller) QueryCrpDemandsQuota(kt *kit.Kit, obsProject []enumor.ObsProject, technicalClasses []string) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	// 转换 ObsProject 类型
	obsNewProjects := make([]string, len(obsProject))
	for i, op := range obsProject {
		obsNewProjects[i] = string(op)
	}

	demandReq := &QueryIEGDemandsReq{
		OpProdNames:      []string{constant.IEGResPlanServiceProductName},
		ObsProjects:      obsNewProjects,
		TechnicalClasses: technicalClasses,
	}
	demands, err := c.QueryIEGDemands(kt, demandReq)
	if err != nil {
		logs.Errorf("failed to query ieg demands from crp, err: %v, demandReq: %+v, rid: %s",
			err, cvt.PtrToVal(demandReq), kt.Rid)
		return nil, err
	}

	return demands, nil
}
