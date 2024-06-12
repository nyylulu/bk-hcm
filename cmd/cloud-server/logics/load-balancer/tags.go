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

package lblogic

import (
	"errors"
	"fmt"

	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// GenTagsForBizs 为负载均衡生成业务标签
func GenTagsForBizs(kt *kit.Kit, ccCli cmdb.Client, bkBizId int64) (tags []*loadbalancer.TagPair, err error) {

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

	tags = []*loadbalancer.TagPair{
		{Key: "运营产品", Value: fmt.Sprintf("%s_%d", cmdbBizInfo.BkProductName, cmdbBizInfo.BkProductID)},
		{Key: "一级业务", Value: fmt.Sprintf("%s_%d", cmdbBizInfo.Bs1Name, cmdbBizInfo.Bs1NameID)},
		{Key: "二级业务", Value: fmt.Sprintf("%s_%d", cmdbBizInfo.Bs2Name, cmdbBizInfo.Bs2NameID)},
		{Key: "运营部门", Value: fmt.Sprintf("%s_%d", cmdbBizInfo.VirtualDeptName, cmdbBizInfo.VirtualDeptID)},
		{Key: "负责人", Value: kt.User},
	}

	return tags, nil
}
