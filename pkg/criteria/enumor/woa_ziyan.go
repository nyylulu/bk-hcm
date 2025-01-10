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
	"fmt"
	"strconv"
	"time"

	"hcm/pkg/tools/converter"
)

// RPTicketType is resource plan ticket type.
type RPTicketType string

const (
	// RPTicketTypeAdd is resource plan ticket status add.
	RPTicketTypeAdd RPTicketType = "add"
	// RPTicketTypeAdjust is resource plan ticket status adjust.
	RPTicketTypeAdjust RPTicketType = "adjust"
	// RPTicketTypeDelete is resource plan ticket status delete.
	RPTicketTypeDelete RPTicketType = "delete"
)

// Validate RPTicketType.
func (t RPTicketType) Validate() error {
	switch t {
	case RPTicketTypeAdd, RPTicketTypeAdjust, RPTicketTypeDelete:
	default:
		return fmt.Errorf("unsupported resource plan type: %s", t)
	}

	return nil
}

// rdTicketTypeNameMap records RPTicketType's name.
var rdTicketTypeNameMap = map[RPTicketType]string{
	RPTicketTypeAdd:    "新增",
	RPTicketTypeAdjust: "调整",
	RPTicketTypeDelete: "取消",
}

// Name return RPTicketType's name.
func (t RPTicketType) Name() string {
	return rdTicketTypeNameMap[t]
}

// GetRPTicketTypeMembers get RPTicketType's members.
func GetRPTicketTypeMembers() []RPTicketType {
	return []RPTicketType{
		RPTicketTypeAdd,
		RPTicketTypeAdjust,
		RPTicketTypeDelete,
	}
}

// RPTicketStatus is resource plan ticket status.
type RPTicketStatus string

const (
	// RPTicketStatusInit is resource plan ticket status init.
	RPTicketStatusInit RPTicketStatus = "init"
	// RPTicketStatusAuditing is resource plan ticket status auditing.
	RPTicketStatusAuditing RPTicketStatus = "auditing"
	// RPTicketStatusRejected is resource plan ticket status rejected.
	RPTicketStatusRejected RPTicketStatus = "rejected"
	// RPTicketStatusDone is resource plan ticket status done.
	RPTicketStatusDone RPTicketStatus = "done"
	// RPTicketStatusFailed is resource plan ticket status failed.
	RPTicketStatusFailed RPTicketStatus = "failed"
	// RPTicketStatusRevoked is resource plan ticket status revoked.
	RPTicketStatusRevoked RPTicketStatus = "revoked"
)

// Validate RPTicketStatus.
func (s RPTicketStatus) Validate() error {
	switch s {
	case RPTicketStatusInit:
	case RPTicketStatusAuditing:
	case RPTicketStatusRejected:
	case RPTicketStatusDone:
	case RPTicketStatusFailed:
	case RPTicketStatusRevoked:
	default:
		return fmt.Errorf("unsupported resource plan status: %s", s)
	}

	return nil
}

// rdTicketStatusNameMap records RPTicketStatus's name.
var rdTicketStatusNameMap = map[RPTicketStatus]string{
	RPTicketStatusInit:     "待审批",
	RPTicketStatusAuditing: "审批中",
	RPTicketStatusRejected: "审批拒绝",
	RPTicketStatusDone:     "审批通过",
	RPTicketStatusFailed:   "审批失败",
	RPTicketStatusRevoked:  "已撤销",
}

// Name return RPTicketStatus's name.
func (s RPTicketStatus) Name() string {
	return rdTicketStatusNameMap[s]
}

// GetRPTicketStatusMembers get RPTicketStatus's members.
func GetRPTicketStatusMembers() []RPTicketStatus {
	return []RPTicketStatus{
		RPTicketStatusInit,
		RPTicketStatusAuditing,
		RPTicketStatusRejected,
		RPTicketStatusDone,
		RPTicketStatusFailed,
		RPTicketStatusRevoked,
	}
}

// DemandClass is resource plan demand class.
type DemandClass string

