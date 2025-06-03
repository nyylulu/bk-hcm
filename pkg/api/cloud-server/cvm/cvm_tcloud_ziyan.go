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

package cscvm

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// CmdbHostListReq 从cc 查询主机api 接口
type CmdbHostListReq struct {
	// 为false则只拉取cmdb信息，为true则进一步拉取对应云上cvm信息
	QueryFromCloud bool           `json:"query_from_cloud"`
	AccountID      string         `json:"account_id" validate:"required"`
	Region         string         `json:"region" validate:"omitempty"`
	Zone           string         `json:"zone" validate:"omitempty"`
	CloudInstIDs   []string       `json:"inst_ids" validate:"omitempty"`
	CloudVpcIDs    []string       `json:"cloud_vpc_ids" validate:"omitempty"`
	CloudSubnetIDs []string       `json:"cloud_subnet_ids" validate:"omitempty"`
	BkSetIDs       []int64        `json:"bk_set_ids" validate:"omitempty"`
	BkModuleIDs    []int64        `json:"bk_module_ids" validate:"omitempty"`
	Page           *cmdb.BasePage `json:"page" validate:"required"`
}

// Validate CloudHostListReq.
func (req CmdbHostListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CmdbHostQueryReq 从cc 查询主机
type CmdbHostQueryReq struct {
	BkBizID        int64
	Vendor         enumor.Vendor  `json:"vendor" validate:"omitempty"`
	AccountID      string         `json:"account_id" validate:"required"`
	Region         string         `json:"region" validate:"omitempty"`
	Zone           string         `json:"zone" validate:"omitempty"`
	CloudVpcIDs    []string       `json:"cloud_vpc_ids" validate:"omitempty"`
	CloudSubnetIDs []string       `json:"cloud_subnet_ids" validate:"omitempty"`
	CloudInstIDs   []string       `json:"inst_ids" validate:"omitempty"`
	BkSetIDs       []int64        `json:"bk_set_ids" validate:"omitempty"`
	BkModuleIDs    []int64        `json:"bk_module_ids" validate:"omitempty"`
	InnerIP        []string       `json:"inner_ip" validate:"omitempty"`
	OuterIP        []string       `json:"outer_ip" validate:"omitempty"`
	InnerIPv6      []string       `json:"inner_ipv6" validate:"omitempty"`
	OuterIPv6      []string       `json:"outer_ipv6" validate:"omitempty"`
	BkHostIDs      []int64        `json:"bk_host_ids" validate:"omitempty"`
	Page           *cmdb.BasePage `json:"page" validate:"required"`
}
