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

// Package launcher implements device launcher which launches device to resource pool
package launcher

import (
	"context"
	"errors"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/utils/wait"
	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/pkg/logs"

	"k8s.io/client-go/util/workqueue"
)

// Launcher dispatch and deal launch task
type Launcher struct {
	esbCli esb.Client
	queue  workqueue.RateLimitingInterface
	ctx    context.Context
}

// New create a dispatcher
func New(ctx context.Context, esbCli esb.Client) *Launcher {
	dispatcher := &Launcher{
		esbCli: esbCli,
		queue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "launcher"),
		ctx:    ctx,
	}

	// TODO: get worker num from config
	go dispatcher.Run(5)

	return dispatcher
}

// Run starts dispatcher
func (l *Launcher) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(l.runWorker, time.Second, l.ctx)
	}

	select {
	case <-l.ctx.Done():
		logs.Infof("dispatcher exits")
	}
}

// Pop gets head of launch task queue
func (l *Launcher) Pop() (uint64, error) {
	obj, shutdown := l.queue.Get()
	if shutdown {
		return 0, nil
	}

	defer l.queue.Done(obj)

	id, ok := obj.(uint64)
	if !ok {
		l.queue.Forget(obj)
		logs.Warnf("Expected uint64 in queue but got %#v", obj)
		return 0, errors.New("got non-uint64 from queue")
	}

	l.queue.Forget(obj)

	return id, nil
}

// Add add launch task to launch task queue
func (l *Launcher) Add(id uint64) {
	l.queue.Add(id)
}

// runWorker deals with launch task
func (l *Launcher) runWorker() error {
	id, err := l.Pop()
	if err != nil {
		logs.Errorf("scheduler:cvm:launcher:runWorker:failed, failed to deal launch task, "+
			"for get launch task from informer err: %v", err)
		return err
	}

	if err = l.dispatchHandler(id); err != nil {
		logs.Errorf("scheduler:cvm:launcher:runWorker:failed, failed to dispatch launch task, "+
			"err: %v, task id: %d", err, id)
		return err
	}

	logs.Infof("scheduler:cvm:launcher:runWorker:success, Successfully dispatch launch task %d", id)

	return nil
}

// dispatchHandler launch task dispatch handler
func (l *Launcher) dispatchHandler(id uint64) error {
	task, err := l.getTask(id)
	if err != nil {
		logs.Errorf("failed to get launch task, err: %v", err)
		return err
	}

	records, err := l.getOpRecords(id)
	if err != nil {
		logs.Errorf("failed to get launch op records, err: %v", err)
		return err
	}

	hostIDs := make([]int64, 0)
	for _, record := range records {
		hostIDs = append(hostIDs, record.HostID)
	}

	if len(hostIDs) == 0 {
		logs.Errorf("scheduler:cvm:launcher:dispatchHandler:failed, id: %dfailed to deal launch task, "+
			"for get no hosts", id)
		return errors.New("failed to deal launch task, for get no hosts")
	}

	if err = l.transferHost2Pool(hostIDs, 931); err != nil {
		logs.Errorf("scheduler:cvm:launcher:dispatchHandler:failed, failed to transfer hosts to pool, id: %d, err: %v",
			id, err)
		return err
	}

	// add host to pool
	if err = l.updatePoolHost(records); err != nil {
		logs.Errorf("scheduler:cvm:launcher:dispatchHandler:failed, failed to update pool host, id: %d, "+
			"records: %+v, err: %v", id, records, err)
		return err
	}

	// update records status
	if err = l.updateOpRecordStatus(id, table.OpTaskPhaseSuccess); err != nil {
		logs.Errorf("scheduler:cvm:launcher:dispatchHandler:failed, failed to update op record status, err: %v", err)
		return err
	}

	task.Status.Phase = table.OpTaskPhaseSuccess
	task.Status.SuccessNum = uint(len(hostIDs))
	task.Status.PendingNum = 0
	task.Status.FailedNum = 0
	// update task status
	if err = l.updateTaskStatus(task); err != nil {
		logs.Errorf("scheduler:cvm:launcher:dispatchHandler:failed, failed to update launch task status, "+
			"id: %s, err: %v", task.ID, err)
		return err
	}

	return nil
}

func (l *Launcher) getTask(id uint64) (*table.LaunchTask, error) {
	filter := &mapstr.MapStr{
		"id": id,
	}

	task, err := dao.Set().LaunchTask().GetLaunchTask(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (l *Launcher) getOpRecords(id uint64) ([]*table.OpRecord, error) {
	filter := map[string]interface{}{
		"op_type": table.OpTypeLaunch,
		"task_id": id,
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}

	records, err := dao.Set().OpRecord().FindManyOpRecord(context.Background(), page, filter)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// updateOpRecordStatus update operation record status
func (l *Launcher) updateOpRecordStatus(id uint64, phase table.OpTaskPhase) error {
	filter := map[string]interface{}{
		"op_type": table.OpTypeLaunch,
		"task_id": id,
	}

	doc := map[string]interface{}{
		"phase":     phase,
		"update_at": time.Now(),
	}

	if err := dao.Set().OpRecord().UpdateOpRecord(context.Background(), filter, doc); err != nil {
		return err
	}

	return nil
}

// updateTaskStatus update launch task status
func (l *Launcher) updateTaskStatus(task *table.LaunchTask) error {
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

	if err := dao.Set().LaunchTask().UpdateLaunchTask(context.Background(), filter, doc); err != nil {
		return err
	}

	return nil
}

// updatePoolHost update pool host info
func (l *Launcher) updatePoolHost(records []*table.OpRecord) error {
	for _, record := range records {
		now := time.Now()
		newHost := &table.PoolHost{
			HostID: record.HostID,
			Labels: record.Labels,
			Status: &table.PoolHostStatus{
				Phase:      table.PoolHostPhaseIdle,
				LaunchID:   record.TaskID,
				LaunchTime: now,
			},
			CreateAt: now,
			UpdateAt: now,
		}
		filter := map[string]interface{}{
			"bk_host_id": record.HostID,
		}

		if err := dao.Set().PoolHost().UpsertPoolHost(context.Background(), filter, newHost); err != nil {
			logs.Errorf("failed to upsert pool host, err: %v", err)
			return err
		}
	}

	return nil
}
