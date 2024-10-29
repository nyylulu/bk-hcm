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

// Package recycler ...
package recycler

import (
	"errors"
	"fmt"
	"strings"

	"hcm/pkg"
	"hcm/pkg/logs"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/thirdparty/esb/cmdb"
)

func (r *Recycler) getIpByHostID(hostID int64) (string, error) {
	hostInfos, err := r.getHostBaseInfo([]int64{hostID})
	if err != nil {
		logs.Errorf("failed to get host base info, err: %v", err)
		return "", err
	}

	cnt := len(hostInfos)
	if cnt != 1 {
		logs.Errorf("get unexpected host base info, for count %d != 1", cnt)
		return "", fmt.Errorf("get unexpected host base info, for count %d != 1", cnt)
	}

	if hostInfos[0].BkHostID != hostID {
		logs.Errorf("get unexpected host base info, for return host id %d != target %d", hostInfos[0].BkHostID, hostID)
		return "", fmt.Errorf("get unexpected host base info, for return host id %d != target %d",
			hostInfos[0].BkHostID, hostID)
	}

	ip := r.getUniqIp(hostInfos[0].BkHostInnerIP)

	return ip, nil
}

// getHostBaseInfo get host base info in cc 3.0
func (r *Recycler) getHostBaseInfo(hostIds []int64) ([]*cmdb.Host, error) {
	if len(hostIds) == 0 {
		return nil, errors.New("host id list is empty")
	}

	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionOr,
		Rules: []querybuilder.Rule{
			querybuilder.AtomRule{
				Field:    "bk_host_id",
				Operator: querybuilder.OperatorIn,
				Value:    hostIds,
			},
		},
	}

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: rule,
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

	resp, err := r.esbCli.Cmdb().ListHost(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// getUniqIp get CC host unique inner ip
func (r *Recycler) getUniqIp(ips string) string {
	// when CC host has multiple inner ips, bk_host_innerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(ips, ",")
	if len(multiIps) == 0 {
		return ""
	}

	return multiIps[0]
}
