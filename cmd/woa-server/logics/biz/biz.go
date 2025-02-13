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

package biz

import (
	"errors"
	"fmt"
	"strconv"

	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/querybuilder"
)

// ListAuthorizedBiz list authorized biz with biz access permission from cmdb.
func (l *logics) ListAuthorizedBiz(kt *kit.Kit) ([]int64, error) {
	authReq := &meta.ListAuthResInput{Type: meta.Biz, Action: meta.Access}
	authResp, err := l.authorizer.ListAuthorizedInstances(kt, authReq)
	if err != nil {
		logs.Errorf("failed to list authorized instance, err: %v, rid: %d", err, kt.Rid)
		return nil, err
	}

	// search cmdb biz with biz access permission.
	cmdbReq := &cmdb.SearchBizReq{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
	}
	if !authResp.IsAny {
		ids := make([]int64, 0, len(authResp.IDs))
		for _, id := range authResp.IDs {
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse id %s failed, err: %v", id, err)
			}
			ids = append(ids, intID)
		}

		cmdbReq.Filter = &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_id",
						Operator: querybuilder.OperatorIn,
						Value:    ids,
					},
				},
			},
		}
	}
	resp, err := l.esbClient.Cmdb().SearchBiz(nil, nil, cmdbReq)
	if err != nil {
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	bkBizIDs := make([]int64, 0, len(resp.Data.Info))
	for _, info := range resp.Data.Info {
		bkBizIDs = append(bkBizIDs, info.BkBizId)
	}

	return bkBizIDs, nil
}

// GetBizOrgRel get biz org relation.
func (l *logics) GetBizOrgRel(kt *kit.Kit, bkBizID int64) (*mtypes.BizOrgRel, error) {
	// search cmdb business belonging.
	req := &cmdb.SearchBizBelongingParams{
		BizIDs: []int64{bkBizID},
	}

	resp, err := l.esbClient.Cmdb().SearchBizBelonging(kt, req)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return nil, errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	// convert search biz belonging response to biz org relation response.
	bizBelong := (*resp)[0]
	rst := &mtypes.BizOrgRel{
		BkBizID:         bizBelong.BizID,
		BkBizName:       bizBelong.BizName,
		OpProductID:     bizBelong.OpProductID,
		OpProductName:   bizBelong.OpProductName,
		PlanProductID:   bizBelong.PlanProductID,
		PlanProductName: bizBelong.PlanProductName,
		VirtualDeptID:   bizBelong.VirtualDeptID,
		VirtualDeptName: bizBelong.VirtualDeptName,
	}

	return rst, nil
}

// BatchCheckUserBizAccessAuth batch check user biz access auth.
func (l *logics) BatchCheckUserBizAccessAuth(kt *kit.Kit, bkBizID int64, userNames []string) (map[string]bool, error) {
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	userAuthMap, err := l.authorizer.AuthorizeByUsers(kt, userNames, authRes)
	if err != nil {
		logs.Errorf("failed to check authorize by users, bkBizID: %d, err: %v, userNames: %v, rid: %s",
			bkBizID, err, userNames, kt.Rid)
		return nil, err
	}

	processorAuth := make(map[string]bool)
	for _, userName := range userNames {
		if authInfo, ok := userAuthMap[userName]; ok {
			processorAuth[userName] = authInfo.IsAuth
		}
	}
	return processorAuth, nil
}
