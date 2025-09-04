/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for
 * the specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package actionlb

import (
	"fmt"
	"hcm/pkg/api/data-service/task"

	actcli "hcm/cmd/task-server/logics/action/cli"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

var _ action.Action = new(CreateTargetGroupWithRelAction)
var _ action.ParameterAction = new(CreateTargetGroupWithRelAction)

// CreateTargetGroupWithRelAction 创建目标组并绑定监听器，同时添加RS
type CreateTargetGroupWithRelAction struct{}

// CreateTargetGroupWithRelOption 创建目标组并绑定监听器，同时添加RS参数
type CreateTargetGroupWithRelOption struct {
	Vendor              enumor.Vendor        `json:"vendor" validate:"required"`
	LoadBalancerID      string               `json:"lb_id" validate:"required"`
	ListenerID          string               `json:"listener_id" validate:"required"`
	ListenerRuleID      string               `json:"listener_rule_id" validate:"omitempty"`
	RuleType            enumor.RuleType      `json:"rule_type" validate:"required"`
	Targets             []*corelb.BaseTarget `json:"targets" validate:"required"`
	ManagementDetailIDs []string             `json:"management_detail_ids" validate:"required"`
}

// Validate 验证参数
func (opt CreateTargetGroupWithRelOption) Validate() error {

	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	for i, target := range opt.Targets {
		if target.InstType == "" {
			return fmt.Errorf("target[%d] inst_type is required", i)
		}
		if target.Port <= 0 {
			return fmt.Errorf("target[%d] port must be greater than 0", i)
		}

		if target.InstType == enumor.EniInstType {
			if target.IP == "" {
				return fmt.Errorf("target[%d] ip is required for ENI type", i)
			}
		} else if target.InstType == enumor.CvmInstType {
			if target.CloudInstID == "" {
				return fmt.Errorf("target[%d] cloud_inst_id is required for CVM type", i)
			}
		}
	}

	return nil
}

// ParameterNew return request params.
func (act CreateTargetGroupWithRelAction) ParameterNew() (params any) {
	return new(CreateTargetGroupWithRelOption)
}

// Name return action name
func (act CreateTargetGroupWithRelAction) Name() enumor.ActionName {
	return enumor.ActionCreateTargetGroupWithRel
}

// Run 执行创建目标组并绑定监听器，同时添加RS
func (act CreateTargetGroupWithRelAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*CreateTargetGroupWithRelOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type not match")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	logs.Infof("start create target group with rel, vendor: %s, listener: %s, targets count: %d, rid: %s",
		opt.Vendor, opt.ListenerID, len(opt.Targets), kt.Kit().Rid)

	switch opt.Vendor {
	case enumor.TCloud:
		return act.createTCloudTargetGroupWithRel(kt, opt)
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", opt.Vendor)
	}
}

// createTCloudTargetGroupWithRel 创建TCloud目标组并绑定监听器，同时添加RS
func (act CreateTargetGroupWithRelAction) createTCloudTargetGroupWithRel(kt run.ExecuteKit, opt *CreateTargetGroupWithRelOption) (any, error) {
	ruleType := enumor.Layer7RuleType
	if opt.ListenerRuleID == "" {
		ruleType = enumor.Layer4RuleType
		logs.Infof("detected Layer4 listener, no rule ID needed, rid: %s", kt.Kit().Rid)
	} else {
		logs.Infof("detected Layer7 listener with rule ID: %s, rid: %s", opt.ListenerRuleID, kt.Kit().Rid)
	}

	targets := act.convertTargetsToRegisterTargets(opt.Targets)
	logs.Infof("converted %d targets to register targets, rid: %s", len(targets), kt.Kit().Rid)

	createReq := &hclb.CreateTargetGroupWithRelReq{
		Vendor:              opt.Vendor,
		LoadBalancerID:      opt.LoadBalancerID,
		ListenerID:          opt.ListenerID,
		ListenerRuleID:      opt.ListenerRuleID,
		RuleType:            ruleType,
		Targets:             targets,
		ManagementDetailIDs: opt.ManagementDetailIDs,
	}

	logs.Infof("creating target group with rel, req: %+v, rid: %s", createReq, kt.Kit().Rid)
	result, err := actcli.GetHCService().TCloud.Clb.CreateTargetGroupWithRel(kt.Kit(), createReq)
	if err != nil {
		logs.Errorf("create tcloud target group with rel failed, req: %+v, err: %v, rid: %s",
			createReq, err, kt.Kit().Rid)
		return nil, fmt.Errorf("create target group with rel failed: %w", err)
	}

	logs.Infof("successfully created target group with rel, target group ID: %s, rid: %s",
		result.TargetGroupID, kt.Kit().Rid)

	err = act.updateTaskDetails(kt, opt.ManagementDetailIDs, result.TargetGroupID, "success")
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Kit().Rid)
	}

	return &CreateTargetGroupWithRelResult{
		TargetGroupID: result.TargetGroupID,
		ListenerID:    opt.ListenerID,
		RuleID:        opt.ListenerRuleID,
		TargetsCount:  len(opt.Targets),
		Status:        "success",
		Message: fmt.Sprintf("Successfully created target group %s and bound %d RS",
			result.TargetGroupID, len(opt.Targets)),
	}, nil
}

