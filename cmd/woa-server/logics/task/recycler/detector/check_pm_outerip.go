/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package detector ...
package detector

import (
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/ngateapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
)

// checkPmOuterIP 标准运维-物理机外网IP回收及清理检查
func (d *Detector) checkPmOuterIP(step *table.DetectStep, retryNum int) (int, string, error) {
	ip := step.IP
	// 根据IP获取主机信息
	hostBase, err := d.getHostBaseInfo([]string{ip})
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkPmOuterIP:failed, failed to get host from cc err: %v, step: %+v",
			err, cvt.PtrToVal(step))
		return 0, "", fmt.Errorf("failed to check pm outer ip: %s, for get host from cc err: %v", ip, err)
	}

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("recycler:logics:cvm:checkPmOuterIP:failed, failed to get invalid host num %d != 1, ip: %s",
			cnt, ip)
		return 0, "", fmt.Errorf("failed to check pm outer ip, for get invalid host num %d != 1, ip: %s", cnt, ip)
	}

	hostInfo := hostBase[0]

	// skip pm outer check if host is not physical machine
	if !hostInfo.IsPmAndOuterIPDevice() {
		logs.Infof("recycler:logics:cvm:checkPmOuterIP:success:CHECK_SKIP, recycle ngate outer ip: %s, hostInfo: %+v",
			step.IP, cvt.PtrToVal(hostInfo))
		return 0, "跳过", nil
	}

	// 标准运维-回收外网IP的任务链接
	jobUrl := ""
	attempt := 0
	exeInfos := make([]string, 0)

	// skip==0表示可以调用标准运维回收IP流程，否则跳过该流程(产品需求-需要支持跳过标准运维的流程，避免因该流程阻塞公司流程)
	if step.Skip == 0 {
		attempt, jobUrl, err = d.recycleSopsTask(step.IP, retryNum)
		if err != nil {
			exeInfos = append(exeInfos, err.Error())
			return attempt, strings.Join(exeInfos, "\n"), err
		}
		exeInfos = append(exeInfos, jobUrl)
	}

	// 调用公司sniper公网IP回收接口
	var ngateExeInfos []string
	rty := retry.NewRetryPolicy(uint(retryNum), [2]uint{3000, 15000})
	err = rty.BaseExec(kit.New(), func() error {
		ngateExeInfos, err = d.recycleNgateIP(step, hostInfo)
		return err
	})
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkPmOuterIP:failed, recycle ngate outer ip: %s, hostOuterIPv4: %s, "+
			"hostOuterIPv6: %s, api return err: %v", step.IP, hostInfo.BkHostOuterIP, hostInfo.BkHostOuterIPv6, err)
		exeInfos = append(exeInfos, err.Error())
		return attempt, strings.Join(exeInfos, "\n"), err
	}

	// 记录ngate执行的信息
	exeInfos = append(exeInfos, ngateExeInfos...)

	return attempt, strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) recycleSopsTask(ip string, retry int) (int, string, error) {
	attempt := 0
	var err error
	var jobUrl string
	for i := 0; i < retry; i++ {
		attempt = i
		jobUrl, err = d.recycleSopsOuterIP(ip)
		if err == nil {
			break
		}

		logs.Errorf("recycler:logics:cvm:checkPmOuterIP:failed, recycle sops outer ip: %s, attempt: %d, "+
			"api return err: %v", ip, attempt, err)

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	return attempt, jobUrl, err
}

func (d *Detector) recycleNgateIP(step *table.DetectStep, hostInfo *cmdb.Host) ([]string, error) {
	exeInfos := make([]string, 0)

	if step.User == "" {
		logs.Errorf("failed to recycle ngate outer ip, for invalid user is empty, step id: %s, stepName: %s",
			step.ID, step.StepName)
		return exeInfos, fmt.Errorf("failed to check pm outerip, for invalid user is empty")
	}

	// 如果外网IP为空，无需处理，直接返回
	if !hostInfo.IsPmAndOuterIPDevice() {
		logs.Errorf("failed to recycle ngate outer ip, for invalid host outer ipv4 or ipv6 is empty, step id: %s, "+
			"stepName: %s, ip: %s, hostInfo: %+v", step.ID, step.StepName, step.IP, cvt.PtrToVal(hostInfo))
		return exeInfos, fmt.Errorf("failed to check pm outerip, for invalid host outer ipv4 or ipv6 is empty")
	}

	for _, ipVersion := range []string{ngateapi.IPv4Version, ngateapi.IPv6Version} {
		addressList := make([]string, 0)
		switch ipVersion {
		case ngateapi.IPv4Version:
			if len(hostInfo.BkHostOuterIP) == 0 {
				continue
			}

			addressList = []string{hostInfo.BkHostOuterIP}
		case ngateapi.IPv6Version:
			if len(hostInfo.BkHostOuterIPv6) == 0 {
				continue
			}

			addressList = []string{hostInfo.BkHostOuterIPv6}
		}

		recycleIPReq := &ngateapi.RecycleIPReq{
			AssertIDList:  []string{hostInfo.BkAssetID},
			DeviceType:    ngateapi.ServerDeviceType,
			AddressList:   addressList,
			IPTypeEnum:    ngateapi.OuterIPType,
			IPVersionEnum: ipVersion,
			User:          step.User,
		}
		recycleIPResp, err := d.ngate.RecycleIP(nil, recycleIPReq)

		recycleIPRespStr := d.structToStr(recycleIPResp)
		ngateReqLogMsg := fmt.Sprintf("ngate recycle outer ip, ipVersion: %s, innerIP: %s, request: %s, response: %s",
			ipVersion, step.IP, d.structToStr(recycleIPReq), recycleIPRespStr)
		exeInfos = append(exeInfos, ngateReqLogMsg)
		logs.Infof("recycler:logics:cvm:checkPmOuterIP:NgateResponse, %s", ngateReqLogMsg)
		if err != nil {
			logs.Errorf("recycler:logics:cvm:checkPmOuterIP:failed, failed to check ngate outer ip, err: %v, "+
				"ipVersion: %s, step: %+v, recycleIPReq: %+v",
				err, ipVersion, cvt.PtrToVal(step), cvt.PtrToVal(recycleIPReq))
			return exeInfos, fmt.Errorf("failed to check pm outer ip: %s, hostOuterIPv4: %s, hostOuterIPv6: %s, "+
				"stepName: %s, err: %v", step.IP, hostInfo.BkHostOuterIP, hostInfo.BkHostOuterIPv6, step.StepName, err)
		}

		if recycleIPResp.ReturnCode != 0 || !recycleIPResp.Success {
			return exeInfos, fmt.Errorf("recycle ngate outer ip: %s, ipVersion: %s, hostOuterIPv4: %s, "+
				"hostOuterIPv6: %s, api return err: %s",
				step.IP, ipVersion, hostInfo.BkHostOuterIP, hostInfo.BkHostOuterIPv6, recycleIPRespStr)
		}
	}

	return exeInfos, nil
}

// recycleSopsOuterIP 回收外网IP
func (d *Detector) recycleSopsOuterIP(ip string) (string, error) {
	// 根据IP获取主机信息
	hostInfo, err := d.cc.GetHostInfoByIP(d.backendKit, ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:recycle outer ip, get host info by ip failed, ip: %s, err: %v", ip, err)
		return "", err
	}

	// 如果外网IP为空，无需处理，直接返回
	if !hostInfo.IsPmAndOuterIPDevice() {
		logs.Errorf("sops:process:check:recycle outer ip, for invalid host outer ipv4 or ipv6 is empty, ip: %s, "+
			"hostInfo: %+v", ip, cvt.PtrToVal(hostInfo))
		return "", fmt.Errorf("failed to check recycle sops outer ip, for invalid host outer ipv4 or ipv6 is empty")
	}

	// 根据bk_host_id，获取bk_biz_id
	bkBizIDs, err := d.cc.GetHostBizIds(d.backendKit, []int64{hostInfo.BkHostID})
	if err != nil {
		logs.Errorf("sops:process:check:recycle outer ip, get host biz id failed, ip: %s, bkHostId: %d, "+
			"err: %v", ip, hostInfo.BkHostID, err)
		return "", err
	}
	bkBizID, ok := bkBizIDs[hostInfo.BkHostID]
	if !ok {
		logs.Errorf("recycleSopsOuterIP:can not find biz id by host id: %d, ip:%s", hostInfo.BkHostID, ip)
		return "", fmt.Errorf("can not find biz id by host id: %d, ip: %s", hostInfo.BkHostID, ip)
	}

	// 1. create job
	jobId, jobUrl, err := sops.CreateRecycleOuterIPSopsTask(d.backendKit, d.sops, ip, bkBizID, hostInfo.BkOsType)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:recycleSopsOuterIP:failed, host %s failed to recycle outer ip, "+
			"bkBizID: %d, err: %v", ip, bkBizID, err)
		return "", fmt.Errorf("failed to recycle outer ip process, ip: %s, err: %v", ip, err)
	}

	// 2. get job status
	if _, err = sops.CheckTaskStatus(d.backendKit, d.sops, jobId, bkBizID); err != nil {
		// if host ping death, go ahead to recycle
		if strings.Contains(err.Error(), "ping death") {
			logs.Infof("sops:process:check:host %s ping death, skip recycle outer ip process step, jobId: %d, "+
				"bkBizID: %d, err: %v", ip, jobId, bkBizID, err)
			return "", nil
		}
		logs.Errorf("recycler:logics:cvm:recycleSopsOuterIP:failed, check job status, host: %s, jobId: %d, "+
			"bkBizID: %d, err: %v", ip, jobId, bkBizID, err)
		return "", fmt.Errorf("host %s failed to recycle outer ip process, job id: %d, bkBizID: %d, err: %v",
			ip, jobId, bkBizID, err)
	}
	return jobUrl, nil
}
