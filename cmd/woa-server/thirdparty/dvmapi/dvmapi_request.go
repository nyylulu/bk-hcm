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

package dvmapi

// OrderCreateReq docker vm apply request
type OrderCreateReq struct {
	Cores         int      `json:"cores"`
	Memory        int      `json:"memory"`
	Disk          int      `json:"disk"`
	Image         string   `json:"image"`
	DisplayName   string   `json:"displayName"`
	AppModuleName string   `json:"appModule"`
	SetId         string   `json:"setId"`
	Module        string   `json:"module"`
	WhiteList     []string `json:"whiteList"`
	HostType      string   `json:"hostType"`
	HostIp        []string `json:"hostIp"`
	Idle          string   `json:"idle"`
	Affinity      int      `json:"affinity,string"`
	MountPath     string   `json:"mountPath"`
	Replicas      uint     `json:"replicas"`
	Operator      string   `json:"operator"`
	Reason        string   `json:"reason"`
	DeliverModule string   `json:"deliverModule"`
	HostRole      string   `json:"hostRole"`
}

// ListHostReq list host in cluster request
type ListHostReq struct {
	SetId       string `json:"set_id"`
	DeviceClass string `json:"device_class"`
	Cores       int    `json:"cores"`
	Memory      int    `json:"memory"`
	Disk        int    `json:"disk"`
	HostRole    string `json:"host_role"`
}
