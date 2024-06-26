/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package matcher provide the matcher for task
package matcher

import (
	"fmt"

	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
)

// transferHostAndSetOperator transfer host and set operator in cc 3.0
func (m *Matcher) transferHostAndSetOperator(info *types.DeviceInfo, order *types.ApplyOrder) error {
	hostId, err := m.cc.GetHostId(nil, nil, info.Ip)
	if err != nil {
		logs.Errorf("failed to get host id by ip: %s, err: %v", info.Ip, err)
		return err
	}

	if err := m.transferHost(info, hostId, order.BkBizId); err != nil {
		logs.Errorf("failed to transfer host, ip: %s, err: %v", info.Ip, err)
		return err
	}

	if err := m.updateHostOperator(info, hostId, order.User); err != nil {
		logs.Errorf("failed to update host operator, ip: %s, err: %v", info.Ip, err)
		return err
	}

	return nil
}

// transferHost transfer host to target business in cc 3.0
func (m *Matcher) transferHost(info *types.DeviceInfo, hostId, bizId int64) error {
	transferReq := &cmdb.TransferHostReq{
		From: cmdb.TransferHostSrcInfo{
			// TODO: use const
			FromBizID: 931,
			HostIDs:   []int64{hostId},
		},
		To: cmdb.TransferHostDstInfo{
			ToBizID: bizId,
		},
	}

	resp, err := m.cc.TransferHost(nil, nil, transferReq)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		return fmt.Errorf("failed to transfer host to target business, ip: %s, biz id: %d, code: %d, msg: %s", info.Ip,
			bizId, resp.Code, resp.ErrMsg)
	}

	return nil
}

// updateHostOperator update host operator in cc 3.0
func (m *Matcher) updateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error {
	update := &cmdb.UpdateHostProperty{
		HostID: hostId,
		Properties: map[string]interface{}{
			"operator":        operator,
			"bk_bak_operator": operator,
		},
	}
	req := &cmdb.UpdateHostsReq{
		Update: []*cmdb.UpdateHostProperty{
			update,
		},
	}

	resp, err := m.cc.UpdateHosts(nil, nil, req)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		return fmt.Errorf("failed to update host operator, ip: %s, code: %d, msg: %s", info.Ip, resp.Code, resp.ErrMsg)
	}

	return nil
}

func (m *Matcher) getBizName(bizId int64) string {
	req := &cmdb.SearchBizReq{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_id",
						Operator: querybuilder.OperatorEqual,
						Value:    bizId,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	resp, err := m.cc.SearchBiz(nil, nil, req)
	if err != nil {
		logs.Warnf("failed to get cc business info, err: %v", err)
		return ""
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Warnf("failed to get cc business info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return ""
	}

	cnt := len(resp.Data.Info)
	if cnt != 1 {
		logs.Warnf("get invalid cc business info count %d != 1", cnt)
		return ""
	}

	return resp.Data.Info[0].BkBizName
}
