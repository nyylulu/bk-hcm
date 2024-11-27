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

// Package alarmapi Alarm API
package alarmapi

// CheckAlarmReq check alarm policy request
type CheckAlarmReq struct {
	Method string            `json:"method"`
	Params *CheckAlarmParams `json:"params"`
}

// CheckAlarmParams check alarm policy parameters
type CheckAlarmParams struct {
	Ip string `json:"ip"`
}

// AddShieldReq add shield alarm config request
type AddShieldReq struct {
	Method string           `json:"method"`
	Params *AddShieldParams `json:"params"`
}

// AddShieldParams add shield alarm config parameters
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

// DelShieldReq del shield alarm config request
type DelShieldReq struct {
	Method string           `json:"method"`
	Params *DelShieldParams `json:"params"`
}

// DelShieldParams del shield alarm config parameters
type DelShieldParams struct {
	ID       string `json:"id"`
	Operator string `json:"operator"`
	// OIp request origin ip
	OIp string `json:"o_ip"`
}
