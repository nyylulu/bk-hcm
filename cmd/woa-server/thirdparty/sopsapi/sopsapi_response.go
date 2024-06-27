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

// Package sopsapi sops api
package sopsapi

// RespMeta sops response meta info
type RespMeta struct {
	Result  bool   `json:"result" mapstructure:"result"`
	Code    int    `json:"code" mapstructure:"code"`
	Message string `json:"message" mapstructure:"message"`
}

// CreateTaskResp create sops task response
type CreateTaskResp struct {
	RespMeta `json:",inline"`
	Data     *CreateTaskRst `json:"data"`
}

// CreateTaskRst create sops task result
type CreateTaskRst struct {
	TaskId  int64  `json:"task_id"`
	TaskUrl string `json:"task_url"`
}

// StartTaskResp start sops task response
type StartTaskResp struct {
	RespMeta `json:",inline"`
	Data     *StartTaskRst `json:"data"`
}

// StartTaskRst start sops task result
type StartTaskRst struct{}

// GetTaskStatusResp get sops task status response
type GetTaskStatusResp struct {
	RespMeta `json:",inline"`
	Data     *GetTaskStatusRst `json:"data"`
}

// GetTaskStatusRst get sops task status result
type GetTaskStatusRst struct {
	// ID 节点ID
	ID string `json:"id"`
	// State (CREATED:未执行 RUNNING:执行中 FAILED:失败 SUSPENDED:暂停 REVOKED:已终止 FINISHED:已完成)
	State    string                            `json:"state"`
	Children map[string]GetTaskNodeChildrenRst `json:"children"`
}

// GetTaskNodeChildrenRst get sops task node children result
type GetTaskNodeChildrenRst struct {
	// ID 节点ID
	ID string `json:"id"`
	// State 最后一次执行状态，CREATED：未执行，RUNNING：执行中，FAILED：失败，NODE_SUSPENDED：暂停，FINISHED：成功
	State string `json:"state"`
	// RootID 根结点ID
	RootID string `json:"root_id"`
	// ParentID 父节点ID
	ParentID string `json:"parent_id"`
	// Retry 重试次数
	Retry int64 `json:"retry"`
	// Skip 是否跳过
	Skip bool `json:"skip"`
	// ElapsedTime elapsed time
	ElapsedTime int64 `json:"elapsed_time"`
	// StartTime 最后一次执行开始时间
	StartTime string `json:"start_time"`
	// FinishTime 最后一次执行结束时间
	FinishTime string `json:"finish_time"`
}

// GetTaskNodeDetailResp get sops task node detail response
type GetTaskNodeDetailResp struct {
	RespMeta `json:",inline"`
	Data     *GetTaskNodeDetailRst `json:"data"`
}

// GetTaskNodeDetailRst get sops task node detail result
type GetTaskNodeDetailRst struct {
	// ID 节点ID
	ID string `json:"id"`
	// StartTime 最后一次执行开始时间
	StartTime string `json:"start_time"`
	// FinishTime 最后一次执行结束时间
	FinishTime string `json:"finish_time"`
	// State 最后一次执行状态，CREATED：未执行，RUNNING：执行中，FAILED：失败，NODE_SUSPENDED：暂停，FINISHED：成功
	State string `json:"state"`
	// Outputs 输出参数
	Outputs []TaskNodeDetailOutput `json:"outputs"`
	// ExData 节点执行失败详情，json字符串或者HTML字符串、普通字符串
	ExData string `json:"ex_data"`
}

// TaskNodeDetailOutput task node detail output
type TaskNodeDetailOutput struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Key    string `json:"key"`
	Preset bool   `json:"preset"`
}

// GetTaskNodeDataResp get sops task node data response
type GetTaskNodeDataResp struct {
	RespMeta `json:",inline"`
	Data     *GetTaskNodeDataRst `json:"data"`
}

// GetTaskNodeDataRst get sops task node data result
type GetTaskNodeDataRst struct {
	// Outputs 输出参数
	Outputs []TaskNodeDetailOutput `json:"outputs"`
	// ExData 节点执行失败详情，json字符串或者HTML字符串、普通字符串
	ExData string `json:"ex_data"`
}
