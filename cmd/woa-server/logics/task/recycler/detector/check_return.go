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

// Package detector ...
package detector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

func (d *Detector) checkReturn(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkIsReturning(step)
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

func (d *Detector) checkIsReturning(step *table.DetectStep) (string, error) {
	exeInfos := make([]string, 0)

	filter := &mapstr.MapStr{
		"suborder_id": step.SuborderID,
		"ip":          step.IP,
	}
	host, err := dao.Set().RecycleHost().GetRecycleHost(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get recycle host %s, err: %v", step.IP, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to get recycle host %s, err: %v", step.IP, err)
	}

	switch host.ResourceType {
	case table.ResourceTypeCvm:
		return d.checkCvmReturn(host.AssetID)
	case table.ResourceTypePm:
		return d.checkErpReturn(host.AssetID)
	default:
		return strings.Join(exeInfos, "\n"), nil
	}
}

func (d *Detector) checkCvmReturn(assetId string) (string, error) {
	exeInfos := make([]string, 0)

	if len(assetId) == 0 {
		logs.Errorf("failed to check cvm return process, for invalid asset id is empty")
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check cvm return process, for invalid asset id is " +
			"empty")
	}

	req := &cvmapi.GetCvmProcessReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmGetProcessMethod,
		},
		Params: &cvmapi.GetCvmProcessParam{
			AssetIds: []string{assetId},
		},
	}

	resp, err := d.cvm.GetCvmProcess(nil, nil, req)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkCvmReturn:failed, failed to check cvm return process, "+
			"err: %v, assetId: %s, req: %+v", err, assetId, cvt.PtrToVal(req))
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check cvm return process, err: %v", err)
	}

	respStr := d.structToStr(resp)
	exeInfo := fmt.Sprintf("yunti response: %s", respStr)
	exeInfos = append(exeInfos, exeInfo)

	if resp.Error.Code != 0 {
		logs.Errorf("recycler:logics:cvm:checkCvmReturn:failed, failed to check cvm return process, code: %d, msg: %s",
			resp.Error.Code, resp.Error.Message)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("check return process api return err: %s", resp.Error.Message)
	}

	processes := make([]string, 0)
	for _, item := range resp.Result.Data {
		if item.AssetId == assetId && len(item.StatusDesc) > 0 {
			process := fmt.Sprintf("%s(%s)", item.OrderId, item.StatusDesc)
			processes = append(processes, process)
		}
	}
	if len(processes) > 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has cvm process: %s", strings.Join(processes, ";"))
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) checkErpReturn(assetId string) (string, error) {
	exeInfos := make([]string, 0)

	if len(assetId) == 0 {
		logs.Errorf("failed to check erp return process, for invalid asset id is empty")
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check erp return process, for invalid asset id is " +
			"empty")
	}

	req := &cvmapi.GetErpProcessReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.GetErpProcessMethod,
		},
		Params: &cvmapi.GetErpProcessParam{
			AssetIds: []string{assetId},
		},
	}

	resp, err := d.cvm.GetErpProcess(nil, nil, req)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkErpReturn:failed, failed to check erp return process, "+
			"err: %v, assetId: %s, req: %+v", err, assetId, cvt.PtrToVal(req))
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check erp return process, err: %v", err)
	}

	respStr := d.structToStr(resp)
	exeInfo := fmt.Sprintf("yunti response: %s", respStr)
	exeInfos = append(exeInfos, exeInfo)

	if resp.Error.Code != 0 {
		logs.Errorf("recycler:logics:cvm:checkErpReturn:failed, failed to check erp return process, code: %d, msg: %s",
			resp.Error.Code, resp.Error.Message)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("check return process api return err: %s", resp.Error.Message)
	}

	for _, item := range resp.Result.Data {
		if item.AssetId == assetId && item.ActionType == "退回" {
			return strings.Join(exeInfos, "\n"), fmt.Errorf("has erp return order %s", item.OrderId)
		}
	}

	return strings.Join(exeInfos, "\n"), nil
}
