/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package detector ...
package detector

import (
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/querybuilder"
)

func (d *Detector) checkOwner(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkHasVm(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	return attempt, exeInfo, err
}

func (d *Detector) checkHasVm(ip string) (string, error) {
	ips := []string{ip}
	hostBase, err := d.getHostBaseInfo(ips)
	if err != nil {
		logs.Errorf("failed to get host from cc, err: %v, ip: %s", err, ip)
		return "", fmt.Errorf("failed to get host from cc err: %v", err)
	}

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("get invalid host num %d != 1", cnt)
		return "", fmt.Errorf("get invalid host num %d != 1", cnt)
	}

	host := hostBase[0]

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_svr_owner_asset_id",
						Operator: querybuilder.OperatorEqual,
						Value:    host.BkAssetID,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	// set rate limit to avoid cc api error "API rate limit exceeded by stage/resource strategy"
	ccLimiter.Take()
	resp, err := d.cc.ListHost(d.backendKit, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return "", err
	}

	respStr := d.structToStr(resp)
	exeInfo := fmt.Sprintf("vm check response: %s", respStr)

	vmNum := len(resp.Info)
	if vmNum > 0 {
		logs.Errorf("host has %d vm", vmNum)
		return exeInfo, fmt.Errorf("host has %d vm", vmNum)
	}

	return exeInfo, nil
}
