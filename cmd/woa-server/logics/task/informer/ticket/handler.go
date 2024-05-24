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

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/watch"
	"hcm/cmd/woa-server/storage/dal"
	"hcm/pkg/logs"
)

// ticketHandler apply ticket informer handler
type ticketHandler struct {
	key     Key
	watchDB dal.DB
}

// newTicketTokenHandler creates an apply ticket informer token handler
func newTicketTokenHandler(key Key, watchDB dal.DB) *ticketHandler {
	return &ticketHandler{
		key:     key,
		watchDB: watchDB,
	}
}

// SetLastWatchToken set watch token and resume time at the same time
func (h *ticketHandler) SetLastWatchToken(ctx context.Context, token string) error {
	return nil
}

// setLastWatchToken set last watch token used after events are successfully inserted
func (h *ticketHandler) setLastWatchToken(ctx context.Context, data map[string]interface{}) error {
	filter := map[string]interface{}{
		"_id": "cr_applyTicket",
	}

	// only update the need fields to avoid erasing the previous exist fields
	tokenInfo := make(mapstr.MapStr)
	for key, value := range data {
		tokenInfo[h.key.Collection()+"."+key] = value
	}

	// update id and cursor field if set, to compensate for the scenario of searching with an outdated but latest cursor
	if id, exists := data[common.BKFieldID]; exists {
		tokenInfo[common.BKFieldID] = id
	}

	if cursor, exists := data[common.BKCursorField]; exists {
		tokenInfo[common.BKCursorField] = cursor
	}

	if err := h.watchDB.Table(common.BKTableNameWatchToken).Update(ctx, filter, tokenInfo); err != nil {
		logs.Errorf("set apply ticket %s last watch token failed, err: %v, data: %+v", h.key.Collection(),
			err, tokenInfo)
		return err
	}
	return nil
}

// GetStartWatchToken get start watch token from watch token db first, if an error occurred, get from chain db
func (h *ticketHandler) GetStartWatchToken(ctx context.Context) (token string, err error) {
	filter := map[string]interface{}{
		"_id": "cr_applyTicket",
	}

	data := make(map[string]watch.LastChainNodeData)
	if err := h.watchDB.Table(common.BKTableNameWatchToken).Find(filter).Fields(h.key.Collection()).
		One(ctx, &data); err != nil {
		if !h.watchDB.IsNotFoundError(err) {
			logs.ErrorJson("get apply ticket start watch token, will get the last event's time and start watch, "+
				"err: %s, filter: %s", err, filter)
		}

		tailNode := new(watch.ChainNode)
		if err := h.watchDB.Table("cr_applyTicket").Find(nil).Fields(common.BKTokenField).
			Sort(common.BKFieldID+":-1").One(context.Background(), tailNode); err != nil {

			if !h.watchDB.IsNotFoundError(err) {
				logs.Errorf("get apply ticket last watch token from mongo failed, err: %v", err)
				return "", err
			}
			// the tail node is not exist.
			return "", nil
		}
		return tailNode.Token, nil
	}

	// check whether this field is exists or not
	node, exists := data[h.key.Collection()]
	if !exists {
		// watch from now on.
		return "", nil
	}

	return node.Token, nil
}
