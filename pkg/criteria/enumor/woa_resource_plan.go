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

import "fmt"

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
)

// Validate RPTicketStatus.
func (s RPTicketStatus) Validate() error {
	switch s {
	case RPTicketStatusInit:
	case RPTicketStatusAuditing:
	case RPTicketStatusRejected:
	case RPTicketStatusDone:
	case RPTicketStatusFailed:
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
}

// Name return RPTicketStatus's name.
func (s RPTicketStatus) Name() string {
	return rdTicketStatusNameMap[s]
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

// ObsProject is obs project.
// TODO this enum will be changed to get from obs api.
type ObsProject string

const (
	// ObsProjectNormal is obs normal project.
	ObsProjectNormal ObsProject = "常规项目"
	// ObsProjectReuse is obs project reuse project.
	ObsProjectReuse ObsProject = "改造复用"
	// ObsProjectCNY is obs project normal Chinese New Year project.
	ObsProjectCNY ObsProject = "2025春节保障"
	// ObsProjectDissolve is obs dissolve project.
	ObsProjectDissolve ObsProject = "2024机房裁撤"
	// ObsProjectMigrate is obs migrate project.
	ObsProjectMigrate ObsProject = "轻量云徙"
)

// Validate ObsProject.
func (o ObsProject) Validate() error {
	switch o {
	case ObsProjectNormal:
	case ObsProjectReuse:
	case ObsProjectCNY:
	case ObsProjectDissolve:
	case ObsProjectMigrate:
	default:
		return fmt.Errorf("unsupported obs project: %s", o)
	}

	return nil
}

// GetObsProjectMembers get ObsProject's members.
func GetObsProjectMembers() []ObsProject {
	return []ObsProject{ObsProjectNormal, ObsProjectReuse, ObsProjectCNY, ObsProjectDissolve, ObsProjectMigrate}
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
		return fmt.Errorf("unsupported obs project: %s", d)
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
