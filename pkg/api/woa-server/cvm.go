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

package woaserver

import (
	"fmt"

	"hcm/pkg"
	"hcm/pkg/criteria/validator"
)

// StartIdleCheckReq ...
type StartIdleCheckReq struct {
	HostIDs  []int64  `json:"bk_host_ids" validate:"required,min=1"`
	AssetIDs []string `json:"bk_asset_ids" validate:"required,min=1"`
	IPs      []string `json:"ips" validate:"required,min=1"`
	BkBizID  int64    `json:"bk_biz_id" validate:"required"`
}

// Validate ...
func (r StartIdleCheckReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.IPs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ips exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(r.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(r.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	// 检查 HostIDs、AssetIDs 和 IPs 的长度是否相等
	if len(r.HostIDs) != len(r.AssetIDs) || len(r.HostIDs) != len(r.IPs) {
		return fmt.Errorf("HostIDs, AssetIDs, and IPs must have the same length")
	}
	return nil
}

// StartIdleCheckRsp 返回mongodb生成的唯一单号和子单号
type StartIdleCheckRsp struct {
	SuborderID string `json:"suborder_id"`
}
