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
	"hcm/pkg/api/core"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkdbm"
)

func (d *Detector) checkDbm(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil
	var isRetry bool

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err, isRetry = d.checkDbmMachinePool(step.IP)
		if err == nil || !isRetry {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	return attempt, exeInfo, err
}

func (d *Detector) checkDbmMachinePool(ip string) (exeInfo string, err error, isRetry bool) {
	// 查询bkdbm的主机池，如果该主机在dbm主机池里面，则不允许回收
	req := &bkdbm.ListMachinePool{IPs: []string{ip}, Offset: 0, Limit: 1}
	resp, err := d.bkDbm.QueryMachinePool(core.NewBackendKit(), req)
	if err != nil {
		logs.Errorf("failed to check bkdbm machine pool, ip: %s, err: %v", ip, err)
		return "", fmt.Errorf("failed to check bkdbm machine pool, err: %v", err), true
	}

	dbmRespStr := d.structToStr(resp)
	exeInfo = fmt.Sprintf("check bkdbm machine pool response: %s", dbmRespStr)

	if len(resp.Results) > 0 {
		logs.Infof("failed to check bkdbm machine pool, ip:%s is in the BK-DBM's machine pool, bkdbm resp: %s",
			ip, dbmRespStr)
		return exeInfo, fmt.Errorf("该主机在DBM中使用，不允许回收"), false
	}

	return exeInfo, nil, false
}
