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

// Package task ...
package task

import "hcm/pkg/criteria/enumor"

// DVMSelector docker vm selector parameter
type DVMSelector struct {
	Cores             int             `json:"cpu"`
	Memory            int             `json:"mem"`
	Disk              int             `json:"disk"`
	DeviceClass       string          `json:"deviceClass"`
	Image             string          `json:"image"`
	Kernel            string          `json:"kernel"`
	DockerType        string          `json:"dockerType"`
	NetworkType       string          `json:"networkType"`
	DataDiskMountPath string          `json:"dataDiskMountPath"`
	DataDiskType      enumor.DiskType `json:"dataDiskType"`
	DataDiskRaid      string          `json:"dataDiskRaid"`
	Region            string          `json:"region"`
	Zone              string          `json:"zone"`
	ExtranetIsp       string          `json:"extranetIsp"`
	HostRole          string          `json:"hostRole"`
	CpuProvider       string          `json:"cpuProvider"`
	AmdDevicePattern  []string        `json:"amdDevicePattern"`
}

// HostPriority docker host priority for schedule
type HostPriority struct {
	IP               string
	DeviceClass      string
	Equipment        string
	ModuleName       string
	SZone            string
	SetId            string
	AllocatableCount int
	ScheduledVMs     int
	Score            float64
}

const (
	// MaxPriority max priority
	MaxPriority = 10
)

// HostPriorityList host priority list
type HostPriorityList []HostPriority

// Len host priority list length
func (h HostPriorityList) Len() int {
	return len(h)
}

// Less compare two host priority
func (h HostPriorityList) Less(i, j int) bool {
	if h[i].Score == h[j].Score {
		return h[i].IP < h[j].IP
	}
	return h[i].Score < h[j].Score
}

// Swap swap two host priority
func (h HostPriorityList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Anti Const
const (
	AntiNone   string = "ANTI_NONE"   //无要求
	AntiRack   string = "ANTI_RACK"   //分机架
	AntiModule string = "ANTI_MODULE" //分Module
	AntiCampus string = "ANTI_CAMPUS" //分Campus
)
