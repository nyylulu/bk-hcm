/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package ziyan

import (
	"errors"
	"fmt"

	apicore "hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// ResourceMeta 资源归属元数据
type ResourceMeta struct {
	// 运营产品ID
	OpProductID int64 `json:"op_product_id,omitempty" `
	// 运营产品名称
	OpProductName string `json:"op_product_name,omitempty"`
	// 一级业务名
	Bs1Name string `json:"bs1_name,omitempty"`
	// 一级业务ID
	Bs1NameID int64 `json:"bs1_name_id,omitempty"`
	// 二级业务名
	Bs2Name string `json:"bs2_name,omitempty"`
	// 二级业务ID
	Bs2NameID int64 `json:"bs2_name_id,omitempty"`

	// 虚拟部门名称
	VirtualDeptName string `json:"virtual_dept_name,omitempty"`
	// 虚拟部门ID
	VirtualDeptID int64 `json:"virtual_dept_id,omitempty"`

	// 负责人
	Manager string `json:"manager,omitempty"`
	// 备份负责人
	BakManager string `json:"bak_manager,omitempty"`
}

// GetSyncFilterTag 按二级业务标签过滤，带`tag:`前缀
func (i *ResourceMeta) GetSyncFilterTag() apicore.TagPair {
	return apicore.TagPair{Key: "二级业务", Value: fmt.Sprintf("%s_%d", i.Bs2Name, i.Bs2NameID)}
}

// GetTagMap ...
func (i *ResourceMeta) GetTagMap() apicore.TagMap {
	return apicore.TagMap{
		"运营产品": fmt.Sprintf("%s_%d", i.OpProductName, i.OpProductID),
		"一级业务": fmt.Sprintf("%s_%d", i.Bs1Name, i.Bs1NameID),
		"二级业务": fmt.Sprintf("%s_%d", i.Bs2Name, i.Bs2NameID),
		"运营部门": fmt.Sprintf("%s_%d", i.VirtualDeptName, i.VirtualDeptID),
		"负责人":  i.Manager,
	}
}

// GetTagPairs ...
func (i *ResourceMeta) GetTagPairs() []apicore.TagPair {
	return []apicore.TagPair{
		{Key: "运营产品", Value: fmt.Sprintf("%s_%d", i.OpProductName, i.OpProductID)},
		{Key: "一级业务", Value: fmt.Sprintf("%s_%d", i.Bs1Name, i.Bs1NameID)},
		{Key: "二级业务", Value: fmt.Sprintf("%s_%d", i.Bs2Name, i.Bs2NameID)},
		{Key: "运营部门", Value: fmt.Sprintf("%s_%d", i.VirtualDeptName, i.VirtualDeptID)},
		{Key: "负责人", Value: i.Manager},
	}
}

// GetResourceMetaByBiz 为负载均衡生成业务标签
func GetResourceMetaByBiz(kt *kit.Kit, ccCli cmdb.Client, bkBizId int64) (tags *ResourceMeta, err error) {

	// 去cc 查询业务信息
	req := &cmdb.SearchBizCompanyCmdbInfoParams{BizIDs: []int64{bkBizId}}
	companyInfoList, err := ccCli.SearchBizCompanyCmdbInfo(kt, req)
	if err != nil {
		logs.Errorf("fail call cc to SearchBizCompanyCmdbInfo for cc biz id: %d, err: %v, rid: %s",
			bkBizId, err, kt.Rid)
		return nil, err
	}
	if companyInfoList == nil || len(*companyInfoList) < 1 {
		return nil, errors.New("no data returned form cc")
	}
	cmdbBizInfo := (*companyInfoList)[0]
	if cmdbBizInfo.BkBizID != bkBizId {
		logs.Errorf("company cmdb biz info from cc mismatch, want: %d, got: %d, rid: %s",
			bkBizId, cmdbBizInfo.BkBizID, kt.Rid)
		return nil, fmt.Errorf("company cmdb biz info from cc mismatch, want: %d, got: %d",
			bkBizId, cmdbBizInfo.BkBizID)
	}
	meta := &ResourceMeta{
		OpProductID:     cmdbBizInfo.BkProductID,
		OpProductName:   cmdbBizInfo.BkProductName,
		Bs1Name:         cmdbBizInfo.Bs1Name,
		Bs1NameID:       cmdbBizInfo.Bs1NameID,
		Bs2Name:         cmdbBizInfo.Bs2Name,
		Bs2NameID:       cmdbBizInfo.Bs2NameID,
		VirtualDeptName: cmdbBizInfo.VirtualDeptName,
		VirtualDeptID:   cmdbBizInfo.VirtualDeptID,
		Manager:         kt.User,
	}

	return meta, nil
}
