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

// Package task ...
package task

import (
	"errors"
	"fmt"

	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// IamOpType ...
const (
	IamOpTypeResourceApply   string = "resource_apply"
	IamOpTypeResourceRecycle string = "resource_recycle"
	IamOpTypeAccessBusiness  string = "access_business"
)

func (s *service) getBizName(kit *kit.Kit, bizId int64) string {
	req := &cmdb.SearchBizReq{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_id",
						Operator: querybuilder.OperatorEqual,
						Value:    bizId,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	resp, err := s.Cc.SearchBiz(kit.Ctx, nil, req)
	if err != nil {
		logs.Warnf("failed to get cc business info, err: %v, rid: %s", err, kit.Rid)
		return ""
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Warnf("failed to get cc business info, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kit.Rid)
		return ""
	}

	cnt := len(resp.Data.Info)
	if cnt != 1 {
		logs.Warnf("get invalid cc business info count %d != 1, rid: %s", cnt, kit.Rid)
		return ""
	}

	return resp.Data.Info[0].BkBizName
}

// TODO 需要替换为海垒的权限Auth模型
func (s *service) checkPermission(kit *kit.Kit, bizId int64, _ string) (*metadata.BaseResp, error) {
	user := kit.User
	if user == "" {
		logs.Errorf("failed to check permission, for invalid user is empty, rid: %s", kit.Rid)
		return nil, errors.New("failed to check permission, for invalid user is empty")
	}

	// TODO 临时测试使用，后续需要删除
	if bizId != types.AuthorizedBizID {
		return nil, fmt.Errorf("不能操作业务id: %d下的机器", bizId)
	}

	//req := &iamapi.AuthVerifyReq{
	//	System: "bk_cr",
	//	Subject: &iamapi.Subject{
	//		Type: "user",
	//		ID:   user,
	//	},
	//	Action: &iamapi.Action{
	//		ID: opType,
	//	},
	//	Resources: []*iamapi.Resource{
	//		{
	//			System: "bk_cmdb",
	//			Type:   "biz",
	//			ID:     strconv.Itoa(int(bizId)),
	//		},
	//	},
	//}
	//resp, err := s.Iam.AuthVerify(nil, nil, req)
	//if err != nil {
	//	logs.Errorf("failed to auth verify, err: %v, rid: %s", err, kit.Rid)
	//	return nil, err
	//}
	//if resp.Code != 0 {
	//	logs.Errorf("failed to auth verify, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, kit.Rid)
	//	return nil, fmt.Errorf("failed to auth verify, err: %s", resp.Message)
	//}
	//
	//bizName := s.getBizName(kit, bizId)
	//if resp.Data.Allowed != true {
	//	permission := &metadata.IamPermission{
	//		SystemID: "bk_cr",
	//		Actions: []metadata.IamAction{
	//			{
	//				ID: opType,
	//				RelatedResourceTypes: []metadata.IamResourceType{
	//					{
	//						SystemID: "bk_cmdb",
	//						Type:     "biz",
	//						Instances: [][]metadata.IamResourceInstance{
	//							{
	//								metadata.IamResourceInstance{
	//									Type: "biz",
	//									ID:   strconv.Itoa(int(bizId)),
	//									Name: bizName,
	//								},
	//							},
	//						},
	//					},
	//				},
	//			},
	//		},
	//	}
	//	authResp := metadata.NewNoPermissionResp(permission)
	//	return &authResp, common.NoAuthorizeError
	//}

	return nil, nil
}
