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

package itsm

import (
	"fmt"

	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway"
)

type actionType = string

const (
	// ActionTypeTERMINATE 终止
	ActionTypeTERMINATE actionType = "TERMINATE"
)

// TerminateTicket 终止单据
// TerminateTicket 不关心单据当前处于什么状态，直接终止，需要单独申请权限，正常用户无法在 ITSM 上执行此操作。
func (i *itsm) TerminateTicket(kt *kit.Kit, sn string, operator string, actionMsg string) error {
	req := map[string]interface{}{
		"sn":             sn,
		"operator":       operator,
		"action_type":    ActionTypeTERMINATE,
		"action_message": actionMsg,
	}
	resp := new(apigateway.BaseResponse)
	err := i.client.Post().
		SubResourcef("/operate_ticket/").
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}
	if !resp.Result || resp.Code != 0 {
		return fmt.Errorf("terminate ticket failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return nil
}
