/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plan

import (
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"
)

// ListDemandClass lists demand class.
func (s *service) ListDemandClass(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.DemandClass]{Details: enumor.GetDemandClassMembers()}, nil
}

// ListResMode lists resource mode.
func (s *service) ListResMode(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.ResMode]{Details: enumor.GetResModeMembers()}, nil
}

// ListDemandSource lists demand source.
func (s *service) ListDemandSource(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.DemandSource]{Details: enumor.GetDemandSourceMembers()}, nil
}

// ListRPTicketStatus lists resource plan ticket status.
func (s *service) ListRPTicketStatus(_ *rest.Contexts) (interface{}, error) {
	// get resource plan ticket status members.
	statuses := enumor.GetRPTicketStatusMembers()
	// convert to ptypes.RPTicketStatusItem slice.
	details := make([]ptypes.RPTicketStatusItem, 0, len(statuses))
	for _, status := range statuses {
		details = append(details, ptypes.RPTicketStatusItem{
			Status:     status,
			StatusName: status.Name(),
		})
	}
	return &core.ListResultT[ptypes.RPTicketStatusItem]{Details: details}, nil
}
