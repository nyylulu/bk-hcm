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

package resplan

import (
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
)

// RPTicketListResult list resource plan ticket result.
type RPTicketListResult struct {
	Count   uint64                   `json:"count"`
	Details []rpt.ResPlanTicketTable `json:"details"`
}

// RPTicketWithStatusListRst list resource plan ticket with status result.
type RPTicketWithStatusListRst struct {
	Count   uint64               `json:"count"`
	Details []RPTicketWithStatus `json:"details"`
}

// RPTicketWithStatusAndResListRst list resource plan ticket with status and resource result.
type RPTicketWithStatusAndResListRst struct {
	Count   uint64                     `json:"count"`
	Details []RPTicketWithStatusAndRes `json:"details"`
}

// RPTicketWithStatus resource plan ticket with status.
type RPTicketWithStatus struct {
	rpt.ResPlanTicketTable
	Status     enumor.RPTicketStatus `json:"status"`
	StatusName string                `json:"status_name"`
}

// RPTicketWithStatusAndRes resource plan ticket with status and resource.
type RPTicketWithStatusAndRes struct {
	RPTicketWithStatus
	TicketTypeName string               `json:"ticket_type_name"`
	OriginalInfo   RPTicketResourceInfo `json:"original_info"`
	UpdatedInfo    RPTicketResourceInfo `json:"updated_info"`
}

// RPTicketResourceInfo resource plan ticket resource info.
type RPTicketResourceInfo struct {
	Cvm cvmInfo `json:"cvm"`
	Cbs cbsInfo `json:"cbs"`
}

// NewResourceInfo new resource info.
func NewResourceInfo(cpuCore, memory, diskSize int64) RPTicketResourceInfo {
	return RPTicketResourceInfo{
		Cvm: cvmInfo{
			CpuCore: &cpuCore,
			Memory:  &memory,
		},
		Cbs: cbsInfo{
			DiskSize: &diskSize,
		},
	}
}

// NewNullResourceInfo new null resource info.
func NewNullResourceInfo() RPTicketResourceInfo {
	return RPTicketResourceInfo{
		Cvm: cvmInfo{
			CpuCore: nil,
			Memory:  nil,
		},
		Cbs: cbsInfo{
			DiskSize: nil,
		},
	}
}

type cvmInfo struct {
	CpuCore *int64 `json:"cpu_core"`
	Memory  *int64 `json:"memory"`
}

type cbsInfo struct {
	DiskSize *int64 `json:"disk_size"`
}
