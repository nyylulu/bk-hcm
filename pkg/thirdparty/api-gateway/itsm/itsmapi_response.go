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

// OperateNodeResp operate itsm ticket node response
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

// TicketStep itsm step
type TicketStep struct {
	Name           string      `json:"name"`
	StateId        int64       `json:"state_id"`
	Processors     string      `json:"processors"`
	Status         string      `json:"status,omitempty"`
	ActionType     string      `json:"action_type,omitempty"`
	Tag            string      `json:"tag,omitempty"`
	ProcessorsType string      `json:"processors_type,omitempty"`
	Fields         []StepField `json:"fields,omitempty"`
	Operations     []Operation `json:"operations,omitempty"`
}

// StepField itsm step field
type StepField struct {
	Id            int           `json:"id"`
	IsReadonly    bool          `json:"is_readonly"`
	SourceType    string        `json:"source_type"`
	SourceUri     string        `json:"source_uri"`
	ApiInstanceId int           `json:"api_instance_id"`
	Type          string        `json:"type"`
	Key           string        `json:"key"`
	Name          string        `json:"name"`
	ValidateType  string        `json:"validate_type"`
	Regex         string        `json:"regex"`
	RegexConfig   any           `json:"regex_config"`
	CustomRegex   string        `json:"custom_regex"`
	Desc          string        `json:"desc"`
	Choice        []Choice      `json:"choice"`
	Meta          StepFieldMeta `json:"meta"`
	WorkflowId    int           `json:"workflow_id"`
	// ITSM BUG state_id 字段 可能会在一次接口返回中同时出现int或字符串类型，这里忽略该字段避免反序列化失败
	// StateId       int           `json:"state_id"`
	Source string `json:"source"`
}

// Operation itsm operation
type Operation struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	CanOperate bool   `json:"can_operate"`
}

// StepFieldMeta itsm step field meta
type StepFieldMeta struct {
	Code string `json:"code,omitempty"`
}

// Choice itsm choice
type Choice struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ApproveNodeResult defines the itsm ticket approve node result response
type ApproveNodeResult struct {
	Name          string `json:"name"`
	ProcessedUser string `json:"processed_user"`
	ApproveResult bool   `json:"approve_result"`
	ApproveRemark string `json:"approve_remark"`
}
