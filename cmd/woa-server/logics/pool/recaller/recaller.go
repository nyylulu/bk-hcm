/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package recaller implements device recaller which deals device recall task
package recaller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/logics/pool/recycler"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	ccapi "hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/utils/wait"

	"k8s.io/client-go/util/workqueue"
)

// Recaller dispatch and deal recall task
type Recaller struct {
	recycler *recycler.Recycler
	cmdbCli  cmdb.Client
	queue    workqueue.RateLimitingInterface
	ctx      context.Context
}

// New create a dispatcher
func New(ctx context.Context, cmdbCli cmdb.Client) *Recaller {
	dispatcher := &Recaller{
		cmdbCli: cmdbCli,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "recaller"),
		ctx:     ctx,
	}

	// TODO: get worker num from config
	go dispatcher.Run(5)

	return dispatcher
}

// SetRecycler set recaller member recycler
func (r *Recaller) SetRecycler(recycler *recycler.Recycler) {
	r.recycler = recycler
}

// Run starts dispatcher
func (r *Recaller) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(r.runWorker, time.Second, r.ctx)
	}

	select {
	case <-r.ctx.Done():
		logs.Infof("dispatcher exits")
	}
}

// Pop gets head of recall task queue
func (r *Recaller) Pop() (uint64, error) {
	obj, shutdown := r.queue.Get()
	if shutdown {
		return 0, nil
	}

	defer r.queue.Done(obj)

	id, ok := obj.(uint64)
	if !ok {
		r.queue.Forget(obj)
		logs.Warnf("Expected uint64 in queue but got %#v", obj)
		return 0, errors.New("got non-uint64 from queue")
	}

	r.queue.Forget(obj)

	return id, nil
}

// Add add recall task to recall task queue
func (r *Recaller) Add(id uint64) {
	r.queue.Add(id)
}

// runWorker deals with recall task
func (r *Recaller) runWorker() error {
	orderId, err := r.Pop()
	if err != nil {
		logs.Errorf("scheduler:cvm:recaller:runWorker:failed, failed to deal recall task, for get recall task "+
			"from informer err: %v", err)
		return err
	}

	if err = r.dispatchHandler(orderId); err != nil {
		logs.Errorf("scheduler:cvm:recaller:runWorker:failed, failed to dispatch recall task, err: %v, order id: %d",
			err, orderId)
		return err
	}

	logs.Infof("scheduler:cvm:recaller:runWorker:success, Successfully dispatch recall task %d", orderId)

	return nil
}

// dispatchHandler recall task dispatch handler
func (r *Recaller) dispatchHandler(id uint64) error {
	// 1. get recycle task
	task, err := r.getTask(id)
	if err != nil {
		logs.Errorf("failed to get recall task, err: %v", err)
		return err
	}

	switch task.Status.Phase {
	case table.OpTaskPhaseSuccess:
		{
			logs.Infof("recall task %d is success, need not handle", task.ID)
			return nil
		}
	case table.OpTaskPhaseFailed:
		{
			logs.Warnf("recall task %d is failed, cannot handle", task.ID)
			return nil
		}
	case table.OpTaskPhaseRunning, table.OpTaskPhaseInit:
		{
			return r.returnHost(task)
		}
	}

	return nil
}

func (r *Recaller) getTask(id uint64) (*table.RecallTask, error) {
	filter := &mapstr.MapStr{
		"id": id,
	}

	task, err := dao.Set().RecallTask().GetRecallTask(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *Recaller) returnHost(task *table.RecallTask) error {
	// 1. get idle hosts
	hosts, err := r.getIdleHosts(task)
	if err != nil {
		logs.Errorf("failed to get idle host for recall task %d, err: %v", task.ID, err)
		return err
	}

	for _, host := range hosts {
		// transfer hosts from 资源运营服务-CR资源池 to 资源运营服务-CR资源下架中
		if err = r.transferHost(host.HostID, types.BizIDPool, types.BizIDPool,
			types.ModuleIDPoolRecalling); err != nil {
			logs.Errorf("scheduler:cvm:recaller:returnHost:failed, transfer host %d, err: %v", host.HostID, err)
			return err
		}

		// update pool host status
		if err = r.updateHostStatus(host.HostID, table.PoolHostPhaseForRecall); err != nil {
			logs.Errorf("scheduler:cvm:recaller:returnHost:failed, update host %d status, err: %v", host.HostID, err)
			return err
		}
	}

	// update op record
	if err = r.createRecallOpRecords(task, hosts); err != nil {
		logs.Errorf("scheduler:cvm:recaller:returnHost:failed, failed to create recall op record, err: %v", err)
		return err
	}

	// 创建下架回收记录，并放入下架回收队列
	if err = r.createRecallDetail(task, hosts); err != nil {
		logs.Errorf("scheduler:cvm:recaller:returnHost:failed, failed to create recall detail, err: %v", err)
		return err
	}

	// update task status
	task.Status.Phase = table.OpTaskPhaseRunning
	task.Status.SuccessNum = task.Status.SuccessNum + uint(len(hosts))
	task.Status.PendingNum = task.Status.TotalNum - task.Status.SuccessNum
	task.Status.FailedNum = 0
	if task.Status.SuccessNum >= task.Status.TotalNum {
		task.Status.Phase = table.OpTaskPhaseSuccess
	}

	if err = r.updateRecallTaskStatus(task); err != nil {
		logs.Errorf("scheduler:cvm:recaller:returnHost:failed, failed to update recall task status, "+
			"id: %d, err: %v", task.ID, err)
		return err
	}

	return nil
}

func (r *Recaller) getIdleHosts(task *table.RecallTask) ([]*table.PoolHost, error) {
	filter := r.getFilter(task)

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKMaxInstanceLimit,
	}

	hosts, err := dao.Set().PoolHost().FindManyPoolHost(context.Background(), page, filter)
	if err != nil {
		return nil, err
	}

	num := task.Status.PendingNum
	candidateHosts := hosts
	if len(hosts) > int(num) {
		candidateHosts = hosts[0:num]
	}

	return candidateHosts, nil
}

