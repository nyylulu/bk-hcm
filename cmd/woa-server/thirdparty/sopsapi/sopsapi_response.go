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
	TaskId  int    `json:"task_id"`
	TaskUrl string `json:"task_url"`
}

// GetTaskStatusResp get sops task status response
type GetTaskStatusResp struct {
	RespMeta `json:",inline"`
	Data     *GetTaskStatusRst `json:"data"`
}

// GetTaskStatusRst get sops task status result
type GetTaskStatusRst struct {
	State string `json:"state"`
}
