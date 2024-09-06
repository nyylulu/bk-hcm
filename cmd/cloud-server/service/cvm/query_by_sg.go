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

package cvm

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// QueryCvmBySGID ...
func (svc *cvmSvc) QueryCvmBySGID(cts *rest.Contexts) (any, error) {
	return svc.queryCvmBySGID(cts, handler.ListResourceAuthRes, constant.UnassignedBiz)
}

// QueryBizCvmBySGID ...
func (svc *cvmSvc) QueryBizCvmBySGID(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.queryCvmBySGID(cts, handler.ListBizAuthRes, bizID)
}

func (svc *cvmSvc) queryCvmBySGID(cts *rest.Contexts, authHandler handler.ListAuthResHandler, bizID int64) (any,
	error) {

	sgID := cts.PathParameter("sg_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	// list authorized instance 安全组权限校验
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SecurityGroup, Action: meta.Find, Filter: tools.EqualExpression("id", sgID)})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	return svc.cvmLgc.QueryCvmBySGID(cts.Kit, bizID, sgID)
}
