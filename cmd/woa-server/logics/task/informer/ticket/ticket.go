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

// Package ticket ...
package ticket

import (
	"context"
	"errors"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/model/task"
	"hcm/cmd/woa-server/storage/dal"
	"hcm/cmd/woa-server/storage/stream"
	"hcm/cmd/woa-server/storage/stream/types"
	tasktype "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"

	"github.com/tidwall/gjson"
	"k8s.io/client-go/util/workqueue"
)

// Interface apply informer interface
type Interface interface {
	// Pop gets head of apply info queue
	Pop() (uint64, error)
}

// ticketInformer apply ticket informer which list and watch database and cache apply ticket info
type ticketInformer struct {
	key     Key
	watchDB dal.DB
	event   stream.LoopInterface
	queue   workqueue.RateLimitingInterface
}

// New creates an ticket informer
func New(loopWatch stream.LoopInterface, watchDB dal.DB) (*ticketInformer, error) {
	ticketInformer := &ticketInformer{
		key:     KeyTicket,
		watchDB: watchDB,
		event:   loopWatch,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apply"),
	}

	if err := ticketInformer.Run(); err != nil {
		logs.Errorf("failed to start ticket informer, err: %v", err)
		return nil, err
	}

	return ticketInformer, nil
}

// Run starts apply informer
func (a *ticketInformer) Run() error {
	return a.listAndWatchApplyTicket()
}

// Pop gets head of apply info queue
func (a *ticketInformer) Pop() (uint64, error) {
	obj, shutdown := a.queue.Get()
	if shutdown {
		return 0, nil
	}

	defer a.queue.Done(obj)

	id, ok := obj.(uint64)
	if !ok {
		a.queue.Forget(obj)
		logs.Warnf("Expected string in queue but got %#v", obj)
		return 0, errors.New("got non-string from queue")
	}

	a.queue.Forget(obj)

	return id, nil
}

// listAndWatchApplyTicket list and watch database and cache apply ticket into queue
func (a *ticketInformer) listAndWatchApplyTicket() error {
	// list apply ticket
	applyTickets, err := a.listApplyTicket()
	if err != nil {
		return err
	}

	for _, ticket := range applyTickets {
		a.queue.Add(ticket)
	}

	// watch apply ticket
	handler := newTicketTokenHandler(a.key, a.watchDB)
	startTime := &types.TimeStamp{Sec: uint32(time.Now().Unix())}

	loopOpts := &types.LoopOneOptions{
		LoopOptions: types.LoopOptions{
			Name: "ticket_info",
			WatchOpt: &types.WatchOptions{
				Options: types.Options{
					EventStruct:     new(map[string]interface{}),
					Collection:      common.BKTableNameApplyTicket,
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
			DoAdd:    a.onUpsert,
			DoUpdate: a.onUpsert,
			DoDelete: a.onDelete,
		},
	}

	return a.event.WithOne(loopOpts)
}

// listApplyTicket gets apply ticket list from database
func (a *ticketInformer) listApplyTicket() ([]uint64, error) {
	filter := map[string]interface{}{
		"stage": tasktype.TicketStageRunning,
	}

	page := metadata.BasePage{
		Limit: common.BKNoLimit,
	}

	tickets, err := model.Operation().ApplyTicket().FindManyApplyTicket(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to list apply ticket by filter: %+v, err: %v", filter, err)
		return nil, err
	}

	ids := make([]uint64, 0)
	for _, ticket := range tickets {
		ids = append(ids, ticket.OrderId)
	}

	return ids, nil
}

// onUpsert set or update apply ticket cache
func (a *ticketInformer) onUpsert(e *types.Event) bool {
	logs.V(5).Infof("received apply ticket event, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes, e.ID())

	// TODO: order_id as const
	id := gjson.GetBytes(e.DocBytes, "order_id").Uint()
	if id <= 0 {
		logs.Errorf("received invalid ticket event, skip, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes, e.ID())
		return false
	}

	a.queue.Add(id)

	return false
}

// onDelete delete apply ticket cache
func (a *ticketInformer) onDelete(e *types.Event) bool {
	// TODO: add exception log
	return false
}
