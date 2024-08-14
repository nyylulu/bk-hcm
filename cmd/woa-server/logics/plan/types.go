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

package plan

import (
	"hcm/pkg/criteria/enumor"
)

// TicketBriefInfo resource plan ticket brief info
type TicketBriefInfo struct {
	ID              string                `json:"id"`
	Applicant       string                `json:"applicant"`
	BkBizID         int64                 `json:"bk_biz_id"`
	BkBizName       string                `json:"bk_biz_name"`
	BkProductName   string                `json:"bk_product_name"`
	PlanProductName string                `json:"plan_product_name"`
	DemandClass     enumor.DemandClass    `json:"demand_class"`
	CpuCore         int64                 `json:"cpu_core"`
	Memory          int64                 `json:"memory"`
	DiskSize        int64                 `json:"disk_size"`
	SubmittedAt     string                `json:"submitted_at"`
	Status          enumor.RPTicketStatus `json:"status"`
	ItsmSn          string                `json:"itsm_sn"`
	ItsmUrl         string                `json:"itsm_url"`
	CrpSn           string                `json:"crp_sn"`
	CrpUrl          string                `json:"crp_url"`
}