const (
	// DemandClassCVM is demand class cvm.
	DemandClassCVM DemandClass = "CVM"
	// DemandClassCA is demand class ca.
	DemandClassCA DemandClass = "CA"
)

// Validate DemandClass.
func (c DemandClass) Validate() error {
	switch c {
	case DemandClassCVM:
	case DemandClassCA:
	default:
		return fmt.Errorf("unsupported demand class: %s", c)
	}

	return nil
}

// GetDemandClassMembers get DemandClass's members.
func GetDemandClassMembers() []DemandClass {
	return []DemandClass{DemandClassCVM, DemandClassCA}
}

// DemandResType is resource plan demand resource type.
type DemandResType string

const (
	// DemandResTypeCVM is demand resource type cvm.
	DemandResTypeCVM DemandResType = "CVM"
	// DemandResTypeCBS is demand resource type cbs.
	DemandResTypeCBS DemandResType = "CBS"
)

// Validate DemandResType.
func (t DemandResType) Validate() error {
	switch t {
	case DemandResTypeCVM:
	case DemandResTypeCBS:
	default:
		return fmt.Errorf("unsupported demand resource type: %s", t)
	}

	return nil
}

// ResModeCode is resource plan res mode code.
type ResModeCode string

const (
	ResModeCodeByDeviceType   ResModeCode = "device_type"
	ResModeCodeByDeviceFamily ResModeCode = "device_family"
)

// Validate ResModeCode
func (r ResModeCode) Validate() error {
	switch r {
	case ResModeCodeByDeviceType:
	case ResModeCodeByDeviceFamily:
	default:
		return fmt.Errorf("unsupported res mode code: %s", r)
	}

	return nil
}

// ResMode is resource plan resource mode.
type ResMode string

const (
	// ResModeByDeviceType is resource mode of by device type.
	ResModeByDeviceType ResMode = "按机型"
	// ResModeByDeviceFamily is resource mode of by device family.
	ResModeByDeviceFamily ResMode = "按机型族"
)

// Validate ResMode.
func (r ResMode) Validate() error {
	switch r {
	case ResModeByDeviceType:
	case ResModeByDeviceFamily:
	default:
		return fmt.Errorf("unsupported res mode: %s", r)
	}

	return nil
}

// GetResModeMembers get ResMode's members.
func GetResModeMembers() []ResMode {
	return []ResMode{ResModeByDeviceType, ResModeByDeviceFamily}
}

var resModeNameMap = map[ResModeCode]ResMode{
	ResModeCodeByDeviceType:   ResModeByDeviceType,
	ResModeCodeByDeviceFamily: ResModeByDeviceFamily,
}

// Name get ResModeCode's name.
func (r ResModeCode) Name() ResMode {
	return resModeNameMap[r]
}

// Code get ResMode's code.
func (r ResMode) Code() (ResModeCode, error) {
	for code, n := range resModeNameMap {
		if n == r {
			return code, nil
		}
	}
	return "", fmt.Errorf("unsupported res mode: %s", r)
}

// ObsProject is obs project.
type ObsProject string

const (
	// ObsProjectNormal is obs normal project.
	ObsProjectNormal ObsProject = "常规项目"
	// ObsProjectReuse is obs project reuse project.
	ObsProjectReuse ObsProject = "改造复用"
	// ObsProjectMigrate is obs migrate project.
	ObsProjectMigrate ObsProject = "轻量云徙"
	// ObsProjectRollServer is obs roll server project.
	ObsProjectRollServer ObsProject = "滚服项目"
)

// GetObsProjectMembers get ObsProject's members.
func GetObsProjectMembers() []ObsProject {
	obsProjects := []ObsProject{ObsProjectNormal, ObsProjectReuse, ObsProjectMigrate, ObsProjectRollServer}
	obsProjects = append(obsProjects, getSpringObsProject(), getDissolveObsProject())

	return obsProjects
}

// GetObsProjectMembersForResPlan get ObsProject's members for resource plan.
func GetObsProjectMembersForResPlan() []ObsProject {
	obsProjects := []ObsProject{ObsProjectNormal, ObsProjectReuse, ObsProjectMigrate}
	obsProjects = append(obsProjects, getSpringObsProjectForResPlan()...)
	obsProjects = append(obsProjects, getDissolveObsProjectForResPlan()...)

	return obsProjects
}

