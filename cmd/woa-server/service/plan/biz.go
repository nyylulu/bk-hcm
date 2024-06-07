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
	"errors"

	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetBizOrgRel get biz org relation.
func (s *service) GetBizOrgRel(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if bkBizID <= 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("bk biz id should be > 0"))
	}

	rst, err := s.getBizOrgRel(cts.Kit, bkBizID)
	if err != nil {
		logs.Errorf("failed to get biz org rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return rst, nil
}

// getBizOrgRel get biz org relation.
func (s *service) getBizOrgRel(kt *kit.Kit, bkBizID int64) (*ptypes.BizOrgRel, error) {
	// search cmdb business belonging.
	req := &cmdb.SearchBizBelongingParams{
		BizIDs: []int64{bkBizID},
	}
	resp, err := s.esbClient.Cmdb().SearchBizBelonging(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || len(resp.Data) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return nil, errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	// convert search biz belonging response to biz org relation response.
	bizBelong := resp.Data[0]
	rst := &ptypes.BizOrgRel{
		BkBizID:         bizBelong.BizID,
		BkBizName:       bizBelong.BizName,
		BkProductID:     bizBelong.BkProductID,
		BkProductName:   bizBelong.BkProductName,
		PlanProductID:   bizBelong.PlanProductID,
		PlanProductName: bizBelong.PlanProductName,
		VirtualDeptID:   bizBelong.VirtualDeptID,
		VirtualDeptName: bizBelong.VirtualDeptName,
	}

	return rst, nil
}
