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

package errf

import (
	"regexp"

	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// 腾讯云bpass审批id匹配正则
var tcloudBPassIDRegexp *regexp.Regexp

func init() {
	tcloudBPassIDRegexp = regexp.MustCompile("ApplicationId: `\\d+`")
}

// GetBPassSNFromErr 如果是BPaas错误，则返回对应sn，如果不符合条件将返回空串
func GetBPassSNFromErr(err error) (bpaasID string) {

	if terr := GetTypedError[*terrors.TencentCloudSDKError](err); terr != nil &&
		(*terr).GetCode() == vpc.INVALIDPARAMETERVALUE_MEMBERAPPROVALAPPLICATIONSTARTED {

		// 	获取审批单号
		msg := (*terr).GetMessage()
		if appIDMsg := tcloudBPassIDRegexp.FindString(msg); len(appIDMsg) > 17 {
			bpaasID = appIDMsg[16 : len(appIDMsg)-1]
		}
		return bpaasID
	}
	return ""
}

// NeedBPassApproval 触发BPass审批流程
const NeedBPassApproval int32 = 2000012
