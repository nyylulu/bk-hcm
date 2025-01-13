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

package table

var (
	// ReqForGetHost 获取裁撤主机数据
	ReqForGetHost = &reqSource{
		source: "host",
		ccHostFields: []string{"bk_host_id", "bk_asset_id", "module_name", "operator", "bk_bak_operator",
			"bk_host_innerip", "bk_host_outerip", "svr_device_class", "idc_unit_name", "bk_os_version",
			"svr_input_time", "raid_name", "logic_domain", "bk_disk", "bk_cpu"},
		esHostFields: []string{"server_asset_id", "ip", "outer_ip", "app_name", "bk_biz_id", "module", "device_type",
			"module_name", "idc_unit_name", "sfw_name_version", "go_up_date", "raid_name", "logic_area",
			"server_bak_operator", "server_operator", "device_layer", "cpu_score", "mem_score",
			"inner_net_traffic_score", "disk_io_score", "disk_util_score", "is_pass", "mem4linux", "inner_net_traffic",
			"outer_net_traffic", "disk_io", "disk_util", "disk_total", "max_cpu_core_amount", "group_name", "center",
		},
	}

	// ReqForGetDissolveTable 获取裁撤表格数据
	ReqForGetDissolveTable = &reqSource{
		source:       "table",
		ccHostFields: []string{"bk_host_id", "bk_asset_id", "module_name", "bk_cpu"},
		esHostFields: []string{"server_asset_id", "bk_biz_id", "app_name", "max_cpu_core_amount"},
	}
)

// ReqSourceI ...
type ReqSourceI interface {
	GetReqSource() string
	GetCCHostFields() []string
	GetEsHostFields() []string
}

type reqSource struct {
	source       string
	ccHostFields []string
	esHostFields []string
}

func (d *reqSource) GetReqSource() string {
	return d.source
}

func (d *reqSource) GetCCHostFields() []string {
	return d.ccHostFields
}

func (d *reqSource) GetEsHostFields() []string {
	return d.esHostFields
}
