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

// Package metadata ...
package metadata

import (
	"time"

	"hcm/cmd/woa-server/common/mapstr"
)

// ID ...
type ID struct {
	ID string `json:"id"`
}

// IDResult ...
type IDResult struct {
	BaseResp `json:",inline"`
	Data     ID `json:"data"`
}

// HostInstanceResult ...
type HostInstanceResult struct {
	BaseResp `json:",inline"`
	Data     map[string]interface{} `json:"data"`
}

// FavoriteResult ...
type FavoriteResult struct {
	Count uint64                   `json:"count"`
	Info  []map[string]interface{} `json:"info"`
}

// GetHostFavoriteResult ...
type GetHostFavoriteResult struct {
	BaseResp `json:",inline"`
	Data     FavoriteResult `json:"data"`
}

// GetHostFavoriteWithIDResult ...
type GetHostFavoriteWithIDResult struct {
	BaseResp `json:",inline"`
	Data     FavouriteMeta `json:"data"`
}

// HistoryContent ...
type HistoryContent struct {
	Content string `json:"content"`
}

// AddHistoryResult ...
type AddHistoryResult struct {
	BaseResp `json:",inline"`
	Data     ID `json:"data"`
}

// HistoryMeta ...
type HistoryMeta struct {
	ID         string    `json:"id,omitempty" bson:"id,omitempty" `
	User       string    `json:"user,omitempty" bson:"user,omitempty"`
	Content    string    `json:"content,omitempty" bson:"content,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty" bson:"create_time,omitempty"`
	OwnerID    string    `json:"bk_supplier_account" bson:"bk_supplier_account"`
}

// HistoryResult ...
type HistoryResult struct {
	Count uint64        `json:"count"`
	Info  []HistoryMeta `json:"info"`
}

// GetHistoryResult ...
type GetHistoryResult struct {
	BaseResp `json:",inline"`
	Data     HistoryResult `json:"data"`
}

// HostInfo ...
type HostInfo struct {
	Count int             `json:"count"`
	Info  []mapstr.MapStr `json:"info"`
}

// GetHostsResult ...
type GetHostsResult struct {
	BaseResp `json:",inline"`
	Data     HostInfo `json:"data"`
}

// GetHostModuleIDsResult ...
type GetHostModuleIDsResult struct {
	BaseResp `json:",inline"`
	Data     []int64 `json:"data"`
}

// ParamData ...
type ParamData struct {
	ApplicationID       int64   `json:"bk_biz_id"`
	HostID              []int64 `json:"bk_host_id"`
	OwnerModuleID       int64   `json:"bk_owner_module_id"`
	OwnerAppplicationID int64   `json:"bk_owner_biz_id"`
}

// AssignHostToAppParams ...
type AssignHostToAppParams struct {
	ApplicationID      int64   `json:"bk_biz_id"`
	HostID             []int64 `json:"bk_host_id"`
	ModuleID           int64   `json:"bk_module_id"`
	OwnerApplicationID int64   `json:"bk_owner_biz_id"`
	OwnerModuleID      int64   `json:"bk_owner_module_id"`
}

// ModuleHost ...
type ModuleHost struct {
	AppID    int64  `json:"bk_biz_id,omitempty" bson:"bk_biz_id"`
	HostID   int64  `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	ModuleID int64  `json:"bk_module_id,omitempty" bson:"bk_module_id"`
	SetID    int64  `json:"bk_set_id,omitempty" bson:"bk_set_id"`
	OwnerID  string `json:"bk_supplier_account,omitempty" bson:"bk_supplier_account"`
}

// HostConfig ...
type HostConfig struct {
	BaseResp `json:",inline"`
	Data     HostConfigData `json:"data"`
}

// HostConfigData ...
type HostConfigData struct {
	Count int64        `json:"count"`
	Info  []ModuleHost `json:"data"`
	Page  BasePage     `json:"page"`
}

// HostModuleResp ...
type HostModuleResp struct {
	BaseResp `json:",inline"`
	Data     []ModuleHost `json:"data"`
}

// ModuleHostConfigParams ...
type ModuleHostConfigParams struct {
	ApplicationID int64   `json:"bk_biz_id"`
	HostID        int64   `json:"bk_host_id"`
	ModuleID      []int64 `json:"bk_module_id"`
	OwnerID       string  `json:"bk_supplier_account" bson:"bk_supplier_account"`
}

// GetUserCustomResult get user custom result
type GetUserCustomResult struct {
	BaseResp `json:",inline"`
	Data     map[string]interface{} `json:"data"`
}

// FavouriteParms get user custom result
type FavouriteParms struct {
	ID          string `json:"id,omitempty"`
	Info        string `json:"info,omitempty"`
	QueryParams string `json:"query_params,omitempty"`
	Name        string `json:"name,omitempty"`
	IsDefault   int    `json:"is_default,omitempty"`
	Count       int    `json:"count,omitempty"`
	BizID       int64  `json:"bk_biz_id"`
}

// FavouriteMeta favourite meta
type FavouriteMeta struct {
	BizID       int64     `json:"bk_biz_id" bson:"bk_biz_id"`
	ID          string    `json:"id,omitempty" bson:"id,omitempty"`
	Info        string    `json:"info,omitempty" bson:"info,omitempty"`
	Name        string    `json:"name,omitempty" bson:"name,omitempty"`
	Count       int       `json:"count,omitempty" bson:"count,omitempty"`
	User        string    `json:"user,omitempty" bson:"user,omitempty"`
	OwnerID     string    `json:"bk_supplier_account,omitempty" bson:"bk_supplier_account,omitempty"`
	QueryParams string    `json:"query_params,omitempty" bson:"query_params,omitempty"`
	CreateTime  time.Time `json:"create_time,omitempty" bson:"create_time,omitempty"`
	UpdateTime  time.Time `json:"last_time,omitempty" bson:"last_time,omitempty"`
}

// TransferHostToInnerModule transfer host to inner module eg:idle module ,fault module
type TransferHostToInnerModule struct {
	ApplicationID int64   `json:"bk_biz_id"`
	ModuleID      int64   `json:"bk_module_id"`
	HostID        []int64 `json:"bk_host_id"`
}

// DistinctIDResponse distinct id response
type DistinctIDResponse struct {
	BaseResp `json:",inline"`
	Data     DistinctID `json:"data"`
}

// DistinctID distinct id
type DistinctID struct {
	IDArr []int64 `json:"id_arr"`
}
