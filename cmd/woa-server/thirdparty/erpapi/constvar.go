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

// Package erpapi ...
package erpapi

const (
	// ReqType 请求类型
	ReqType = "Json"
	// ReqVersion 请求版本
	ReqVersion = "1.0"
	// ReqKey 请求key
	ReqKey = "20140228001"
	// ReqModule 请求模块
	ReqModule = "quota"
	// ReqOperator 操作人
	ReqOperator = "forestchen"
	// IEGDeptId IEG部门ID
	IEGDeptId = 3
	// ReturnReasonRegular 退回原因 - 常规回收 类型
	ReturnReasonRegular = "常规回收"
	// ReturnReasonDissolve 退回原因 - 机房裁撤退回 类型
	ReturnReasonDissolve = "机房裁撤退回"
	// ReturnReasonExpired 退回原因 - 过保故障退回 类型，要求有uwork故障单据
	ReturnReasonExpired = "过保故障退回"

	// ReturnOrderLinkPrefix ERP退回单据详情链接前缀
	ReturnOrderLinkPrefix = "https://cloud.erp.woa.com/return/order/detail/"

	// DeviceReturnMethod 创建设备退回订单方法
	DeviceReturnMethod = "deviceReturn"
	// QueryReturnMethod 设备退回状态查询方法
	QueryReturnMethod = "queryReturnDevicesStatus"
)
