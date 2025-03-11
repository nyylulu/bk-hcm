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

// Package detector ...
package detector

import (
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/util"
)

func (d *Detector) preCheck(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkRecyclability(step)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	return attempt, exeInfo, err
}

// RecycleCheck check whether hosts can be recycled or not
func (d *Detector) checkRecyclability(step *table.DetectStep) (string, error) {

	exeInfos := make([]string, 0)

	if step.User == "" {
		logs.Errorf("failed to recycle check, for invalid user is empty, step id: %s", step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for invalid user is empty")
	}

	// 1. check host's operator
	ips := []string{step.IP}
	hostBase, err := d.getHostBaseInfo(ips)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host err: %v, step id: %s", err, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for get host from cc err: %v", err)
	}

	hostBaseStr := d.structToStr(hostBase)
	exeInfo := fmt.Sprintf("host base info: %s", hostBaseStr)
	exeInfos = append(exeInfos, exeInfo)

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("recycler:logics:cvm:checkRecyclability:failed, failed to recycle check, "+
			"for get invalid host num %d != 1", cnt)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for get invalid host num %d != 1",
			cnt)
	}

	if strings.Contains(hostBase[0].Operator, step.User) == false &&
		strings.Contains(hostBase[0].BkBakOperator, step.User) == false {
		logs.Errorf("recycler:logics:cvm:checkRecyclability:failed, failed to recycle check, for %s is not "+
			"operator or bak operator of host %s", step.User, step.IP)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for %s is not operator or bak "+
			"operator of host %s", step.User, step.IP)
	}

	// 2. check module
	hostIds := []int64{hostBase[0].BkHostID}
	relations, err := d.getHostTopoInfo(hostIds)
	if err != nil {
		logs.Errorf("failed to recycle check, for get host topo err: %v, step id: %s", err, step.ID)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for get host topo from cc err: %v",
			err)
	}

	hostTopoStr := d.structToStr(relations)
	exeInfo = fmt.Sprintf("host topo info: %s", hostTopoStr)
	exeInfos = append(exeInfos, exeInfo)

	if len(relations) <= 0 {
		logs.Errorf("failed to recycle check, for get no host topo")
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for get no host topo")
	}

	mapBizToModule := make(map[int64][]int64)
	mapHostToRel := make(map[int64]*cmdb.HostBizRel)
	for _, rel := range relations {
		mapHostToRel[rel.BkHostId] = rel
		if _, ok := mapBizToModule[rel.BkBizId]; !ok {
			mapBizToModule[rel.BkBizId] = []int64{rel.BkModuleId}
		} else {
			mapBizToModule[rel.BkBizId] = append(mapBizToModule[rel.BkBizId], rel.BkModuleId)
		}
	}

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := util.IntArrayUnique(moduleIds)
		moduleList, err := d.getModuleInfo(bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to recycle check, for get module info err: %v, step id: %s", err, step.ID)
			return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to recycle check, for get module info err: %v", err)
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}

	moduleId := int64(0)
	if rel, ok := mapHostToRel[hostBase[0].BkHostID]; ok {
		moduleId = rel.BkModuleId
	}
	var moduleDefaultVal int64
	if module, ok := mapModuleIdToModule[moduleId]; ok {
		moduleDefaultVal = module.Default
	}

	if moduleDefaultVal != cmdb.DftModuleRecycle {
		logs.Errorf("failed to recycle check, for host %s module is not 待回收", step.IP)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("主机(%s)不在空闲机池下的待回收模块", step.IP)
	}

	return strings.Join(exeInfos, "\n"), nil

}
