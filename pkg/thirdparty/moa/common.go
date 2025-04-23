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

package moa

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// InitiateVerificationResp the response of the verification request
type InitiateVerificationResp struct {
	SessionId string `json:"sessionId"`
}

// InitiateVerificationReq 发起验证的请求体
type InitiateVerificationReq struct {
	Username      string               `json:"username" validate:"required"`
	Channel       enumor.Moa2FAChannel `json:"channel" validate:"required"`
	Language      string               `json:"language" validate:"required"`
	PromptPayload string               `json:"promptPayload" validate:"required"`
}

// Validate ...
func (v *InitiateVerificationReq) Validate() error {
	return validator.Validate.Struct(v)
}

// VerificationReq ...
type VerificationReq struct {
	SessionId string `json:"sessionId" validate:"required"`
	Username  string `json:"username" validate:"required"`
}

// Validate ...
func (v *VerificationReq) Validate() error {
	return validator.Validate.Struct(v)
}

// VerificationResp ...
type VerificationResp struct {
	SessionId  string               `json:"sessionId"`
	Status     enumor.MoaStatus     `json:"status"`
	ButtonType enumor.MoaButtonType `json:"buttonType"`
}
