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

// Package meta is meta logics related package.
package meta

import (
	"fmt"
	"strconv"

	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
)

// GetBizsByOpProd get bizs by op product.
func (l *logics) GetBizsByOpProd(kt *kit.Kit, prodID int64) ([]mtypes.Biz, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]mtypes.Biz, 0)
	for _, bizBelong := range allBizBelonging {
		if bizBelong.BkProductID == prodID {
			result = append(result, mtypes.Biz{
				BkBizID:   bizBelong.BkBizID,
				BkBizName: bizBelong.BizName,
			})
		}
	}

	return result, nil
}

// getCmdbAllBizBelonging get cmdb all biz belonging.
func (l *logics) getCmdbAllBizBelonging(kt *kit.Kit) ([]cmdb.CompanyCmdbInfo, error) {
	result := make([]cmdb.CompanyCmdbInfo, 0)
	batch := constant.SearchBizBelongingMaxLimit
	start := 0
	for {
		req := &cmdb.SearchBizCompanyCmdbInfoParams{
			Page: &cmdb.BasePage{
				Limit: int64(batch),
				Start: int64(start),
			},
		}

		resp, err := l.cmdbCli.SearchBizCompanyCmdbInfo(kt, req)
		if err != nil {
			logs.Errorf("failed to search biz belonging, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result = append(result, cvt.PtrToVal(resp)...)

		if len(*resp) <= 0 {
			break
		}

		start += len(*resp)
	}

	return result, nil
}

// GetOpProducts get op products.
func (l *logics) GetOpProducts(kt *kit.Kit) ([]mtypes.OpProduct, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// use map to unique op product.
	opProdMap := make(map[int64]mtypes.OpProduct)
	for _, bizBelong := range allBizBelonging {
		opProdMap[bizBelong.BkProductID] = mtypes.OpProduct{
			OpProductID:   bizBelong.BkProductID,
			OpProductName: bizBelong.BkProductName,
		}
	}

	result := maps.Values(opProdMap)
	return result, nil
}

// GetPlanProducts get op products.
func (l *logics) GetPlanProducts(kt *kit.Kit) ([]mtypes.PlanProduct, error) {
	allBizBelonging, err := l.getCmdbAllBizBelonging(kt)
	if err != nil {
		logs.Errorf("failed to get cmdb all biz belonging, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// use map to unique plan product.
	planProdMap := make(map[int64]mtypes.PlanProduct)
	for _, bizBelong := range allBizBelonging {
		planProdMap[bizBelong.PlanProductID] = mtypes.PlanProduct{
			PlanProductID:   bizBelong.PlanProductID,
			PlanProductName: bizBelong.PlanProductName,
		}
	}

	result := maps.Values(planProdMap)
	return result, nil
}

// GetOrgTopo get org topo
func (l *logics) GetOrgTopo(kt *kit.Kit, orgView enumor.View) (*mtypes.OrgInfo, error) {
	switch orgView {
	case enumor.IEGView:
		iegOrgMap, err := l.getIEGOrgTopos(kt)
		if err != nil {
			return nil, err
		}
		return l.buildOrgTopo(iegOrgMap), nil
	default:
		return nil, fmt.Errorf("unsupported org topo view: %s", orgView)
	}
}

// getIEGOrgTopos get ieg org topos
func (l *logics) getIEGOrgTopos(kt *kit.Kit) (map[string]*mtypes.OrgInfo, error) {
	depts, err := l.dao.OrgTopo().ListAllDepartment(kt)
	if err != nil {
		logs.Errorf("failed to get department info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	topos := make(map[string]*mtypes.OrgInfo)
	for _, v := range depts {
		// 过滤腾讯公司及IEG的部门列表
		if (v.Level == constant.TencentDeptLevel && v.DeptID == constant.TencentDeptID) ||
			(v.Level == constant.IEGDeptLevel && v.DeptID == constant.IEGDeptID) || v.Level > constant.IEGDeptLevel {
			topos[v.DeptID] = &mtypes.OrgInfo{
				ID:          v.DeptID,
				Name:        v.DeptName,
				FullName:    v.FullName,
				Level:       v.Level,
				Parent:      v.Parent,
				TofDeptID:   v.TofDeptID,
				HasChildren: cvt.PtrToVal(v.HasChildren) != 0,
				Children:    nil,
			}
		}
	}

	return topos, nil
}

func (l *logics) buildOrgTopo(topos map[string]*mtypes.OrgInfo) *mtypes.OrgInfo {
	var root *mtypes.OrgInfo
	for _, org := range topos {
		if org.ID == constant.TencentDeptID && org.Parent == strconv.Itoa(int(constant.TencentDeptLevel)) {
			root = org
		} else {
			if parent, ok := topos[org.Parent]; ok && org.Level != 0 {
				parent.Children = append(parent.Children, org)
			}
		}
	}
	return root
}
