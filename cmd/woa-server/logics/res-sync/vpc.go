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

// Package ressync ...
package ressync

import (
	"fmt"
	"time"

	configTypes "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	"hcm/pkg/logs"
)

// SyncVpc sync vpc
func (l *logics) SyncVpc() error {
	startTime := time.Now()
	kt := core.NewBackendKit()
	logs.Infof("start to sync vpc, startTime: %v, rid: %s", startTime, kt.Rid)

	regions, err := l.configLogics.Region().GetRegion(kt)
	if err != nil {
		logs.Errorf("failed to get all region list, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("failed to get all region list, err: %v", err)
	}

	success := 0
	failed := 0
	for _, region := range regions.Info {
		subKt := kt.NewSubKit()
		req := &configTypes.GetVpcParam{
			Region: region.Region,
		}

		if err = l.configLogics.Vpc().SyncVpc(subKt, req); err != nil {
			failed++
			logs.Warnf("failed to sync vpc, region: %+v, err: %v, rid: %s", region, err, subKt.Rid)
			// continue when error occurs
			continue
		}
		success++
	}
	endTime := time.Now()
	logs.Infof("end to sync vpc, count: %d, success: %d, failed: %d, endTime: %v, cost: %fs, rid: %s",
		len(regions.Info), success, failed, endTime, endTime.Sub(startTime).Seconds(), kt.Rid)

	return nil
}
