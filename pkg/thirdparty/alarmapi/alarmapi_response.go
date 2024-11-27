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

// AlarmShieldConfig alarm shield config
type AlarmShieldConfig struct {
	CiSetInfo  string `json:"ciset_info"`
	ShieldRule string `json:"shield_rule"`
	CycleEnd   string `json:"cycle_end"`
}

// AddShieldResp add shield alarm config response
type AddShieldResp struct {
	Code int    `json:"returnCode"`
	Msg  string `json:"msg"`
	Data int64  `json:"data"`
}
