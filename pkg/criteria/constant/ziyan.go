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

package constant

import "time"

// 自研账号需要使用内部域名
const (
	// InternalVpcEndpoint vpc 内部域名
	InternalVpcEndpoint = "vpc.internal.tencentcloudapi.com"
	// InternalCvmEndpoint cvm 内部域名
	InternalCvmEndpoint = "cvm.internal.tencentcloudapi.com"
	// InternalCbsEndpoint cbs 内部域名
	InternalCbsEndpoint = "cbs.internal.tencentcloudapi.com"
	// InternalCamEndpoint cam 内部域名
	InternalCamEndpoint = "cam.internal.tencentcloudapi.com"
	// InternalBillingEndpoint 账单内部域名
	InternalBillingEndpoint = "billing.internal.tencentcloudapi.com"
	// InternalCertEndpoint 证书内部域名
	InternalCertEndpoint = "ssl.internal.tencentcloudapi.com"
	// InternalClbEndpoint 负载均衡内部域名
	InternalClbEndpoint = "clb.internal.tencentcloudapi.com"
	// InternalTagEndpoint 标签内部域名
	InternalTagEndpoint = "tag.internal.tencentcloudapi.com"

	// InternalBPaasEndpoint bpaas 内部域名
	InternalBPaasEndpoint = "bpaas.internal.tencentcloudapi.com"

	// IEGDeptName 互动娱乐事业部
	IEGDeptName = "互动娱乐事业部"
)

// 等待时间
const (
	// IntervalWaitTaskStart 任务启动前的等待时间
	IntervalWaitTaskStart = time.Second * 5
)