// Validate ObsProject.
func (o ObsProject) Validate() error {
	obsProjects := GetObsProjectMembers()
	obsProjectMap := converter.SliceToMap(obsProjects, func(obj ObsProject) (ObsProject, struct{}) {
		return obj, struct{}{}
	})

	if _, ok := obsProjectMap[o]; !ok {
		return fmt.Errorf("unsupported obs project: %s", o)
	}

	return nil
}

// ValidateResPlan validate obs project used in resource plan.
func (o ObsProject) ValidateResPlan() error {
	obsProjects := GetObsProjectMembersForResPlan()
	obsProjectMap := converter.SliceToMap(obsProjects, func(obj ObsProject) (ObsProject, struct{}) {
		return obj, struct{}{}
	})

	if _, ok := obsProjectMap[o]; !ok {
		return fmt.Errorf("unsupported obs project: %s", o)
	}

	return nil
}

// getSpringObsProject get spring obs project.
func getSpringObsProject() ObsProject {
	// 春保窗口期：12月1日～次年3月15日
	// 12月1日～12月31日提单的春保项目前缀为次年
	year := time.Now().Local().Year()
	if time.Now().Month() == time.December {
		year += 1
	}

	prefixYear := strconv.Itoa(year)
	project := ObsProject(prefixYear + "春节保障")

	return project
}

// getDissolveObsProject get dissolve obs project.
func getDissolveObsProject() ObsProject {
	// 按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	year := time.Now().Local().Year()
	prefixYear := strconv.Itoa(year)
	project := ObsProject(prefixYear + "机房裁撤")

	return project
}

// getSpringObsProjectForResPlan get spring obs project for resource plan.
func getSpringObsProjectForResPlan() []ObsProject {
	projects := make([]ObsProject, 0)
	nowYear := time.Now().Year()
	// 春保窗口期：12月1日～次年3月25日

	// 因预测的提前性，1月1日～次年3月25日均允许提次年的春保项目预测单
	prefixYear := strconv.Itoa(nowYear + 1)
	projects = append(projects, ObsProject(prefixYear+"春节保障"))

	// 3月25日前允许申请当年的春保项目预测单
	ddl := time.Date(nowYear, time.March, 25, 0, 0, 0, 0, time.Local)
	if time.Now().Before(ddl) {
		prefixYear := strconv.Itoa(nowYear)
		projects = append(projects, ObsProject(prefixYear+"春节保障"))
	}

	return projects
}

// getDissolveObsProjectForResPlan get dissolve obs project for resource plan.
func getDissolveObsProjectForResPlan() []ObsProject {
	projects := make([]ObsProject, 0)

	// 按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	nowYear := time.Now().Local().Year()
	prefixYear := strconv.Itoa(nowYear)
	projects = append(projects, ObsProject(prefixYear+"机房裁撤"))
	nextPrefixYear := strconv.Itoa(nowYear + 1)
	projects = append(projects, ObsProject(nextPrefixYear+"机房裁撤"))

	return projects
}

// RequireType is resource apply require type.
type RequireType int64

const (
	// RequireTypeRegular 常规项目
	RequireTypeRegular RequireType = 1
	// RequireTypeSpring 春节保障
	RequireTypeSpring RequireType = 2
	// RequireTypeDissolve 机房裁撤
	RequireTypeDissolve RequireType = 3
	// RequireTypeRollServer 滚服项目
	RequireTypeRollServer RequireType = 6
	//	RequireTypeGreenChannel 小额绿通
	RequireTypeGreenChannel RequireType = 7
	// RequireTypeSpringResPool 春保资源池
	RequireTypeSpringResPool RequireType = 8
)

var requireTypeNameMap = map[RequireType]string{
	RequireTypeRegular:       "常规项目",
	RequireTypeSpring:        "春节保障",
	RequireTypeDissolve:      "机房裁撤",
	RequireTypeRollServer:    "滚服项目",
	RequireTypeGreenChannel:  "小额绿通",
	RequireTypeSpringResPool: "春保资源池",
}

