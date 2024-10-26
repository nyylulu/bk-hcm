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

package record

import (
	"context"
	"time"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CreateInitRecord create resource apply init record
func CreateInitRecord(suborderId, ip string) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"ip":          ip,
	}
	cnt, err := model.Operation().InitRecord().CountInitRecord(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to create init record, err: %v", err)
		return err
	}
	if cnt > 0 {
		return nil
	}

	now := time.Now()
	record := &types.InitRecord{
		SubOrderId: suborderId,
		Ip:         ip,
		TaskId:     "",
		TaskLink:   "",
		Status:     types.InitStatusInit,
		Message:    "initing",
		CreateAt:   now,
		UpdateAt:   now,
		StartAt:    now,
		EndAt:      now,
	}
	if err := model.Operation().InitRecord().CreateInitRecord(context.Background(), record); err != nil {
		logs.Errorf("failed to create init record, err: %v", err)
		return err
	}

	return nil
}

// UpdateInitRecord update resource apply init record
func UpdateInitRecord(suborderId, ip, taskId, taskUrl, message string, status types.InitStepStatus) error {
	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"ip":          ip,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"message":   message,
		"update_at": now,
		"end_at":    now,
	}

	if taskId != "" {
		doc["task_id"] = taskId
		doc["task_link"] = taskUrl
	}

	if err := model.Operation().InitRecord().UpdateInitRecord(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update init record, err: %v", err)
		return err
	}

	return nil
}

// CreateDiskCheckRecord create resource apply disk check record
func CreateDiskCheckRecord(suborderId, ip string) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"ip":          ip,
	}
	cnt, err := model.Operation().DiskCheckRecord().CountDiskCheckRecord(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to create disk check record, err: %v", err)
		return err
	}
	if cnt > 0 {
		return nil
	}

	now := time.Now()
	record := &types.DiskCheckRecord{
		SubOrderId: suborderId,
		Ip:         ip,
		TaskId:     "",
		TaskLink:   "",
		Status:     types.DiskCheckStatusHandling,
		Message:    "handling",
		CreateAt:   now,
		UpdateAt:   now,
		StartAt:    now,
		EndAt:      now,
	}
	if err := model.Operation().DiskCheckRecord().CreateDiskCheckRecord(context.Background(), record); err != nil {
		logs.Errorf("failed to create disk check record, err: %v", err)
		return err
	}

	return nil
}

// CreateDeliverRecord create resource apply deliver record
func CreateDeliverRecord(info *types.DeviceInfo) error {
	filter := map[string]interface{}{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}
	cnt, err := model.Operation().DeliverRecord().CountDeliverRecord(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to create deliver record, err: %v", err)
		return err
	}
	if cnt > 0 {
		return nil
	}

	now := time.Now()
	record := &types.DeliverRecord{
		SubOrderId:       info.SubOrderId,
		Ip:               info.Ip,
		AssetId:          info.AssetId,
		Status:           types.DeliverStatusHandling,
		Message:          "handling",
		Deliverer:        info.Deliverer,
		GenerateTaskId:   info.GenerateTaskId,
		GenerateTaskLink: info.GenerateTaskLink,
		InitTaskId:       info.InitTaskId,
		InitTaskLink:     info.InitTaskLink,
		CreateAt:         now,
		UpdateAt:         now,
		StartAt:          now,
	}
	if err := model.Operation().DeliverRecord().CreateDeliverRecord(context.Background(), record); err != nil {
		logs.Errorf("failed to create deliver record, err: %v", err)
		return err
	}

	return nil
}

// UpdateDeliverRecord update resource apply deliver record
func UpdateDeliverRecord(info *types.DeviceInfo, message string, status types.DeliverStepStatus) error {
	filter := mapstr.MapStr{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"message":   message,
		"update_at": now,
		"end_at":    now,
	}

	if err := model.Operation().DeliverRecord().UpdateDeliverRecord(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update deliver record, err: %v", err)
		return err
	}

	return nil
}

// GetDeliverRecord get resource apply deliver record
func GetDeliverRecord(kt *kit.Kit, subOrderId string, ip string, assetId string) (*types.DeliverRecord, error) {
	filter := mapstr.MapStr{
		"suborder_id": subOrderId,
		"ip":          ip,
		"asset_id":    assetId,
	}

	record, err := model.Operation().DeliverRecord().GetDeliverRecord(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get deliver record, ip: %s, err: %v, rid: %s", ip, err, kt.Rid)
		return nil, err
	}

	return record, nil
}

// GetInitRecords get init records
func GetInitRecords(kt *kit.Kit, subOrderId string) ([]*types.InitRecord, error) {
	records := make([]*types.InitRecord, 0)
	startIndex := 0
	filter := mapstr.MapStr{
		"suborder_id": subOrderId,
	}
	for {
		page := metadata.BasePage{
			Start: startIndex,
			Limit: constant.BatchOperationMaxLimit,
		}

		record, err := model.Operation().InitRecord().FindManyInitRecord(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("failed to get init record, err: %v, subOrderId: %s, rid: %s", err, subOrderId, kt.Rid)
			return nil, err
		}
		records = append(records, record...)
		if len(record) < constant.BatchOperationMaxLimit {
			break
		}
		startIndex += constant.BatchOperationMaxLimit
	}

	return records, nil
}

// GetInitRecord get init record by ip
func GetInitRecord(kt *kit.Kit, subOrderId string, ip string) (*types.InitRecord, error) {
	filter := mapstr.MapStr{
		"suborder_id": subOrderId,
		"ip":          ip,
	}

	record, err := model.Operation().InitRecord().GetInitRecord(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get init record, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return record, nil
}
