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

package dissolve

import (
	"hcm/cmd/woa-server/logics/dissolve/host"
	"hcm/cmd/woa-server/logics/dissolve/module"
	dissolvetable "hcm/cmd/woa-server/logics/dissolve/table"
	esCli "hcm/cmd/woa-server/thirdparty/es"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/pkg/cc"
	"hcm/pkg/dal/dao"
)

// Logics provides resource dissolve logics
type Logics interface {
	RecycledModule() module.RecycledModule
	RecycledHost() host.RecycledHost
	Table() dissolvetable.Table
}

type logics struct {
	recycledModule module.RecycledModule
	recycledHost   host.RecycledHost
	table          dissolvetable.Table
}

// New create a logics manager
func New(dao dao.Set, esbCli esb.Client, esCli *esCli.EsCli, conf cc.WoaServerSetting) Logics {
	recycledModule := module.New(dao)
	recycledHost := host.New(dao)
	originDate := conf.ResDissolve.OriginDate
	blacklist := conf.Blacklist

	return &logics{
		recycledModule: recycledModule,
		recycledHost:   recycledHost,
		table:          dissolvetable.New(recycledModule, recycledHost, esbCli, esCli, originDate, blacklist),
	}
}

// RecycledModule recycled module interface
func (l *logics) RecycledModule() module.RecycledModule {
	return l.recycledModule
}

// RecycledHost recycled host interface
func (l *logics) RecycledHost() host.RecycledHost {
	return l.recycledHost
}

// Table resource dissolve table interface
func (l *logics) Table() dissolvetable.Table {
	return l.table
}
