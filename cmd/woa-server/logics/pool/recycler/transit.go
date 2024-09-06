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

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	ccapi "hcm/cmd/woa-server/thirdparty/esb/cmdb"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/logs"
)

func (r *Recycler) dealTransitTask(task *table.RecallDetail) error {
	// transfer hosts from 资源运营服务-CR资源下架中 to 资源运营服务-SA云化池
	if err := r.transferHost(task.HostID, types.BizIDPool, types.BizIDPool, types.ModuleIDPoolMatch); err != nil {
		logs.Errorf("failed to transfer host %d, err: %v", task.HostID, err)

		errUpdate := r.updateTaskTransitStatus(task, err.Error(), table.RecallStatusTransitFailed)
		if errUpdate != nil {
			logs.Warnf("failed to update recall task status, err: %v", errUpdate)
		}

		return err
	}

	// update task status
	if err := r.updateTaskTransitStatus(task, "", table.RecallStatusDone); err != nil {
		logs.Errorf("failed to update recall task status, err: %v", err)
		return err
	}

	return nil
}

// transferHost transfer host to target business in cc 3.0
func (r *Recycler) transferHost(hostID, fromBizID, toBizID, toModuleId int64) error {
	transferReq := &ccapi.TransferHostReq{
		From: ccapi.TransferHostSrcInfo{
			FromBizID: fromBizID,
			HostIDs:   []int64{hostID},
		},
		To: ccapi.TransferHostDstInfo{
			ToBizID: toBizID,
		},
	}

	// if destination module id is 0, transfer host to idle module of business
	// otherwise, transfer host to input module
	if toModuleId > 0 {
		transferReq.To.ToModuleID = toModuleId
	}

	resp, err := r.esbCli.Cmdb().TransferHost(nil, nil, transferReq)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		return fmt.Errorf("failed to transfer host from biz %d to %d, host id: %d, code: %d, msg: %s", fromBizID,
			toBizID, hostID, resp.Code, resp.ErrMsg)
	}

	return nil
}

func (r *Recycler) updateTaskTransitStatus(task *table.RecallDetail, msg string, status table.RecallStatus) error {

	filter := map[string]interface{}{
		"id": task.ID,
	}

	now := time.Now()
	update := map[string]interface{}{
		"status":    status,
		"update_at": now,
	}

	if msg != "" {
		update["message"] = msg
	}

	if err := dao.Set().RecallDetail().UpdateRecallDetail(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}
