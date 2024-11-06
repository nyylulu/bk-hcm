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

// Package record provides record functions
package record

import (
	"context"
	"time"

	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
)

// CreateCommitStep init apply order commit step info
func CreateCommitStep(ctx context.Context, suborderId string, replicas uint, stepID int) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"step_name":   types.StepNameCommit,
	}
	stepCnt, err := model.Operation().ApplyStep().CountApplyStep(ctx, filter)
	if err != nil {
		logs.Errorf("failed to create commit step, err: %v", err)
		return err
	}
	if stepCnt > 0 {
		return nil
	}

	// create step if no record in db
	now := time.Now()
	step := &types.ApplyStep{
		SubOrderId: suborderId,
		StepId:     stepID,
		StepName:   types.StepNameCommit,
		Status:     types.StepStatusSuccess,
		Message:    types.StepMsgSuccess,
		TotalNum:   replicas,
		SuccessNum: replicas,
		FailedNum:  0,
		RunningNum: 0,
		CreateAt:   now,
		UpdateAt:   now,
		StartAt:    now,
		EndAt:      now,
	}
	if err := model.Operation().ApplyStep().CreateApplyStep(ctx, step); err != nil {
		logs.Errorf("failed to create commit step, err: %v", err)
		return err
	}

	return nil
}

// CreateGenerateStep init apply order generate step info
func CreateGenerateStep(ctx context.Context, suborderId string, replicas uint, stepID int) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"step_name":   types.StepNameGenerate,
	}
	stepCnt, err := model.Operation().ApplyStep().CountApplyStep(ctx, filter)
	if err != nil {
		logs.Errorf("failed to create generate step, err: %v", err)
		return err
	}
	if stepCnt > 0 {
		return nil
	}

	// create step if no record in db
	now := time.Now()
	step := &types.ApplyStep{
		SubOrderId: suborderId,
		StepId:     stepID,
		StepName:   types.StepNameGenerate,
		Status:     types.StepStatusInit,
		Message:    types.StepMsgInit,
		TotalNum:   replicas,
		SuccessNum: 0,
		FailedNum:  0,
		RunningNum: 0,
		CreateAt:   now,
		UpdateAt:   now,
	}
	if err := model.Operation().ApplyStep().CreateApplyStep(ctx, step); err != nil {
		logs.Errorf("failed to create generate step, err: %v", err)
		return err
	}

	return nil
}

// UpdateGenerateStep update apply order generate step info
func UpdateGenerateStep(suborderId string, total uint, errStep error) error {
	now := time.Now()
	if errStep != nil {
		filter := mapstr.MapStr{
			"suborder_id": suborderId,
			"step_name":   types.StepNameGenerate,
		}
		doc := mapstr.MapStr{
			"status":    types.StepStatusFailed,
			"message":   errStep.Error(),
			"update_at": now,
			"end_at":    now,
		}

		if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
			logs.Errorf("failed to update generate step, err: %v", err)
			return err
		}
		return nil
	}

	devices, err := getUnreleasedDevice(suborderId)
	if err != nil {
		logs.Errorf("failed to update generate step, err: %v", err)
		return err
	}

	status := types.StepStatusHandling
	message := types.StepMsgHandling
	count := uint(len(devices))
	if count >= total {
		status = types.StepStatusSuccess
		message = types.StepMsgSuccess
	}

	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"step_name":   types.StepNameGenerate,
	}

	doc := mapstr.MapStr{
		"status":      status,
		"message":     message,
		"success_num": count,
		"update_at":   now,
	}

	if status == types.StepStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update generate step, err: %v", err)
		return err
	}

	return nil
}

// CreateInitStep init apply order init step info
func CreateInitStep(ctx context.Context, suborderId string, replicas uint, stepID int) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"step_name":   types.StepNameInit,
	}
	stepCnt, err := model.Operation().ApplyStep().CountApplyStep(ctx, filter)
	if err != nil {
		logs.Errorf("failed to create init step, err: %v", err)
		return err
	}
	if stepCnt > 0 {
		return nil
	}

	// create step if no record in db
	now := time.Now()
	step := &types.ApplyStep{
		SubOrderId: suborderId,
		StepId:     stepID,
		StepName:   types.StepNameInit,
		Status:     types.StepStatusInit,
		Message:    types.StepMsgInit,
		TotalNum:   replicas,
		SuccessNum: 0,
		FailedNum:  0,
		RunningNum: 0,
		CreateAt:   now,
		UpdateAt:   now,
	}
	if err := model.Operation().ApplyStep().CreateApplyStep(ctx, step); err != nil {
		logs.Errorf("failed to create init step, err: %v", err)
		return err
	}

	return nil
}

