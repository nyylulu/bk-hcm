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

// Package task ...
package task

import (
	"hcm/cmd/woa-server/common/time"
	"hcm/cmd/woa-server/storage/dal/mongo"
	"hcm/pkg/cc"
)

// EventInfo event info
type EventInfo struct {
	EventId        int       `json:"event_id"`
	OrderId        int       `json:"order_id"`
	SubOrderId     string    `json:"suborder_id"`
	NotifyStrategy string    `json:"notify_strategy"`
	Receiver       string    `json:"receiver"`
	Message        string    `json:"message"`
	Status         string    `json:"status"`
	CreateAt       time.Time `json:"create_at"`
	UpdateAt       time.Time `json:"update_at"`
}

// Config server configs
type Config struct {
	Mongo      mongo.Config
	WatchMongo mongo.Config
	ClientConf cc.ClientConfig
}

// AuthorizedBizID 临时给前端同学进行联调的可操作的业务id reborn (213)  todo 待删除
const AuthorizedBizID = 213
