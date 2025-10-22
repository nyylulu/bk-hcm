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

package recycle

import (
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// recycleRecoverer 回收恢复服务
type recycleRecoverer struct {
	recyclerIf recycler.Interface
	itsmCli    itsm.Client
	cmdbCli    cmdb.Client
}

// Interface recycle recoverer interface
type Interface interface {
	recoverRecycleTickets(kt *kit.Kit) error
	recoverRunningOrders(kt *kit.Kit, orders []*table.RecycleOrder)
}

// StartRecover 开启回收服务
func StartRecover(kt *kit.Kit, itsmCli itsm.Client, recycler recycler.Interface, cmdbCli cmdb.Client,
	sd serviced.State) error {

	recycleRecoverer := &recycleRecoverer{
		recyclerIf: recycler,
		itsmCli:    itsmCli,
		cmdbCli:    cmdbCli,
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
		logs.Infof("start recycle recover logic, rid: %s", kt.Rid)

		if err := recycleRecoverer.recoverRecycleTickets(subKit); err != nil {
			logs.Errorf("recycle recover: failed to start recycle recover, err: %v, rid: %s", err, subKit.Rid)
		}
	}()

	return nil
}

// RecoverRecycleTickets 恢复服务重启前未完成订单
func (r *recycleRecoverer) recoverRecycleTickets(kt *kit.Kit) error {
	// get recover time and expire time
	restartTime := time.Now()
	expireTime := restartTime.AddDate(0, 0, recovertask.ExpireDays)
	runningOrders, err := r.getRecycleRunningOrders(kt, expireTime, restartTime)
	if err != nil {
		logs.Errorf("failed to get recycle orders by status and limit time, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	r.recoverRunningOrders(kt, runningOrders)
	return nil
}

// recoverRunningOrders 恢复未完成回收订单
func (r *recycleRecoverer) recoverRunningOrders(kt *kit.Kit, orders []*table.RecycleOrder) {
	defer func() {
		if err := recover(); err != nil {
			logs.Errorf("[hcm server panic], err: %v, rid: %s, debug strace: %s", err, kt.Rid, debug.Stack())
		}
	}()

	logs.Infof("start recover running recycle orders, orderNum: %d, rid: %s", len(orders), kt.Rid)

	dealChan := make(chan struct{}, recovertask.ApplyGoroutinesNum)
	wg := sync.WaitGroup{}
	successNum, failedNum := int64(0), int64(0)
	for _, order := range orders {
		wg.Add(1)
		dealChan <- struct{}{}

		go func(kt *kit.Kit, order *table.RecycleOrder) {
			defer func() {
				wg.Done()
				<-dealChan
			}()
			logs.Infof("start recover recycle order, subOrderId: %s, status: %s, rid: %s", order.SuborderID,
				order.Status, kt.Rid)
			if err := r.dealOrder(kt, order); err != nil {
				logs.Errorf("failed to recover recycle order, subOrderId: %s, err: %v, rid: %s", order.SuborderID,
					err, kt.Rid)
				atomic.AddInt64(&failedNum, 1)
				return
			}
			atomic.AddInt64(&successNum, 1)
		}(kt, order)
	}
	wg.Wait()

	logs.Infof("finish recover running recycle orders, successNum: %d, failedNum: %d, totalNum: %d, rid: %s",
		successNum, failedNum, len(orders), kt.Rid)
}

// dealOrder 根据订单状态处理回收订单
func (r *recycleRecoverer) dealOrder(kt *kit.Kit, order *table.RecycleOrder) error {
	switch order.Status {
	case table.RecycleStatusCommitted:
		logs.Infof("start to recover COMMITTED recycle order, the first step of recycle, subOrderId: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return r.recoverCommittedOrder(kt, order)
	case table.RecycleStatusDetecting:
		logs.Infof("start to recover DETECTING recycle order, the second step of recycle subOrderId: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return r.recoverDetectedOrder(kt, order)
	case table.RecycleStatusTransiting:
		logs.Infof("start to recover TRANSITING recycle order, the third step of recycle, subOrderId: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return r.recoverTransitedOrder(kt, order)
	case table.RecycleStatusReturning:
		logs.Infof("start to recover RETURNING recycle order, the last step of recycle subOrderId: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return r.recoverReturnedOrder(kt, order)
	case table.RecycleStatusReturningPlan:
		logs.Infof("start to recover RETURNING_PLAN recycle order, the last step of recycle subOrderId: %s, rid: %s",
			order.SuborderID, kt.Rid)
		return r.recoverReturningPlanOrder(kt, order)
	default:
		logs.Errorf("unknown recycle order status: %s, subOrderId: %s, rid: %s", order.Status, order.SuborderID, kt.Rid)
		return fmt.Errorf("unknown recycle order status: %s, subOrderId: %s", order.Status, order.SuborderID)
	}
}