// convertTargetsToRegisterTargets
func (act CreateTargetGroupWithRelAction) convertTargetsToRegisterTargets(targets []*corelb.BaseTarget) []*hclb.RegisterTarget {
	result := make([]*hclb.RegisterTarget, 0, len(targets))

	for i, target := range targets {
		registerTarget := &hclb.RegisterTarget{
			TargetType: target.InstType,
			Port:       target.Port,
			Weight:     target.Weight,
		}

		switch target.InstType {
		case enumor.EniInstType:
			registerTarget.EniIp = target.IP
			logs.V(4).Infof("converted ENI target[%d]: IP=%s, Port=%d, Weight=%v",
				i, target.IP, target.Port, target.Weight)

		case enumor.CvmInstType:
			registerTarget.CloudInstID = target.CloudInstID
			registerTarget.InstName = target.InstName
			registerTarget.PrivateIPAddress = target.PrivateIPAddress
			registerTarget.PublicIPAddress = target.PublicIPAddress
			registerTarget.Zone = target.Zone
			logs.V(4).Infof("converted CVM target[%d]: ID=%s, Name=%s, Port=%d, Weight=%v",
				i, target.CloudInstID, target.InstName, target.Port, target.Weight)

		default:
			logs.Warnf("unknown target type: %s for target[%d], rid: %s", target.InstType, i, "unknown")
		}

		result = append(result, registerTarget)
	}

	return result
}

// updateTaskDetails 更新任务详情状态
func (act CreateTargetGroupWithRelAction) updateTaskDetails(kt run.ExecuteKit, detailIDs []string, targetGroupID string, status string) error {
	items := make([]task.UpdateTaskDetailField, 0, len(detailIDs))

	for _, detailID := range detailIDs {
		item := task.UpdateTaskDetailField{
			ID:    detailID,
			State: enumor.TaskDetailSuccess,
			Result: fmt.Sprintf("Target group created and RS bound successfully. Target Group ID: %s, Status: %s",
				targetGroupID, status),
		}
		items = append(items, item)
	}

	updateReq := &task.UpdateDetailReq{
		Items: items,
	}

	err := actcli.GetDataService().Global.TaskDetail.Update(kt.Kit(), updateReq)
	if err != nil {
		logs.Errorf("update task detail failed, detail IDs: %v, err: %v, rid: %s",
			detailIDs, err, kt.Kit().Rid)
		return err
	}

	logs.Infof("successfully updated %d task details, target group ID: %s, rid: %s",
		len(detailIDs), targetGroupID, kt.Kit().Rid)
	return nil
}

// CreateTargetGroupWithRelResult 创建目标组并绑定RS的结果
type CreateTargetGroupWithRelResult struct {
	TargetGroupID string `json:"target_group_id"` // 创建的目标组ID
	ListenerID    string `json:"listener_id"`     // 监听器ID
	RuleID        string `json:"rule_id"`         // 规则ID（七层）
	TargetsCount  int    `json:"targets_count"`   // RS数量
	Status        string `json:"status"`          // 操作状态
	Message       string `json:"message"`
}