// GetName get name of RequireType.
func (t RequireType) GetName() string {
	if name, ok := requireTypeNameMap[t]; ok {
		return name
	}

	return "Unknown"
}

// GetRequireTypeMembers get members of RequireType.
func GetRequireTypeMembers() []RequireType {
	return []RequireType{
		RequireTypeRegular,
		RequireTypeSpring,
		RequireTypeDissolve,
		RequireTypeRollServer,
		RequireTypeGreenChannel,
		RequireTypeSpringResPool,
	}
}

// Validate RequireType.
func (t RequireType) Validate() error {
	requireTypeMembers := GetRequireTypeMembers()
	requireTypeMemberMap := converter.SliceToMap(requireTypeMembers, func(member RequireType) (RequireType, struct{}) {
		return member, struct{}{}
	})
	if _, ok := requireTypeMemberMap[t]; !ok {
		return fmt.Errorf("unsupported require type: %d", t)
	}

	return nil
}

// NeedVerifyResPlan need verify resource plan.
func (t RequireType) NeedVerifyResPlan() bool {
	switch t {
	// 常规项目、春节保障、机房裁撤、春保资源池需要校验预测
	case RequireTypeRegular, RequireTypeSpring, RequireTypeDissolve, RequireTypeSpringResPool:
		return true
	default:
		return false
	}
}

// ToObsProject ObsProject.
func (t RequireType) ToObsProject() ObsProject {
	if obsProject, ok := requireTypeObsProjectMap[t]; ok {
		return obsProject
	}

	// 默认是常规项目
	return ObsProjectNormal
}

// IsNeedQuotaManage 是否在主机申请时使用额度管理
func (t RequireType) IsNeedQuotaManage() bool {
	if t == RequireTypeRollServer || t == RequireTypeSpringResPool {
		return true
	}

	return false
}

// IsUseManageBizPlan 是否使用管理业务的运营产品去申请主机，扣减预测
func (t RequireType) IsUseManageBizPlan() bool {
	if t == RequireTypeRollServer || t == RequireTypeSpringResPool {
		return true
	}

	return false
}

// ToRequireTypeWhenGetDevice 查询机型时使用的需求类型
func (t RequireType) ToRequireTypeWhenGetDevice() RequireType {
	requireTypeMap := map[RequireType]RequireType{
		RequireTypeRegular:    RequireTypeRegular,
		RequireTypeRollServer: RequireTypeRollServer,
		// "小额绿通"使用"常规项目"类型查询机型
		RequireTypeGreenChannel: RequireTypeRegular,
		RequireTypeSpring:       RequireTypeSpring,
		RequireTypeDissolve:     RequireTypeDissolve,
		// "春保资源池"使用"常规项目"类型查询机型
		RequireTypeSpringResPool: RequireTypeRegular,
	}

	requireType, ok := requireTypeMap[t]
	if !ok {
		return t
	}
	return requireType
}

var requireTypeObsProjectMap = map[RequireType]ObsProject{
	RequireTypeRegular:    ObsProjectNormal,
	RequireTypeRollServer: ObsProjectRollServer,
	// "小额绿通"使用"常规项目"的 obs project
	RequireTypeGreenChannel: ObsProjectNormal,
	RequireTypeSpring:       getSpringObsProject(),
	RequireTypeDissolve:     getDissolveObsProject(),
	// "春保资源池"使用"常规项目"的 obs project
	RequireTypeSpringResPool: ObsProjectNormal,
}

// DemandSource is demand source.
// TODO this enum will be changed to get from obs api.
type DemandSource string

