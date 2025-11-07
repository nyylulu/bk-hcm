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

// Package classifier implements device classifier which helps to
// divide recycle order into suborders.
package classifier

import (
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
)

// RecycleGrpType recycle hosts group type
type RecycleGrpType int

// definition of various recycle host group type
const (
	RecycleGrpCVMPriRegularImmediate     RecycleGrpType = iota
	RecycleGrpCVMPriRegularDelay         RecycleGrpType = iota
	RecycleGrpCVMPriDissolveImmediate    RecycleGrpType = iota
	RecycleGrpCVMPriDissolveDelay        RecycleGrpType = iota
	RecycleGrpCVMPriSpringImmediate      RecycleGrpType = iota
	RecycleGrpCVMPriSpringDelay          RecycleGrpType = iota
	RecycleGrpCVMPriRentImmediate        RecycleGrpType = iota
	RecycleGrpCVMPriRentDelay            RecycleGrpType = iota
	RecycleGrpCVMPubRegularImmediate     RecycleGrpType = iota
	RecycleGrpCVMPubRegularDelay         RecycleGrpType = iota
	RecycleGrpCVMPubDissolveImmediate    RecycleGrpType = iota
	RecycleGrpCVMPubDissolveDelay        RecycleGrpType = iota
	RecycleGrpCVMPubSpringImmediate      RecycleGrpType = iota
	RecycleGrpCVMPubSpringDelay          RecycleGrpType = iota
	RecycleGrpCVMPubRentImmediate        RecycleGrpType = iota
	RecycleGrpCVMPubRentDelay            RecycleGrpType = iota
	RecycleGrpPMRegularImmediate         RecycleGrpType = iota
	RecycleGrpPMRegularDelay             RecycleGrpType = iota
	RecycleGrpPMDissolveImmediate        RecycleGrpType = iota
	RecycleGrpPMDissolveDelay            RecycleGrpType = iota
	RecycleGrpPMExpiredImmediate         RecycleGrpType = iota
	RecycleGrpPMExpiredDelay             RecycleGrpType = iota
	RecycleGrpOthers                     RecycleGrpType = iota
	RecycleGrpCVMPubRollServerImmediate  RecycleGrpType = iota
	RecycleGrpCVMPubRollServerDelay      RecycleGrpType = iota
	RecycleGrpCVMPriRollServerImmediate  RecycleGrpType = iota
	RecycleGrpCVMPriRollServerDelay      RecycleGrpType = iota
	RecycleGrpCVMPubShortRentalImmediate RecycleGrpType = iota
	RecycleGrpCVMPubShortRentalDelay     RecycleGrpType = iota
	RecycleGrpCVMPriShortRentalImmediate RecycleGrpType = iota
	RecycleGrpCVMPriShortRentalDelay     RecycleGrpType = iota
)

// IsRollingServerType check if the recycle group type is rolling server type
func (r RecycleGrpType) IsRollingServerType() bool {
	switch r {
	case RecycleGrpCVMPubRollServerImmediate, RecycleGrpCVMPubRollServerDelay,
		RecycleGrpCVMPriRollServerImmediate, RecycleGrpCVMPriRollServerDelay:
		return true
	default:
		return false
	}
}

// IsShortRentalType check if the recycle group type is short rental type
func (r RecycleGrpType) IsShortRentalType() bool {
	switch r {
	case RecycleGrpCVMPubShortRentalImmediate, RecycleGrpCVMPubShortRentalDelay,
		RecycleGrpCVMPriShortRentalImmediate, RecycleGrpCVMPriShortRentalDelay:
		return true
	default:
		return false
	}
}

// RecycleGroupProperty recycle strategy properties of recycle group
type RecycleGroupProperty struct {
	ResourceType  table.ResourceType
	RecycleType   table.RecycleType
	ReturnType    table.RetPlanType
	Pool          table.PoolType
	CostConcerned bool
}

