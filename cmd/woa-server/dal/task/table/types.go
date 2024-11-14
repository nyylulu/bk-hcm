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

// Package table provide definition for table
package table

import (
	"fmt"
	"strconv"
	"time"
)

// RecycleType resource recycle type
type RecycleType string

// definition of various recycle type
const (
	RecycleTypeRegular    RecycleType = "常规项目"
	RecycleTypeDissolve   RecycleType = "机房裁撤"
	RecycleTypeExpired    RecycleType = "过保裁撤"
	RecycleTypeSpring     RecycleType = "春节保障"
	RecycleTypeRent       RecycleType = "短租项目"
	RecycleTypeRollServer RecycleType = "滚服项目"
)

// ToObsProject convert recycle type to OBS project name
func (rt RecycleType) ToObsProject() string {
	switch rt {
	case RecycleTypeSpring:
		return rt.getSpringObsProject()
	case RecycleTypeRent:
		return string(RecycleTypeRent)
	case RecycleTypeDissolve:
		return rt.getDissolveObsProject()
	case RecycleTypeRollServer:
		return string(RecycleTypeRollServer)
	default:
		return string(RecycleTypeRegular)
	}
}

func (rt RecycleType) getSpringObsProject() string {
	// 资源回收的春保窗口期：12月1日～次年4月20日
	// 12月1日～12月31日提单的春保项目前缀为次年
	year := time.Now().Local().Year()
	if time.Now().Month() == time.December {
		year += 1
	}

	prefixYear := strconv.Itoa(year)
	project := prefixYear + string(RecycleTypeSpring)

	return project
}

func (rt RecycleType) getDissolveObsProject() string {
	// TODO:
	// 暂定按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	year := time.Now().Local().Year()
	prefixYear := strconv.Itoa(year)
	project := prefixYear + string(RecycleTypeDissolve)

	return project
}

// Validate validate
func (rt RecycleType) Validate() error {
	switch rt {
	case RecycleTypeRegular, RecycleTypeDissolve, RecycleTypeExpired,
		RecycleTypeSpring, RecycleTypeRent, RecycleTypeRollServer:
	default:
		return fmt.Errorf("validate unknown recycle type: %s", rt)
	}
	return nil
}

// RecycleStatus recycle status
type RecycleStatus string

// definition of various recycle status
const (
	RecycleStatusDefault       RecycleStatus = "DEFAULT"
	RecycleStatusUncommit      RecycleStatus = "UNCOMMIT"
	RecycleStatusCommitted     RecycleStatus = "COMMITTED"
	RecycleStatusDetecting     RecycleStatus = "DETECTING"
	RecycleStatusDetectFailed  RecycleStatus = "DETECT_FAILED"
	RecycleStatusAudit         RecycleStatus = "FOR_AUDIT"
	RecycleStatusRejected      RecycleStatus = "REJECTED"
	RecycleStatusTransiting    RecycleStatus = "TRANSITING"
	RecycleStatusTransitFailed RecycleStatus = "TRANSIT_FAILED"
	RecycleStatusReturning     RecycleStatus = "RETURNING"
	RecycleStatusReturnFailed  RecycleStatus = "RETURN_FAILED"
	RecycleStatusDone          RecycleStatus = "DONE"
	RecycleStatusTerminate     RecycleStatus = "TERMINATE"
)

// definition of various recycle status description
const (
	RecycleStatusDescUncommit      string = "未提单"
	RecycleStatusDescCommitted     string = "已提单"
	RecycleStatusDescDetecting     string = "预检中"
	RecycleStatusDescDetectFailed  string = "预检失败"
	RecycleStatusDescAudit         string = "待审核"
	RecycleStatusDescRejected      string = "已驳回"
	RecycleStatusDescTransiting    string = "中转中"
	RecycleStatusDescTransitFailed string = "中转失败"
	RecycleStatusDescReturning     string = "退回中"
	RecycleStatusDescReturnFailed  string = "退回失败"
	RecycleStatusDescDone          string = "已完成"
	RecycleStatusDescTerminate     string = "终止"
)

// RecycleStage recycle stage
type RecycleStage string