const (
	// DemandSourceIndChg is demand source indicator changes.
	DemandSourceIndChg DemandSource = "指标变化"
	// DemandSourceArchAdj is demand source architecture adjustment.
	DemandSourceArchAdj DemandSource = "架构调整"
	// DemandSourceOpt is demand source cost optimization and increased usage rate.
	DemandSourceOpt DemandSource = "成本优化&利用率提升"
	// DemandSourceBufferAdj is demand source buffer adjustment.
	DemandSourceBufferAdj DemandSource = "Buffer调整"
	// DemandSourceForced is demand source forced.
	DemandSourceForced DemandSource = "不可抗力(政策/合规)"
	// DemandSourceSupply is demand source supply issues.
	DemandSourceSupply DemandSource = "供应问题"
	// DemandSourceSysProcess is demand source system process issues.
	DemandSourceSysProcess DemandSource = "系统流程问题"
)

// Validate DemandSource.
func (d DemandSource) Validate() error {
	switch d {
	case DemandSourceIndChg:
	case DemandSourceArchAdj:
	case DemandSourceOpt:
	case DemandSourceBufferAdj:
	case DemandSourceForced:
	case DemandSourceSupply:
	case DemandSourceSysProcess:
	default:
		return fmt.Errorf("unsupported demand source: %s", d)
	}

	return nil
}

// GetDemandSourceMembers get DemandSource's members.
func GetDemandSourceMembers() []DemandSource {
	return []DemandSource{
		DemandSourceIndChg,
		DemandSourceArchAdj,
		DemandSourceOpt,
		DemandSourceBufferAdj,
		DemandSourceForced,
		DemandSourceSupply,
		DemandSourceSysProcess,
	}
}

// CrpDemandLockStatus is resource plan crp demand lock status.
type CrpDemandLockStatus int8

const (
	// CrpDemandUnLocked is resource plan crp demand unlocked.
	CrpDemandUnLocked CrpDemandLockStatus = 0
	// CrpDemandLocked is resource plan crp demand locked.
	CrpDemandLocked CrpDemandLockStatus = 1
)

// Validate CrpDemandLockStatus.
func (s CrpDemandLockStatus) Validate() error {
	switch s {
	case CrpDemandUnLocked:
	case CrpDemandLocked:
	default:
		return fmt.Errorf("unsupported crp demand lock status: %d", s)
	}

	return nil
}

// RPDemandAdjustType is resource plan demand adjust type.
type RPDemandAdjustType string

const (
	// RPDemandAdjustTypeUpdate is resource plan demand adjust type update.
	RPDemandAdjustTypeUpdate RPDemandAdjustType = "update"
	// RPDemandAdjustTypeDelay is resource plan demand adjust type delay.
	RPDemandAdjustTypeDelay RPDemandAdjustType = "delay"
)

// Validate RPDemandAdjustType.
func (t RPDemandAdjustType) Validate() error {
	switch t {
	case RPDemandAdjustTypeUpdate:
	case RPDemandAdjustTypeDelay:
	default:
		return fmt.Errorf("unsupported resource plan demand adjust type: %s", t)
	}

	return nil
}

// CrpAdjustType crp adjust type.
type CrpAdjustType string

const (
	// CrpAdjustTypeUpdate is crp adjust type update.
	CrpAdjustTypeUpdate CrpAdjustType = "常规修改"
	// CrpAdjustTypeDelay is crp adjust type delay.
	CrpAdjustTypeDelay CrpAdjustType = "加急延期"
	// CrpAdjustTypeCancel is crp adjust type cancel.
	CrpAdjustTypeCancel CrpAdjustType = "需求取消"
)

// DemandStatus is resource plan demand status.
type DemandStatus string

const (
	// DemandStatusCanApply 预测需求可申领.
	DemandStatusCanApply DemandStatus = "can_apply"
	// DemandStatusNotReady 预测需求未到申领时间.
	DemandStatusNotReady DemandStatus = "not_ready"
	// DemandStatusExpired 预测需求已过期.
	DemandStatusExpired DemandStatus = "expired"
	// DemandStatusSpentAll 预测需求已耗尽.
	DemandStatusSpentAll DemandStatus = "spent_all"
	// DemandStatusLocked 预测需求变更中.
	DemandStatusLocked DemandStatus = "locked"
)

