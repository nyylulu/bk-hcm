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

// Package itsmapi ...
package itsmapi

// 资源申请单字段
const (
	// TicketKeyTitle 标题
	TicketKeyTitle string = "title"
	// TicketKeyApplyId 资源申请单号
	TicketKeyApplyId string = "ZIYUANSHENQINGDANHAO"
	// TicketKeyApplyLink 资源申请单链接
	TicketKeyApplyLink string = "ZIYUANSHENQINGDANLIANJIE"
	// TicketKeyNeedSysAudit 是否需要系统审核
	TicketKeyNeedSysAudit string = "SHIFOUXUYAOXITONGSHENHE"

	TicketValTitleFormat     string = "资源申请单据审核[order_id:%d]"
	TicketValNeedSysAuditYes string = "SHI"
	TicketValNeedSysAuditNo  string = "FOU"

	// ActionTypeTransition 审批类型
	ActionTypeTransition string = "TRANSITION"
)

// MapStateKey 资源申请单的状态
var MapStateKey = map[int64][]string{
	// stage环境
	// leader审核
	1957: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	1955: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},

	// prod环境
	// leader审核
	1798: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	1797: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},
}
