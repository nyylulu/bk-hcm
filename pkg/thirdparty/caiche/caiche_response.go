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

package caiche

import "hcm/pkg/criteria/enumor"

// GetTokenResp get token response
type GetTokenResp struct {
	ID      string          `json:"id"`
	JsonRPC string          `json:"jsonrpc"`
	Result  *GetTokenResult `json:"result"`
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
}

// GetTokenResult get token result
type GetTokenResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// ListDeviceResp list device response
type ListDeviceResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data *DeviceListData `json:"data"`
}

// DeviceListData device list data
type DeviceListData struct {
	DataList   []Device `json:"dataList"`
	Total      int      `json:"total"`
	UpdateTime string   `json:"update_time"`
}

// Device device
type Device struct {
	SvrAssetId            string              `json:"svr_assetId"`
	EqsName               string              `json:"eqs_name"`
	ServerLanIP           string              `json:"server_lan_ip"`
	ServerWanIP           string              `json:"server_wan_ip"`
	DeptName              string              `json:"dept_name"`
	SvrOperator           string              `json:"svr_operator"`
	SvrBakOperator        string              `json:"svr_bak_operator"`
	BsiPath               string              `json:"bsi_path"`
	IdcName               string              `json:"Idc_name"`
	Region                string              `json:"region"`
	ZoneName              string              `json:"zone_name"`
	Campus                string              `json:"campus"`
	Module                string              `json:"module"`
	ServerLogicDomain     string              `json:"server_logic_domain"`
	SvrDeviceClassName    string              `json:"svr_device_class_name"`
	ProjectName           string              `json:"project_name"`
	AbolishStatus         string              `json:"abolish_status"`
	AbolishPrincipal      string              `json:"abolish_principal"`
	PlanProduct           string              `json:"plan_product"`
	AbolishDate           string              `json:"abolish_date"`
	ProjectID             string              `json:"project_id"`
	AlterAbolishPrincipal string              `json:"alter_abolish_principal"`
	ExpectAbolishDate     string              `json:"expect_abolish_date"`
	AckStatus             string              `json:"ack_status"`
	AbolishPhase          enumor.AbolishPhase `json:"abolish_phase"`
	DefaultBsiGroup       string              `json:"default_bsi_group"`
	DefaultBG             string              `json:"default_bg"`
	AckEndDate            string              `json:"ack_end_date"`
	HasIPRelatedInfo      bool                `json:"has_ip_related_info"`
	SvrTypeName           string              `json:"svr_type_name"`
	BsiGroupL2            string              `json:"bsi_group_l2"`
	SvrOwnerAssetID       string              `json:"svr_owner_asset_id"`
	Down                  string              `json:"down"`
	SelfSvrTypeName       string              `json:"self_svr_type_name"`
	HighDispersion        string              `json:"high_dispersion"`
	UpdatedAt             string              `json:"updated_at"`
	ID                    string              `json:"id"`
	IdcParentName         string              `json:"idc_parent_name"`
	ServerRack            string              `json:"server_rack"`
	RckID                 string              `json:"rck_id"`
	PosCode               string              `json:"pos_code"`
	YunxiDst              string              `json:"yunxi_dst"`
	PlanDestination       string              `json:"plan_destination"`
	PlanFinishTime        string              `json:"plan_finish_time"`
	InnerSwitchAssetID    string              `json:"inner_switch_asset_id"`
	InnerSwitchIP         string              `json:"inner_switch_ip"`
	NeedUser              string              `json:"need_user"`
	YunxiPriority         string              `json:"yunxi_priority"`
	YunxiType             string              `json:"yunxi_type"`
	MoveDest              string              `json:"move_dest"`
	ProjectCode           string              `json:"project_code"`
	VirtualDepartmentName string              `json:"virtual_department_name"`
	ObgBG                 string              `json:"obg_bg"`
}
