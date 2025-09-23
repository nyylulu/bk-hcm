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
	ptypes "hcm/cmd/woa-server/types/plan"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// unlockTicketOriginalDemands 解锁订单中的原始预测需求，用于预测修改失败等特殊情况，避免死锁
func (d *Dispatcher) unlockTicketOriginalDemands(kt *kit.Kit, ticket *ptypes.TicketInfo) error {
	allDemandIDs := make([]string, 0)
	for _, demand := range ticket.Demands {
		if demand.Original != nil {
			allDemandIDs = append(allDemandIDs, demand.Original.DemandID)
		}
	}

	if len(allDemandIDs) == 0 {
		return nil
	}

	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(allDemandIDs, 0)
	if err := d.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Errorf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// BatchUpsertResPlanDemand batch upsert res plan demand and unlock res plans.
func (d *Dispatcher) BatchUpsertResPlanDemand(kt *kit.Kit, upsertReq *rpproto.ResPlanDemandBatchUpsertReq,
	updatedIDs []string) ([]string, error) {

	unlockDemandIDs := make([]string, 0)
	// 批量创建和更新预测
	createdRst, err := d.client.DataService().Global.ResourcePlan.BatchUpsertResPlanDemand(kt, upsertReq)
	if err != nil {
		logs.Errorf("failed to batch upsert res plan demand, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	unlockDemandIDs = append(unlockDemandIDs, createdRst.IDs...)
	unlockDemandIDs = append(unlockDemandIDs, updatedIDs...)
	if len(unlockDemandIDs) == 0 {
		return createdRst.IDs, nil
	}

	// unlock all crp demands.
	unlockReq := rpproto.NewResPlanDemandLockOpReqBatch(unlockDemandIDs, 0)
	if err := d.client.DataService().Global.ResourcePlan.UnlockResPlanDemand(kt, unlockReq); err != nil {
		logs.Warnf("failed to unlock all resource plan demand, err: %v, rid: %s", err, kt.Rid)
	}

	return createdRst.IDs, nil
}
