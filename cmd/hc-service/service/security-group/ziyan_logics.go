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
	"hcm/pkg/api/core/ziyan"
	dataproto "hcm/pkg/api/data-service"
	datacli "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

func parseAndSaveBPaasApplication(kt *kit.Kit, dataCli *datacli.Client,
	content *coreziyan.BPaasApplicationContent) error {

	if err := content.Validate(); err != nil {
		return err
	}

	logs.Infof("bpaas approval triggered, action: %s, sn: %v, account id: %s, rid: %s",
		content.Action, content.SN, content.AccountID, kt.Rid)
	// 保存本地申请单
	contentStr, err := json.MarshalToString(content)
	if err != nil {
		logs.Errorf("fail to marshal content to string, err: %v, action: %s, rid: %s", err, content.Action, kt.Rid)
		return err
	}

	applicationReq := &dataproto.ApplicationCreateReq{
		Source:         enumor.ApplicationSourceBPaas,
		SN:             content.SN,
		Type:           content.Action,
		Status:         enumor.Pending,
		Applicant:      kt.User,
		Content:        contentStr,
		DeliveryDetail: "{}",
		Memo:           nil,
		BkBizIDs:       []int64{content.BkBizID},
	}
	resp, err := dataCli.Global.Application.CreateApplication(kt.Ctx, kt.Header(), applicationReq)
	if err != nil {
		logs.Errorf("fail to create application for bpaas(sn: %s), err: %v, action: %s, rid: %s",
			content.SN, err, content.Action, kt.Rid)
		return err
	}
	// 重新返回bpaas错误以触发前端提示, 这里直接在错误message中写入申请单id
	return errf.New(errf.NeedBPaasApproval, resp.ID)
}
