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

package launcher

import (
	"fmt"

	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg/logs"
	ccapi "hcm/pkg/thirdparty/esb/cmdb"
)

// transferHost2Pool transfer hosts to CR pool module
func (l *Launcher) transferHost2Pool(hostIds []int64, srcBizId int64) error {
	// TODO: get from config
	// transfer hosts to 资源运营服务-CR资源池
	destBiz := types.BizIDPool
	destModule := types.ModuleIDPool
	return l.transferHost(hostIds, srcBizId, destBiz, destModule)
}

func (l *Launcher) transferHost(hostIds []int64, srcBizId, destBizId, destModuleId int64) error {

	transferReq := &ccapi.TransferHostReq{
		From: ccapi.TransferHostSrcInfo{
			FromBizID: srcBizId,
			HostIDs:   hostIds,
		},
		To: ccapi.TransferHostDstInfo{
			ToBizID: destBizId,
		},
	}

	// if destination module id is 0, transfer host to idle module of business
	// otherwise, transfer host to input module
	if destModuleId > 0 {
		transferReq.To.ToModuleID = destModuleId
	}

	resp, err := l.esbCli.Cmdb().TransferHost(nil, nil, transferReq)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("scheduler:cvm:launcher:transferHost:failed, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return fmt.Errorf("failed to transfer host to target business, code: %d, msg: %s", resp.Code, resp.ErrMsg)
	}
	return nil
}
