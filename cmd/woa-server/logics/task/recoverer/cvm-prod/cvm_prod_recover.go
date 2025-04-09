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

package cvmprod

import (
	"runtime/debug"
	"time"

	"hcm/cmd/woa-server/logics/cvm"
	types "hcm/cmd/woa-server/types/cvm"
	"hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"

	"golang.org/x/sync/errgroup"
)

// StartRecover 开启cvm生产恢复逻辑
func StartRecover(kt *kit.Kit, logics cvm.Logics, sd serviced.State) {
	restorer := &cvmProdRestorer{
		logics: logics,
	}
	subKit := kt.NewSubKit()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Errorf("[hcm server panic], err: %v, rid: %s, debug strace: %s", err, kt.Rid, debug.Stack())
			}
		}()

		for !sd.IsMaster() {
			logs.Warnf("current server is not master, sleep one minute, rid: %s", kt.Rid)
			time.Sleep(time.Minute)
		}
		logs.Infof("start cvm product recover logic, rid: %s", kt.Rid)

		if err := restorer.recover(subKit); err != nil {
			logs.Errorf("failed to start cvm product recover, err: %v, rid: %s", err, subKit.Rid)
		}
	}()
}

type cvmProdRestorer struct {
	logics cvm.Logics
}

func (c *cvmProdRestorer) recover(kt *kit.Kit) error {
	orders, err := c.getNeedRecoverOrders(kt)
	if err != nil {
		logs.Errorf("failed to get need recover orders, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	var eg errgroup.Group
	eg.SetLimit(task.CvmProdRecoverGoroutinesNum)

	for _, order := range orders {
		curOrder := order
		eg.Go(func() error {
			logs.Infof("start to recover order, order id: %d, status: %s, rid: %s", curOrder.OrderId, curOrder.Status,
				kt.Rid)

			switch curOrder.Status {
			case types.ApplyStatusInit:
				if err = c.logics.ExecuteApplyOrder(kt, curOrder); err != nil {
					logs.Errorf("failed to execute apply order, err: %v, order id: %d, rid: %s", err, curOrder.OrderId,
						kt.Rid)
				}
			case types.ApplyStatusRunning:
				if err = c.logics.CreateCvmFromTaskResult(kt, curOrder); err != nil {
					logs.Errorf("failed to create cvm from task result, err: %v, order id: %d, rid: %s", err,
						curOrder.OrderId, kt.Rid)
				}
			default:
				logs.Warnf("order in this status is not supported, order id: %d, status: %s, rid: %s",
					curOrder.OrderId, curOrder.Status, kt.Rid)
			}
			return nil
		})
	}

	_ = eg.Wait()

	return nil
}

func (c *cvmProdRestorer) getNeedRecoverOrders(kt *kit.Kit) ([]*types.ApplyOrder, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, task.ExpireDays)
	status := []types.ApplyStatus{types.ApplyStatusInit, types.ApplyStatusRunning}
	param := &types.GetApplyParam{Status: status, Start: startTime.Format(time.DateOnly)}
	res, err := c.logics.GetApplyOrder(kt, param)
	if err != nil {
		logs.Errorf("failed to get cvm product orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return res.Info, nil
}