// MapGroupProperty map of recycle group type to recycle group property
var MapGroupProperty = map[RecycleGrpType]RecycleGroupProperty{
	RecycleGrpCVMPriRegularImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPrivate,
		CostConcerned: true,
	},
	RecycleGrpCVMPriRegularDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPrivate,
		CostConcerned: true,
	},
	RecycleGrpCVMPriDissolveImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPriDissolveDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPriSpringImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeSpring,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPriSpringDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeSpring,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPubRegularImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPublic,
		CostConcerned: true,
	},
	RecycleGrpCVMPubRegularDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPublic,
		CostConcerned: true,
	},
	RecycleGrpCVMPubDissolveImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPubDissolveDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPubSpringImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeSpring,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPubSpringDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeSpring,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpPMRegularImmediate: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanImmediate,
		CostConcerned: true,
	},
	RecycleGrpPMRegularDelay: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeRegular,
		ReturnType:    table.RetPlanDelay,
		CostConcerned: true,
	},
	RecycleGrpPMDissolveImmediate: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanImmediate,
		CostConcerned: false,
	},
	RecycleGrpPMDissolveDelay: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeDissolve,
		ReturnType:    table.RetPlanDelay,
		CostConcerned: false,
	},
	RecycleGrpPMExpiredImmediate: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeExpired,
		ReturnType:    table.RetPlanImmediate,
		CostConcerned: false,
	},
	RecycleGrpPMExpiredDelay: {
		ResourceType:  table.ResourceTypePm,
		RecycleType:   table.RecycleTypeExpired,
		ReturnType:    table.RetPlanDelay,
		CostConcerned: false,
	},
	RecycleGrpOthers: {
		ResourceType:  table.ResourceTypeOthers,
		CostConcerned: false,
	},
	RecycleGrpCVMPubRollServerImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRollServer,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPubRollServerDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRollServer,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPriRollServerImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRollServer,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPriRollServerDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeRollServer,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPubShortRentalImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeShortRental,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPubShortRentalDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeShortRental,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPublic,
		CostConcerned: false,
	},
	RecycleGrpCVMPriShortRentalImmediate: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeShortRental,
		ReturnType:    table.RetPlanImmediate,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
	RecycleGrpCVMPriShortRentalDelay: {
		ResourceType:  table.ResourceTypeCvm,
		RecycleType:   table.RecycleTypeShortRental,
		ReturnType:    table.RetPlanDelay,
		Pool:          table.PoolPrivate,
		CostConcerned: false,
	},
}

// RecycleGroup recycle group of hosts with the same resource type, recycle type and return plan
type RecycleGroup map[RecycleGrpType][]*table.RecycleHost

// ClassifyRecycleGroups classify hosts into groups with different recycle strategies
func ClassifyRecycleGroups(bkBizHostsMap map[int64][]*table.RecycleHost, plan *types.ReturnPlan) (
	map[int64]RecycleGroup, error) {

	groups := make(map[int64]RecycleGroup)
	for bkBizID, hosts := range bkBizHostsMap {
		groups[bkBizID] = RecycleGroup{}
		for _, host := range hosts {
			grpType := getRecycleGrpType(host, plan)
			if _, ok := groups[bkBizID][grpType]; !ok {
				groups[bkBizID][grpType] = make([]*table.RecycleHost, 0)
			}
			groups[bkBizID][grpType] = append(groups[bkBizID][grpType], host)
		}
	}

	return groups, nil
}

// FillClassifyInfo fill recycle host with classification info
func FillClassifyInfo(hosts []*table.RecycleHost, plan *types.ReturnPlan) {

	for _, host := range hosts {
		// fill resource type
		resType := getResType(host)
		host.ResourceType = resType

		// fill return plan
		switch resType {
		case table.ResourceTypeCvm:
			host.ReturnPlan = plan.CvmPlan
		case table.ResourceTypePm:
			host.ReturnPlan = plan.PmPlan
		}
	}
}

// getResType get host resource type
func getResType(host *table.RecycleHost) table.ResourceType {
	if IsUnsupportedDevice(host.AssetID, host.IP) {
		return table.ResourceTypeUnsupported
	}

	if isSpecialCvm(host.DeviceType) {
		return table.ResourceTypeOthers
	}

	if IsQcloudCvm(host.AssetID) {
		return table.ResourceTypeCvm
	}

	if isIdcPm(host.AssetID) {
		return table.ResourceTypePm
	}

	return table.ResourceTypeOthers
}

