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
	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/ziyan"
)

// GenTagsForBizs 为自研云资源生成业务标签，负责人标签从kit中获取
func GenTagsForBizs(kt *kit.Kit, ccCli cmdb.Client, bkBizId int64) (tags []apicore.TagPair, err error) {
	meta, err := ziyan.GetResourceMetaByBiz(kt, ccCli, bkBizId)
	if err != nil {
		logs.Errorf("fail to get resource meta for bk biz id: %d, err: %v, rid: %s",
			bkBizId, err, kt.Rid)
		return nil, err
	}
	return meta.GetTagPairs(), nil
}

// GenTagsForBizsWithManager 为自研云资源生成业务标签，负责人通过参数提供
// 允许业务、主备负责人不全部提供，此时仅更新部分标签到云上
func GenTagsForBizsWithManager(kt *kit.Kit, ccCli cmdb.Client, bkBizId int64, manager, bakManager string) (
	tags []apicore.TagPair, err error) {

	if bkBizId == constant.UnassignedBiz {
		meta := &ziyan.ResourceMeta{
			Manager:    manager,
			BakManager: bakManager,
		}

		return meta.GetTagPairs(), nil
	}

	meta, err := ziyan.GetResourceMetaByBizWithManager(kt, ccCli, bkBizId, manager, bakManager)
	if err != nil {
		logs.Errorf("fail to get resource meta for bk biz id: %d, err: %v, rid: %s",
			bkBizId, err, kt.Rid)
		return nil, err
	}
	return meta.GetTagPairs(), nil
}
