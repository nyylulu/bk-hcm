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
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/kit"
)

// ListResPlanSubTicket list resource plan sub_ticket.
func (c *Controller) ListResPlanSubTicket(kt *kit.Kit, req *ptypes.ListResPlanSubTicketReq) (
	*ptypes.ListResPlanSubTicketResp, error) {

	return c.resFetcher.ListResPlanSubTicket(kt, req)
}

// GetResPlanSubTicketDetail get resource plan sub_ticket detail.
func (c *Controller) GetResPlanSubTicketDetail(kt *kit.Kit, subTicketID string) (*ptypes.GetSubTicketDetailResp, string,
	error) {

	return c.resFetcher.GetResPlanSubTicketDetail(kt, subTicketID)
}

// GetResPlanSubTicketAudit get res plan sub ticket audit
func (c *Controller) GetResPlanSubTicketAudit(kt *kit.Kit, bizID int64, subTicketID string) (
	*ptypes.GetSubTicketAuditResp, string, error) {

	return c.resFetcher.GetResPlanSubTicketAudit(kt, bizID, subTicketID)
}

// RetryResPlanFailedSubTickets 重试失败的子单
func (c *Controller) RetryResPlanFailedSubTickets(kt *kit.Kit, ticketID string, subTicketIDs []string) error {
	// TODO 等待拆单逻辑完成 先用同步方式验证速度，拆单速度太慢的话需要转异步

	return nil
}
