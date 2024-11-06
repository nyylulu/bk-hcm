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

package rollingserver

import (
	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// IsResPoolBiz 判断业务是否是资源池业务
func (l *logics) IsResPoolBiz(kt *kit.Kit, bizID int64) (bool, error) {
	listReq := &rsproto.ResourcePoolBusinessListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
			},
		},
		Page: &core.BasePage{Count: true},
	}
	result, err := l.client.DataService().Global.RollingServer.ListResPoolBiz(kt, listReq)
	if err != nil {
		logs.Errorf("list rolling resource pool business failed, err: %v, req: %+v, rid: %s", err, *listReq, kt.Rid)
		return false, err
	}

	return len(result.Details) != 0, nil
}
