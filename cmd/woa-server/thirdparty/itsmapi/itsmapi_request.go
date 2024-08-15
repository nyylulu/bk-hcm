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

package itsmapi

// CreateTicketReq create itsm ticket request
type CreateTicketReq struct {
	ServiceId int            `json:"service_id"`
	Creator   string         `json:"creator"`
	Fields    []*TicketField `json:"fields"`
}

// TicketField itsm ticket field
type TicketField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// OperateNodeReq operate itsm ticket node request
type OperateNodeReq struct {
	Sn         string         `json:"sn"`
	StateId    int64          `json:"state_id"`
	Operator   string         `json:"operator"`
	ActionType string         `json:"action_type"`
	Fields     []*TicketField `json:"fields"`
}
