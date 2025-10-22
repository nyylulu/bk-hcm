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

package plan

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/plan/dispatcher"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ApplyDestroyOrderToResPlanDemand 将销毁单返还的预测预算保存到本地
func (c *Controller) ApplyDestroyOrderToResPlanDemand(kt *kit.Kit, destroyOrderID string) error {

	returnTask, err := getReturnTaskByTaskID(kt, destroyOrderID)
	if err != nil {
		logs.Errorf("get return task failed, err: %v, destroyOrderID: %s, rid: %s", err, destroyOrderID, kt.Rid)
		return err
	}

	bizOrgRel, err := c.bizLogics.GetBizOrgRel(kt, returnTask.BkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, biz: %d, rid: %s", err, returnTask.BkBizID, kt.Rid)
		return err
	}

	returnOrders, err := c.resFetcher.GetOrderList(kt, destroyOrderID)
	if err != nil {
		logs.Errorf("get order list failed, err: %v, destroyOrderID: %s, rid: %s", err, destroyOrderID, kt.Rid)
		return err
	}

	for _, order := range returnOrders {
		if order.Status != enumor.QueryOrderInfoStatusSuccess {
			logs.Errorf("order status is not success, order: %v, statusCode: %d, statusMsg: %s, rid: %s", order,
				order.Status, order.StatusMsg, kt.Rid)
			continue
		}

		changeInfos, err := c.dispatcher.QueryCrpOrderChangeInfo(kt, order.OrderID)
		if err != nil {
			logs.Errorf("failed to query crp order change info, err: %v, destroyOrderID: %s, rid: %s",
				err, order.OrderID, kt.Rid)
			return err
		}

		ctx := &dispatcher.ApplyTicketCtx{
			ID:              destroyOrderID,
			Applicant:       returnTask.User,
			BkBizID:         returnTask.BkBizID,
			BkBizName:       bizOrgRel.BkBizName,
			OpProductID:     bizOrgRel.OpProductID,
			OpProductName:   bizOrgRel.OpProductName,
			PlanProductID:   bizOrgRel.PlanProductID,
			PlanProductName: bizOrgRel.PlanProductName,
			VirtualDeptID:   bizOrgRel.VirtualDeptID,
			VirtualDeptName: bizOrgRel.VirtualDeptName,
			Remark:          "",
			CrpSN:           order.OrderID,
			DemandClass:     enumor.DemandClassCVM,
		}

		err = c.dispatcher.ApplyResPlanDemandChange(kt, ctx, changeInfos)
		if err != nil {
			logs.Errorf("failed to apply res plan demand change, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func getReturnTaskByTaskID(kt *kit.Kit, taskID string) (*table.ReturnTask, error) {
	filter := &mapstr.MapStr{
		"task_id": taskID,
	}
	task, err := dao.Set().ReturnTask().GetReturnTask(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if task == nil {
		logs.Errorf("recycle order not found, task_id: %s, rid: %s", taskID, kt.Rid)
		return nil, fmt.Errorf("recycle order not found")
	}
	return task, nil
}