// UpdateInitStep update apply order init step info
func UpdateInitStep(suborderId string, total uint) error {
	devices, err := getUnreleasedDevice(suborderId)
	if err != nil {
		logs.Errorf("failed to update init step, err: %v", err)
		return err
	}

	status := types.StepStatusHandling
	message := types.StepMsgHandling
	count := uint(0)
	for _, device := range devices {
		if device.IsInited {
			count++
		}
	}
	if count >= total {
		status = types.StepStatusSuccess
		message = types.StepMsgSuccess
	}

	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"step_name":   types.StepNameInit,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":      status,
		"message":     message,
		"success_num": count,
		"update_at":   now,
	}

	if status == types.StepStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update init step, err: %v", err)
		return err
	}

	return nil
}

// CreateDiskCheckStep init apply order disk check step info
func CreateDiskCheckStep(ctx context.Context, suborderId string, replicas uint, stepID int) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"step_name":   types.StepNameDiskCheck,
	}
	stepCnt, err := model.Operation().ApplyStep().CountApplyStep(ctx, filter)
	if err != nil {
		logs.Errorf("failed to create disk check step, err: %v", err)
		return err
	}
	if stepCnt > 0 {
		return nil
	}

	// create step if no record in db
	now := time.Now()
	step := &types.ApplyStep{
		SubOrderId: suborderId,
		StepId:     stepID,
		StepName:   types.StepNameDiskCheck,
		Status:     types.StepStatusInit,
		Message:    types.StepMsgInit,
		TotalNum:   replicas,
		SuccessNum: 0,
		FailedNum:  0,
		RunningNum: 0,
		CreateAt:   now,
		UpdateAt:   now,
	}
	if err := model.Operation().ApplyStep().CreateApplyStep(ctx, step); err != nil {
		logs.Errorf("failed to create disk check step, err: %v", err)
		return err
	}

	return nil
}

// UpdateDiskCheckStep update apply order disk check step info
func UpdateDiskCheckStep(suborderId string, total uint) error {
	devices, err := getUnreleasedDevice(suborderId)
	if err != nil {
		logs.Errorf("failed to update disk check step, err: %v", err)
		return err
	}

	status := types.StepStatusHandling
	message := types.StepMsgHandling
	count := uint(0)
	for _, device := range devices {
		if device.IsDiskChecked {
			count++
		}
	}
	if count >= total {
		status = types.StepStatusSuccess
		message = types.StepMsgSuccess
	}

	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"step_name":   types.StepNameDiskCheck,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":      status,
		"message":     message,
		"success_num": count,
		"update_at":   now,
	}

	if status == types.StepStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update disk check step, err: %v", err)
		return err
	}

	return nil
}

// CreateDeliverStep init apply order deliver step info
func CreateDeliverStep(ctx context.Context, suborderId string, replicas uint, stepID int) error {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"step_name":   types.StepNameDeliver,
	}
	stepCnt, err := model.Operation().ApplyStep().CountApplyStep(ctx, filter)
	if err != nil {
		logs.Errorf("failed to create deliver step, err: %v", err)
		return err
	}
	if stepCnt > 0 {
		return nil
	}

	// create step if no record in db
	now := time.Now()
	step := &types.ApplyStep{
		SubOrderId: suborderId,
		StepId:     stepID,
		StepName:   types.StepNameDeliver,
		Status:     types.StepStatusInit,
		Message:    types.StepMsgInit,
		TotalNum:   replicas,
		SuccessNum: 0,
		FailedNum:  0,
		RunningNum: 0,
		CreateAt:   now,
		UpdateAt:   now,
	}
	if err := model.Operation().ApplyStep().CreateApplyStep(ctx, step); err != nil {
		logs.Errorf("failed to create deliver step, err: %v", err)
		return err
	}

	return nil
}

// UpdateDeliverStep update apply order deliver step info
func UpdateDeliverStep(suborderId string, total uint) error {
	devices, err := getUnreleasedDevice(suborderId)
	if err != nil {
		logs.Errorf("failed to update deliver step, err: %v", err)
		return err
	}

	status := types.StepStatusHandling
	message := types.StepMsgHandling
	count := uint(0)
	for _, device := range devices {
		if device.IsDelivered {
			count++
		}
	}
	if count >= total {
		status = types.StepStatusSuccess
		message = types.StepMsgSuccess
	}

	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"step_name":   types.StepNameDeliver,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":      status,
		"message":     message,
		"success_num": count,
		"update_at":   now,
	}

	if status == types.StepStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to update deliver step, err: %v", err)
		return err
	}

	return nil
}

// StartStep update apply order step with start info
func StartStep(suborderId string, stepName string) error {
	filter := mapstr.MapStr{
		"suborder_id": suborderId,
		"step_name":   stepName,
		"status":      types.StepStatusInit,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    types.StepStatusHandling,
		"message":   types.StepMsgHandling,
		"start_at":  now,
		"update_at": now,
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), &filter, &doc); err != nil {
		logs.Errorf("failed to start order %s step name %s, err: %v", suborderId, stepName, err)
		return err
	}

	return nil
}

// getUnreleasedDevice gets unreleased devices binding to given apply order
func getUnreleasedDevice(orderId string) ([]*types.DeviceInfo, error) {
	filter := &mapstr.MapStr{
		"suborder_id": orderId,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to order %s, err: %v", orderId, err)
	}

	return devices, nil
}
