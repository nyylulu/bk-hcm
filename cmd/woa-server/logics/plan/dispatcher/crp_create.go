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

package dispatcher

import (
	"errors"
	"fmt"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// createAddCrpTicket create add crp ticket.
func (c *CrpTicketCreator) createAddCrpTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) (string, error) {
	addReq, err := c.constructAddReq(kt, ticket)
	if err != nil {
		logs.Errorf("failed to construct add cvm & cbs plan order request, err: %v, ticket_id: %s, rid: %s", err,
			ticket.ID, kt.Rid)
		return "", err
	}

	resp, err := c.crpCli.AddCvmCbsPlan(kt.Ctx, kt.Header(), addReq)
	if err != nil {
		logs.Errorf("failed to add cvm & cbs plan order, err: %v, ticket_id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to add cvm & cbs plan order, code: %d, msg: %s, crp_tran_id: %s, ticket_id: %s, rid: %s",
			resp.Error.Code,
			resp.Error.Message, resp.TraceId, ticket.ID, kt.Rid)
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to add cvm & cbs plan order, for return empty order id, ticket_id: %s, rid: %s", ticket.ID,
			kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// constructAddReq construct cvm cbs plan add request.
func (c *CrpTicketCreator) constructAddReq(kt *kit.Kit, ticket *ptypes.TicketInfo) (*cvmapi.AddCvmCbsPlanReq, error) {
	addReq := &cvmapi.AddCvmCbsPlanReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAddMethod,
		},
		Params: &cvmapi.AddCvmCbsPlanParam{
			Operator: ticket.Applicant,
			DeptName: ticket.VirtualDeptName,
			Desc:     "",
			Items:    make([]*cvmapi.AddPlanItem, 0),
		},
	}

	switch ticket.DemandClass {
	case enumor.DemandClassCVM:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCvmDesc
	case enumor.DemandClassCA:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCADesc
	default:
		logs.Warnf("failed to construct add desc, unsupported demand class: %s, rid: %s", ticket.DemandClass, kt.Rid)
	}

	for _, demand := range ticket.Demands {
		if demand.Updated == nil {
			logs.Errorf("failed to create add crp ticket, demand updated is nil, rid: %s", kt.Rid)
			return nil, errors.New("demand updated is nil")
		}

		planItem := &cvmapi.AddPlanItem{
			UseTime:         demand.Updated.ExpectTime,
			ProjectName:     string(demand.Updated.ObsProject),
			PlanProductName: ticket.PlanProductName,
			ProductName:     ticket.OpProductName,
			CityName:        demand.Updated.RegionName,
			ZoneName:        demand.Updated.ZoneName,
			CoreTypeName:    demand.Updated.Cvm.CoreType,
			InstanceModel:   demand.Updated.Cvm.DeviceType,
			CvmAmount:       demand.Updated.Cvm.Os.InexactFloat64(),
			CoreAmount:      int(demand.Updated.Cvm.CpuCore),
			Desc:            demand.Updated.Remark,
			InstanceIO:      int(demand.Updated.Cbs.DiskIo),
			DiskTypeName:    demand.Updated.Cbs.DiskType.Name(),
			DiskAmount:      int(demand.Updated.Cbs.DiskSize),
		}

		addReq.Params.Items = append(addReq.Params.Items, planItem)
	}

	return addReq, nil
}
