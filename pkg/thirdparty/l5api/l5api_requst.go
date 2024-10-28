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

// Package l5api l5 api
package l5api

// L5Req l5 request
type L5Req struct {
	Version       int          `json:"version"`
	ComponentName string       `json:"componentName"`
	User          string       `json:"user"`
	EventId       int          `json:"eventId"`
	Interface     *L5Interface `json:"interface"`
}

// L5Interface l5 interface
type L5Interface struct {
	InterfaceName string   `json:"interfaceName"`
	Para          *L5Param `json:"para"`
}

// L5Param l5 param
type L5Param struct {
	Ip string `json:"ip"`
}
