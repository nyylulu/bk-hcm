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
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
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

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := generator()

		targets, concurrentErr := concurrence.BaseExecWithResult(
			cc.CloudServer().CLBImportConfig.ConcurrentCount, taskDetails,
			func(detail *layer4ListenerBindRSTaskDetail) (*hclb.RegisterTarget, error) {

				target := &hclb.RegisterTarget{
					TargetType: detail.InstType,
					Port:       int64(detail.RsPort[0]),
					Weight:     converter.ValToPtr(int64(converter.PtrToVal(detail.Weight))),
				}
				if detail.InstType == enumor.EniInstType {
					target.EniIp = detail.RsIp
				}

				if detail.InstType == enumor.CvmInstType {
					cvm, err := validateCvmExist(kt,
						c.dataServiceCli, detail.RsIp, c.vendor, c.bkBizID, c.accountID, lb)
					if err != nil {
						logs.Errorf("validate cvm exist failed, ip: %s, err: %v, rid: %s", detail.RsIp, err, kt.Rid)
						return nil, err
					}

					target.CloudInstID = cvm.CloudID
					target.InstName = cvm.Name
					target.PrivateIPAddress = cvm.PrivateIPv4Addresses
					target.PublicIPAddress = cvm.PublicIPv4Addresses
					target.Zone = cvm.Zone
				}
				return target, nil
			})
		if concurrentErr != nil {
			logs.Errorf("batch validate cvm exist failed, err: %v, rid: %s", concurrentErr, kt.Rid)
			return nil, concurrentErr
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

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := generator()

		targets, concurrentErr := concurrence.BaseExecWithResult(
			cc.CloudServer().CLBImportConfig.ConcurrentCount, taskDetails,
			func(detail *layer7ListenerBindRSTaskDetail) (*hclb.RegisterTarget, error) {
				target := &hclb.RegisterTarget{
					TargetType: detail.InstType,
					Port:       int64(detail.RsPort[0]),
					Weight:     converter.ValToPtr(int64(converter.PtrToVal(detail.Weight))),
				}
				if detail.InstType == enumor.EniInstType {
					target.EniIp = detail.RsIp
				}

				if detail.InstType == enumor.CvmInstType {
					cvm, err := validateCvmExist(kt,
						c.dataServiceCli, detail.RsIp, c.vendor, c.bkBizID, c.accountID, lb)
					if err != nil {
						logs.Errorf("validate cvm exist failed, ip: %s, err: %v, rid: %s", detail.RsIp, err, kt.Rid)
						return nil, err
					}

					target.CloudInstID = cvm.CloudID
					target.InstName = cvm.Name
					target.PrivateIPAddress = cvm.PrivateIPv4Addresses
					target.PublicIPAddress = cvm.PublicIPv4Addresses
					target.Zone = cvm.Zone
				}
				return target, nil
			})
		if concurrentErr != nil {
			logs.Errorf("batch register tcloud target failed, err: %v, rid: %s", concurrentErr, kt.Rid)
			return nil, concurrentErr
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