// Validate DemandStatus.
func (d DemandStatus) Validate() error {
	switch d {
	case DemandStatusCanApply:
	case DemandStatusNotReady:
	case DemandStatusExpired:
	case DemandStatusSpentAll:
	case DemandStatusLocked:
	default:
		return fmt.Errorf("unsupported demand status: %s", d)
	}

	return nil
}

var demandStatusNameMaps = map[DemandStatus]string{
	DemandStatusCanApply: "可申领",
	DemandStatusNotReady: "未到申领时间",
	DemandStatusExpired:  "已过期",
	DemandStatusSpentAll: "已耗尽",
	DemandStatusLocked:   "变更中",
}

// Name return the name of DemandStatus.
func (d DemandStatus) Name() string {
	return demandStatusNameMaps[d]
}

// PlanTypeCode is resource plan type code.
type PlanTypeCode string

const (
	// PlanTypeCodeInPlan is in plan.
	PlanTypeCodeInPlan PlanTypeCode = "in_plan"
	// PlanTypeCodeOutPlan is out plan.
	PlanTypeCodeOutPlan PlanTypeCode = "out_plan"
)

// Validate PlanTypeCode.
func (p PlanTypeCode) Validate() error {
	switch p {
	case PlanTypeCodeInPlan:
	case PlanTypeCodeOutPlan:
	default:
		return fmt.Errorf("unsupported plan type code: %s", p)
	}

	return nil
}

// GetPlanTypeCodeHcmMembers get hcm PlanTypeCode's members.
func GetPlanTypeCodeHcmMembers() []PlanTypeCode {
	return []PlanTypeCode{PlanTypeCodeInPlan, PlanTypeCodeOutPlan}
}

// PlanTypeMaps is plan type maps.
var PlanTypeMaps = map[PlanTypeCode]PlanType{
	PlanTypeCodeInPlan:  PlanTypeHcmInPlan,
	PlanTypeCodeOutPlan: PlanTypeHcmOutPlan,
}

// Name return the name of PlanTypeCode.
func (p PlanTypeCode) Name() PlanType {
	return PlanTypeMaps[p]
}

// InPlan return true if the plan type is in plan.
func (p PlanTypeCode) InPlan() bool {
	switch p {
	case PlanTypeCodeInPlan:
		return true
	case PlanTypeCodeOutPlan:
		return false
	default:
		return false
	}
}

// PlanType is resource plan type.
type PlanType string

const (
	// PlanTypeCrpInPlan is crp in plan.
	PlanTypeCrpInPlan PlanType = "计划内"
	// PlanTypeCrpOutPlan is crp out plan.
	PlanTypeCrpOutPlan PlanType = "计划外"
	// PlanTypeHcmInPlan is hcm in plan.
	PlanTypeHcmInPlan PlanType = "预测内"
	// PlanTypeHcmOutPlan is hcm out plan.
	PlanTypeHcmOutPlan PlanType = "预测外"
)

// Validate PlanType.
func (p PlanType) Validate() error {
	switch p {
	case PlanTypeCrpInPlan:
	case PlanTypeCrpOutPlan:
	case PlanTypeHcmInPlan:
	case PlanTypeHcmOutPlan:
	default:
		return fmt.Errorf("unsupported plan type: %s", p)
	}

	return nil
}

// PlanTypeNameMaps is plan type name maps.
var PlanTypeNameMaps = map[PlanType]PlanTypeCode{
	PlanTypeCrpInPlan:  PlanTypeCodeInPlan,
	PlanTypeCrpOutPlan: PlanTypeCodeOutPlan,
	PlanTypeHcmInPlan:  PlanTypeCodeInPlan,
	PlanTypeHcmOutPlan: PlanTypeCodeOutPlan,
}

// GetCode get plan type code.
func (p PlanType) GetCode() PlanTypeCode {
	return PlanTypeNameMaps[p]
}

// GetPlanTypeHcmMembers get hcm PlanType's members.
func GetPlanTypeHcmMembers() []PlanType {
	return []PlanType{PlanTypeHcmInPlan, PlanTypeHcmOutPlan}
}

