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
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg/logs"
)

func (d *Detector) checkSecurityBaseline(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		err = d.checkLog4j(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		exeInfo = err.Error()
	}

	return attempt, exeInfo, err
}

func (d *Detector) checkLog4j(ip string) error {
	ips := []string{ip}
	hostBase, err := d.getHostBaseInfo(ips)
	if err != nil {
		logs.Errorf("failed to check log4j, for list host err: %v, step id: %s", err, ip)
		return fmt.Errorf("failed to check log4j, for get host from cc err: %v", err)
	}

	cnt := len(hostBase)
	if cnt != 1 {
		logs.Errorf("failed to check log4j, for get invalid host num %d != 1", cnt)
		return fmt.Errorf("failed to check log4j, for get invalid host num %d != 1", cnt)
	}

	// check log4j for host
	if !d.isDockerVM(hostBase[0]) {
		pass, err := d.safety.CheckLog4jHost(nil, nil, ip)
		if err != nil {
			return fmt.Errorf("failed to check log4j, err: %v", err)
		} else if pass == false {
			return fmt.Errorf("failed to check log4j, not pass")
		}
		return nil
	}

	// check log4j for container
	parentIp, err := d.getContainerParentIp(hostBase[0])
	if err != nil {
		logs.Errorf("recycler:logics:cvm:checkLog4j:failed, failed to check log4j, for get container "+
			"parent ip: %s, err: %v", ip, err)
		return fmt.Errorf("failed to check log4j, for get container parent ip err: %v", err)
	}

	pass, err := d.safety.CheckLog4jContainer(nil, nil, ip, parentIp)
	if err != nil {
		return fmt.Errorf("failed to check log4j, err: %v", err)
	} else if pass == false {
		return fmt.Errorf("failed to check log4j, not pass")
	}

	return nil
}
