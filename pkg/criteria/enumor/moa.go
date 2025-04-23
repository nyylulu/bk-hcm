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

import (
	"errors"

	"hcm/pkg/iam/meta"
)

// MoaStatus MOA验证状态
type MoaStatus string

const (
	// MoaStatusPending ...
	MoaStatusPending MoaStatus = "pending"
	// MoaStatusFinish ...
	MoaStatusFinish MoaStatus = "finish"
)

// Moa2FAChannel 2FA渠道
type Moa2FAChannel string

const (
	// Moa2FAChannelMOA MOA弹窗确认
	Moa2FAChannelMOA Moa2FAChannel = "moa"
	// Moa2FAChannelSMS 短信验证码
	Moa2FAChannelSMS Moa2FAChannel = "sms"
)

// Validate ...
func (m Moa2FAChannel) Validate() error {
	switch m {
	case Moa2FAChannelMOA, Moa2FAChannelSMS:
		return nil
	default:
		return errors.New("invalid moa 2fa channel: " + string(m))
	}
}

// MoaButtonType Moa 操作按钮类型
type MoaButtonType string

const (
	// MoaButtonTypeCancel 取消按钮
	MoaButtonTypeCancel MoaButtonType = "cancel"
	// MoaButtonTypeConfirm 确定按钮
	MoaButtonTypeConfirm MoaButtonType = "confirm"
)

// Validate ...
func (m MoaButtonType) Validate() error {
	switch m {
	case MoaButtonTypeCancel, MoaButtonTypeConfirm:
		return nil
	default:
		return errors.New("invalid moa button type: " + string(m))
	}
}

// MoaScene 需要MOA验证的操作场景
type MoaScene string

const (
	// MoaSceneSGDelete 安全组删除
	MoaSceneSGDelete MoaScene = "sg_delete"
	// MoaSceneCVMStart CVM开机
	MoaSceneCVMStart MoaScene = "cvm_start"
	// MoaSceneCVMStop CVM关机
	MoaSceneCVMStop MoaScene = "cvm_stop"
	// MoaSceneCVMReset CVM重装
	MoaSceneCVMReset MoaScene = "cvm_reset"
	// MoaSceneCVMReboot CVM重启
	MoaSceneCVMReboot MoaScene = "cvm_reboot"
)

// GetResType ...
func (s MoaScene) GetResType() meta.ResourceType {
	switch s {
	case MoaSceneSGDelete:
		return meta.SecurityGroup
	case MoaSceneCVMStart, MoaSceneCVMStop, MoaSceneCVMReset, MoaSceneCVMReboot:
		return meta.Cvm
	default:
		return "unknown_moa_scene"
	}
}

// MoaVerifyStatus MOA验证状态
type MoaVerifyStatus string

const (
	// MoaVerifyPending 等待用户验证
	MoaVerifyPending MoaVerifyStatus = "pending"
	// MoaVerifyConfirmed 用户已确认
	MoaVerifyConfirmed MoaVerifyStatus = "confirmed"
	// MoaVerifyRejected 用户已拒绝
	MoaVerifyRejected MoaVerifyStatus = "rejected"
	// MoaVerifyNotFound 相关信息未找到或已失效
	MoaVerifyNotFound MoaVerifyStatus = "expired_or_not_found"
)
