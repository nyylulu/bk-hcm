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

// Package algorithm ...
package algorithm

import (
	"regexp"
	"strings"

	"hcm/cmd/woa-server/thirdparty/dvmapi"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
)

// VMFitHostVirtualRatio check whether host match virtual ratio condition
func VMFitHostVirtualRatio(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error) {
	if host.ScheduledVMs >= 3 {
		logs.Infof("host %s match virtual ratio failed, current virtual ratio: %d", host.IP, host.ScheduledVMs)
		return false, nil
	}

	return true, nil
}

// VMFitRegion check whether host match region condition
func VMFitRegion(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error) {
	if selector.Region == "" {
		return true, nil
	}
	if !strings.HasPrefix(host.SZone, selector.Region+"-") {
		logs.Infof("host %s match region (%v) failed", host.IP, host.SZone)
		return false, nil
	}
	return true, nil
}

// VMFitCampus check whether host match campus condition
func VMFitCampus(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error) {
	if selector.Zone == "" {
		return true, nil
	}
	if selector.Zone != host.SZone {
		logs.Infof("host %s match zone (%v) failed", host.IP, host.SZone)
		return false, nil
	}
	return true, nil
}

// VMFitKernel check whether host match kernel condition
func VMFitKernel(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error) {
	if selector.Kernel != "" && !ValidateOSVersion(selector.Kernel, host.OSVersion) {
		logs.Infof("host %s match kernel (%v) failed", host.IP, host.OSVersion)
		return false, nil
	}
	return true, nil
}

// VMFitCpuProvider check whether host match cpu provider condition
func VMFitCpuProvider(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error) {
	if len(selector.AmdDevicePattern) == 0 {
		return true, nil
	}

	switch selector.CpuProvider {
	case "Intel":
		var count int = 0
		for _, item := range selector.AmdDevicePattern {
			matched, err := regexp.MatchString(item, host.DeviceClass)
			if err == nil && !matched {
				count++
			}
		}
		if count == len(selector.AmdDevicePattern) {
			return true, nil
		}
	case "AMD":
		for _, item := range selector.AmdDevicePattern {
			matched, err := regexp.MatchString(item, host.DeviceClass)
			if err == nil && matched {
				return true, nil
			}
		}
	}

	logs.Infof("host %s match cpu provider (%s - %v) failed", host.IP, selector.CpuProvider, selector.AmdDevicePattern)
	return false, nil
}
