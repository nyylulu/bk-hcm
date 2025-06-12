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

package coreziyan

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
)

// BPaasApplicationContent BPaas application content
type BPaasApplicationContent struct {
	Action               enumor.ApplicationType `json:"action"`
	SN                   string                 `json:"sn"`
	AccountID            string                 `json:"account_id"`
	SecurityGroupID      string                 `json:"sg_id"`
	BkBizID              int64                  `json:"bk_biz_id"`
	Region               string                 `json:"region"`
	SecurityGroupCloudID string                 `json:"sg_cloud_id"`
	Params               any                    `json:"params"`
}

// Validate validate
func (c *BPaasApplicationContent) Validate() error {
	if c == nil {
		return fmt.Errorf("content of bpaas cannot be empty")
	}
	if c.SN == "" {
		return fmt.Errorf("sn is empty for bpaas")
	}
	if c.Action == "" {
		return fmt.Errorf("action is empty for bpaas")
	}
	if c.AccountID == "" {
		return fmt.Errorf("account id is empty for bpaas")
	}
	if c.BkBizID == 0 {
		return fmt.Errorf("bk biz id is empty for bpaas")
	}
	if c.Region == "" {
		return fmt.Errorf("region is empty for bpaas")
	}
	return nil
}
