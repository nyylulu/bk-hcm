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

// Package sampwdapi sampwd api
package sampwdapi

import "hcm/pkg/criteria/validator"

// UpdateHostPwdReq 更新密码库中的主机密码-请求体
type UpdateHostPwdReq struct {
	BkHostID     int64  `json:"bk_host_id" validate:"required"`
	Password     string `json:"password" validate:"required"`
	UserName     string `json:"username" validate:"required"`
	GenerateTime string `json:"generate_time" validate:"omitempty"` // 格式为YYYY-MM-DD HH:MM:SS
}

// Validate ...
func (v *UpdateHostPwdReq) Validate() error {
	return validator.Validate.Struct(v)
}

// UpdateHostPwdResp the response of the request
type UpdateHostPwdResp struct {
	ID int `json:"id"`
}
