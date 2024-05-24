/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package recycle recycle informer
package recycle

import (
	"errors"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/storage/dal"
	"hcm/cmd/woa-server/storage/stream"
	"hcm/cmd/woa-server/storage/stream/types"
	"hcm/pkg/logs"

	"github.com/tidwall/gjson"
	"k8s.io/client-go/util/workqueue"
)

// Interface recycle informer interface
type Interface interface {
	// Pop gets head of recycle order queue
	Pop() (string, error)
}

// recycleInformer recycle informer which list and watch database and cache recycle order info
type recycleInformer struct {
	key     Key
	watchDB dal.DB
	event   stream.LoopInterface

	queue workqueue.RateLimitingInterface
}

// New create a recycle informer
func New(loopWatch stream.LoopInterface, watchDB dal.DB) (*recycleInformer, error) {
	recycleInformer := &recycleInformer{
		key:     KeyRecycle,
		watchDB: watchDB,
		event:   loopWatch,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "recycle"),
	}

	if err := recycleInformer.Run(); err != nil {
		logs.Errorf("failed to start recycle informer, err: %v", err)
		return nil, err
	}

	return recycleInformer, nil
}

// Run run recycle informer
func (i *recycleInformer) Run() error {
	return i.listAndWatchRecycleOrder()
}

// Pop gets head of recycle order queue
func (i *recycleInformer) Pop() (string, error) {
	obj, shutdown := i.queue.Get()
	if shutdown {
		return "", nil
	}

	defer i.queue.Done(obj)

	id, ok := obj.(string)
	if !ok {
		i.queue.Forget(obj)
		logs.Warnf("Expected string in queue but got %#v", obj)
		return "", errors.New("got non-int from queue")
	}

	i.queue.Forget(obj)

	return id, nil
}

func (i *recycleInformer) listAndWatchRecycleOrder() error {
	// list recycle order
	events, err := i.listRecycleOrder()
	if err != nil {
		return err
	}

	for _, event := range events {
		i.queue.Add(event)
	}

	// watch recycle order
	handler := newRecycleTokenHandler(i.key, i.watchDB)
	startTime := &types.TimeStamp{Sec: uint32(time.Now().Unix())}

	loopOpts := &types.LoopOneOptions{
		LoopOptions: types.LoopOptions{
			Name: "recycle_info",
			WatchOpt: &types.WatchOptions{
				Options: types.Options{
					EventStruct:     new(map[string]interface{}),
					Collection:      table.RecycleOrderTable,
					StartAfterToken: nil,
					StartAtTime:     startTime,
					// TODO: add failure callback
					WatchFatalErrorCallback: nil,
				},
			},
			TokenHandler: handler,
			RetryOptions: &types.RetryOptions{
				MaxRetryCount: 4,
				RetryDuration: 500 * time.Millisecond,
			},
		},
		EventHandler: &types.OneHandler{
			DoAdd:    i.onUpsert,
			DoUpdate: i.onUpsert,
			DoDelete: i.onDelete,
		},
	}

	return i.event.WithOne(loopOpts)
}

func (i *recycleInformer) listRecycleOrder() ([]string, error) {
	// TODO: query from db
	return nil, nil
}

// onInsert set recycle cache
func (i *recycleInformer) onUpsert(e *types.Event) bool {
	logs.V(5).Infof("received recycle event, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes, e.ID())

	id := gjson.GetBytes(e.DocBytes, "suborder_id").String()
	if len(id) <= 0 {
		logs.Errorf("received invalid recycle event, skip, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes,
			e.ID())
		return false
	}

	i.queue.Add(id)

	return false
}

// onDelete delete recycle cache
func (i *recycleInformer) onDelete(e *types.Event) bool {
	// TODO
	return false
}
