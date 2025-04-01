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
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
)

// SyncCapacity sync capacity
func (l *logics) SyncCapacity() error {
	startTime := time.Now()
	kt := core.NewBackendKit()
	logs.Infof("start to sync device capacity, startTime: %v, rid: %s", startTime, kt.Rid)

	req := &configTypes.GetDeviceParam{
		Page: metadata.BasePage{
			Sort:  "id",
			Start: 0,
			Limit: pkg.BKMaxPageSize,
		},
	}

	total := int64(0)
	success := 0
	failed := 0
	for {
		rst, err := l.configLogics.Device().GetDevice(kt, req)
		if err != nil {
			logs.Errorf("failed to get device info, err: %v, start: %d, rid: %s", err, req.Page.Start, kt.Rid)
			return fmt.Errorf("failed to get device info, err: %v", err)
		}

		total = rst.Count
		for _, device := range rst.Info {
			subKt := kt.NewSubKit()
			reqUpdate := &configTypes.UpdateCapacityParam{
				RequireType: device.RequireType,
				DeviceType:  device.DeviceType,
				Region:      device.Region,
				Zone:        device.Zone,
			}

			// 这里面会调用crp接口获取库存
			if err = l.configLogics.Capacity().UpdateCapacity(subKt, reqUpdate); err != nil {
				failed++
				logs.Warnf("failed to update device capacity info, err: %v, device: %+v, rid: %s",
					err, cvt.PtrToVal(device), subKt.Rid)
				// continue when error occurs
				continue
			}
			success++
		}

		if len(rst.Info) < pkg.BKMaxPageSize {
			break
		}
		req.Page.Start += pkg.BKMaxPageSize
		time.Sleep(time.Second)
	}
	endTime := time.Now()
	logs.Infof("end to sync device capacity, total: %d, success: %d, failed: %d, endTime: %v, cost: %fs, rid: %s",
		total, success, failed, endTime, endTime.Sub(startTime).Seconds(), kt.Rid)

	return nil
}
