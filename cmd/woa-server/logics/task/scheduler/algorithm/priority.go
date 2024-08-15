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

// Package algorithm contains all kinds of priority functors
package algorithm

import (
	"math"

	"hcm/cmd/woa-server/thirdparty/dvmapi"
	types "hcm/cmd/woa-server/types/task"
)

// CalculateBalancedResourceAllocation resource allocation balance priority functor
func CalculateBalancedResourceAllocation(selector *types.DVMSelector, hosts []*dvmapi.DockerHost) (
	types.HostPriorityList, error) {

	result := make(types.HostPriorityList, 0, len(hosts))
	for _, host := range hosts {
		cpuRequested := selector.Cores + (host.CPUCapacity - host.AllocatableCPU)
		memoryRequested := selector.Memory + (host.MemoryCapacity - host.AllocatableMem)
		cpuFraction := fractionOfCapacity(cpuRequested, host.CPUCapacity)
		memoryFraction := fractionOfCapacity(memoryRequested, host.MemoryCapacity)
		var score float64 = 0
		if cpuFraction >= 1 && memoryFraction >= 1 {
			score = 1
		} else if cpuFraction >= 1 || memoryFraction >= 1 {
			score = 0
		} else {
			score = 1 - math.Abs(cpuFraction-memoryFraction)
		}
		result = append(result, types.HostPriority{
			IP:               host.IP,
			DeviceClass:      host.DeviceClass,
			Equipment:        host.Equipment,
			ModuleName:       host.ModuleName,
			SZone:            host.SZone,
			SetId:            host.SetId,
			AllocatableCount: host.AllocatableCount,
			ScheduledVMs:     host.ScheduledVMs,
			Score:            score * types.MaxPriority,
		})
	}
	return result, nil
}

// fractionOfCapacity fraction functor
func fractionOfCapacity(requested, capacity int) float64 {
	if capacity == 0 {
		return 1
	}
	return float64(requested) / float64(capacity)
}
