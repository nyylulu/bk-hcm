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

// Package notice ...
package notice

import (
	"errors"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/storage/dal"
	"hcm/cmd/woa-server/storage/stream"
	"hcm/cmd/woa-server/storage/stream/types"
	"hcm/pkg/logs"

	"github.com/tidwall/gjson"
	"k8s.io/client-go/util/workqueue"
)

// Interface notice informer interface
type Interface interface {
	// Pop gets head of notice queue
	Pop() (string, error)
}

// noticeInformer notice informer which list and watch database and cache notice info
type noticeInformer struct {
	key     Key
	watchDB dal.DB
	event   stream.LoopInterface

	queue workqueue.RateLimitingInterface
}

// New create an notice informer
func New(loopWatch stream.LoopInterface, watchDB dal.DB) (*noticeInformer, error) {
	noticeInformer := &noticeInformer{
		key:     KeyNotice,
		watchDB: watchDB,
		event:   loopWatch,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "generate"),
	}

	if err := noticeInformer.Run(); err != nil {
		logs.Errorf("failed to start generate informer, err: %v", err)
		return nil, err
	}

	return noticeInformer, nil
}

// Run starts notice informer
func (i *noticeInformer) Run() error {
	return i.listAndWatchNotice()
}

// Pop gets head of notice queue
func (i *noticeInformer) Pop() (string, error) {
	obj, shutdown := i.queue.Get()
	if shutdown {
		return "", nil
	}

	defer i.queue.Done(obj)

	id, ok := obj.(string)
	if !ok {
		i.queue.Forget(obj)
		logs.Warnf("expected string in queue but got %#v", obj)
		return "", errors.New("got non-string from queue")
	}

	i.queue.Forget(obj)

	return id, nil
}

// listAndWatchNotice list and watch database and cache notice into queue
func (i *noticeInformer) listAndWatchNotice() error {
	// list notice
	notices, err := i.listNotice()
	if err != nil {
		return err
	}

	for _, notice := range notices {
		i.queue.Add(notice)
	}

	// watch notice
	handler := newEventTokenHandler(i.key, i.watchDB)
	startTime := &types.TimeStamp{Sec: uint32(time.Now().Unix())}

	loopOpts := &types.LoopOneOptions{
		LoopOptions: types.LoopOptions{
			Name: "notice_info",
			WatchOpt: &types.WatchOptions{
				Options: types.Options{
					EventStruct:     new(map[string]interface{}),
					Collection:      common.BKTableNameNoticeInfo,
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

// listNotice gets notice list from database
func (i *noticeInformer) listNotice() ([]string, error) {
	// TODO: query from db
	return nil, nil
}

// onUpsert set or update notice cache
func (i *noticeInformer) onUpsert(e *types.Event) bool {
	logs.V(5).Infof("received notify notice, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes, e.ID())

	id := gjson.GetBytes(e.DocBytes, "notice_id").String()
	if len(id) <= 0 {
		logs.Errorf("received invalid notice event, skip, op: %s, doc: %s, rid: %s", e.OperationType, e.DocBytes,
			e.ID())
		return false
	}

	i.queue.Add(id)

	return false
}

// onDelete delete notice cache
func (i *noticeInformer) onDelete(e *types.Event) bool {
	// TODO
	return false
}
