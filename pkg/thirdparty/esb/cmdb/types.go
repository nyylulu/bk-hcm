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

package cmdb

import (
	"encoding/json"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/thirdparty/esb/types"
	"hcm/pkg/tools/querybuilder"
)

// ----------------------------- biz -----------------------------

// SearchBizParams is esb search cmdb business parameter.
type esbSearchBizParams struct {
	*types.CommParams
	*SearchBizParams
}

// SearchBizParams is cmdb search business parameter.
type SearchBizParams struct {
	Fields            []string     `json:"fields"`
	Page              BasePage     `json:"page"`
	BizPropertyFilter *QueryFilter `json:"biz_property_filter,omitempty"`
}

// BizIDField cmdb 业务字段
const BizIDField = "bk_biz_id"

// QueryFilter is cmdb common query filter.
type QueryFilter struct {
	Rule `json:",inline"`
}

// Rule is cmdb common query rule type.
type Rule interface {
	GetDeep() int
}

// CombinedRule is cmdb query rule that is combined by multiple AtomRule.
type CombinedRule struct {
	Condition Condition `json:"condition"`
	Rules     []Rule    `json:"rules"`
}

// Condition cmdb condition
type Condition string

const (
	// ConditionAnd and
	ConditionAnd = Condition("AND")
)

// GetDeep get query rule depth.
func (r CombinedRule) GetDeep() int {
	maxChildDeep := 1
	for _, child := range r.Rules {
		childDeep := child.GetDeep()
		if childDeep > maxChildDeep {
			maxChildDeep = childDeep
		}
	}
	return maxChildDeep + 1
}

// Combined ...
func Combined(cond Condition, rules ...Rule) CombinedRule {
	return CombinedRule{
		Condition: cond,
		Rules:     rules,
	}
}

// Equal ...
func Equal(field string, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: OperatorEqual,
		Value:    value,
	}
}

// In ...
func In(field string, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: OperatorIn,
		Value:    value,
	}
}

// Atomic ...
func Atomic(field string, op Operator, value any) AtomRule {
	return AtomRule{
		Field:    field,
		Operator: op,
		Value:    value,
	}
}

// AtomRule is cmdb atomic query rule.
type AtomRule struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// Operator cmdb operator
type Operator string

var (
	// OperatorEqual ...
	OperatorEqual = Operator("equal")
	// OperatorIn ...
	OperatorIn = Operator("in")
)

// GetDeep get query rule depth.
func (r AtomRule) GetDeep() int {
	return 1
}

// MarshalJSON marshal QueryFilter to json.
func (qf *QueryFilter) MarshalJSON() ([]byte, error) {
	if qf.Rule != nil {
		return json.Marshal(qf.Rule)
	}
	return make([]byte, 0), nil
}

