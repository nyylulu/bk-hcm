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

// Package xrayapi xray api
package xrayapi

import "hcm/pkg/criteria/enumor"

// QueryFaultTicketResponse query fault ticket response
type QueryFaultTicketResponse struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	TraceID string    `json:"traceId"`
	Data    []*Ticket `json:"data"`
}

// Ticket ticket
type Ticket struct {
	AdjustClassifyLevel      string                      `json:"adjustClassifyLevel"`
	AdjustClassifyLevelName  string                      `json:"adjustClassifyLevelName"`
	AlarmId                  string                      `json:"alarmId"`
	AlarmOccurTime           string                      `json:"alarmOccurTime"`
	BusinessDeptName         string                      `json:"businessDeptName"`
	BusinessPath             string                      `json:"businessPath"`
	CancelId                 string                      `json:"cancelId"`
	CancelReason             string                      `json:"cancelReason"`
	ClassifyLevel            string                      `json:"classifyLevel"`
	ClassifyLevelName        string                      `json:"classifyLevelName"`
	CreateTime               string                      `json:"createTime"`
	CurrentTask              string                      `json:"currentTask"`
	EndTime                  string                      `json:"endTime"`
	EventDescription         string                      `json:"eventDescription"`
	EventPartType            string                      `json:"eventPartType"`
	FaultCodeFromTitan       string                      `json:"faultCodeFromTitan"`
	FaultTypeId              string                      `json:"faultTypeId"`
	FaultTypeName            string                      `json:"faultTypeName"`
	ID                       int                         `json:"id"`         // 故障单id
	InstanceID               int                         `json:"instanceId"` // 故障单号
	IsEnd                    enumor.XrayFaultTicketIsEnd `json:"isEnd"`      // 是否结单(0:未结单 1:已结单)
	IsSystemDisk             bool                        `json:"isSystemDisk"`
	Origin                   int                         `json:"origin"`
	ProcessDescription       string                      `json:"processDescription"`
	ProcessDetailDescription string                      `json:"processDetailDescription"`
	ServerAssetId            string                      `json:"serverAssetId"`
	ServerIp                 string                      `json:"serverIp"`
	Source                   string                      `json:"source"`
	Starter                  string                      `json:"starter"`
	UnitParameter            string                      `json:"unitParameter"`
	UpdateTime               string                      `json:"updateTime"`
	IsRepairFailed           string                      `json:"isRepairFailed"`
}
