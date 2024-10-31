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

package securitygroup

import (
	dataproto "hcm/pkg/api/data-service"
	datacli "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

func parseAndSaveBPaasApplication(kt *kit.Kit, dataCli *datacli.Client, accountID string, bkBizID int64,
	action enumor.ApplicationType, content any, bpaasSN string) error {

	logs.Infof("bpaas approval triggered, action: %s, application id: %v, account id: %s, rid: %s",
		action, bpaasSN, accountID, kt.Rid)
	// 保存本地申请单
	contentStr, err := json.MarshalToString(content)
	if err != nil {
		logs.Errorf("fail to marshal content to string, err: %v, action: %s, rid: %s", err, action, kt.Rid)
		return err
	}

	contentStr, err = json.UpdateMerge(map[string]interface{}{"account_id": accountID}, contentStr)
	if err != nil {
		logs.Errorf("fail to merge account id(%s) into content(%s), err: %v, action: %s, rid: %s",
			accountID, contentStr, err, action, kt.Rid)
		return err
	}

	applicationReq := &dataproto.ApplicationCreateReq{
		Source:         enumor.ApplicationSourceBPaas,
		SN:             bpaasSN,
		Type:           action,
		Status:         enumor.Pending,
		Applicant:      kt.User,
		Content:        contentStr,
		DeliveryDetail: "{}",
		Memo:           nil,
		BkBizIDs:       []int64{bkBizID},
	}
	resp, err := dataCli.Global.Application.CreateApplication(kt.Ctx, kt.Header(), applicationReq)
	if err != nil {
		logs.Errorf("fail to create application for bpaas(id: %s), err: %v, action: %s, rid: %s",
			bpaasSN, err, action, kt.Rid)
		return err
	}
	// 重新返回bpaas错误以触发前端提示, 这里直接
	return errf.New(errf.NeedBPassApproval, resp.ID)
}
