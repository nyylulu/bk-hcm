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

// QueryServerEventReq query server uwork event request
type QueryServerEventReq struct {
	Action   string                  `json:"Action"`
	FlowId   string                  `json:"FlowId"`
	Starter  string                  `json:"Starter"`
	SystemId string                  `json:"SystemId"`
	Data     *QueryServerEventParams `json:"Data"`
}

// QueryServerEventParams query server uwork event request parameters
type QueryServerEventParams struct {
	ResultColumns   *ResultColumns   `json:"ResultColumns"`
	SearchCondition *SearchCondition `json:"SearchCondition"`
}

// ResultColumns query server uwork event result fields
type ResultColumns struct {
	TicketNo    string `json:"TicketNo"`
	ProcessDesc string `json:"ProcessDesc"`
	IsEnd       string `json:"IsEnd"`
}

// SearchCondition query server uwork event search condition
type SearchCondition struct {
	ServerIP string `json:"ServerIP"`
}

// QueryServerProcessReq query server uwork process request
type QueryServerProcessReq struct {
	Action   string                    `json:"Action"`
	Method   string                    `json:"Method"`
	SystemId string                    `json:"SystemId"`
	Data     *QueryServerProcessParams `json:"Data"`
}

// QueryServerProcessParams query server uwork process request parameters
type QueryServerProcessParams struct {
	AssetID string `json:"ServerAssetId"`
}
