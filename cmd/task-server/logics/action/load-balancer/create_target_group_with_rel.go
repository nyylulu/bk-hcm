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
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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

// CreateTargetGroupWithRelAction 创建目标组并绑定监听器
type CreateTargetGroupWithRelAction struct{}

// CreateTargetGroupWithRelOption 创建目标组及其关联关系的选项结构体
type CreateTargetGroupWithRelOption struct {
	Vendor              enumor.Vendor          `json:"vendor" validate:"required"`
	LoadBalancerID      string                 `json:"lb_id" validate:"required"`
	ListenerCloudID     string                 `json:"listener_cloud_id" validate:"required"`
	RuleCloudID         string                 `json:"rule_cloud_id" validate:"required"`
	Targets             []*hclb.RegisterTarget `json:"targets" validate:"required"`
	ManagementDetailIDs []string               `json:"management_detail_ids" validate:"required"`
}

// Validate 验证
func (opt CreateTargetGroupWithRelOption) Validate() error {

	switch opt.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for create target group with rel: %s", opt.Vendor)
	}

	if len(opt.Targets) != len(opt.ManagementDetailIDs) {
		return errf.Newf(errf.InvalidParameter, "targets and management_detail_ids length not match, %d != %d",
			len(opt.Targets), len(opt.ManagementDetailIDs))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act CreateTargetGroupWithRelAction) ParameterNew() (params any) {
	return new(CreateTargetGroupWithRelOption)
}

// Name return action name
func (act CreateTargetGroupWithRelAction) Name() enumor.ActionName {
	return enumor.ActionCreateTargetGroupWithRel
}

// Run 执行创建目标组和关联关系
func (act CreateTargetGroupWithRelAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*CreateTargetGroupWithRelOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type not match")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch opt.Vendor {
	case enumor.TCloud:
		return act.createTCloudTargetGroupWithRel(kt, opt)
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", opt.Vendor)
	}
}

// createTCloudTargetGroupWithRel 创建TCloud目标组并绑定监听器
func (act CreateTargetGroupWithRelAction) createTCloudTargetGroupWithRel(kt run.ExecuteKit, opt *CreateTargetGroupWithRelOption) (any, error) {

	ruleType := enumor.Layer7RuleType
	if opt.RuleCloudID == "" {
		ruleType = enumor.Layer4RuleType
	}

	req := &hclb.CreateTargetGroupWithRelReq{
		Vendor:              opt.Vendor,
		LoadBalancerID:      opt.LoadBalancerID,
		ListenerID:          opt.ListenerCloudID,
		ListenerRuleID:      opt.RuleCloudID,
		RuleType:            ruleType,
		Targets:             opt.Targets,
		ManagementDetailIDs: opt.ManagementDetailIDs,
	}

	result, err := actcli.GetHCService().TCloud.Clb.CreateTargetGroupWithRel(kt.Kit(), req)
	if err != nil {
		logs.Errorf("create tcloud target group with rel failed, req: %+v, err: %v, rid: %s", req, err, kt.Kit().Rid)
		return nil, err
	}
	err = act.updateTaskDetails(kt, opt.ManagementDetailIDs, result.TargetGroupID)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}

	return result, nil
}

// updateTaskDetails 更新任务详情状态
func (act CreateTargetGroupWithRelAction) updateTaskDetails(kt run.ExecuteKit, detailIDs []string, targetGroupID string) error {
	items := make([]task.UpdateTaskDetailField, 0, len(detailIDs))
	for _, detailID := range detailIDs {
		items = append(items, task.UpdateTaskDetailField{
			ID:     detailID,
			State:  enumor.TaskDetailSuccess,
			Result: fmt.Sprintf("Target group created successfully, ID: %s", targetGroupID),
		})
	}
	updateReq := &task.UpdateDetailReq{
		Items: items,
	}
	err := actcli.GetDataService().Global.TaskDetail.Update(kt.Kit(), updateReq)
	if err != nil {
		logs.Errorf("update task detail failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return err
	}
	return nil
}
