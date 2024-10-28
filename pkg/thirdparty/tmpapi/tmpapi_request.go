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

// Package tmpapi TMP API
package tmpapi

// CheckTMPReq check TMP alarm policy request
type CheckTMPReq struct {
	Method string          `json:"method"`
	Params *CheckTMPParams `json:"params"`
}

// CheckTMPParams check TMP alarm policy parameters
type CheckTMPParams struct {
	Ip string `json:"ip"`
}

// AddShieldReq add shield TMP alarm config request
type AddShieldReq struct {
	Method string           `json:"method"`
	Params *AddShieldParams `json:"params"`
}

// AddShieldParams add shield TMP alarm config parameters
type AddShieldParams struct {
	// once 100 ips at most
	Ip       []string `json:"ip"`
	Operator string   `json:"operator"`
	// OIp request origin ip
	OIp         string `json:"o_ip"`
	Reason      string `json:"reason"`
	ShieldStart string `json:"shield_start"`
	ShieldEnd   string `json:"shield_end"`
}
