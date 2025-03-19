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

package loadbalancer

import (
	"hcm/pkg/criteria/validator"
)

// ClbZiyanMianliuTgwGroupName 免流参数指定的Tgw 集群名参数
const ClbZiyanMianliuTgwGroupName = "ziyan_mianliu"

// TCloudZiyanCreateClbOption 自研云负载均衡创建参数
type TCloudZiyanCreateClbOption struct {
	TCloudCreateClbOption `json:",inline"`
	// 直通参数
	ZhiTong *bool `json:"zhi_tong"`
	// 内网多可用区
	Zones []string `json:"zones"`
	// 免流时使用 TgwGroupName='ziyan_mianliu'
	TgwGroupName *string `json:"tgw_group_name"`
}

// Validate tcloud clb create option.
func (opt TCloudZiyanCreateClbOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDescribeSlaCapacityOption 定义查询性能保障规格参数
type TCloudDescribeSlaCapacityOption struct {
	Region   string   `json:"region" validate:"required"`
	SlaTypes []string `json:"sla_types" validate:"omitempty"`
}

// Validate ...
func (opt TCloudDescribeSlaCapacityOption) Validate() error {
	return validator.Validate.Struct(opt)
}
