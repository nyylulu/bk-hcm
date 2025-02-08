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

package itsm

import (
	"fmt"
	"strconv"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway"
)

// ApproveReq define approve req.
type ApproveReq struct {
	Sn       string `json:"sn"`
	StateID  int    `json:"state_id"`
	Approver string `json:"approver"`
	Action   string `json:"action"`
	Remark   string `json:"remark"`
}

// Approve 快捷审批接口。
func (i *itsm) Approve(kt *kit.Kit, req *ApproveReq) error {

	resp := new(apigateway.BaseResponse)
	err := i.client.Post().
		SubResourcef("/approve/").
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if !resp.Result || resp.Code != 0 {
		return fmt.Errorf("approve failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return nil
}

// ApproveNodeOpt 通过/拒绝 ITSM指定节点参数 节点需要为审批节点，自动识别 审批意见、备注字段
type ApproveNodeOpt struct {
	SN       string `json:"itsm_sn"`
	StateId  int64  `json:"state_id"`
	Operator string `json:"operator"`
	// 审批意见
	Approval bool `json:"approval"`
	// 备注
	Remark string `json:"remark"`
}

// ApproveNode 通过/拒绝 ITSM指定节点 自动识别 `审批意见`、`备注` 字段
func (i *itsm) ApproveNode(kt *kit.Kit, param *ApproveNodeOpt) error {

	// 1. 查找节点
	statusResp, err := i.GetTicketStatus(kt, param.SN)
	if err != nil {
		logs.Errorf("fail to get itsm ticket status by sn %s, err: %v, rid: %s", param.SN, err, kt.Rid)
		return err
	}
	status := statusResp.Data
	if status.CurrentStatus != string(StatusRunning) {
		logs.Errorf("try to approve itsm ticket %s, but state is %s, rid: %s", param.SN, status.CurrentStatus,
			kt.Rid)
		return fmt.Errorf("itsm state not RUNNING, but %s", status.CurrentStatus)
	}

	var curStep *TicketStep
	for i := range status.CurrentSteps {
		step := status.CurrentSteps[i]
		if step.StateId == param.StateId {
			curStep = step
			break
		}
	}
	if curStep == nil {
		logs.Errorf("fail to approve itsm ticket %s, state not found, state: %s, rid: %s",
			param.SN, status.CurrentStatus, kt.Rid)
		return fmt.Errorf("fail to approve, state not found")
	}

	req := &OperateNodeReq{
		Sn:         param.SN,
		StateId:    param.StateId,
		Operator:   param.Operator,
		ActionType: ActionTypeTransition,
		Fields:     make([]*TicketField, 0),
	}

	var approvalResultField, remarkField *TicketField
	// 查找field
	for i := range curStep.Fields {
		field := curStep.Fields[i]
		switch {
		// 审批意见
		case field.Name == FieldNameApprovalResultName:
			approvalResultField = &TicketField{
				Key:   field.Key,
				Value: strconv.FormatBool(param.Approval),
			}
		// 备注
		case field.Name == FieldNameApprovalRemarkName:
			// 可能有多个`备注`字段，填充所有
			remarkField = &TicketField{
				Key:   field.Key,
				Value: param.Remark,
			}
		default:
			// 	ignore other fields
		}
	}
	if approvalResultField == nil {
		logs.Errorf("fail to approve, approval result field not found, fields: %+v, param: %+v, rid: %s",
			curStep.Fields, param, kt.Rid)
		return fmt.Errorf("fail to approve, approval result field not found")
	}
	req.Fields = append(req.Fields, approvalResultField)

	if remarkField == nil && param.Remark != "" {
		logs.Errorf("fail to approve, remark field not found, fields: %+v, param: %+v, rid: %s",
			curStep.Fields, param, kt.Rid)
		return fmt.Errorf("fail to approve, remark field not found")
	}
	if remarkField != nil {
		req.Fields = append(req.Fields, remarkField)
	}

	resp, err := i.OperateNode(kt, req)
	if err != nil {
		logs.Errorf("failed to call itsm to approve ticket, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return err
	}

	if resp.Code != 0 {
		logs.Errorf("failed to approve itsm ticket, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kt.Rid)
		return fmt.Errorf("failed to approve itsm ticket, sn: %s, code: %d, msg: %s",
			param.SN, resp.Code, resp.ErrMsg)
	}

	return nil
}

// FieldNameApprovalResultName ITSM 审批意见
const FieldNameApprovalResultName = "审批意见"

// FieldNameApprovalRemarkName ITSM 备注
const FieldNameApprovalRemarkName = "备注"
