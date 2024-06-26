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

package es

const (
	// Organization field of the request
	Organization = "organization"
	// Operator field of the request
	Operator = "operator"
	// NotInCCIPs 在es中，但是不在cc中的ip数组的条件
	NotInCCIPs = "not_in_cc_ips"

	// AppName field of the host
	AppName = "app_name"
	// ModuleName field of the host
	ModuleName = "module_name"
	// AssetID field of the host
	AssetID           = "server_asset_id"
	serverOperator    = "server_operator"
	serverBakOperator = "server_bak_operator"
	department        = "department"
	center            = "center"
	groupName         = "group_name"
	innerIP           = "ip"

	indexPrefix = "app_device_pass_dtl_"
)

// Host detail info
type Host struct {
	ServerAssetID        string  `json:"server_asset_id"`
	InnerIP              string  `json:"ip"`
	OuterIP              string  `json:"outer_ip"`
	AppName              string  `json:"app_name"`
	Module               string  `json:"module"`
	DeviceType           string  `json:"device_type"`
	ModuleName           string  `json:"module_name"`
	IdcUnitName          string  `json:"idc_unit_name"`
	SfwNameVersion       string  `json:"sfw_name_version"`
	GoUpDate             string  `json:"go_up_date"`
	RaidName             string  `json:"raid_name"`
	LogicArea            string  `json:"logic_area"`
	ServerBakOperator    string  `json:"server_bak_operator"`
	ServerOperator       string  `json:"server_operator"`
	DeviceLayer          string  `json:"device_layer"`
	CPUScore             float64 `json:"cpu_score"`
	MemScore             float64 `json:"mem_score"`
	InnerNetTrafficScore float64 `json:"inner_net_traffic_score"`
	DiskIoScore          float64 `json:"disk_io_score"`
	DiskUtilScore        float64 `json:"disk_util_score"`
	IsPass               string  `json:"is_pass"`
	Mem4linux            float64 `json:"mem4linux"`
	InnerNetTraffic      float64 `json:"inner_net_traffic"`
	OuterNetTraffic      float64 `json:"outer_net_traffic"`
	DiskIo               float64 `json:"disk_io"`
	DiskUtil             float64 `json:"disk_util"`
	DiskTotal            float64 `json:"disk_total"`
	MaxCPUCoreAmount     int64   `json:"max_cpu_core_amount"`
	GroupName            string  `json:"group_name"`
	Center               string  `json:"center"`
}