// IsUnsupportedDevice verify if given host is unsupported device
// 固资号非 TC、TYSV或TDKIEG 开头的均为不支持资源类型
// devcloud机器：a) ip开头为"9.134"或"9.135" 或 b) 固资号开头为"TCDEV"
func IsUnsupportedDevice(assetId, ip string) bool {
	// devcloud机器
	if strings.HasPrefix(ip, "9.134") || strings.HasPrefix(ip, "9.135") || strings.HasPrefix(assetId, "TCDEV") {
		return true
	}

	// cvm
	if strings.HasPrefix(assetId, "TC") {
		return false
	}

	// idc physical machine
	if strings.HasPrefix(assetId, "TYSV") {
		return false
	}

	// 算力特殊机型
	if strings.HasPrefix(assetId, "TDKIEG") {
		return false
	}

	return true
}

// IsQcloudCvm verify if given host is qcloud cvm device
// 固资号为 TC*** （排除掉 TC***-VM****) 的是CVM机型
func IsQcloudCvm(assetId string) bool {
	if !strings.HasPrefix(assetId, "TC") {
		return false
	}

	dashIdx := strings.Index(assetId, "-")
	if dashIdx < 0 {
		return true
	}

	// exclude qcloud docker vm
	if strings.HasPrefix(assetId[dashIdx+1:], "VM") {
		return false
	}

	return true
}

// isIdcPm verify if given host is idc physical machine
// 固资号为 TYSV*** （排除掉 TYSV***-VM****) 的是物理机机型
func isIdcPm(assetId string) bool {
	if !strings.HasPrefix(assetId, "TYSV") {
		return false
	}

	dashIdx := strings.Index(assetId, "-")
	if dashIdx < 0 {
		return true
	}

	// exclude idc docker vm and kvm
	if strings.HasPrefix(assetId[dashIdx+1:], "VM") {
		return false
	}

	return true
}

// isSpecialCvm verify if given host is cvm special case which skip return step
func isSpecialCvm(deviceType string) bool {
	if deviceType == "IT5nt.21XLARGE208" ||
		deviceType == "S5nt.21XLARGE206" ||
		deviceType == "S3ne.17XLARGE210" {
		return true
	}

	return false
}

// GetFixedRecycleType get host recycle type if it's a fixed type, otherwise return table.RecycleTypeRegular
func GetFixedRecycleType(host *table.RecycleHost, isDissolveHost bool) table.RecycleType {
	// 机房裁撤
	if isDissolveHost {
		return table.RecycleTypeDissolve
	}

	// 过保裁撤
	if isExpiredPm(host) {
		return table.RecycleTypeExpired
	}

	// 春节保障
	if isSpringCvm(host) {
		return table.RecycleTypeSpring
	}

	return table.RecycleTypeRegular
}

// isExpiredPm verify if given host is expired physical machine
func isExpiredPm(host *table.RecycleHost) bool {
	if host.ResourceType != table.ResourceTypePm {
		return false
	}

	if len(host.InputTime) > 0 {
		// input time in format like 2017-12-07T00:00:00+08:00
		inTime, err := time.Parse("2006-01-02T15:04:05+08:00", host.InputTime)
		if err != nil {
			logs.Warnf("failed to parse input time %s", host.InputTime)
			return false
		}
		// 4 year expiration
		expiredTime := inTime.AddDate(4, 0, 0)
		now := time.Now()
		if now.Year() > expiredTime.Year() || (now.Year() == expiredTime.Year() && now.Month() > expiredTime.Month()) {
			return true
		}
	}

	return false
}

// isSpringCvm verify if given host is cvm with obs project "春节保障"
func isSpringCvm(host *table.RecycleHost) bool {
	if host.ResourceType != table.ResourceTypeCvm {
		return false
	}

	if !strings.Contains(host.ObsProject, string(table.RecycleTypeSpring)) {
		return false
	}

	// 非春保窗口期发起的资源回收单据，按常规项目分类
	if !isSpringWindow() {
		return false
	}

	projName := getSpringObsProject()
	if host.ObsProject == projName {
		return true
	}

	return false
}