// definition of various recycle stage
const (
	RecycleStageCommit    RecycleStage = "COMMIT"
	RecycleStageDetect    RecycleStage = "DETECT"
	RecycleStageAudit     RecycleStage = "AUDIT"
	RecycleStageTransit   RecycleStage = "TRANSIT"
	RecycleStageReturn    RecycleStage = "RETURN"
	RecycleStageDone      RecycleStage = "DONE"
	RecycleStageTerminate RecycleStage = "TERMINATE"
)

// definition of various recycle stage description
const (
	RecycleStageDescCommit    string = "提单"
	RecycleStageDescDetect    string = "预检"
	RecycleStageDescAudit     string = "审核"
	RecycleStageDescTransit   string = "中转"
	RecycleStageDescReturn    string = "退回"
	RecycleStageDescDone      string = "完成"
	RecycleStageDescTerminate string = "终止"
)

// ResourceType resource type
type ResourceType string

// definition of various recycle resource type
const (
	ResourceTypePm          ResourceType = "IDCPM"
	ResourceTypeCvm         ResourceType = "QCLOUDCVM"
	ResourceTypeOthers      ResourceType = "OTHERS"
	ResourceTypeUnsupported ResourceType = "UNSUPPORTED"
)

// RetPlanType resource return plan type
type RetPlanType string

// definition of various resource return plan type
const (
	RetPlanImmediate RetPlanType = "IMMEDIATE"
	RetPlanDelay     RetPlanType = "DELAY"
)

// DetectStatus recycle detection status
type DetectStatus string

// definition of various detection task status
const (
	DetectStatusInit    DetectStatus = "INIT"
	DetectStatusRunning DetectStatus = "RUNNING"
	DetectStatusPaused  DetectStatus = "PAUSED"
	DetectStatusSuccess DetectStatus = "SUCCESS"
	DetectStatusFailed  DetectStatus = "FAILED"
)

// DetectStatusSeq recycle detection status sequence, for recycle detection task ordering
type DetectStatusSeq int

// definition of various detection task status sequence
const (
	DetectStatusSeqFailed  DetectStatusSeq = 1
	DetectStatusSeqPaused  DetectStatusSeq = 2
	DetectStatusSeqRunning DetectStatusSeq = 3
	DetectStatusSeqInit    DetectStatusSeq = 4
	DetectStatusSeqSuccess DetectStatusSeq = 5
)

// DetectStatus2Seq map of recycle detection status to recycle detection status sequence
var DetectStatus2Seq = map[DetectStatus]DetectStatusSeq{
	DetectStatusFailed:  DetectStatusSeqFailed,
	DetectStatusPaused:  DetectStatusSeqPaused,
	DetectStatusRunning: DetectStatusSeqRunning,
	DetectStatusInit:    DetectStatusSeqInit,
	DetectStatusSuccess: DetectStatusSeqSuccess,
}

// DetectStepName detection task step name
type DetectStepName string

// definition of various detection task step name
const (
	StepPreCheck       DetectStepName = "PRE_CHECK"
	StepCheckUwork     DetectStepName = "CHECK_UWORK"
	StepCheckGCS       DetectStepName = "CHECK_GCS"
	StepBasicCheck     DetectStepName = "BASIC_CHECK"
	StepCheckOwner     DetectStepName = "CHECK_OWNER"
	StepCvmCheck       DetectStepName = "CVM_CHECK"
	StepCheckSafety    DetectStepName = "CHECK_SAFETY"
	StepCheckReturn    DetectStepName = "CHECK_RETURN"
	StepCheckProcess   DetectStepName = "CHECK_PROCESS"
	StepCheckPmOuterIP DetectStepName = "CHECK_PM_OUTERIP" // 物理机外网IP检查
)

// ReturnStatus recycle detection status
type ReturnStatus string

// definition of various resource return task status
const (
	ReturnStatusInit    ReturnStatus = "INIT"
	ReturnStatusRunning ReturnStatus = "RUNNING"
	ReturnStatusPaused  ReturnStatus = "PAUSED"
	ReturnStatusSuccess ReturnStatus = "SUCCESS"
	ReturnStatusFailed  ReturnStatus = "FAILED"
)

// PoolType cvm resource pool type
type PoolType int

// definition of various cvm resource pool type
const (
	PoolPrivate PoolType = 0
	PoolPublic  PoolType = 1
)
