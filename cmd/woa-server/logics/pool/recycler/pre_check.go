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

package recycler

import (
	"context"
	"fmt"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	ccapi "hcm/cmd/woa-server/thirdparty/esb/cmdb"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/logs"
)

func (r *Recycler) dealPreCheckTask(task *table.RecallDetail) error {
	// get host from module 资源运营服务-CR资源下架中
	host, err := r.getHostByIDFromPool(task.HostID)
	if err != nil {
		logs.Errorf("failed to get host by id from pool, err: %v", task.HostID, err)

		errUpdate := r.updateTaskPreCheckStatus(task, "", "", err.Error(), table.RecallStatusPreCheckFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// update task status
	err = r.updateTaskPreCheckStatus(task, host.BkHostInnerIp, host.BkAssetId, "", table.RecallStatusClearChecking)
	if err != nil {
		logs.Errorf("failed to update recall task status, err: %v", err)
		return err
	}

	go func() {
		r.Add(task.ID)
	}()

	return nil
}

// getHostFromPool get hosts by ID from resource pool
func (r *Recycler) getHostByIDFromPool(hostID int64) (*ccapi.HostInfo, error) {
	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules: []querybuilder.Rule{
			querybuilder.AtomRule{
				Field:    "bk_host_id",
				Operator: querybuilder.OperatorEqual,
				Value:    hostID,
			},
		},
	}

	req := &ccapi.ListBizHostReq{
		// check whether host is in module 资源运营服务-CR资源下架中
		BkBizId:     types.BizIDPool,
		BkModuleIds: []int64{types.ModuleIDPoolRecalling},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
		},
		Page: ccapi.BasePage{
			Start: 0,
			Limit: common.BKMaxInstanceLimit,
		},
	}
	if len(rule.Rules) > 0 {
		req.HostPropertyFilter = &querybuilder.QueryFilter{
			Rule: rule,
		}
	}

	resp, err := r.esbCli.Cmdb().ListBizHost(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	num := len(resp.Data.Info)
	if num != 1 {
		err := fmt.Errorf("get invalid hosts num %d != 1", num)
		logs.Errorf("get invalid hosts num %d != 1", num)
		return nil, err
	}

	host := resp.Data.Info[0]
	if host.BkHostId != hostID {
		err := fmt.Errorf("get invalid host id %d != targe %d", host.BkHostId, hostID)
		logs.Errorf("get invalid host id %d != targe %d", host.BkHostId, hostID)
		return nil, err
	}

	return host, nil
}

func (r *Recycler) updateTaskPreCheckStatus(task *table.RecallDetail, ip, asset, msg string,
	status table.RecallStatus) error {

	filter := map[string]interface{}{
		"id": task.ID,
	}

	now := time.Now()
	update := map[string]interface{}{
		"status":    status,
		"update_at": now,
	}

	if ip != "" {
		update["labels.ip"] = ip
	}

	if asset != "" {
		update["labels.bk_asset_id"] = asset
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
