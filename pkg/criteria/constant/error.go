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

// WarnSign 告警标识
type WarnSign string

const (
	// AccountSyncFailed account sync failed.
	AccountSyncFailed WarnSign = "account_sync_failed"
	// CmdbSyncFailed cmdb sync failed.
	CmdbSyncFailed WarnSign = "cmdb_sync_failed"
	// DeleteCvmStartScriptFailed delete cvm start script failed.
	DeleteCvmStartScriptFailed WarnSign = "delete_cvm_start_script_failed"
	// CvmHasMultipleVpcs cvm has multiple vpc. 因为蓝鲸体现中一台主机只能属于一个云区域，如果主机有多个Vpv，
	// 就可能属于多个不同的云区域，蓝鲸概念冲突，这类主机暂不同步进海垒。
	CvmHasMultipleVpcs WarnSign = "cvm_has_multiple_vpc"
	// AccountBillConfigFailed account bill config failed.
	AccountBillConfigFailed WarnSign = "account_bill_config_failed"
	// RecycleUpdateRecordFailed 后台回收任务过程中更新记录失败
	RecycleUpdateRecordFailed WarnSign = "recycle_update_record_failed"
	// AsyncTaskWarnSign 异步任务框架执行异常告警
	AsyncTaskWarnSign = "async_task_exec_exception"
	// ApplicationDeliverFailed 申请单交付失败告警
	ApplicationDeliverFailed WarnSign = "application_deliver_failed"

	// DemandChangeAppliedFailed demand change applied failed
	DemandChangeAppliedFailed WarnSign = "demand_change_applied_failed"
)

const (
	// TCloudLimitExceededErrCode 腾讯云限频错误码
	TCloudLimitExceededErrCode = "RequestLimitExceeded"

	// TCloudNetworkErrorErrCode 腾讯云网络错误 错误码
	TCloudNetworkErrorErrCode = "ClientError.NetworkError"
)

// TCloudClientErrRetryTimes ..
const TCloudClientErrRetryTimes = 5

const (
	// CRPResPlanDemandIsInProcessing CRP的预测资源已被其他单据修改，正在审批中
	CRPResPlanDemandIsInProcessing = "AdjustDemandIsInProcessingException"
	// CRPResPlanDemandIsOverLimit CRP的预测资源超上限问题
	CRPResPlanDemandIsOverLimit = "已超上限，请联系规划经理沟通"
	// CRPResPlanDemandIsOverLimitMessage CRP的预测资源超上限问题, 提示信息
	CRPResPlanDemandIsOverLimitMessage = "部门资源预测数据超过规划上限，请企微联系%s，并提供相关信息：1.运营产品  2.地域 3.机型4.CPU总核数 5.业务需求原因 "
)
