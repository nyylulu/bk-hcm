/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// RespMeta itsm response meta info
type RespMeta struct {
	Result bool   `json:"result" mapstructure:"result"`
	Code   int    `json:"code" mapstructure:"code"`
	ErrMsg string `json:"message" mapstructure:"message"`
}

// CreateTicketResp create itsm ticket response
type CreateTicketResp struct {
	RespMeta `json:",inline"`
	Data     *CreateTicketRst `json:"data"`
}

// CreateTicketRst create itsm ticket result
type CreateTicketRst struct {
	Sn string `json:"sn"`
}

// OperateTicketResp operate itsm ticket node response
type OperateNodeResp struct {
	RespMeta `json:",inline"`
	Data     interface{} `json:"data"`
}

// GetTicketStatusRst get itsm ticket status result
type GetTicketStatusRst struct {
	CurrentStatus string        `json:"current_status"`
	TicketUrl     string        `json:"ticket_url"`
	CurrentSteps  []*TicketStep `json:"current_steps"`
}

// TicketStep itsm ticket step
type TicketStep struct {
	Name       string `json:"name"`
	Processors string `json:"processors"`
	StateId    int64  `json:"state_id"`
}

// GetTicketLogResp get itsm ticket logs response
type GetTicketLogResp struct {
	RespMeta `json:",inline"`
	Data     *GetTicketLogRst `json:"data"`
}

// GetTicketLogRst get itsm ticket logs result
type GetTicketLogRst struct {
	Sn       string      `json:"sn"`
	Title    string      `json:"title"`
	CreateAt string      `json:"create_at"`
	Creator  string      `json:"creator"`
	Logs     []*AuditLog `json:"logs"`
}

// AuditLog itsm ticket log
type AuditLog struct {
	Operator  string `json:"operator"`
	Message   string `json:"message"`
	OperateAt string `json:"operate_at"`
	Source    string `json:"source"`
}
