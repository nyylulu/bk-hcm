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

package global

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ResourcePlanClient is data service resource plan api client.
type ResourcePlanClient struct {
	client rest.ClientInterface
}

// NewResourcePlanClient create a new resource plan api client.
func NewResourcePlanClient(client rest.ClientInterface) *ResourcePlanClient {
	return &ResourcePlanClient{
		client: client,
	}
}

// --- resource plan demand ---

// ListResPlanDemand list resource plan demand
func (b *ResourcePlanClient) ListResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandListReq) (
	*rpproto.ResPlanDemandListResult, error) {

	return common.Request[rpproto.ResPlanDemandListReq, rpproto.ResPlanDemandListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/list")
}

// BatchCreateResPlanDemand batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/batch/create")
}

// BatchUpdateResPlanDemand update resource plan demand
func (b *ResourcePlanClient) BatchUpdateResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/batch")
}

// DeleteResPlanDemand delete resource plan demand
func (b *ResourcePlanClient) DeleteResPlanDemand(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_demands/batch")
}

// LockResPlanDemand lock resource plan demand
func (b *ResourcePlanClient) LockResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandLockOpReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandLockOpReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/lock")
}

// UnlockResPlanDemand unlock resource plan demand
func (b *ResourcePlanClient) UnlockResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandLockOpReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandLockOpReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/unlock")
}

// BatchUpsertResPlanDemand upsert resource plan demand
func (b *ResourcePlanClient) BatchUpsertResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchUpsertReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandBatchUpsertReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/batch/upsert")
}

// --- demand penalty base ---

// ListDemandPenaltyBase list resource plan demand
func (b *ResourcePlanClient) ListDemandPenaltyBase(kt *kit.Kit, req *rpproto.DemandPenaltyBaseListReq) (
	*rpproto.DemandPenaltyBaseListResult, error) {

	return common.Request[rpproto.DemandPenaltyBaseListReq, rpproto.DemandPenaltyBaseListResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_penalty_bases/list")
}

// BatchCreateDemandPenaltyBase batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateDemandPenaltyBase(kt *kit.Kit, req *rpproto.DemandPenaltyBaseCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.DemandPenaltyBaseCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_penalty_bases/batch/create")
}

// BatchUpdateDemandPenaltyBase update resource plan demand
func (b *ResourcePlanClient) BatchUpdateDemandPenaltyBase(kt *kit.Kit,
	req *rpproto.DemandPenaltyBaseBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.DemandPenaltyBaseBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/demand_penalty_bases/batch")
}

// DeleteDemandPenaltyBase delete resource plan demand
func (b *ResourcePlanClient) DeleteDemandPenaltyBase(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/demand_penalty_bases/batch")
}

// --- demand changelog ---

// ListDemandChangelog list resource plan demand
func (b *ResourcePlanClient) ListDemandChangelog(kt *kit.Kit, req *rpproto.DemandChangelogListReq) (
	*rpproto.DemandChangelogListResult, error) {

	return common.Request[rpproto.DemandChangelogListReq, rpproto.DemandChangelogListResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_changelogs/list")
}

// BatchCreateDemandChangelog batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateDemandChangelog(kt *kit.Kit, req *rpproto.DemandChangelogCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.DemandChangelogCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_changelogs/batch/create")
}

// BatchUpdateDemandChangelog update resource plan demand
func (b *ResourcePlanClient) BatchUpdateDemandChangelog(kt *kit.Kit,
	req *rpproto.DemandChangelogBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.DemandChangelogBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/demand_changelogs/batch")
}

// DeleteDemandChangelog delete resource plan demand
func (b *ResourcePlanClient) DeleteDemandChangelog(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/demand_changelogs/batch")
}
