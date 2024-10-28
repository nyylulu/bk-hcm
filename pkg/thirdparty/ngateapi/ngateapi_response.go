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

// Package ngateapi ngate api
package ngateapi

// RecycleIPResponse recycle ip response
type RecycleIPResponse struct {
	ReturnCode int            `json:"returnCode"`
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	TraceID    string         `json:"traceId"`
	Data       []*RecycleItem `json:"data"`
}

// RecycleItem recycle item
type RecycleItem struct {
	AssertID string `json:"assertId"`
	IP       string `json:"ip"`
}
