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
	"math"
	"strconv"
	"time"
)

// RecycleType resource recycle type
type RecycleType string

// definition of various recycle type
const (
	RecycleTypeRegular     RecycleType = "常规项目"
	RecycleTypeDissolve    RecycleType = "机房裁撤"
	RecycleTypeExpired     RecycleType = "过保裁撤"
	RecycleTypeSpring      RecycleType = "春节保障"
	RecycleTypeRollServer  RecycleType = "滚服项目"
	RecycleTypeShortRental RecycleType = "短租项目"
)

// CanUpdateRecycleType 输入当前的回收类型，以及想要更新的回收类型，根据回收类型的优先级进行判断，
// 如果当前回收类型优先级高于想要更新的回收类型，则返回false，否则返回true
func (rt RecycleType) CanUpdateRecycleType(recycleTypeSeq []RecycleType, desired RecycleType) bool {
	curPriority := rt.getRecycleTypePriority(recycleTypeSeq)
	desiredPriority := desired.getRecycleTypePriority(recycleTypeSeq)

	// 值越小，优先级越高
	if curPriority < desiredPriority {
		return false
	}

	return true
}

// getRecycleTypePriority 返回值越小，优先级越高
func (rt RecycleType) getRecycleTypePriority(recycleTypeSeq []RecycleType) int {
	if rt.IsFixedType() {
		return math.MinInt
	}

	// 如果提供了可变回收类型的优先级序列，则按照提供的优先级排序；未提及的回收类型，优先级最低
	seqNum := len(recycleTypeSeq)
	for i, r := range recycleTypeSeq {
		if r == rt {
			return i
		}
	}

	switch rt {
	case RecycleTypeRollServer:
		return seqNum + 0
	case RecycleTypeShortRental:
		return seqNum + 1
	default:
		return math.MaxInt
	}
}

// IsFixedType return whether recycle type is fixed
func (rt RecycleType) IsFixedType() bool {
	switch rt {
	// 机房裁撤和春节保障属于固定回收类型，优先级最高
	case RecycleTypeDissolve, RecycleTypeSpring:
		return true
	default:
		return false
	}
}

// ToObsProject convert recycle type to OBS project name
func (rt RecycleType) ToObsProject() string {
	switch rt {
	case RecycleTypeSpring:
		return rt.getSpringObsProject()
	case RecycleTypeDissolve:
		return rt.getDissolveObsProject()
	case RecycleTypeRollServer:
		return string(RecycleTypeRollServer)
	case RecycleTypeShortRental:
		return string(RecycleTypeShortRental)
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
		RecycleTypeSpring, RecycleTypeRollServer, RecycleTypeShortRental:
	default:
		return fmt.Errorf("validate unknown recycle type: %s", rt)
	}
	return nil
}

// RecycleStatus recycle status
type RecycleStatus string

// definition of various recycle status
const (
	RecycleStatusDefault          RecycleStatus = "DEFAULT"
	RecycleStatusUncommit         RecycleStatus = "UNCOMMIT"
	RecycleStatusCommitted        RecycleStatus = "COMMITTED"
	RecycleStatusDetecting        RecycleStatus = "DETECTING"
	RecycleStatusDetectFailed     RecycleStatus = "DETECT_FAILED"
	RecycleStatusAudit            RecycleStatus = "FOR_AUDIT"
	RecycleStatusRejected         RecycleStatus = "REJECTED"
	RecycleStatusTransiting       RecycleStatus = "TRANSITING"
	RecycleStatusTransitFailed    RecycleStatus = "TRANSIT_FAILED"
	RecycleStatusReturning        RecycleStatus = "RETURNING"
	RecycleStatusReturnFailed     RecycleStatus = "RETURN_FAILED"
	RecycleStatusReturningPlan    RecycleStatus = "RETURNING_PLAN"
	RecycleStatusReturnPlanFailed RecycleStatus = "RETURN_PLAN_FAILED"
	RecycleStatusDone             RecycleStatus = "DONE"
	RecycleStatusTerminate        RecycleStatus = "TERMINATE"
)

// definition of various recycle status description
const (
	RecycleStatusDescUncommit         string = "未提单"
	RecycleStatusDescCommitted        string = "已提单"
	RecycleStatusDescDetecting        string = "预检中"
	RecycleStatusDescDetectFailed     string = "预检失败"
	RecycleStatusDescAudit            string = "待审核"
	RecycleStatusDescRejected         string = "已驳回"
	RecycleStatusDescTransiting       string = "中转中"
	RecycleStatusDescTransitFailed    string = "中转失败"
	RecycleStatusDescReturning        string = "退回中"
	RecycleStatusDescReturnFailed     string = "退回失败"
	RecycleStatusDescReturningPlan    string = "返还预测中"
	RecycleStatusDescReturnPlanFailed string = "返还预测失败"
	RecycleStatusDescDone             string = "已完成"
	RecycleStatusDescTerminate        string = "终止"
)

// RecycleStage recycle stage
type RecycleStage string

// definition of various recycle stage
const (
	RecycleStageCommit     RecycleStage = "COMMIT"
	RecycleStageDetect     RecycleStage = "DETECT"
	RecycleStageAudit      RecycleStage = "AUDIT"
	RecycleStageTransit    RecycleStage = "TRANSIT"
	RecycleStageReturn     RecycleStage = "RETURN"
	RecycleStageReturnPlan RecycleStage = "RETURN_PLAN" // 返还预测
	RecycleStageDone       RecycleStage = "DONE"
	RecycleStageTerminate  RecycleStage = "TERMINATE"
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

// DetectStepName detection task step name
type DetectStepName string

// definition of various detection task step name
const (
	StepPreCheck       DetectStepName = "PRE_CHECK"
	StepCheckUwork     DetectStepName = "CHECK_UWORK"
	StepCheckTcaplus   DetectStepName = "CHECK_TCAPLUS"
	StepCheckDBM       DetectStepName = "CHECK_DBM"
	StepBasicCheck     DetectStepName = "BASIC_CHECK"
	StepCheckOwner     DetectStepName = "CHECK_OWNER"
	StepCvmCheck       DetectStepName = "CVM_CHECK"
	StepCheckReturn    DetectStepName = "CHECK_RETURN"
	StepCheckProcess   DetectStepName = "CHECK_PROCESS"
	StepCheckPmOuterIP DetectStepName = "CHECK_PM_OUTERIP" // 物理机外网IP检查
)

const (
	// DetectStepsPerTask 一个空闲检查任务包含的步骤数
	DetectStepsPerTask = 10
	// DetectTaskMaxPageLimit 空闲检查任务执行详情查询分页最大数量
	// 1台待空闲检查的主机->1个detectTask->10个detectStep，因为500/10=50，所以限制查询主机数为50
	DetectTaskMaxPageLimit = 500 / DetectStepsPerTask
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
