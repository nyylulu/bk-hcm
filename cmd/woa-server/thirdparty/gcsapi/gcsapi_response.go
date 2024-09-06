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

// Package gcsapi ...
package gcsapi

// GCSResp GCS check response
type GCSResp struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    GCSResult `json:"data"`
}

// GCSResult GCS check result
type GCSResult struct {
	Detail  []interface{} `json:"detail"`
	RowsNum int           `json:"rowsNum"`
}
