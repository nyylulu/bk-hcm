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

package handlers

import (
	"runtime/debug"
	"time"

	"hcm/cmd/cloud-server/logics/ziyan"
	"hcm/pkg/api/core"
	ds "hcm/pkg/api/data-service"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
)

// TimingCheckBPaaSApplication 定时检查BPaaS单据。
func TimingCheckBPaaSApplication(cliSet *client.ClientSet, sd serviced.ServiceDiscover, interval time.Duration) {

	defer func() {
		err := recover()
		if err != nil {
			logs.Errorf("%s panic, err: %s, stack:\n%v", constant.WaitAndCheckBPaasFailed, err, debug.Stack())
		}
	}()

	if interval == 0 {
		interval = time.Minute
	}

	for {
		time.Sleep(interval)

		if !sd.IsMaster() {
			logs.Infof("check bpaas application, but is not master, skip")
			continue
		}

		kt := core.NewBackendKit()
		if err := WaitAndCheckBPaasApplication(kt, cliSet.DataService(), cliSet.HCService()); err != nil {
			logs.Errorf("%s, err: %v, rid: %s", constant.WaitAndCheckBPaasFailed, err, kt.Rid)
		}
	}
}

// WaitAndCheckBPaasApplication wait deliver cvm.
func WaitAndCheckBPaasApplication(kt *kit.Kit, dsCli *dataservice.Client, hcCli *hcservice.Client) error {

	// 查询交付状态中的BPaas单据
	apps, err := queryPendingBPaasApplication(kt, dsCli)
	if err != nil {
		return err
	}

	// 如果没有需要处理的单据跳过即可
	if len(apps) == 0 {
		logs.V(5).Infof("pending bpaas application not found, skip handle, rid: %s", kt.Rid)
		return nil
	}

	for _, app := range apps {
		err := ziyan.CheckAndUpdateBPaasStatus(kt, dsCli, hcCli, app)
		if err != nil {
			logs.Errorf("%s, check and update bpaas application failed, err: %v, rid: %s",
				constant.WaitAndCheckBPaasFailed, err, kt.Rid)
			// continue handle next application
			continue
		}
	}

	return nil
}

func queryPendingBPaasApplication(kt *kit.Kit, cli *dataservice.Client) (apps []*ds.ApplicationResp, err error) {

	req := &ds.ApplicationListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "status",
					Op:    filter.Equal.Factory(),
					Value: enumor.Pending,
				},
				&filter.AtomRule{
					Field: "source",
					Op:    filter.Equal.Factory(),
					Value: enumor.ApplicationSourceBPaas,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.Application.ListApplication(kt, req)
	if err != nil {
		logs.Errorf("list bpaas application failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return result.Details, nil

}
