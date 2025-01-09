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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/xshipapi"
)

func (d *Detector) checkUwork(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkUworkPass(step)
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

func (d *Detector) checkUworkPass(step *table.DetectStep) (string, error) {
	exeInfos := make([]string, 0)

	if step.User == "" {
		logs.Errorf("failed to check uwork-xray ticket, for invalid user is empty, step id: %s", step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork-xray, for invalid user is empty")
	}

	filter := &mapstr.MapStr{
		"suborder_id": step.SuborderID,
		"ip":          step.IP,
	}
	host, err := dao.Set().RecycleHost().GetRecycleHost(context.Background(), filter)
	if err != nil {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to get host asset id, err: %s, subOrderID: %s, "+
			"stepIP: %s", err.Error(), step.SuborderID, step.IP)
	}

	// 获取尚未结单的故障单
	respTicket, err := d.xray.CheckXrayFaultTickets(nil, nil, host.AssetID, enumor.XrayFaultTicketNotEnd)
	if err != nil {
		logs.Errorf("failed to check uwork-xray ticket, err: %v, stepIP: %s, assetID: %s, step id: %s",
			err, step.IP, host.AssetID, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork-xray, err: %v", err)
	}

	ticketRespStr := d.structToStr(respTicket)
	exeInfo := fmt.Sprintf("uwork-xray ticket response: %s", ticketRespStr)
	exeInfos = append(exeInfos, exeInfo)

	ticketIds := make([]string, 0)
	for _, ticket := range respTicket.Data {
		if ticket.IsEnd == enumor.XrayFaultTicketNotEnd {
			ticketIds = append(ticketIds, strconv.Itoa(ticket.InstanceID))
		}
	}

	if len(ticketIds) > 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has uwork-xray tickets: %s", strings.Join(ticketIds, ";"))
	}

	respProcess, err := d.xship.GetXServerProcess(nil, nil, host.AssetID)
	if err != nil {
		logs.Errorf("failed to check uwork-xship process, err: %v, stepIP: %s, assetID: %s, step id: %s",
			err, step.IP, host.AssetID, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork-xship process, err: %v", err)
	}

	processRespStr := d.structToStr(respProcess)
	exeInfoProcess := fmt.Sprintf("uwork-xship process response: %s", processRespStr)
	exeInfos = append(exeInfos, exeInfoProcess)

	if respProcess.Code != xshipapi.CodeSuccess {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("check uwork-xship process api return err: %s",
			respProcess.Message)
	}

	if respProcess.Data == nil {
		return strings.Join(exeInfos, "\n"), errors.New("check uwork-xship process api return data is nil")
	}

	processes := make([]string, 0)
	for _, process := range respProcess.Data.Processes {
		processes = append(processes, fmt.Sprintf("%d(%s)", process.ID, process.Name))
	}

	if len(processes) > 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has uwork-xship process: %s", strings.Join(processes, ";"))
	}

	return strings.Join(exeInfos, "\n"), nil
}
