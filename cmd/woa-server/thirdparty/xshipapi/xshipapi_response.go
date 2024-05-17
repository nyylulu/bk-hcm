/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xshipapi

type AcceptStatus int

// RespMeta xship response meta info
type RespMeta struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	TranceID string `json:"traceId"`
}

// ReinstallResp create host reinstall task response
type ReinstallResp struct {
	RespMeta `json:",inline"`
	Data     *ReinstallRst `json:"data"`
}

// ReinstallRst create host reinstall task result
type ReinstallRst struct {
	AcceptOrders []*AcceptOrder `json:"processAcceptList"`
}

// AcceptOrder xship accept order info
type AcceptOrder struct {
	OrderID      string `json:"acceptOrderId"`
	AcceptStatus int    `json:"acceptStatus"`
	AcceptMsg    string `json:"acceptMessage"`
	AssetID      string `json:"assetId"`
	Uuid         string `json:"uuid"`
}

// ReinstallStatusResp get host reinstall task status response
type ReinstallStatusResp struct {
	RespMeta `json:",inline"`
	Data     *ReinstallStatusRst `json:"data"`
}

// ReinstallStatusRst host reinstall task status result
type ReinstallStatusRst struct {
	TotalNum       int                `json:"totalNum"`
	ReinstallInfos []*ReinstallStatus `json:"reinstallInfos"`
}

// ReinstallStatus host reinstall task status
type ReinstallStatus struct {
	OrderID    string       `json:"acceptOrderId"`
	AssetID    string       `json:"serverAssetId"`
	IP         string       `json:"ip"`
	Starter    string       `json:"starter"`
	Status     AcceptStatus `json:"acceptStatus"`
	ErrMsg     string       `json:"errMessage"`
	RejectCode string       `json:"rejectCode"`
	SourceOS   string       `json:"sourceOs"`
	ReqOS      string       `json:"reqOs"`
	SourceRaid string       `json:"sourceRaid"`
	ReqRaid    string       `json:"reqRaid"`
	CreateTime string       `json:"createTime"`
	EndTime    string       `json:"endTime"`
}
