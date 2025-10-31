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

package cscvm

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
)

// BatchIdleCheckReq ...
type BatchIdleCheckReq struct {
	BkHostIDs []int64 `json:"bk_host_ids" validate:"required"`
}

// Validate ...
func (req BatchIdleCheckReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.BkHostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}
	return nil
}

// IdleCheckResultReq ...
type IdleCheckResultReq struct {
	Page *core.BasePage `json:"page" validate:"required"`
}

// Validate validate IdleCheckResultReq
func (req IdleCheckResultReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Page.Validate(); err != nil {
		return err
	}
	return nil
}

// IdleCheckResultRsp ...
type IdleCheckResultRsp types.ListResult[*IdleCheckResultRspItem]

// IdleCheckResultRspItem ...
type IdleCheckResultRspItem struct {
	DetectTask  table.DetectTask   `json:"detect_task"`
	DetectSteps []table.DetectStep `json:"detect_steps"`
}

// BatchIdleCheckCvmRsp ...
type BatchIdleCheckCvmRsp struct {
	TaskManagementID    string `json:"task_management_id"`
	IdleCheckSuborderID string `json:"idle_check_suborder_id"`
}
