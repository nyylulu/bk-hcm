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

// Package cmdb CC API response
package cmdb

import "strings"

// RespMeta cc response meta info
type RespMeta struct {
	Result bool   `json:"result" mapstructure:"result"`
	Code   int    `json:"code" mapstructure:"code"`
	ErrMsg string `json:"message" mapstructure:"message"`
}

// AddHostResp add host to cc response
type AddHostResp struct {
	RespMeta `json:",inline"`
}

// TransferHostResp transfer host to another business response
type TransferHostResp struct {
	RespMeta `json:",inline"`
}

// ListHostResp list host response
type ListHostResp struct {
	RespMeta `json:",inline"`
	Data     ListHostResult `json:"data"`
}

// ListHostResult list host result
type ListHostResult struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// UpdateHostsResp update hosts response
type UpdateHostsResp struct {
	RespMeta `json:",inline"`
}

// HostModuleResp host module relation response
type HostModuleResp struct {
	RespMeta `json:",inline"`
	Data     []ModuleHost `json:"data"`
}

// ModuleHost host module relation result
type ModuleHost struct {
	AppID    int64  `json:"bk_biz_id,omitempty" bson:"bk_biz_id"`
	HostID   int64  `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	ModuleID int64  `json:"bk_module_id,omitempty" bson:"bk_module_id"`
	SetID    int64  `json:"bk_set_id,omitempty" bson:"bk_set_id"`
	OwnerID  string `json:"bk_supplier_account,omitempty" bson:"bk_supplier_account"`
}

// DeviceTopoInfo topo info
type DeviceTopoInfo struct {
	InnerIP      string `json:"innerIP"`
	AssetID      string `json:"assetID"`
	DeviceClass  string `json:"deviceClass"`
	Raid         string `json:"raid"`
	OSName       string `json:"osName"`
	OSVersion    string `json:"osVersion"`
	IdcArea      string `json:"idcArea"`
	SZone        string `json:"sZone"`
	ModuleName   string `json:"moduleName"`
	Equipment    string `json:"equipment"`
	IdcLogicArea string `json:"idcLogicArea"`
}

// ListBizHostResp list host response
type ListBizHostResp struct {
	RespMeta `json:",inline"`
	Data     ListBizHostResult `json:"data"`
}

// ListBizHostResult list host result
type ListBizHostResult struct {
	Count int         `json:"count"`
	Info  []*HostInfo `json:"info"`
}

// OsType 操作系统类型
type OsType string

const (
	// LinuxOsType 操作系统类型-Linux
	LinuxOsType OsType = "1"
	// WindowsOsType 操作系统类型-Windows
	WindowsOsType OsType = "2"
)

// HostInfo host info
type HostInfo struct {
	BkHostId        int64  `json:"bk_host_id"`
	BkCloudId       int64  `json:"bk_cloud_id"`
	BkAssetId       string `json:"bk_asset_id"`
	BkHostInnerIP   string `json:"bk_host_innerip"`
	BkHostOuterIP   string `json:"bk_host_outerip"`
	BkHostOuterIPv6 string `json:"bk_host_outerip_v6"`
	// 外网运营商
	BkIpOerName string `json:"bk_ip_oper_name"`
	// 机型
	SvrDeviceClass string `json:"svr_device_class"`
	// 操作系统名称
	BkOsName string `json:"bk_os_name"`
	// 操作系统类型
	BkOsType OsType `json:"bk_os_type"`
	// 操作系统版本
	BkOsVersion string `json:"bk_os_version"`
	// IDC区域
	BkIdcArea string `json:"bk_idc_area"`
	// 地域
	BkZoneName string `json:"bk_zone_name"`
	// 可用区(子Zone)
	SubZone string `json:"sub_zone"`
	// 子ZoneID
	SubZoneId  string `json:"sub_zone_id"`
	ModuleName string `json:"module_name"`
	// 机架号
	RackId      string `json:"rack_id"`
	IdcUnitName string `json:"idc_unit_name"`
	// 逻辑区域
	LogicDomain string `json:"logic_domain"`
	RaidName    string `json:"raid_name"`
	// 机器上架时间，格式如"2018-05-07T00:00:00+08:00"
	SvrInputTime string `json:"svr_input_time"`
	// 主要维护人
	Operator string `json:"operator"`
	// 备份维护人
	BakOperator string `json:"bk_bak_operator"`
	// 状态
	SvrStatus string `json:"srv_status"`
	// 磁盘容量
	BkDisk float64 `json:"bk_disk"`
	// CPU逻辑核心数
	BkCpu int64 `json:"bk_cpu"`
	// 服务器来源类型ID(未知(0, 默认值) 自有(1) 托管(2) 租用(3) 虚拟机(4) 容器(5))
	BkSvrSourceTypeID string `json:"bk_svr_source_type_id"`
}

// GetUniqIp get CC host unique inner ip
func (h *HostInfo) GetUniqIp() string {
	// when CC host has multiple inner ips, bk_host_innerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(h.BkHostInnerIP, ",")
	if len(multiIps) == 0 {
		return ""
	}

	return multiIps[0]
}

// GetUniqOuterIp get CC host unique outer ip
func (h *HostInfo) GetUniqOuterIp() string {
	// when CC host has multiple outer ips, bk_host_outerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(h.BkHostOuterIP, ",")
	if len(multiIps) == 0 {
		return ""
	}

	return multiIps[0]
}

// IsPmAndOuterIPDevice 检查是否物理机，是否有外网IP
func (h *HostInfo) IsPmAndOuterIPDevice() bool {
	// 服务器来源类型ID(未知(0, 默认值) 自有(1) 托管(2) 租用(3) 虚拟机(4) 容器(5))
	if h.BkSvrSourceTypeID != BkSvrSourceTypeIDSelf && h.BkSvrSourceTypeID != BkSvrSourceTypeIDDeposit &&
		h.BkSvrSourceTypeID != BkSvrSourceTypeIDLease {
		return false
	}

	if h.BkHostOuterIP == "" && h.BkHostOuterIPv6 == "" {
		return false
	}

	return true
}

// HostBizRelResp find host business relation response
type HostBizRelResp struct {
	RespMeta `json:",inline"`
	Data     []*HostBizRel `json:"data"`
}

// HostBizRel host business relation
type HostBizRel struct {
	BkHostId     int64  `json:"bk_host_id"`
	BkBizId      int64  `json:"bk_biz_id"`
	BkSetId      int64  `json:"bk_set_id"`
	BkModuleId   int64  `json:"bk_module_id"`
	BkSupplierId string `json:"bk_supplier_account"`
}

// SearchBizResp search business response
type SearchBizResp struct {
	RespMeta `json:",inline"`
	Data     *SearchBizRst `json:"data"`
}

// SearchBizRst search business result
type SearchBizRst struct {
	Count int        `json:"count"`
	Info  []*BizInfo `json:"info"`
}

// BizInfo business info
type BizInfo struct {
	BkBizId         int64  `json:"bk_biz_id"`
	BkBizName       string `json:"bk_biz_name"`
	BkOperGrpNameID int64  `json:"bk_oper_grp_name_id"`
}

// SearchModuleResp search module response
type SearchModuleResp struct {
	RespMeta `json:",inline"`
	Data     *SearchModuleRst `json:"data"`
}

// SearchModuleRst search module result
type SearchModuleRst struct {
	Count int           `json:"count"`
	Info  []*ModuleInfo `json:"info"`
}

// ModuleInfo module info
type ModuleInfo struct {
	BkModuleId   int64  `json:"bk_module_id"`
	BkModuleName string `json:"bk_module_name"`
	Default      int    `json:"default"`
}

// BizInternalModuleResp search business's internal module response
type BizInternalModuleResp struct {
	RespMeta `json:",inline"`
	Data     *BizInternalModuleRespRst `json:"data"`
}

// BizInternalModuleRespRst search business's internal module result
type BizInternalModuleRespRst struct {
	BkSetID   int64         `json:"bk_set_id"`
	BkSetName string        `json:"bk_set_name"`
	Module    []*ModuleInfo `json:"module"`
}

// CrTransitResp transfer host to CR transit module response
type CrTransitResp struct {
	RespMeta `json:",inline"`
	Data     *CrTransitRst `json:"data"`
}

// CrTransitRst transfer host to CR transit module result
type CrTransitRst struct {
	AssetIds []string `json:"asset_ids"`
}

// SearchBizBelongingRst is search cmdb business belonging result.
type SearchBizBelongingRst struct {
	Data []SearchBizBelonging `json:"data"`
}

// SearchBizBelonging is search cmdb business belonging element of result.
type SearchBizBelonging struct {
	BizID            int64  `json:"bk_biz_id"`
	BizName          string `json:"bk_biz_name"`
	OpProductID      int64  `json:"bsi_product_id"`
	OpProductName    string `json:"bsi_product_name"`
	PlanProductID    int64  `json:"plan_product_id"`
	PlanProductName  string `json:"plan_product_name"`
	BusinessDeptID   int64  `json:"business_dept_id"`
	BusinessDeptName string `json:"business_dept_name"`
	Bs1Name          string `json:"bs1_name"`
	Bs1NameID        int64  `json:"bs1_name_id"`
	Bs2Name          string `json:"bs2_name"`
	Bs2NameID        int64  `json:"bs2_name_id"`
	VirtualDeptID    int64  `json:"virtual_dept_id"`
	VirtualDeptName  string `json:"virtual_dept_name"`
}
