/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package actionsg

import (
	"fmt"

	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/errf"
)

func handleBPaasErr(kt run.ExecuteKit, actErr error) (caught bool, err error) {
	ef := errf.Error(actErr)
	if ef == nil || ef.Code != errf.NeedBPaasApproval {
		return false, nil
	}
	// 捕获bpaas错误，写入结果中
	err = kt.ShareData().Set(kt.Kit(), "bpaas_sn", ef.Message)
	if err != nil {
		return true, fmt.Errorf("fail to set bpaas_sn: %s, err: %w", ef.Message, err)
	}
	return true, nil

}
