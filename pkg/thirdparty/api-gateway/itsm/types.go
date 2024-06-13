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

package itsm

type Status string

const (
	// StatusRunning 处理中
	StatusRunning Status = "RUNNING"
	// StatusFinished 已结束
	StatusFinished Status = "FINISHED"
	// StatusTerminated 被终止
	StatusTerminated Status = "TERMINATED"
	// StatusSuspended 被挂起
	StatusSuspended Status = "SUSPENDED"
)

// GetTicketStatusResp get itsm ticket status response
type GetTicketStatusResp struct {
	CurrentStatus Status        `json:"current_status"`
	TicketUrl     string        `json:"ticket_url"`
	CurrentSteps  []*TicketStep `json:"current_steps"`
}

// TicketStep itsm ticket step
type TicketStep struct {
	Name       string `json:"name"`
	Processors string `json:"processors"`
	StateID    int64  `json:"state_id"`
}