// getFilter get mgo filter
func (r *Recaller) getFilter(task *table.RecallTask) map[string]interface{} {
	filter := make(map[string]interface{})

	// get idle host only
	filter["status.phase"] = table.PoolHostPhaseIdle

	for _, selector := range task.Spec.Selector {
		key := fmt.Sprintf("labels.%s", selector.Key)
		switch selector.Operator {
		case table.SelectOpEqual:
			filter[key] = selector.Value
		case table.SelectOpIn:
			filter[key] = mapstr.MapStr{
				pkg.BKDBIN: selector.Value,
			}
		}
	}

	return filter
}

// transferHost transfer host to target business in cc 3.0
func (r *Recaller) transferHost(hostID, fromBizID, toBizID, toModuleId int64) error {
	transferReq := &ccapi.TransferHostReq{
		From: ccapi.TransferHostSrcInfo{
			FromBizID: fromBizID,
			HostIDs:   []int64{hostID},
		},
		To: ccapi.TransferHostDstInfo{
			ToBizID: toBizID,
		},
	}

	// if destination module id is 0, transfer host to idle module of business
	// otherwise, transfer host to input module
	if toModuleId > 0 {
		transferReq.To.ToModuleID = toModuleId
	}

	kt := core.NewBackendKit()
	kt.Ctx = r.ctx
	err := r.cmdbCli.TransferHost(kt, transferReq)
	if err != nil {
		logs.Errorf("scheduler:cvm:recaller:transferHost:failed, err: %v, req: %+v", err, transferReq)
		return err
	}
	return nil
}

func (r *Recaller) updateHostStatus(hostID int64, phase table.PoolHostPhase) error {
	filter := map[string]interface{}{
		"bk_host_id": hostID,
	}

	now := time.Now()
	update := map[string]interface{}{
		"status.phase": phase,
		"update_at":    now,
	}

	if err := dao.Set().PoolHost().UpdatePoolHost(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}

func (r *Recaller) createRecallOpRecords(task *table.RecallTask, hosts []*table.PoolHost) error {
	now := time.Now()
	for _, host := range hosts {
		id, err := dao.Set().OpRecord().NextSequence(context.Background())
		if err != nil {
			logs.Errorf("failed to create op record, err: %v", err)
			return err
		}
		record := &table.OpRecord{
			ID:       id,
			HostID:   host.HostID,
			Labels:   host.Labels,
			OpType:   table.OpTypeRecall,
			TaskID:   task.ID,
			Phase:    table.OpTaskPhaseSuccess,
			Message:  "",
			Operator: "icr",
			CreateAt: now,
			UpdateAt: now,
		}

		if err = dao.Set().OpRecord().CreateOpRecord(context.Background(), record); err != nil {
			logs.Errorf("scheduler:cvm:recaller:createRecallOpRecords:failed, failed to save op record, "+
				"host id: %d, err: %v", host.HostID, err)
			return fmt.Errorf("failed to save op record, host id: %d, err: %v", host.HostID, err)
		}
	}

	return nil
}

func (r *Recaller) createRecallDetail(task *table.RecallTask, hosts []*table.PoolHost) error {
	now := time.Now()
	for _, host := range hosts {
		detail := &table.RecallDetail{
			ID:            fmt.Sprintf("%d-%d", task.ID, host.HostID),
			RecallID:      task.ID,
			HostID:        host.HostID,
			Labels:        host.Labels,
			Status:        table.RecallStatusReturned,
			Message:       "",
			ReinstallID:   "",
			ReinstallLink: "",
			ConfCheckID:   "",
			ConfCheckLink: "",
			Operator:      "icr",
			CreateAt:      now,
			UpdateAt:      now,
		}

		if err := dao.Set().RecallDetail().CreateRecallDetail(context.Background(), detail); err != nil {
			logs.Errorf("scheduler:cvm:recaller:createRecallDetail:failed, failed to save recall detail, "+
				"host id: %d, err: %v", host.HostID, err)
			return fmt.Errorf("failed to save recall detail, host id: %d, err: %v", host.HostID, err)
		}

		// add recall task to dispatch queue
		r.recycler.Add(detail.ID)
	}

	return nil
}

// updateRecallTaskStatus update recall task status
func (r *Recaller) updateRecallTaskStatus(task *table.RecallTask) error {
	filter := map[string]interface{}{
		"id": task.ID,
	}

	doc := map[string]interface{}{
		"status.phase":       task.Status.Phase,
		"status.success_num": task.Status.SuccessNum,
		"status.pending_num": task.Status.PendingNum,
		"status.failed_num":  task.Status.FailedNum,
		"update_at":          time.Now(),
	}

	if err := dao.Set().RecallTask().UpdateRecallTask(context.Background(), filter, doc); err != nil {
		return err
	}

	return nil
}
