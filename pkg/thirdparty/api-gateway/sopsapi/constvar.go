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

// Package sopsapi ...
package sopsapi

const (
	// TaskStateCreated 未执行
	TaskStateCreated string = "CREATED"
	// TaskStateRunning 执行中
	TaskStateRunning string = "RUNNING"
	// TaskStateFailed 失败
	TaskStateFailed string = "FAILED"
	// TaskStateSuspended 暂停
	TaskStateSuspended string = "SUSPENDED"
	// TaskStateRevoked 已终止
	TaskStateRevoked string = "REVOKED"
	// TaskStateFinished 已完成
	TaskStateFinished string = "FINISHED"
)

// bk-ops标准运维插件
const (
	// CommonTemplateSource bk-sops标准运维-普通模版来源
	CommonTemplateSource = "common"
	// InitLinuxTemplateID 初始化-Linux-的流程模版ID
	InitLinuxTemplateID int64 = 10078
	// InitLinuxTaskNamePrefix 初始化-Linux-新建任务名称的前缀
	InitLinuxTaskNamePrefix = "【常用】【SA】【Linux】初始化-%s-%s"
	// InitWindowsTemplateID 初始化-Windows-的流程模版ID
	InitWindowsTemplateID int64 = 10082
	// InitWindowsTaskNamePrefix 初始化-Windows-新建任务名称的前缀
	InitWindowsTaskNamePrefix = "【常用】【SA】【Windows】初始化-%s-%s"
	// ConfigCheckLinux 配置检查-Linux（已确认:不需要Windows）
	ConfigCheckLinux int64 = 10069
	// ConfigCheckLinuxTaskNamePrefix 配置检查-Linux-新建任务名称的前缀
	ConfigCheckLinuxTaskNamePrefix = "【常用】【SA】【Linux】配置检查-%s"
	// DataClearLinux 数据清理-Linux（已确认:不需要Windows）
	DataClearLinux int64 = 10201
	// DataClearLinuxTaskNamePrefix 数据清理-Linux-新建任务名称的前缀
	DataClearLinuxTaskNamePrefix = "【危险】【SA】【Linux】数据清理-%s"
	// IdleCheckLinux 空闲检查-Linux
	IdleCheckLinux int64 = 10102
	// IdleCheckLinuxTaskNamePrefix 空闲检查-Linux-新建任务名称的前缀
	IdleCheckLinuxTaskNamePrefix = "【回收接口调用】【SA】【Linux】空闲检查-%s"
	// IdleCheckWindows 空闲检查-Windows
	IdleCheckWindows int64 = 10103
	// IdleCheckWindowsTaskNamePrefix 空闲检查-Windows-新建任务名称的前缀
	IdleCheckWindowsTaskNamePrefix = "【回收接口调用】【SA】【Windows】空闲检查-%s"
	// RecycleOuterIPLinux 回收外网IP-Linux
	RecycleOuterIPLinux int64 = 10206
	// RecycleOuterIPLinuxTaskNamePrefix 回收外网IP-Linux-新建任务名称的前缀
	RecycleOuterIPLinuxTaskNamePrefix = "【危险】【SA】【Linux】回收外网IP-%s"
	// RecycleOuterIPWindows 回收外网IP-Windows
	RecycleOuterIPWindows int64 = 10207
	// RecycleOuterIPWindowsTaskNamePrefix 回收外网IP-Windows-新建任务名称的前缀
	RecycleOuterIPWindowsTaskNamePrefix = "【危险】【SA】【Windows】回收外网IP-%s"
)

// SopsCreateType 非scr申请主机创建sops任务
const SopsCreateType = "NonOrderType"