// ToAnotherPlanType the plan type of crp to the plan type of hcm, or vice versa.
// TODO: 这个方法的功能是不明确的，使用者难以确定什么情况下使用该方法
func (p PlanType) ToAnotherPlanType() PlanType {
	switch p {
	case PlanTypeCrpInPlan:
		return PlanTypeHcmInPlan
	case PlanTypeCrpOutPlan:
		return PlanTypeHcmOutPlan
	case PlanTypeHcmInPlan:
		return PlanTypeCrpInPlan
	case PlanTypeHcmOutPlan:
		return PlanTypeCrpOutPlan
	default:
		return p
	}
}

// InPlan return the plan type in plan or not.
func (p PlanType) InPlan() bool {
	switch p {
	case PlanTypeCrpInPlan, PlanTypeHcmInPlan:
		return true
	case PlanTypeCrpOutPlan, PlanTypeHcmOutPlan:
		return false
	default:
		return false
	}
}

// GetCrpConsumeResPlanSourceTypes get crp system source types of consuming resource plan.
func GetCrpConsumeResPlanSourceTypes() []string {
	return []string{"申领划扣", "申领划扣退回"}
}

// VerifyResPlanRst is verify resource plan result.
type VerifyResPlanRst string

const (
	// VerifyResPlanRstPass is resource plan result pass.
	VerifyResPlanRstPass VerifyResPlanRst = "PASS"
	// VerifyResPlanRstFailed is resource plan result failed.
	VerifyResPlanRstFailed VerifyResPlanRst = "FAILED"
	// VerifyResPlanRstNotInvolved is resource plan result not involved.
	VerifyResPlanRstNotInvolved VerifyResPlanRst = "NOT_INVOLVED"
)

// Validate VerifyResPlanRst.
func (r VerifyResPlanRst) Validate() error {
	switch r {
	case VerifyResPlanRstPass:
	case VerifyResPlanRstFailed:
	case VerifyResPlanRstNotInvolved:
	default:
		return fmt.Errorf("unsupported verify resource plan result: %s", r)
	}

	return nil
}

// CrpOrderStatus is crp order status.
type CrpOrderStatus int

const (
	// CrpOrderStatusDeptApprove 部门管理员审批
	CrpOrderStatusDeptApprove CrpOrderStatus = 0
	// CrpOrderStatusDirectorApprove 业务总监审批
	CrpOrderStatusDirectorApprove CrpOrderStatus = 1
	// CrpOrderStatusPlanApprove 规划经理审批
	CrpOrderStatusPlanApprove CrpOrderStatus = 2
	// CrpOrderStatusResourceApprove 资源经理审批
	CrpOrderStatusResourceApprove CrpOrderStatus = 3
	// CrpOrderStatusCloudApprove 等待云上审批
	CrpOrderStatusCloudApprove CrpOrderStatus = 14
	// CrpOrderStatusWaitDeliver 等待交付
	CrpOrderStatusWaitDeliver CrpOrderStatus = 4
	// CrpOrderStatusDelivering 交付队列中
	CrpOrderStatusDelivering CrpOrderStatus = 5
	// CrpOrderStatusResource 资源准备中
	CrpOrderStatusResource CrpOrderStatus = 6
	// CrpOrderStatusCvm CVM 生成中
	CrpOrderStatusCvm CrpOrderStatus = 7
	// CrpOrderStatusFinish 执行完成
	CrpOrderStatusFinish CrpOrderStatus = 8
	// CrpOrderStatusReject 驳回
	CrpOrderStatusReject CrpOrderStatus = 127
	// CrpOrderStatusFailed CVM创建失败
	CrpOrderStatusFailed CrpOrderStatus = 129
)

// CrpOrderStatusCanRevoke 目前 crp 只支持单据处于下列状态时可以发起撤单
var CrpOrderStatusCanRevoke = []CrpOrderStatus{
	CrpOrderStatusDeptApprove,
	CrpOrderStatusDirectorApprove,
	CrpOrderStatusPlanApprove,
	CrpOrderStatusResourceApprove,
	CrpOrderStatusWaitDeliver,
	CrpOrderStatusDelivering,
}

// CoreType 核心类型
type CoreType string

