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

package itsm

import (
	"fmt"
	"sync"
)

type Status string

const (
	// StatusRunning 处理中
	StatusRunning Status = "RUNNING"
	// StatusFinished 已结束
	StatusFinished Status = "FINISHED"
	// StatusTerminated 被终止
	StatusTerminated Status = "TERMINATED"
	// StatusRevoked 已撤销
	StatusRevoked Status = "REVOKED"
	// StatusSuspended 被挂起
	StatusSuspended Status = "SUSPENDED"
)

// GetTicketStatusResp get itsm ticket status response
type GetTicketStatusResp struct {
	RespMeta `json:",inline"`
	Data     *GetTicketStatusRst `json:"data"`
}

// ServerDiscovery server discovery
type ServerDiscovery struct {
	name    string
	servers []string
	index   int
	sync.RWMutex
}

// GetServers return server instance address
func (s *ServerDiscovery) GetServers() ([]string, error) {
	if s == nil {
		return []string{}, nil
	}
	s.RLock()
	defer s.RUnlock()

	num := len(s.servers)
	if num == 0 {
		return []string{}, fmt.Errorf("oops, there is no %s server can be used", s.name)
	}

	if s.index < num-1 {
		s.index = s.index + 1
		return append(s.servers[s.index-1:], s.servers[:s.index-1]...), nil
	} else {
		s.index = 0
		return append(s.servers[num-1:], s.servers[:num-1]...), nil
	}
}

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
	// TicketKeyApplyReason 申请理由
	TicketKeyApplyReason string = "apply_reason"

	TicketValTitleFormat     string = "资源申请单据审核[order_id:%d]"
	TicketValNeedSysAuditYes string = "SHI"
	TicketValNeedSysAuditNo  string = "FOU"

	// ActionTypeTransition 审批类型
	ActionTypeTransition string = "TRANSITION"
)

// MapStateKey 申请单对应的审批意见节点key
var MapStateKey = map[int64][]string{
	// devhk环境
	// leader审核
	7185: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	7184: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},
	// stage环境
	// leader审核
	7250: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	7249: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},

	// grey环境
	// leader审核
	6241: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	6240: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},
	// prod环境
	// leader审核
	6204: []string{
		// 审批意见key
		"eb81f856de91db69e5d2d2c5a7c45c40",
		// 备注key
		"5357038ea9c2b82567166bb872e59b2d"},
	// 管理员审核
	6203: []string{
		// 审批意见key
		"811a0bd2bf65a754b522bc1e48e1c91e",
		// 备注key
		"e3dda6544a3a67f67d85737a9027e4e5"},
}

// 审批流程中的特殊节点
const (
	// AuditNodeStart 流程开始
	AuditNodeStart string = "流程开始."
	// AuditNodeEnd 流程结束
	AuditNodeEnd string = "流程结束."
)

// ValidateTypeRequire itsm 必填节点
const ValidateTypeRequire = "REQUIRE"
