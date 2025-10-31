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

// Package config config
package config

import (
	"encoding/json"
	"errors"
	"time"

	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	globalconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Config provides interface for operations of dissolve config.
type Config interface {
	GetDissolveHostApplyTime(kt *kit.Kit) (*time.Time, error)
	UpsertDissolveHostApplyTime(kt *kit.Kit, time *time.Time) error
}

type logics struct {
	cliSet *client.ClientSet
}

// New create dissolve config logics.
func New(client *client.ClientSet) Config {
	return &logics{
		cliSet: client,
	}
}

// GetDissolveHostApplyTime get dissolve host apply time.
func (l *logics) GetDissolveHostApplyTime(kt *kit.Kit) (*time.Time, error) {
	config, exist, err := l.getDissolveHostApplyTimeConfig(kt)
	if err != nil {
		logs.Errorf("failed to get dissolve host apply time config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !exist {
		logs.Errorf("dissolve host apply time config not exist, rid: %s", kt.Rid)
		return nil, errors.New("dissolve host apply time config not exist")
	}

	applyTime := new(time.Time)
	if err = json.Unmarshal([]byte(config.ConfigValue), &applyTime); err != nil {
		logs.Errorf("failed to unmarshal config value, err: %v, value: %s, rid: %s", err, config.ConfigValue, kt.Rid)
		return nil, err
	}

	return applyTime, nil
}

func (l *logics) getDissolveHostApplyTimeConfig(kt *kit.Kit) (*globalconf.GlobalConfigTable, bool, error) {
	req := core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigResDissolve),
			tools.RuleJSONEqual("config_key", enumor.GlobalConfigDissolveHostApplyTime),
		),
		Page: core.NewDefaultBasePage(),
	}
	cfgResp, err := l.cliSet.DataService().Global.GlobalConfig.List(kt, &req)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, false, err
	}
	if len(cfgResp.Details) == 0 {
		return nil, false, nil
	}

	return &cfgResp.Details[0], true, nil
}

// UpsertDissolveHostApplyTime upsert dissolve host apply time.
func (l *logics) UpsertDissolveHostApplyTime(kt *kit.Kit, time *time.Time) error {
	if time == nil {
		logs.Errorf("time is nil, rid: %s", kt.Rid)
		return errors.New("time is nil")
	}

	oldConf, exist, err := l.getDissolveHostApplyTimeConfig(kt)
	if err != nil {
		logs.Errorf("failed to get dissolve host apply time config, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	conf := cgconf.GlobalConfigT[any]{
		ConfigType:  string(enumor.GlobalConfigResDissolve),
		ConfigKey:   string(enumor.GlobalConfigDissolveHostApplyTime),
		ConfigValue: time,
	}
	if !exist {
		createReq := datagconf.BatchCreateReqT[any]{Configs: []cgconf.GlobalConfigT[any]{conf}}
		if _, err = l.cliSet.DataService().Global.GlobalConfig.BatchCreate(kt, &createReq); err != nil {
			logs.Errorf("failed to create dissolve host apply time global config, err: %v, req: %+v, rid: %s", err,
				createReq, kt.Rid)
			return err
		}
		return nil
	}

	conf.ID = oldConf.ID
	updateReq := datagconf.BatchUpdateReq{Configs: []cgconf.GlobalConfigT[any]{conf}}
	if err = l.cliSet.DataService().Global.GlobalConfig.BatchUpdate(kt, &updateReq); err != nil {
		logs.Errorf("failed to update dissolve host apply time  global config, err: %v, req: %+v, rid: %s", err,
			updateReq, kt.Rid)
		return err
	}

	return nil
}
