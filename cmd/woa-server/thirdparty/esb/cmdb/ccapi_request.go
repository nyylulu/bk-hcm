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

// Package cmdb CC API response
package cmdb

import (
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/querybuilder"
)

// AddHostReq add host to cc request
type AddHostReq struct {
	// to be added hosts' asset id list, max length is 10
	AssetIDs []string `json:"asset_ids"`
	InnerIps []string `json:"inner_ips"`
}

// TransferHostReq transfer host to another business request
type TransferHostReq struct {
	From TransferHostSrcInfo `json:"bk_from"`
	To   TransferHostDstInfo `json:"bk_to"`
}

// TransferHostSrcInfo transfer host source info
type TransferHostSrcInfo struct {
	FromBizID int64   `json:"bk_biz_id"`
	HostIDs   []int64 `json:"bk_host_ids"`
}

// TransferHostDstInfo transfer host destination info
type TransferHostDstInfo struct {
	ToBizID    int64 `json:"bk_biz_id"`
	ToModuleID int64 `json:"bk_module_id,omitempty"`
}

// ListHostReq list host request
type ListHostReq struct {
	HostPropertyFilter *querybuilder.QueryFilter `json:"host_property_filter"`
	Fields             []string                  `json:"fields"`
	Page               BasePage                  `json:"page"`
}

// BasePage for paging query
type BasePage struct {
	Sort  string `json:"sort,omitempty" mapstructure:"sort"`
	Limit int    `json:"limit,omitempty" mapstructure:"limit"`
	Start int    `json:"start" mapstructure:"start"`
}

// UpdateHostsReq update hosts request
type UpdateHostsReq struct {
	Update []*UpdateHostProperty `json:"update"`
}

// UpdateHostProperty update hosts property
type UpdateHostProperty struct {
	HostID     int64                  `json:"bk_host_id"`
	Properties map[string]interface{} `json:"properties"`
}

// HostModuleRelationParameter get host and module relation parameter
type HostModuleRelationParameter struct {
	HostID []int64 `json:"bk_host_id"`
}

// ListBizHostReq list certain business host request
type ListBizHostReq struct {
	BkBizId            int64                     `json:"bk_biz_id"`
	BkModuleIds        []int64                   `json:"bk_module_ids"`
	ModuleCond         []ConditionItem           `json:"module_cond"`
	HostPropertyFilter *querybuilder.QueryFilter `json:"host_property_filter,omitempty"`
	Fields             []string                  `json:"fields"`
	Page               BasePage                  `json:"page"`
}

// ConditionItem cc query condition item
type ConditionItem struct {
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// HostBizRelReq find host business relation request
type HostBizRelReq struct {
	BkHostId []int64 `json:"bk_host_id"`
}

// SearchBizReq search business request
type SearchBizReq struct {
	Filter *querybuilder.QueryFilter `json:"biz_property_filter,omitempty"`
	Fields []string                  `json:"fields"`
	Page   BasePage                  `json:"page"`
}

// SearchModuleReq search module request
type SearchModuleReq struct {
	BkBizId   int64         `json:"bk_biz_id"`
	Condition mapstr.MapStr `json:"condition"`
	Fields    []string      `json:"fields"`
	Page      BasePage      `json:"page"`
}

// GetBizInternalModuleReq get business's internal module request
type GetBizInternalModuleReq struct {
	BkBizID int64 `json:"bk_biz_id"`
}

// CrTransitReq transfer host to CR transit module request
type CrTransitReq struct {
	From CrTransitSrcInfo `json:"bk_from"`
	To   CrTransitDstInfo `json:"bk_to"`
}

// CrTransitSrcInfo transfer host source info
type CrTransitSrcInfo struct {
	FromBizID    int64 `json:"bk_biz_id"`
	FromModuleID int64 `json:"bk_module_id"`
	// max size is 10
	AssetIDs []string `json:"asset_ids"`
}

// CrTransitDstInfo transfer host destination info
type CrTransitDstInfo struct {
	ToBizID int64 `json:"bk_biz_id"`
}

// CrTransitIdleReq transfer host from CR transit module back to idle module request
type CrTransitIdleReq struct {
	BkBizId  int64    `json:"bk_biz_id"`
	AssetIDs []string `json:"asset_ids"`
}
