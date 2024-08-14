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

// Package detector ...
package detector

import (
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/pkg/logs"
)

func (d *Detector) setHostModule(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		err = d.transferHost(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		exeInfo = err.Error()
	}

	return attempt, exeInfo, err
}

func (d *Detector) transferHost(ip string) error {
	hostId, err := d.cc.GetHostId(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to get host id by ip: %s, err: %v", ip, err)
		return err
	}

	srcBizMap, err := d.cc.GetHostBizIds(nil, nil, []int64{hostId})
	if err != nil {
		logs.Errorf("failed to get host id by ip: %s, err: %v", ip, err)
		return err
	}

	srcBiz, ok := srcBizMap[hostId]
	if !ok {
		logs.Errorf("can not find host bizID from cc ip: %s", ip)
		return fmt.Errorf("can not find host bizID from cc ip: %s", ip)
	}

	destBiz := 213
	destModuleId := 16679
	transferReq := &cmdb.TransferHostReq{
		From: cmdb.TransferHostSrcInfo{
			FromBizID: srcBiz,
			HostIDs:   []int64{hostId},
		},
		To: cmdb.TransferHostDstInfo{
			ToBizID:    int64(destBiz),
			ToModuleID: int64(destModuleId),
		},
	}

	resp, err := d.cc.TransferHost(nil, nil, transferReq)
	if err != nil {
		return err
	}

	if resp.Result == false || resp.Code != 0 {
		return fmt.Errorf("failed to transfer host to target business, ip: %s, biz id: %d, code: %d, msg: %s", ip,
			destBiz, resp.Code, resp.ErrMsg)
	}
	return nil
}
