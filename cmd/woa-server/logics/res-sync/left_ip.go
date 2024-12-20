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

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"
)

// SyncLeftIP sync left ip info collection
func (l *logics) SyncLeftIP() error {
	startTime := time.Now()
	kt := core.NewBackendKit()
	logs.Infof("start to sync left ip, startTime: %v, rid: %s", startTime, kt.Rid)

	param := &types.GetLeftIPParam{
		Page: metadata.BasePage{
			Start: 0,
			Limit: pkg.BKMaxPageSize,
		},
	}

	total := 0
	success := 0
	failed := 0
	for {
		rst, err := l.configLogics.LeftIP().GetLeftIP(kt, param)
		if err != nil {
			logs.Errorf("failed to get all left ip info, err: %v, rid: %s", err, kt.Rid)
			return fmt.Errorf("failed to get all left ip info, err: %v", err)
		}

		total += len(rst.Info)
		for _, zone := range rst.Info {
			syncReq := &types.SyncLeftIPParam{
				Region: zone.Region,
				Zone:   zone.Zone,
			}

			if err = l.configLogics.LeftIP().SyncLeftIP(kt, syncReq); err != nil {
				failed++
				logs.Warnf("failed to sync zone left ip info, err: %v, region: %s, zone: %s, rid: %s",
					err, zone.Region, zone.Zone, kt.Rid)
				// continue when error occurs
				continue
			}
			success++
		}
		if len(rst.Info) < pkg.BKMaxPageSize {
			break
		}
		param.Page.Start += pkg.BKMaxPageSize
	}
	endTime := time.Now()
	logs.Infof("end to sync left ip, total: %d, success: %d, failed: %d, endTime: %v, cost: %fs, rid: %s",
		total, success, failed, endTime, endTime.Sub(startTime).Seconds(), kt.Rid)

	return nil
}
