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

// Package safetyapi provides ...
package safetyapi

const (
	appKey    = "xTwXcRDoKcskVbOp"
	appSecret = "eagD7mQqHUOwUYY6fZKJKrX9Zo3f7JF9"
	apiName   = "baseline_get_task_data_new"
)

// BaseLineReq is struct of security baseline request
type BaseLineReq struct {
	TaskId        int    `json:"task_id"`
	Ip            string `json:"ip"`
	BusinessGroup string `json:"bg"`
	Department    string `json:"dept"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}
