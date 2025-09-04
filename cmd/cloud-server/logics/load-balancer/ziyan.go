/*
 *
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

package lblogic

import (
	"fmt"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"strings"
)

func (c *CreateLayer7ListenerPreviewExecutor) getTCloudZiyanListenersByPort(kt *kit.Kit, lbCloudID string, port int) (
	[]corelb.Listener[corelb.TCloudListenerExtension], error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", c.accountID),
			tools.RuleEqual("bk_biz_id", c.bkBizID),
			tools.RuleEqual("cloud_lb_id", lbCloudID),
			tools.RuleEqual("port", port),
			tools.RuleEqual("vendor", c.vendor),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := c.dataServiceCli.TCloudZiyan.LoadBalancer.ListListener(kt, req)
	if err != nil {
		logs.Errorf("list listener failed, port: %d, cloudLBID: %s, err: %v, rid: %s",
			port, lbCloudID, err, kt.Rid)
		return nil, err
	}
	if len(resp.Details) > 0 {
		return resp.Details, nil
	}
	return nil, nil
}

func (c *Layer4ListenerBindRSExecutor) buildTCloudZiyanFlowTask(kt *kit.Kit, lb corelb.LoadBalancerRaw,
	targetGroupID string, details []*layer4ListenerBindRSTaskDetail,
	generator func() (cur string, prev string), tgToListenerCloudIDs map[string]string) ([]ts.CustomFlowTask, error) {

	//有目标组ID，直接绑定RS
	if targetGroupID != "" && !strings.HasPrefix(targetGroupID, "auto_") && !strings.HasPrefix(targetGroupID, "temp_tg_") {
		logs.Infof("using existing target group: %s, will bind RS directly, rid: %s", targetGroupID, kt.Rid)
		return c.bindTCloudZiyanRSTask(lb, targetGroupID, details, generator, tgToListenerCloudIDs)
	}

	//没有目标组ID，自动创建目标组并绑定RS
	logs.Infof("listener has no target group or using temp target group,"+
		" will auto-create target group and bind RS, rid: %s", kt.Rid)
	return c.autoCreateTCloudZiyanTargetGroupAndBindRS(lb, details, generator, tgToListenerCloudIDs)
}

// autoCreateTCloudZiyanTargetGroupAndBindRS 自动创建TCloudZiyan目标组并绑定RS
func (c *Layer4ListenerBindRSExecutor) autoCreateTCloudZiyanTargetGroupAndBindRS(lb corelb.LoadBalancerRaw,
	details []*layer4ListenerBindRSTaskDetail, generator func() (cur string, prev string),
	tgToListenerCloudIDs map[string]string) ([]ts.CustomFlowTask, error) {

	if len(details) == 0 {
		return nil, fmt.Errorf("details cannot be empty for auto-create target group")
	}

	autoKey := fmt.Sprintf("auto_%s", details[0].listenerCloudID)
	listenerCloudID, exists := tgToListenerCloudIDs[autoKey]
	if !exists {
		return nil, fmt.Errorf("listener cloud ID not found for auto key: %s", autoKey)
	}

	targets := make([]*corelb.BaseTarget, 0, len(details))
	managementDetailIDs := make([]string, 0, len(details))

	for _, detail := range details {

		if len(detail.RsPort) == 0 {
			return nil, fmt.Errorf("RS port cannot be empty for detail: %s", detail.taskDetailID)
		}

		target := &corelb.BaseTarget{
			InstType: detail.InstType,
			Port:     int64(detail.RsPort[0]),
			Weight:   converter.ValToPtr(converter.PtrToVal(detail.Weight)),
		}

		if detail.InstType == enumor.EniInstType {
			if detail.RsIp == "" {
				return nil, fmt.Errorf("ENI IP cannot be empty for detail: %s", detail.taskDetailID)
			}
			target.IP = detail.RsIp
		} else if detail.InstType == enumor.CvmInstType {
			if detail.cvm == nil {
				return nil, fmt.Errorf("CVM info not found for detail: %s", detail.taskDetailID)
			}
			target.CloudInstID = detail.cvm.CloudID
			target.InstName = detail.cvm.Name
			target.PrivateIPAddress = detail.cvm.PrivateIPv4Addresses
			target.PublicIPAddress = detail.cvm.PublicIPv4Addresses
			target.Zone = detail.cvm.Zone
		}

		targets = append(targets, target)
		managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
	}

	cur, prev := generator()
	createTGTask := ts.CustomFlowTask{
		ActionID:   action.ActIDType(cur),
		ActionName: enumor.ActionCreateTargetGroupWithRel,
		Params: &actionlb.CreateTargetGroupWithRelOption{
			Vendor:              c.vendor,
			LoadBalancerID:      lb.ID,
			ListenerID:          listenerCloudID,
			ListenerRuleID:      "",
			RuleType:            enumor.Layer4RuleType,
			Targets:             targets,
			ManagementDetailIDs: managementDetailIDs,
		},
		DependOn: []action.ActIDType{action.ActIDType(prev)},
	}

	return []ts.CustomFlowTask{createTGTask}, nil
}

// bindTCloudZiyanRSTask 绑定TCloudZiyan RS任务
func (c *Layer4ListenerBindRSExecutor) bindTCloudZiyanRSTask(lb corelb.LoadBalancerRaw,
	targetGroupID string, details []*layer4ListenerBindRSTaskDetail,
	generator func() (cur string, prev string), tgToListenerCloudIDs map[string]string) ([]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := generator()

		targets := make([]*hclb.RegisterTarget, 0, len(taskDetails))
		for _, detail := range taskDetails {
			target := &hclb.RegisterTarget{
				TargetType: detail.InstType,
				Port:       int64(detail.RsPort[0]),
				Weight:     converter.ValToPtr(int64(converter.PtrToVal(detail.Weight))),
			}
			if detail.InstType == enumor.EniInstType {
				target.EniIp = detail.RsIp
			} else if detail.InstType == enumor.CvmInstType {
				if detail.cvm == nil {
					return nil, fmt.Errorf("rs ip(%s) not found", detail.RsIp)
				}

				target.CloudInstID = detail.cvm.CloudID
				target.InstName = detail.cvm.Name
				target.PrivateIPAddress = detail.cvm.PrivateIPv4Addresses
				target.PublicIPAddress = detail.cvm.PublicIPv4Addresses
				target.Zone = detail.cvm.Zone
			}
			targets = append(targets, target)
		}

		req := &hclb.BatchRegisterTCloudTargetReq{
			CloudListenerID: tgToListenerCloudIDs[targetGroupID],
			TargetGroupID:   targetGroupID,
			RuleType:        enumor.Layer4RuleType,
			Targets:         targets,
		}
		managementDetailIDs := slice.Map(taskDetails, func(detail *layer4ListenerBindRSTaskDetail) string {
			return detail.taskDetailID
		})
		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudBindTarget,
			Params: &actionlb.BatchTaskBindTargetOption{
				Vendor:                       c.vendor,
				LoadBalancerID:               lb.ID,
				ManagementDetailIDs:          managementDetailIDs,
				BatchRegisterTCloudTargetReq: req,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)
		// update taskDetail.actionID
		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}

func (c *Layer7ListenerBindRSExecutor) buildTCloudZiyanFlowTask(kt *kit.Kit, lb corelb.LoadBalancerRaw,
	targetGroupID string, details []*layer7ListenerBindRSTaskDetail, generator func() (cur string, prev string),
	tgToListenerCloudIDs map[string]string, tgToCloudRuleIDs map[string]string) ([]ts.CustomFlowTask, error) {

	//有目标组ID，直接绑定RS
	if targetGroupID != "" && !strings.HasPrefix(targetGroupID, "auto_") && !strings.HasPrefix(targetGroupID, "temp_tg_") {
		logs.Infof("using existing target group: %s, will bind RS directly, rid: %s", targetGroupID, kt.Rid)
		return c.bindTCloudZiyanRSTask(lb, targetGroupID, details, generator, tgToListenerCloudIDs, tgToCloudRuleIDs)
	}

	//没有目标组ID，自动创建目标组并绑定RS
	logs.Infof("listener has no target group or using temp target group,"+
		" will auto-create target group and bind RS, rid: %s", kt.Rid)
	return c.createTCloudZiyanTargetGroupTask(lb, targetGroupID, details, generator, tgToListenerCloudIDs, tgToCloudRuleIDs)
}

// createTCloudZiyanTargetGroupTask 创建TCloudZiyan目标组任务
func (c *Layer7ListenerBindRSExecutor) createTCloudZiyanTargetGroupTask(lb corelb.LoadBalancerRaw,
	targetGroupID string, details []*layer7ListenerBindRSTaskDetail, generator func() (cur string, prev string),
	tgToListenerCloudIDs map[string]string, tgToCloudRuleIDs map[string]string) ([]ts.CustomFlowTask, error) {

	listenerCloudID := tgToListenerCloudIDs[targetGroupID]
	urlRuleCloudID := tgToCloudRuleIDs[targetGroupID]

	if strings.HasPrefix(targetGroupID, "auto_") {
		if len(details) > 0 && details[0].urlRuleCloudID != "" {
			urlRuleCloudID = details[0].urlRuleCloudID
		} else {
			logs.Warnf("URL rule cloud ID is empty for auto-created target group, listener: %s, rid: %s",
				listenerCloudID, details[0].taskDetailID)
		}
	}

	targets := make([]*corelb.BaseTarget, 0, len(details))
	managementDetailIDs := make([]string, 0, len(details))

	for _, detail := range details {
		target := &corelb.BaseTarget{
			InstType: detail.InstType,
			Port:     int64(detail.RsPort[0]),
			Weight:   converter.ValToPtr(converter.PtrToVal(detail.Weight)),
		}

		if detail.InstType == enumor.EniInstType {
			target.IP = detail.RsIp
		} else if detail.InstType == enumor.CvmInstType {
			if detail.cvm == nil {
				return nil, fmt.Errorf("rs ip(%s) not found", detail.RsIp)
			}
			target.CloudInstID = detail.cvm.CloudID
			target.InstName = detail.cvm.Name
			target.PrivateIPAddress = detail.cvm.PrivateIPv4Addresses
			target.PublicIPAddress = detail.cvm.PublicIPv4Addresses
			target.Zone = detail.cvm.Zone
		}

		targets = append(targets, target)
		managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
	}

	cur, prev := generator()
	createTGTask := ts.CustomFlowTask{
		ActionID:   action.ActIDType(cur),
		ActionName: enumor.ActionCreateTargetGroupWithRel,
		Params: &actionlb.CreateTargetGroupWithRelOption{
			Vendor:              c.vendor,
			LoadBalancerID:      lb.ID,
			ListenerID:          listenerCloudID,
			ListenerRuleID:      urlRuleCloudID,
			RuleType:            enumor.Layer7RuleType,
			Targets:             targets,
			ManagementDetailIDs: managementDetailIDs,
		},
		DependOn: []action.ActIDType{action.ActIDType(prev)},
	}

	return []ts.CustomFlowTask{createTGTask}, nil
}

// bindTCloudZiyanRSTask 绑定TCloudZiyan RS任务
func (c *Layer7ListenerBindRSExecutor) bindTCloudZiyanRSTask(lb corelb.LoadBalancerRaw,
	targetGroupID string, details []*layer7ListenerBindRSTaskDetail, generator func() (cur string, prev string),
	tgToListenerCloudIDs map[string]string, tgToCloudRuleIDs map[string]string) ([]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := generator()

		targets := make([]*hclb.RegisterTarget, 0, len(taskDetails))
		for _, detail := range taskDetails {
			target := &hclb.RegisterTarget{
				TargetType: detail.InstType,
				Port:       int64(detail.RsPort[0]),
				Weight:     converter.ValToPtr(int64(converter.PtrToVal(detail.Weight))),
			}
			if detail.InstType == enumor.EniInstType {
				target.EniIp = detail.RsIp
			} else if detail.InstType == enumor.CvmInstType {
				if detail.cvm == nil {
					return nil, fmt.Errorf("rs ip(%s) not found", detail.RsIp)
				}

				target.CloudInstID = detail.cvm.CloudID
				target.InstName = detail.cvm.Name
				target.PrivateIPAddress = detail.cvm.PrivateIPv4Addresses
				target.PublicIPAddress = detail.cvm.PublicIPv4Addresses
				target.Zone = detail.cvm.Zone
			}
			targets = append(targets, target)
		}

		req := &hclb.BatchRegisterTCloudTargetReq{
			CloudListenerID: tgToListenerCloudIDs[targetGroupID],
			CloudRuleID:     tgToCloudRuleIDs[targetGroupID],
			TargetGroupID:   targetGroupID,
			RuleType:        enumor.Layer7RuleType,
			Targets:         targets,
		}
		managementDetailIDs := slice.Map(taskDetails, func(detail *layer7ListenerBindRSTaskDetail) string {
			return detail.taskDetailID
		})
		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudBindTarget,
			Params: &actionlb.BatchTaskBindTargetOption{
				Vendor:                       c.vendor,
				LoadBalancerID:               lb.ID,
				ManagementDetailIDs:          managementDetailIDs,
				BatchRegisterTCloudTargetReq: req,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)
		// update taskDetail.actionID
		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}
