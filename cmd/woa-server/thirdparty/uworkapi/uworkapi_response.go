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

// Package uworkapi uwork api
package uworkapi

// UworkTicketResponse uwork response
type UworkTicketResponse struct {
	Return int       `json:"Return"`
	Detail string    `json:"Details"`
	Data   []*Ticket `json:"Data"`
}

// Ticket uwork ticket
type Ticket struct {
	TicketNo    string `json:"TicketNo"`
	ProcessDesc string `json:"ProcessDesc"`
	IsEnd       string `json:"IsEnd"`
}

// UworkProcessResponse uwork process response
type UworkProcessResponse struct {
	Return int              `json:"Return"`
	Detail string           `json:"Details"`
	Data   []*ProcessResult `json:"Data"`
}

// ProcessResult query server uwork process result
type ProcessResult struct {
	IsExist     int        `json:"IsExist"`
	AssetID     string     `json:"ServerAssetId"`
	ProcessList []*Process `json:"ProcessList"`
}

// Process uwork process info
type Process struct {
	ID   int    `json:"InstanceId"`
	Name string `json:"ProcessName"`
}
