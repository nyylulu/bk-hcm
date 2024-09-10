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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/thirdparty/esb/types"
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

// SearchBizResp is cmdb search business response.
type SearchBizResp struct {
	types.BaseResponse
	SearchBizResult `json:"data"`
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
	BizID              int64        `json:"bk_biz_id"`
	BkSetIDs           []int64      `json:"bk_set_ids"`
	BkModuleIDs        []int64      `json:"bk_module_ids"`
	Fields             []string     `json:"fields"`
	Page               BasePage     `json:"page"`
	HostPropertyFilter *QueryFilter `json:"host_property_filter,omitempty"`
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
}

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
	"bk_asset_id",
	"bk_svr_device_cls_name",

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

// SearchBizBelonging is search cmdb business belonging element of result.
type SearchBizBelonging struct {
	BizID            int64  `json:"bk_biz_id"`
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