const (
	// CoreTypeBig 大核心
	CoreTypeBig CoreType = "大核心"
	// CoreTypeMedium 中核心
	CoreTypeMedium CoreType = "中核心"
	// CoreTypeSmall 小核心
	CoreTypeSmall CoreType = "小核心"
)

// Validate CoreType.
func (r CoreType) Validate() error {
	switch r {
	case CoreTypeBig:
	case CoreTypeMedium:
	case CoreTypeSmall:
	default:
		return fmt.Errorf("unsupported verify core type result: %s", r)
	}

	return nil
}

// ItsmServiceNameApply ITSM中的资源申请流程在 HCM 中的名称（此处名称并非与 ITSM 中的流程名称完全一致）
const ItsmServiceNameApply = "资源申领流程"

// DemandPenaltyBaseSource is demand penalty base source.
type DemandPenaltyBaseSource string

const (
	DemandPenaltyBaseSourceLocal DemandPenaltyBaseSource = "local"
	DemandPenaltyBaseSourceCrp   DemandPenaltyBaseSource = "crp"
)

// Validate DemandPenaltyBaseSource.
func (d DemandPenaltyBaseSource) Validate() error {
	switch d {
	case DemandPenaltyBaseSourceLocal:
	case DemandPenaltyBaseSourceCrp:
	default:
		return fmt.Errorf("unsupported demand penalty base source: %s", d)
	}

	return nil
}

type DemandChangelogType string

const (
	DemandChangelogTypeAppend DemandChangelogType = "append"
	DemandChangelogTypeAdjust DemandChangelogType = "adjust"
	DemandChangelogTypeDelete DemandChangelogType = "delete"
	DemandChangelogTypeExpend DemandChangelogType = "expend"
)

// Validate DemandChangelogType.
func (d DemandChangelogType) Validate() error {
	switch d {
	case DemandChangelogTypeAppend:
	case DemandChangelogTypeAdjust:
	case DemandChangelogTypeDelete:
	case DemandChangelogTypeExpend:
	default:
		return fmt.Errorf("unsupported demand changelog type: %s", d)
	}

	return nil
}

var demandChangelogTypeNameMap = map[DemandChangelogType]string{
	DemandChangelogTypeAppend: "追加",
	DemandChangelogTypeAdjust: "调整",
	DemandChangelogTypeDelete: "删除",
	DemandChangelogTypeExpend: "消耗",
}

// Name get demand changelog type name.
func (d DemandChangelogType) Name() string {
	return demandChangelogTypeNameMap[d]
}

// ResPlanWeekHolidayStatus is resource plan week holiday status.
type ResPlanWeekHolidayStatus int8

const (
	// ResPlanWeekIsHoliday resource plan week is holiday.
	ResPlanWeekIsHoliday ResPlanWeekHolidayStatus = 0
	// ResPlanWeekIsNotHoliday resource plan week is not holiday.
	ResPlanWeekIsNotHoliday ResPlanWeekHolidayStatus = 1
)

// Validate ResPlanWeekHolidayStatus.
func (s ResPlanWeekHolidayStatus) Validate() error {
	switch s {
	case ResPlanWeekIsHoliday:
	case ResPlanWeekIsNotHoliday:
	default:
		return fmt.Errorf("unsupported res plan week holiday status: %d", s)
	}

	return nil
}

const (
	// ResourcePoolBiz 资源池业务
	ResourcePoolBiz = 931
)

// AbolishPhase 裁撤阶段
type AbolishPhase string

const (
	// Incomplete 裁撤未完成
	Incomplete AbolishPhase = "incomplete"
	// Complete 裁撤完成
	Complete AbolishPhase = "complete"
	// BsiComplete 业务退回
	BsiComplete AbolishPhase = "bsiComplete"
	// Retain 保留暂不裁撤
	Retain AbolishPhase = "retain"
)

// XrayFaultTicketIsEnd xray故障单是否结单
type XrayFaultTicketIsEnd int

const (
	XrayFaultTicketNotEnd XrayFaultTicketIsEnd = 0
	XrayFaultTicketHasEnd XrayFaultTicketIsEnd = 1
)
