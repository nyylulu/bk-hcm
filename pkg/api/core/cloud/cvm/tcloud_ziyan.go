/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

package cvm

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// QueryCloudCvmReq 查询云上cvm信息
type QueryCloudCvmReq struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"required"`
	CvmIDs    []string      `json:"cvm_ids"`
	// 安全组id
	SGIDs []string       `json:"security_groups_ids"`
	Page  *core.BasePage `json:"page" validate:"required"`
}

// Validate ...
func (r QueryCloudCvmReq) Validate() error {

	return validator.Validate.Struct(r)
}

// TCloudZiyanCvmExtension 自研云cvm拓展
type TCloudZiyanCvmExtension struct {
	*TCloudCvmExtension `json:",inline"`
	SecurityGroupNames  []string `json:"security_group_names"`
}

// TCloudZiyanHostExtension 内部版从cc同步的自研云的主机
type TCloudZiyanHostExtension struct {
	*TCloudCvmExtension `json:",inline"`
	HostName            string               `json:"bk_host_name"`       // CC主机名称
	SvrSourceTypeID     cmdb.SvrSourceTypeID `json:"svr_source_type_id"` // 服务器来源类型ID
	SrvStatus           string               `json:"srv_status"`         // CC的运营状态
	SvrDeviceClass      string               `json:"svr_device_class"`   // 机型
	BkDisk              float64              `json:"bk_disk"`            // 磁盘容量(GB)
	BkCpu               int64                `json:"bk_cpu"`             // CPU逻辑核心数
	BkOSName            string               `json:"bk_os_name"`         // 操作系统名称
	Operator            string               `json:"operator"`           // 主负责人
	BkBakOperator       string               `json:"bk_bak_operator"`    // 备份负责人
}
