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
	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/pkg/logs"
)

func (d *Detector) checkProcess(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	// 标准运维-空闲检查的任务链接
	jobUrl := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		jobUrl, err = d.checkIsClear(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	exeInfo = jobUrl
	if err != nil {
		exeInfo = err.Error()
	}

	return attempt, exeInfo, err
}

// checkIsClear 空闲检查
func (d *Detector) checkIsClear(ip string) (string, error) {
	// 根据IP获取主机信息
	hostInfo, err := d.cc.GetHostInfoByIP(d.kt.Ctx, d.kt.Header(), ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:idle check, get host info by ip failed, ip: %s, err: %v", ip, err)
		return "", err
	}

	// 根据bk_host_id，获取bk_biz_id
	bkBizIDs, err := d.cc.GetHostBizIds(d.kt.Ctx, d.kt.Header(), []int64{hostInfo.BkHostId})
	if err != nil {
		logs.Errorf("sops:process:check:idle check process, get host biz id failed, ip: %s, bkHostId: %d, "+
			"err: %v", ip, hostInfo.BkHostId, err)
		return "", err
	}
	bkBizID, ok := bkBizIDs[hostInfo.BkHostId]
	if !ok {
		logs.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
		return "", fmt.Errorf("can not find biz id by host id: %d", hostInfo.BkHostId)
	}

	// 1. create job
	jobId, jobUrl, err := sops.CreateIdleCheckSopsTask(d.kt, d.sops, ip, bkBizID, hostInfo.BkOsType)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkIsClear:failed, host %s failed to check process, bkBizID: %d, err: %v", ip, bkBizID, err)
		return "", fmt.Errorf("failed to check process, err: %v", err)
	}

	// 2. get job status
	if err = sops.CheckTaskStatus(d.kt, d.sops, jobId, bkBizID); err != nil {
		// if host ping death, go ahead to recycle
		if strings.Contains(err.Error(), "ping death") {
			logs.Infof("sops:process:check:host %s ping death, skip check process step, jobId: %d, bkBizID: %d, "+
				"err: %v", ip, jobId, bkBizID, err)
			return "", nil
		}
		logs.Errorf("recycler:logics:cvm:checkIsClear:failed, check job status, host: %s, jobId: %d, bkBizID: %d, err: %v",
			ip, jobId, bkBizID, err)
		return "", fmt.Errorf("host %s failed to check process, job id: %d, bkBizID: %d, err: %v",
			ip, jobId, bkBizID, err)
	}
	return jobUrl, nil
}

func (d *Detector) removeUnusedComment(comment string) string {
	var msg []string
	for _, line := range strings.Split(comment, "\n") {
		if strings.Contains(line, "STATUS") && !strings.Contains(line, "_num:0") {
			index := strings.Index(line, "STATUS")
			msg = append(msg, line[index:])
		}
	}

	return strings.Join(msg, "\n")
}
