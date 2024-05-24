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

// AuthVerifyReq auth policy verify request
type AuthVerifyReq struct {
	System    string      `json:"system"`
	Subject   *Subject    `json:"subject"`
	Action    *Action     `json:"action"`
	Resources []*Resource `json:"resources"`
}

// Subject auth policy verify subject parameter
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Action auth policy verify action parameter
type Action struct {
	ID string `json:"id"`
}

// Resource auth policy verify resource parameter
type Resource struct {
	System string `json:"system"`
	Type   string `json:"type"`
	ID     string `json:"id"`
}

// GetAuthUrlReq get auth url request
type GetAuthUrlReq struct {
	SystemId string        `json:"system_id"`
	Actions  []*AuthAction `json:"actions"`
}

// AuthAction get auth url action
type AuthAction struct {
	ID        string          `json:"id"`
	Resources []*AuthResource `json:"related_resource_types"`
}

// AuthResource get auth url resource
type AuthResource struct {
	SystemId  string        `json:"system_id"`
	Type      string        `json:"type"`
	Instances [][]*Instance `json:"instances"`
}

// Instance get auth url resource instance
type Instance struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
