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

// Package recycler implements device recycler which deals device recycle task
package recycler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/tjjapi"
	"hcm/pkg/thirdparty/xshipapi"
	"hcm/pkg/tools/utils/wait"
	"hcm/pkg/tools/uuid"

	"k8s.io/client-go/util/workqueue"
)

// Recycler dispatch and deal recycle task
type Recycler struct {
	esbCli  esb.Client
	cvm     cvmapi.CVMClientInterface
	tjj     tjjapi.TjjClientInterface
	xship   xshipapi.XshipClientInterface
	tcOpt   cc.TCloudCli
	queue   workqueue.RateLimitingInterface
	sops    sopsapi.SopsClientInterface
	sopsOpt cc.SopsCli
	ctx     context.Context
	kt      *kit.Kit
}

// New create a dispatcher
func New(ctx context.Context, cliConf cc.ClientConfig, thirdCli *thirdparty.Client, esbCli esb.Client) *Recycler {
	dispatcher := &Recycler{
		esbCli:  esbCli,
		cvm:     thirdCli.CVM,
		tjj:     thirdCli.Tjj,
		xship:   thirdCli.Xship,
		tcOpt:   cliConf.TCloudOpt,
		sops:    thirdCli.Sops,
		sopsOpt: cliConf.Sops,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "recaller"),
		ctx:     ctx,
		kt:      &kit.Kit{Ctx: ctx, Rid: uuid.UUID()},
	}

	// TODO: get worker num from config
	go dispatcher.Run(20)

	return dispatcher
}

// Run starts dispatcher
func (r *Recycler) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(r.runWorker, time.Second, r.ctx)
	}

	select {
	case <-r.ctx.Done():
		logs.Infof("dispatcher exits")
	}
}

// Pop gets head of recall task queue
func (r *Recycler) Pop() (string, error) {
	obj, shutdown := r.queue.Get()
	if shutdown {
		return "", nil
	}

	defer r.queue.Done(obj)

	id, ok := obj.(string)
	if !ok {
		r.queue.Forget(obj)
		logs.Warnf("Expected uint64 in queue but got %#v", obj)
		return "", errors.New("got non-uint64 from queue")
	}

	r.queue.Forget(obj)

	return id, nil
}

// Add add recall task to recall task queue
func (r *Recycler) Add(id string) {
	r.queue.Add(id)
}

// runWorker deals with recall task
func (r *Recycler) runWorker() error {
	id, err := r.Pop()
	if err != nil {
		logs.Errorf("failed to deal recycle task, for get recycle task from informer err: %v", err)
		return err
	}

	if err := r.dispatchHandler(id); err != nil {
		logs.Errorf("failed to dispatch recycle task, err: %v, id: %s", err, id)
		return err
	}

	logs.Infof("Successfully dispatch recycle task %s", id)

	return nil
}

// dispatchHandler recall task dispatch handler
func (r *Recycler) dispatchHandler(id string) error {
	// 1. get recycle task
	task, err := r.getTask(id)
	if err != nil {
		logs.Errorf("failed to get recycle task, err: %v", err)
		return err
	}

	switch task.Status {
	case table.RecallStatusReturned, table.RecallStatusPreChecking:
		{
			err = r.dealPreCheckTask(task)
		}
	case table.RecallStatusClearChecking:
		{
			err = r.dealClearCheckTask(task)
		}
	case table.RecallStatusReinstalling:
		{
			err = r.dealReinstallTask(task)
		}
	case table.RecallStatusInitializing:
		{
			err = r.dealInitializeTask(task)
		}
	case table.RecallStatusDataDeleting:
		{
			err = r.dealDataDeleteTask(task)
		}
	case table.RecallStatusConfChecking:
		{
			err = r.dealConfCheckTask(task)
		}
	case table.RecallStatusTransiting:
		{
			err = r.dealTransitTask(task)
		}
	case table.RecallStatusDone:
		{
			logs.Infof("recall task %s is done, need not handle", task.ID)
			err = nil
		}
	case table.RecallStatusTerminate:
		{
			logs.Infof("recall task %s is terminate, need not handle", task.ID)
			err = nil
		}
	case table.RecallStatusPreCheckFailed, table.RecallStatusClearCheckFailed, table.RecallStatusReinstallFailed,
		table.RecallStatusInitializeFailed, table.RecallStatusDataDeleteFailed, table.RecallStatusConfCheckFailed,
		table.RecallStatusTransitFailed:
		{
			logs.Warnf("recall task %s is failed, need not handle", task.ID)
			err = nil
		}
	default:
		{
			logs.Warnf("recall taskID:%s, unsupported recall task status:%s, cannot handle", task.ID, task.Status)
			err = fmt.Errorf("recall taskID:%s unsupported recall task status:%s, cannot handle", task.ID, task.Status)
		}
	}

	return err
}

func (r *Recycler) getTask(id string) (*table.RecallDetail, error) {
	filter := &mapstr.MapStr{
		"id": id,
	}

	task, err := dao.Set().RecallDetail().GetRecallDetail(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	return task, nil
}
