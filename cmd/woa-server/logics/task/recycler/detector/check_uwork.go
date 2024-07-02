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
	"strings"
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/thirdparty/xshipapi"
	"hcm/pkg/logs"
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
		logs.Errorf("failed to check uwork ticket, for invalid user is empty, step id: %s", step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork, for invalid user is empty")
	}

	respTicket, err := d.uwork.CheckUworkTicket(nil, nil, step.User, step.IP)
	if err != nil {
		logs.Errorf("failed to check uwork ticket, err: %v, step id: %s", err, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork, err: %v", err)
	}

	ticketRespStr := d.structToStr(respTicket)
	exeInfo := fmt.Sprintf("uwork ticket response: %s", ticketRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respTicket.Return != 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("check uwork ticket api return err: %s", respTicket.Detail)
	}

	ticketIds := make([]string, 0)
	for _, ticket := range respTicket.Data {
		if ticket.IsEnd == "0" {
			ticketIds = append(ticketIds, ticket.TicketNo)
		}
	}

	if len(ticketIds) > 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has uwork tickets: %s", strings.Join(ticketIds, ";"))
	}

	filter := &mapstr.MapStr{
		"suborder_id": step.SuborderID,
		"ip":          step.IP,
	}
	host, err := dao.Set().RecycleHost().GetRecycleHost(context.Background(), filter)
	if err != nil {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to get host asset id: err: %s", err.Error())
	}

	respProcess, err := d.xship.GetXServerProcess(nil, nil, host.AssetID)
	if err != nil {
		logs.Errorf("failed to check uwork process, err: %v, step id: %s", err, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check uwork process, err: %v", err)
	}

	processRespStr := d.structToStr(respProcess)
	exeInfoProcess := fmt.Sprintf("uwork process response: %s", processRespStr)
	exeInfos = append(exeInfos, exeInfoProcess)

	if respProcess.Code != xshipapi.CodeSuccess {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("check uwork process api return err: %s", respProcess.Message)
	}

	if respProcess.Data == nil {
		return strings.Join(exeInfos, "\n"), errors.New("check uwork process api return data is nil")
	}

	processes := make([]string, 0)
	for _, process := range respProcess.Data.Processes {
		processes = append(processes, fmt.Sprintf("%d(%s)", process.ID, process.Name))
	}

	if len(processes) > 0 {
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has uwork process: %s", strings.Join(processes, ";"))
	}

	return strings.Join(exeInfos, "\n"), nil
}
