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

package ziyan

import (
	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/tcloud"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/kit"
	bpaas "hcm/pkg/thirdparty/tencentcloud/bpaas/v20181217"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// TCloudZiyan 自研云adaptor接口，基本同步腾讯云接口，避免冗余
type TCloudZiyan interface {
	tcloud.TCloud
	GetBPaasApplicationDetail(kt *kit.Kit, applicationID uint64) (*bpaas.GetBpaasApplicationDetailResponseParams, error)
	CreateZiyanLoadBalancer(kt *kit.Kit, opt *typelb.TCloudZiyanCreateClbOption) (*poller.BaseDoneResult, error)
	DescribeSlaCapacity(kt *kit.Kit, opt *typelb.TCloudDescribeSlaCapacityOption) (
		*clb.DescribeSlaCapacityResponseParams, error)
}
