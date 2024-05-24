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

// L5Resp response
type L5Resp struct {
	TimeStamp     int64    `json:"timestamp"`
	EventId       int64    `json:"eventId"`
	ReturnCode    int      `json:"returnCode"`
	ReturnMessage string   `json:"returnMessage"`
	Data          *SIDList `json:"data"`
}

// SIDList sid info list
type SIDList struct {
	SIDList []*SIDInfo `json:"sidList"`
}

// SIDInfo sid info
type SIDInfo struct {
	ModId string `json:"modId"`
	CmdId string `json:"cmdId"`
}
