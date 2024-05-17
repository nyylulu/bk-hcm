/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xshipapi

// ReinstallReq create reinstall task request
type ReinstallReq struct {
	Assets  []*Asset `json:"assetList"`
	Starter string   `json:"starter"`
}

// Asset asset info
type Asset struct {
	AssetID   string     `json:"assetId"`
	Variables *Variables `json:"variables"`
}

// Variables reinstall variables specification
type Variables struct {
	OsVersion string `json:"osVersion"`
	Raid      string `json:"raid"`
	Password  string `json:"password"`
}

// GetReinstallStatusReq get host reinstall task request
type GetReinstallStatusReq struct {
	OrderIDs []string `json:"orderList"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}
