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
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/logs"
)

func (d *Detector) checkGCS(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkGCSAndTcaplus(step.IP)
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

func (d *Detector) checkGCSAndTcaplus(ip string) (string, error) {
	exeInfos := make([]string, 0)

	respGcs, err := d.gcs.CheckGCS(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check gcs, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check gcs, err: %v", err)
	}

	gcsRespStr := d.structToStr(respGcs)
	exeInfo := fmt.Sprintf("gcs response: %s", gcsRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respGcs.Code != 0 {
		logs.Infof("failed to check gcs, ip: %s, code: %d, msg: %s", ip, respGcs.Code, respGcs.Message)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check gcs, code: %d, msg: %s", respGcs.Code,
			respGcs.Message)
	}

	if respGcs.Data.RowsNum > 0 || len(respGcs.Data.Detail) > 0 {
		logs.Infof("%s has gcs records, gcs resp: %+v", ip, respGcs)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has %d gcs records", respGcs.Data.RowsNum)
	}

	respTcaplus, err := d.tcaplus.CheckTcaplus(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check tcaplus, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tcaplus, err: %v", err)
	}

	tcaplusRespStr := d.structToStr(respTcaplus)
	exeInfo = fmt.Sprintf("tcaplus response: %s", tcaplusRespStr)
	exeInfos = append(exeInfos, exeInfo)

	cnt := len(respTcaplus.Data)
	if cnt > 0 {
		logs.Infof("%s has tcaplus records, resp: %+v", ip, respTcaplus)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has %d tcaplus records", cnt)
	}

	return strings.Join(exeInfos, "\n"), nil
}
