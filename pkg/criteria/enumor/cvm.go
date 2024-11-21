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

package enumor

// CvmResetStatus define cvm reset status
type CvmResetStatus int

// ResetStatus cvm重装状态
const (
	// NormalCvmResetStatus 状态-正常
	NormalCvmResetStatus CvmResetStatus = 0
	// NoOperatorCvmResetStatus 状态-不是主备负责人
	NoOperatorCvmResetStatus CvmResetStatus = 1
	// NoIdleCvmResetStatus 状态-不在空闲机模块
	NoIdleCvmResetStatus CvmResetStatus = 2
)

// 任务类型
const (
	// ResetCvmTaskType 任务类型-CVM重装
	ResetCvmTaskType = TaskType(FlowResetCvm)
	// StartCvmTaskType 任务类型-启动云服务器
	StartCvmTaskType = TaskType(FlowStartCvm)
	// StopCvmTaskType 任务类型-停止云服务器
	StopCvmTaskType = TaskType(FlowStopCvm)
	// RebootCvmTaskType 任务类型-重启云服务器
	RebootCvmTaskType = TaskType(FlowRebootCvm)
)

// CvmOperateStatus define cvm operate status
type CvmOperateStatus int

// OperateStatus cvm 电源操作状态
const (
	// CvmOperateStatusNormal 状态-正常
	CvmOperateStatusNormal CvmOperateStatus = 0
	// CvmOperateStatusNoOperator 状态-不是主备负责人
	CvmOperateStatusNoOperator CvmOperateStatus = 1
	// CvmOperateStatusNoIdle 状态-不在空闲机模块
	CvmOperateStatusNoIdle CvmOperateStatus = 2
)

// CvmOperateType define cvm operate type
type CvmOperateType string

const (
	// CvmOperateTypeStart 启动云服务器
	CvmOperateTypeStart = "start"
	// CvmOperateTypeStop 停止云服务器
	CvmOperateTypeStop = "stop"
	// CvmOperateTypeReboot 重启云服务器
	CvmOperateTypeReboot = "reboot"
	// CvmOperateTypeReset 重装云服务器
	CvmOperateTypeReset = "reset"
)
