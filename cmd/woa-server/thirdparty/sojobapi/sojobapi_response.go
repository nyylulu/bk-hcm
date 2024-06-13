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

package sojobapi

// RespMeta cc response meta info
type RespMeta struct {
	// return 0 if success
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CreateJobResp create so job response
type CreateJobResp struct {
	RespMeta `json:",inline"`
	// return job id if create so job success
	Data interface{} `json:"data"`
}

// GetJobStatusResp get so job status response
type GetJobStatusResp struct {
	RespMeta `json:",inline"`
	Data     *JobRst `json:"data"`
}

// JobRst job result
type JobRst struct {
	// 0: success, 1: handling, 2: failed
	CheckResultCode JobStatus    `json:"check_result_code"`
	SubJob          []*SubJobRst `json:"sub_job"`
}

type JobStatus int

const (
	JobStatusSuccess  JobStatus = 0
	JobStatusHandling JobStatus = 1
	JobStatusFailed   JobStatus = 2
)

// SubJobRst sub job result
type SubJobRst struct {
	Ip string `json:"ip"`
	// "doing", "success", "failed"
	Result    string `json:"result"`
	Comment   string `json:"comment"`
	OtherInfo string `json:"other_info"`
}

// GetJobStatusDetailResp get so job status detail info response
type GetJobStatusDetailResp struct {
	RespMeta `json:",inline"`
	Data     *JobDetailRst `json:"data"`
}

// JobDetailRst job detail result
type JobDetailRst struct {
	Status     JobDetailStatus    `json:"status"`
	SubJobNum  int                `json:"subjob_num"`
	SubJobSucc int                `json:"subjob_succ"`
	SubJob     []*SubJobDetailRst `json:"subJobList"`
}

type JobDetailStatus string

const (
	JobDetailStatusTodo  JobDetailStatus = "todo"
	JobDetailStatusStart JobDetailStatus = "start"
	JobDetailStatusDoing JobDetailStatus = "doing"
	JobDetailStatusDone  JobDetailStatus = "done"
)

// SubJobDetailRst sub job detail result
type SubJobDetailRst struct {
	Ip string `json:"ip"`
	// "doing", "success", "failed"
	Status      JobDetailStatus `json:"status"`
	Result      string          `json:"result"`
	Comment     string          `json:"comment"`
	CommentMore string          `json:"comment_more"`
	IsFailed    bool            `json:"is_failed"`
}
