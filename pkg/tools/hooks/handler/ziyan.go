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

package handler

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iam/meta"
)

// 如果是对自研云的vpc或subnet资源的查询操作，则取消业务ID的相等检查，表示自研云的这两种资源对所有业务都可见，为临时解决方案，后期需去除
func needDisableBizEqualForZiyan(opt *ValidWithAuthOption) bool {
	if opt.Action != meta.Find {
		return false
	}
	if opt.BasicInfo == nil {
		return false
	}
	if opt.BasicInfo.Vendor != enumor.TCloudZiyan {
		return false
	}
	return opt.ResType == meta.Vpc || opt.ResType == meta.Subnet

}
