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

// Package bkdbm ...
package bkdbm

// ListMachinePool list machine pool.
type ListMachinePool struct {
	IPs    []string `json:"ips"`
	Offset int64    `json:"offset,required"`
	Limit  int64    `json:"limit,required"`
}

// ListMachinePoolResp the response of the list machine pool.
type ListMachinePoolResp struct {
	Count   int64               `json:"count"`
	Results []MachinePoolResult `json:"results"`
}

// MachinePoolResult the result of the machine pool.
type MachinePoolResult struct {
	BkHostID    int64       `json:"bk_host_id"`
	BkCloudID   int64       `json:"bk_cloud_id"`
	IP          string      `json:"ip"`
	Ticket      int64       `json:"ticket"` // 关联单据ID
	City        string      `json:"city"`
	SubZone     string      `json:"sub_zone"`
	AgentStatus int64       `json:"agent_status"`
	Pool        string      `json:"pool"` // 所处主机池
	LatestEvent LatestEvent `json:"latest_event"`
	Creator     string      `json:"creator"`
	CreateAt    string      `json:"create_at"`
	Updater     string      `json:"updater"`
	UpdateAt    string      `json:"update_at"`
}

// LatestEvent the latest event
type LatestEvent struct {
	ID       int64  `json:"id"`
	BkBizID  int64  `json:"bk_biz_id"`
	BkHostID int64  `json:"bk_host_id"`
	IP       string `json:"ip"`
	Event    string `json:"event"`
	Ticket   int64  `json:"ticket"`
	Creator  string `json:"creator"`
	Updater  string `json:"updater"`
}
