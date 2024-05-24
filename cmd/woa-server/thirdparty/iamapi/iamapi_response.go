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

package iamapi

// RespMeta cc response meta info
type RespMeta struct {
	Result  bool   `json:"result" mapstructure:"result"`
	Code    int    `json:"code" mapstructure:"code"`
	Message string `json:"message" mapstructure:"message"`
}

// AuthVerifyResp auth policy verify response
type AuthVerifyResp struct {
	RespMeta `json:",inline"`
	Data     *Decision `json:"data"`
}

// Decision auth policy verify decision
type Decision struct {
	Allowed bool `json:"allowed"`
}

// GetAuthUrlResp get auth url response
type GetAuthUrlResp struct {
	RespMeta `json:",inline"`
	Data     *AuthUrl `json:"data"`
}

// AuthUrl get auth url
type AuthUrl struct {
	Url string `json:"url"`
}