// BasePage is cmdb paging parameter.
type BasePage struct {
	Sort        string `json:"sort,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Start       int64  `json:"start"`
	EnableCount bool   `json:"enable_count,omitempty"`
}

// SearchBizResult is cmdb search business response.
type SearchBizResult struct {
	Count int64 `json:"count"`
	Info  []Biz `json:"info"`
}

// Biz is cmdb biz info.
type Biz struct {
	BizID   int64  `json:"bk_biz_id"`
	BizName string `json:"bk_biz_name"`
	// 二级业务id
	BsName2ID int64 `json:"bs2_name_id"`
}

// -------------------------- cloud area --------------------------

// SearchCloudAreaParams is esb search cmdb cloud area parameter.
type esbSearchCloudAreaParams struct {
	*types.CommParams
	*SearchCloudAreaParams
}

// SearchCloudAreaParams is cmdb search cloud area parameter.
type SearchCloudAreaParams struct {
	Fields    []string               `json:"fields"`
	Page      BasePage               `json:"page"`
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// SearchCloudAreaResp is cmdb search cloud area response.
type SearchCloudAreaResp struct {
	types.BaseResponse `json:",inline"`
	Data               *SearchCloudAreaResult `json:"data"`
}

// SearchCloudAreaResult is cmdb search cloud area result.
type SearchCloudAreaResult struct {
	Count int64       `json:"count"`
	Info  []CloudArea `json:"info"`
}

// CloudArea is cmdb cloud area info.
type CloudArea struct {
	CloudID   int64  `json:"bk_cloud_id"`
	CloudName string `json:"bk_cloud_name"`
}

// ---------------------------- create ----------------------------

// BatchCreateResp cmdb's basic batch create resource response.
type BatchCreateResp struct {
	types.BaseResponse `json:",inline"`
	Data               *BatchCreateResult `json:"data"`
}

// BatchCreateResult cmdb's basic batch create resource result.
type BatchCreateResult struct {
	IDs []int64 `json:"ids"`
}

// ----------------------------- host -----------------------------

// esbAddCloudHostToBizParams is esb add cmdb cloud host to biz parameter.
type esbAddCloudHostToBizParams struct {
	*types.CommParams
	*AddCloudHostToBizParams
}

// AddCloudHostToBizParams is esb add cloud host to biz parameter.
type AddCloudHostToBizParams struct {
	BizID    int64  `json:"bk_biz_id"`
	HostInfo []Host `json:"host_info"`
}

// esbDeleteCloudHostFromBizParams is esb delete cmdb cloud host from biz parameter.
type esbDeleteCloudHostFromBizParams struct {
	*types.CommParams
	*DeleteCloudHostFromBizParams
}

// DeleteCloudHostFromBizParams is esb delete cloud host from biz parameter.
type DeleteCloudHostFromBizParams struct {
	BizID   int64   `json:"bk_biz_id"`
	HostIDs []int64 `json:"bk_host_ids"`
}

// esbListBizHostParams is esb list cmdb host in biz parameter.
type esbListBizHostParams struct {
	*types.CommParams
	*ListBizHostParams
}

// ListBizHostParams is esb list cmdb host in biz parameter.
type ListBizHostParams struct {
	BizID              int64           `json:"bk_biz_id"`
	BkSetIDs           []int64         `json:"bk_set_ids"`
	BkModuleIDs        []int64         `json:"bk_module_ids"`
	ModuleCond         []ConditionItem `json:"module_cond"`
	Fields             []string        `json:"fields"`
	Page               BasePage        `json:"page"`
	HostPropertyFilter *QueryFilter    `json:"host_property_filter,omitempty"`
}

// ListHostReq list host request
type ListHostReq struct {
	HostPropertyFilter *QueryFilter `json:"host_property_filter"`
	Fields             []string     `json:"fields"`
	Page               BasePage     `json:"page"`
	EnableCount        bool         `json:"enable_count,omitempty" mapstructure:"enable_count,omitempty"`
}

// ListHostResp list host response
type ListHostResp struct {
	RespMeta `json:",inline"`
	Data     ListHostResult `json:"data"`
}

// RespMeta cc response meta info
type RespMeta struct {
	Result bool   `json:"result" mapstructure:"result"`
	Code   int    `json:"code" mapstructure:"code"`
	ErrMsg string `json:"message" mapstructure:"message"`
}

// ListHostResult list host result
type ListHostResult struct {
	Count int64   `json:"count"`
	Info  []*Host `json:"info"`
}

// ListBizHostResp is cmdb list cmdb host in biz response.
type ListBizHostResp struct {
	types.BaseResponse
	*ListBizHostResult `json:"data"`
}

// ListBizHostResult is cmdb list cmdb host in biz result.
type ListBizHostResult struct {
	Count int64  `json:"count"`
	Info  []Host `json:"info"`
}

// Host defines cmdb host info.
type Host struct {
	BkHostID          int64           `json:"bk_host_id"`
	BkCloudVendor     CloudVendor     `json:"bk_cloud_vendor"`
	BkCloudInstID     string          `json:"bk_cloud_inst_id"`
	BkCloudHostStatus CloudHostStatus `json:"bk_cloud_host_status"`
	BkCloudID         int64           `json:"bk_cloud_id"`
	// 云上地域，如 "ap-guangzhou"
	BkCloudRegion      string  `json:"bk_cloud_region"`
	BkHostInnerIP      string  `json:"bk_host_innerip"`
	BkHostOuterIP      string  `json:"bk_host_outerip"`
	BkHostInnerIPv6    string  `json:"bk_host_innerip_v6"`
	BkHostOuterIPv6    string  `json:"bk_host_outerip_v6"`
	Operator           string  `json:"operator"`
	BkBakOperator      string  `json:"bk_bak_operator"`
	BkHostName         string  `json:"bk_host_name"`
	BkComment          *string `json:"bk_comment,omitempty"`
	BkOSName           string  `json:"bk_os_name"`
	SvrSourceTypeID    string  `json:"bk_svr_source_type_id"`
	BkAssetID          string  `json:"bk_asset_id"`
	SvrDeviceClassName string  `json:"bk_svr_device_cls_name"`

	// 以下字段仅内部版支持，由cc从云梯获取
	BkCloudZone     string `json:"bk_cloud_zone"`
	BkCloudVpcID    string `json:"bk_cloud_vpc_id"`
	BkCloudSubnetID string `json:"bk_cloud_subnet_id"`
	// 外网运营商
	BkIpOerName string `json:"bk_ip_oper_name"`
	// 机型
	SvrDeviceClass string `json:"svr_device_class"`
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
	// 状态
	SvrStatus string `json:"srv_status"`
	// 磁盘容量
	BkDisk float64 `json:"bk_disk"`
	// CPU逻辑核心数
	BkCpu int64 `json:"bk_cpu"`
	// 实例计费模式
	InstanceChargeType string `json:"instance_charge_type"`
	// 套餐计费起始时间
	BillingStartTime time.Time `json:"billing_start_time"`
	// 套餐计费过期时间
	BillingExpireTime time.Time `json:"billing_expire_time"`
}

// OsType 操作系统类型
type OsType string

// SvrSourceTypeID 服务器来源类型
type SvrSourceTypeID string

const (
	// Own 自有, 物理机
	Own SvrSourceTypeID = "1"
	// Hosting 托管, 物理机
	Hosting SvrSourceTypeID = "2"
	// Rent 租用, 物理机
	Rent = "3"
	// CVM 虚拟机
	CVM = "4"
	// Container 容器
	Container = "5"
)

// HostFields cmdb common fields
var HostFields = []string{
	"bk_cloud_inst_id",
	"bk_host_id",
	"bk_asset_id",
	// 云地域
	"bk_cloud_region",
	// 云厂商
	"bk_cloud_vendor",
	"bk_host_innerip",
	"bk_host_outerip",
	"bk_host_innerip_v6",
	"bk_host_outerip_v6",
	"bk_cloud_host_status",
	"bk_host_name",
	"bk_cloud_id",
	"bk_os_name",
	"bk_svr_source_type_id",
	"bk_svr_device_cls_name",
	"svr_source_type_id", // 服务器来源类型ID
	"srv_status",         // CC的运营状态
	"svr_device_class",   // 机型
	"bk_disk",            // 磁盘容量(GB)
	"bk_cpu",             // CPU逻辑核心数
	"operator",           // 主要维护人
	"bk_bak_operator",    // 备份维护人

	// 以下字段仅内部版支持，由cc从云梯获取
	"bk_cloud_vpc_id",
	"bk_cloud_subnet_id",
	"bk_cloud_zone",
}

type esbFindHostTopoRelationParams struct {
	*types.CommParams
	*FindHostTopoRelationParams
}

// FindHostTopoRelationParams cmdb find host topo request params
type FindHostTopoRelationParams struct {
	BizID       int64    `json:"bk_biz_id"`
	BkSetIDs    []int64  `json:"bk_set_ids,omitempty"`
	BkModuleIDs []int64  `json:"bk_module_ids,omitempty"`
	HostIDs     []int64  `json:"bk_host_ids"`
	Page        BasePage `json:"page"`
}

type findHostTopoRelationResp struct {
	types.BaseResponse `json:",inline"`
	Data               *HostTopoRelationResult `json:"data"`
}

// HostTopoRelationResult cmdb host topo relation result warp
type HostTopoRelationResult struct {
	Count int64              `json:"count"`
	Page  BasePage           `json:"page"`
	Data  []HostTopoRelation `json:"data"`
}

// HostTopoRelation cmdb host topo relation
type HostTopoRelation struct {
	BizID             int64  `json:"bk_biz_id"`
	BkSetID           int64  `json:"bk_set_id"`
	BkModuleID        int64  `json:"bk_module_id"`
	HostID            int64  `json:"bk_host_id"`
	BkSupplierAccount string `json:"bk_supplier_account"`
}

type esbSearchModuleParams struct {
	*types.CommParams
	*SearchModuleParams
}

// SearchModuleParams cmdb module search parameter.
type SearchModuleParams struct {
	BizID             int64  `json:"bk_biz_id"`
	BkSetID           int64  `json:"bk_set_id,omitempty"`
	BkSupplierAccount string `json:"bk_supplier_account,omitempty"`

	Fields    []string               `json:"fields"`
	Page      BasePage               `json:"page"`
	Condition map[string]interface{} `json:"condition"`
}

type searchModuleResp struct {
	types.BaseResponse `json:",inline"`
	Permission         interface{}       `json:"permission"`
	Data               *ModuleInfoResult `json:"data"`
}

// ModuleInfoResult cmdb module info list result
type ModuleInfoResult struct {
	Count int64         `json:"count"`
	Info  []*ModuleInfo `json:"info"`
}

// ModuleInfo cmdb module info
type ModuleInfo struct {
	BkSetID      int64  `json:"bk_set_id"`
	BkModuleName string `json:"bk_module_name"`
	BkModuleId   int64  `json:"bk_module_id"`
	Default      int64  `json:"default"`
}

// CloudVendor defines cmdb cloud vendor type.
type CloudVendor string

const (
	// AwsCloudVendor cmdb aws vendor
	AwsCloudVendor CloudVendor = "1"
	// TCloudCloudVendor cmdb cloud vendor
	TCloudCloudVendor CloudVendor = "2"
	// GcpCloudVendor cmdb gcp vendor
	GcpCloudVendor CloudVendor = "3"
	// AzureCloudVendor cmdb azure vendor
	AzureCloudVendor CloudVendor = "4"
	// HuaWeiCloudVendor cmdb huawei vendor
	HuaWeiCloudVendor CloudVendor = "15"
	// TCloudZiyanCloudVendor 腾讯自研云厂商
	TCloudZiyanCloudVendor CloudVendor = "17"
)

// HcmCmdbVendorMap is hcm vendor to cmdb cloud vendor map.
var HcmCmdbVendorMap = map[enumor.Vendor]CloudVendor{
	enumor.Aws:         AwsCloudVendor,
	enumor.TCloud:      TCloudCloudVendor,
	enumor.Gcp:         GcpCloudVendor,
	enumor.Azure:       AzureCloudVendor,
	enumor.HuaWei:      HuaWeiCloudVendor,
	enumor.TCloudZiyan: TCloudZiyanCloudVendor,
}

// CmdbHcmVendorMap cmdb vendor to hcm vendor
var CmdbHcmVendorMap = map[CloudVendor]enumor.Vendor{
	AwsCloudVendor:         enumor.Aws,
	TCloudCloudVendor:      enumor.TCloud,
	GcpCloudVendor:         enumor.Gcp,
	AzureCloudVendor:       enumor.Azure,
	HuaWeiCloudVendor:      enumor.HuaWei,
	TCloudZiyanCloudVendor: enumor.TCloudZiyan,
}

// CloudHostStatus defines cmdb cloud host status type.
type CloudHostStatus string

const (
	// UnknownCloudHostStatus ...
	UnknownCloudHostStatus CloudHostStatus = "1"
	// StartingCloudHostStatus ...
	StartingCloudHostStatus CloudHostStatus = "2"
	// RunningCloudHostStatus ...
	RunningCloudHostStatus CloudHostStatus = "3"
	// StoppingCloudHostStatus ...
	StoppingCloudHostStatus CloudHostStatus = "4"
	// StoppedCloudHostStatus ...
	StoppedCloudHostStatus CloudHostStatus = "5"
	// TerminatedCloudHostStatus ...
	TerminatedCloudHostStatus CloudHostStatus = "6"
)

// HcmCmdbHostStatusMap is hcm vendor to cmdb cloud host status map.
var HcmCmdbHostStatusMap = map[enumor.Vendor]map[string]CloudHostStatus{
	enumor.TCloud: TCloudCmdbStatusMap,
	enumor.Aws:    AwsCmdbStatusMap,
	enumor.Gcp:    GcpCmdbStatusMap,
	enumor.Azure:  AzureCmdbStatusMap,
	enumor.HuaWei: HuaWeiCmdbStatusMap,
}

// TCloudCmdbStatusMap is tcloud status to cmdb cloud host status map.
var TCloudCmdbStatusMap = map[string]CloudHostStatus{
	"PENDING":       UnknownCloudHostStatus,
	"LAUNCH_FAILED": UnknownCloudHostStatus,
	"RUNNING":       RunningCloudHostStatus,
	"STOPPED":       StoppedCloudHostStatus,
	"STARTING":      StartingCloudHostStatus,
	"STOPPING":      StoppingCloudHostStatus,
	"REBOOTING":     UnknownCloudHostStatus,
	"SHUTDOWN":      StoppedCloudHostStatus,
	"TERMINATING":   TerminatedCloudHostStatus,
}

// AwsCmdbStatusMap is aws status to cmdb cloud host status map.
var AwsCmdbStatusMap = map[string]CloudHostStatus{
	"pending":       UnknownCloudHostStatus,
	"running":       RunningCloudHostStatus,
	"shutting-down": StoppingCloudHostStatus,
	"terminated":    TerminatedCloudHostStatus,
	"stopping":      StoppingCloudHostStatus,
	"stopped":       StoppedCloudHostStatus,
}

// GcpCmdbStatusMap is gcp status to cmdb cloud host status map.
var GcpCmdbStatusMap = map[string]CloudHostStatus{
	"PROVISIONING": UnknownCloudHostStatus,
	"STAGING":      StartingCloudHostStatus,
	"RUNNING":      RunningCloudHostStatus,
	"STOPPING":     StoppingCloudHostStatus,
	"SUSPENDING":   StoppingCloudHostStatus,
	"SUSPENDED":    StoppedCloudHostStatus,
	"REPAIRING":    UnknownCloudHostStatus,
	"TERMINATED":   TerminatedCloudHostStatus,
}

// AzureCmdbStatusMap is azure status to cmdb cloud host status map.
var AzureCmdbStatusMap = map[string]CloudHostStatus{
	"PowerState/running":      RunningCloudHostStatus,
	"PowerState/stopped":      StoppedCloudHostStatus,
	"PowerState/deallocating": StoppingCloudHostStatus,
	"PowerState/deallocated":  StoppedCloudHostStatus,
}

// HuaWeiCmdbStatusMap is huawei status to cmdb cloud host status map.
var HuaWeiCmdbStatusMap = map[string]CloudHostStatus{
	"BUILD":             UnknownCloudHostStatus,
	"REBOOT":            UnknownCloudHostStatus,
	"HARD_REBOOT":       UnknownCloudHostStatus,
	"REBUILD":           UnknownCloudHostStatus,
	"MIGRATING":         UnknownCloudHostStatus,
	"RESIZE":            UnknownCloudHostStatus,
	"ACTIVE":            RunningCloudHostStatus,
	"SHUTOFF":           StoppedCloudHostStatus,
	"REVERT_RESIZE":     UnknownCloudHostStatus,
	"VERIFY_RESIZE":     UnknownCloudHostStatus,
	"ERROR":             UnknownCloudHostStatus,
	"DELETED":           TerminatedCloudHostStatus,
	"SHELVED":           UnknownCloudHostStatus,
	"SHELVED_OFFLOADED": UnknownCloudHostStatus,
	"UNKNOWN":           UnknownCloudHostStatus,
}

// SearchBizCompanyCmdbInfoParams is search cmdb business belonging parameter.
type SearchBizCompanyCmdbInfoParams struct {
	BizIDs   []int64  `json:"bk_biz_ids,omitempty"`
	BizNames []string `json:"bk_biz_names,omitempty"`
	Page     BasePage `json:"page,omitempty"`
}

// CompanyCmdbInfoResult is search cmdb business belonging result.
type CompanyCmdbInfoResult struct {
	Data []CompanyCmdbInfo `json:"data"`
}

// CompanyCmdbInfo is search cmdb business belonging element of result.
type CompanyCmdbInfo struct {
	BkBizID          int64  `json:"bk_biz_id"`
	BizName          string `json:"bk_biz_name"`
	BkProductID      int64  `json:"bsi_product_id"`
	BkProductName    string `json:"bsi_product_name"`
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

// SearchBizBelongingParams is search cmdb business belonging parameter.
type SearchBizBelongingParams struct {
	BizIDs   []int64                `json:"bk_biz_ids,omitempty"`
	BizNames []string               `json:"bk_biz_names,omitempty"`
	Page     SearchBizBelongingPage `json:"page,omitempty"`
}

// SearchBizBelongingPage is search cmdb business belonging paging info.
type SearchBizBelongingPage struct {
	Limit int `json:"limit"`
	Start int `json:"start"`
}

// SearchBizBelongingResult is search cmdb business belonging result.
type SearchBizBelongingResult struct {
	Data []SearchBizBelonging `json:"data"`
}

// EventType is cmdb watch event type.
type EventType string

const (
	// Create is cmdb watch event create type.
	Create EventType = "create"
	// Update is cmdb watch event update type.
	Update EventType = "update"
	// Delete is cmdb watch event delete type.
	Delete EventType = "delete"
)

// CursorType is cmdb watch event cursor type.
type CursorType string

const (
	// HostType is cmdb watch event host cursor type.
	HostType CursorType = "host"
	// HostRelation is cmdb watch event host relation cursor type.
	HostRelation CursorType = "host_relation"
)

// WatchEventParams is esb watch cmdb event parameter.
type WatchEventParams struct {
	// event types you want to care, empty means all.
	EventTypes []EventType `json:"bk_event_types"`
	// the fields you only care, if nil, means all.
	Fields []string `json:"bk_fields"`
	// unix seconds timesss to where you want to watch from.
	// it's like Cursor, but StartFrom and Cursor can not use at the same time.
	StartFrom int64 `json:"bk_start_from"`
	// the cursor you hold previous, means you want to watch event form here.
	Cursor string `json:"bk_cursor"`
	// the resource kind you want to watch
	Resource CursorType       `json:"bk_resource"`
	Filter   WatchEventFilter `json:"bk_filter"`
}

// WatchEventFilter watch event filter
type WatchEventFilter struct {
	// SubResource the sub resource you want to watch, eg. object ID of the instance resource, watch all if not set
	SubResource string `json:"bk_sub_resource,omitempty"`
}

// CCErrEventChainNodeNotExist 如果事件节点不存在，cc会返回该错误码
var CCErrEventChainNodeNotExist = "1103007"

// WatchEventResult is cmdb watch event result.
type WatchEventResult struct {
	// watched events or not
	Watched bool               `json:"bk_watched"`
	Events  []WatchEventDetail `json:"bk_events"`
}

// WatchEventDetail is cmdb watch event detail.
type WatchEventDetail struct {
	Cursor    string          `json:"bk_cursor"`
	Resource  CursorType      `json:"bk_resource"`
	EventType EventType       `json:"bk_event_type"`
	Detail    json.RawMessage `json:"bk_detail"`
}

// HostModuleRelationParams get host and module relation parameter
type HostModuleRelationParams struct {
	BizID  int64   `json:"bk_biz_id,omitempty"`
	HostID []int64 `json:"bk_host_id"`
}

// 分割线，下面是woa内部cmdb esb

// AddHostResp add host to cc response
type AddHostResp struct {
	RespMeta `json:",inline"`
}

// TransferHostResp transfer host to another business response
type TransferHostResp struct {
	RespMeta `json:",inline"`
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

const (
	// LinuxOsType 操作系统类型-Linux
	LinuxOsType OsType = "1"
	// WindowsOsType 操作系统类型-Windows
	WindowsOsType OsType = "2"
)

// GetUniqOuterIp get CC host unique outer ip
func (h *Host) GetUniqOuterIp() string {
	// when CC host has multiple outer ips, bk_host_outerip is like "10.0.0.1,10.0.0.2"
	// return the first ip as host unique ip
	multiIps := strings.Split(h.BkHostOuterIP, ",")
	if len(multiIps) == 0 {
		return ""
	}

	return multiIps[0]
}

// IsPmAndOuterIPDevice 检查是否物理机，是否有外网IP
func (h *Host) IsPmAndOuterIPDevice() bool {
	// 服务器来源类型ID(未知(0, 默认值) 自有(1) 托管(2) 租用(3) 虚拟机(4) 容器(5))
	if h.SvrSourceTypeID != BkSvrSourceTypeIDSelf && h.SvrSourceTypeID != BkSvrSourceTypeIDDeposit &&
		h.SvrSourceTypeID != BkSvrSourceTypeIDLease {
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

// HostBizRelReq find host business relation request
type HostBizRelReq struct {
	BkHostId []int64 `json:"bk_host_id"`
}

// ccapi request

// AddHostReq add host to cc request
type AddHostReq struct {
	// to be added hosts' asset id list, max length is 10
	AssetIDs []string `json:"asset_ids"`
	InnerIps []string `json:"inner_ips"`
}

// TransferHostReq transfer host to another business request
type TransferHostReq struct {
	From TransferHostSrcInfo `json:"bk_from"`
	To   TransferHostDstInfo `json:"bk_to"`
}

// TransferHostSrcInfo transfer host source info
type TransferHostSrcInfo struct {
	FromBizID int64   `json:"bk_biz_id"`
	HostIDs   []int64 `json:"bk_host_ids"`
}

// TransferHostDstInfo transfer host destination info
type TransferHostDstInfo struct {
	ToBizID    int64 `json:"bk_biz_id"`
	ToModuleID int64 `json:"bk_module_id,omitempty"`
}

// UpdateHostsReq update hosts request
type UpdateHostsReq struct {
	Update []*UpdateHostProperty `json:"update"`
}

// UpdateHostProperty update hosts property
type UpdateHostProperty struct {
	HostID     int64                  `json:"bk_host_id"`
	Properties map[string]interface{} `json:"properties"`
}

// HostModuleRelationParameter get host and module relation parameter
type HostModuleRelationParameter struct {
	HostID []int64 `json:"bk_host_id"`
}

// ListBizHostReq list certain business host request
type ListBizHostReq struct {
	BkBizId            int64           `json:"bk_biz_id"`
	BkModuleIds        []int64         `json:"bk_module_ids"`
	ModuleCond         []ConditionItem `json:"module_cond"`
	HostPropertyFilter *QueryFilter    `json:"host_property_filter,omitempty"`
	Fields             []string        `json:"fields"`
	Page               BasePage        `json:"page"`
}

// ConditionItem cc query condition item
type ConditionItem struct {
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// SearchBizReq search business request
type SearchBizReq struct {
	Filter *querybuilder.QueryFilter `json:"biz_property_filter,omitempty"`
	Fields []string                  `json:"fields"`
	Page   BasePage                  `json:"page"`
}

// SearchModuleReq search module request
type SearchModuleReq struct {
	BkBizId   int64         `json:"bk_biz_id"`
	Condition mapstr.MapStr `json:"condition"`
	Fields    []string      `json:"fields"`
	Page      BasePage      `json:"page"`
}

// GetBizInternalModuleReq get business's internal module request
type GetBizInternalModuleReq struct {
	BkBizID int64 `json:"bk_biz_id"`
}

// CrTransitReq transfer host to CR transit module request
type CrTransitReq struct {
	From CrTransitSrcInfo `json:"bk_from"`
	To   CrTransitDstInfo `json:"bk_to"`
}

// CrTransitSrcInfo transfer host source info
type CrTransitSrcInfo struct {
	FromBizID    int64 `json:"bk_biz_id"`
	FromModuleID int64 `json:"bk_module_id"`
	// max size is 10
	AssetIDs []string `json:"asset_ids"`
}

// CrTransitDstInfo transfer host destination info
type CrTransitDstInfo struct {
	ToBizID int64 `json:"bk_biz_id"`
}

// CrTransitIdleReq transfer host from CR transit module back to idle module request
type CrTransitIdleReq struct {
	BkBizId  int64    `json:"bk_biz_id"`
	AssetIDs []string `json:"asset_ids"`
}
