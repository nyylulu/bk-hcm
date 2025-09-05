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

package actcli

import (
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	woaserver "hcm/pkg/client/woa-server"
	"hcm/pkg/dal/dao"
	"hcm/pkg/thirdparty/alarmapi"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/sampwdapi"
)

var (
	cliSet    *client.ClientSet
	daoSet    dao.Set
	obsDaoSet dao.Set
	cmdbCli   cmdb.Client
	alarmCli  alarmapi.AlarmClientInterface
	samPwdCli sampwdapi.Client
)

// SetClientSet set client set.
func SetClientSet(cli *client.ClientSet) {
	cliSet = cli
}

// GetClientSet get client set.
func GetClientSet() *client.ClientSet {
	return cliSet
}

// GetHCService get hc service.
func GetHCService(labels ...string) *hcservice.Client {
	return cliSet.HCService(labels...)
}

// GetDataService get data service.
func GetDataService() *dataservice.Client {
	return cliSet.DataService()
}

// GetWoaServer get woa server.
func GetWoaServer() *woaserver.Client {
	return cliSet.WoaServer()
}

// SetDaoSet set dao set.
func SetDaoSet(cli dao.Set) {
	daoSet = cli
}

// GetDaoSet get dao set.
func GetDaoSet() dao.Set {
	return daoSet
}

// SetObsDaoSet set dao set.
func SetObsDaoSet(cli dao.Set) {
	obsDaoSet = cli
}

// GetObsDaoSet get dao set.
func GetObsDaoSet() dao.Set {
	return obsDaoSet
}

// SetCMDBClient set cmdb client.
func SetCMDBClient(cli cmdb.Client) {
	cmdbCli = cli
}

// GetCMDBCli get cmdb client.
func GetCMDBCli() cmdb.Client {
	return cmdbCli
}

// SetAlarmCli set alarm cli.
func SetAlarmCli(cli alarmapi.AlarmClientInterface) {
	alarmCli = cli
}

// GetAlarmCli get alarm cli.
func GetAlarmCli() alarmapi.AlarmClientInterface {
	return alarmCli
}

// SetSamPwdCli set sampwd cli.
func SetSamPwdCli(cli sampwdapi.Client) {
	samPwdCli = cli
}

// GetSamPwdCli get sampwd cli.
func GetSamPwdCli() sampwdapi.Client {
	return samPwdCli
}
