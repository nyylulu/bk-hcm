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

import "hcm/pkg/criteria/validator"

// InitiateVerificationReq ...
type InitiateVerificationReq struct {
	Username      string `json:"username" validate:"required"`
	Channel       string `json:"channel" validate:"required"`
	Language      string `json:"language" validate:"required"`
	PromptPayload string `json:"prompt_payload" validate:"required"`
}

// Validate ...
func (m *InitiateVerificationReq) Validate() error {
	return validator.Validate.Struct(m)
}

// InitiateVerificationResp the response of the verification request
type InitiateVerificationResp struct {
	SessionId string `json:"session_id"`
}

// VerificationReq ...
type VerificationReq struct {
	SessionId string `json:"session_id" validate:"required"`
	Username  string `json:"username" validate:"required"`
}

// Validate ...
func (v *VerificationReq) Validate() error {
	return validator.Validate.Struct(v)
}

// VerificationResp ...
type VerificationResp struct {
	SessionId  string `json:"session_id"`
	Status     string `json:"status"`
	ButtonType string `json:"button_type"`
}