func getSpringObsProject() string {
	// 资源回收的春保窗口期：12月1日～次年4月20日
	// 12月1日～12月31日提单的春保项目前缀为次年
	year := time.Now().Local().Year()
	if time.Now().Month() == time.December {
		year += 1
	}

	prefixYear := strconv.Itoa(year)
	project := prefixYear + string(table.RecycleTypeSpring)

	return project
}

func isSpringWindow() bool {
	now := time.Now().Local()

	year := now.Year()
	if now.Month() == time.December {
		year += 1
	}

	// 资源回收的春保窗口期：12月1日～次年4月20日
	start := time.Date(year-1, 12, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, 4, 21, 0, 0, 0, 0, now.Location())
	if !now.Before(start) && now.Before(end) {
		return true
	}

	return false
}

// getRecycleGrpType get host recycle group type
func getRecycleGrpType(host *table.RecycleHost, plan *types.ReturnPlan) RecycleGrpType {
	switch host.ResourceType {
	case table.ResourceTypeCvm:
		if host.Pool == table.PoolPublic {
			return cvmPoolPublic(host, plan)
		} else {
			return cvmPoolPrivate(host, plan)
		}
	case table.ResourceTypePm:
		return idcPm(host, plan)
	case table.ResourceTypeOthers:
		return RecycleGrpOthers
	}

	return RecycleGrpOthers
}

func cvmPoolPublic(host *table.RecycleHost, plan *types.ReturnPlan) RecycleGrpType {
	if host.RecycleType == table.RecycleTypeRegular {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPubRegularImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPubRegularDelay
		}
	}
	if host.RecycleType == table.RecycleTypeDissolve {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPubDissolveImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPubDissolveDelay
		}
	}
	if host.RecycleType == table.RecycleTypeSpring {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPubSpringImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPubSpringDelay
		}
	}
	if host.RecycleType == table.RecycleTypeRollServer { // 滚服项目
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPubRollServerImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPubRollServerDelay
		}
	}
	if host.RecycleType == table.RecycleTypeShortRental { // 短租项目
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPubShortRentalImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPubShortRentalDelay
		}
	}
	return RecycleGrpOthers
}

func cvmPoolPrivate(host *table.RecycleHost, plan *types.ReturnPlan) RecycleGrpType {
	if host.RecycleType == table.RecycleTypeRegular {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPriRegularImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPriRegularDelay
		}
	}
	if host.RecycleType == table.RecycleTypeDissolve {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPriDissolveImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPriDissolveDelay
		}
	}
	if host.RecycleType == table.RecycleTypeSpring {
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPriSpringImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPriSpringDelay
		}
	}
	if host.RecycleType == table.RecycleTypeRollServer { // 滚服项目
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPriRollServerImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPriRollServerDelay
		}
	}
	if host.RecycleType == table.RecycleTypeShortRental { // 短租项目
		if plan.CvmPlan == table.RetPlanImmediate {
			return RecycleGrpCVMPriShortRentalImmediate
		}
		if plan.CvmPlan == table.RetPlanDelay {
			return RecycleGrpCVMPriShortRentalDelay
		}
	}
	return RecycleGrpOthers
}

func idcPm(host *table.RecycleHost, plan *types.ReturnPlan) RecycleGrpType {
	if host.RecycleType == table.RecycleTypeRegular {
		if plan.PmPlan == table.RetPlanImmediate {
			return RecycleGrpPMRegularImmediate
		}
		if plan.PmPlan == table.RetPlanDelay {
			return RecycleGrpPMRegularDelay
		}
	}
	if host.RecycleType == table.RecycleTypeDissolve {
		if plan.PmPlan == table.RetPlanImmediate {
			return RecycleGrpPMDissolveImmediate
		}
		if plan.PmPlan == table.RetPlanDelay {
			return RecycleGrpPMDissolveDelay
		}
	}
	if host.RecycleType == table.RecycleTypeExpired {
		if plan.PmPlan == table.RetPlanImmediate {
			return RecycleGrpPMExpiredImmediate
		}
		if plan.PmPlan == table.RetPlanDelay {
			return RecycleGrpPMExpiredDelay
		}
	}
	return RecycleGrpOthers
}
