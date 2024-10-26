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

package task

// ExpireDays 订单捞取恢复的过期时间, 大于ExpireDays订单不做恢复处理
const ExpireDays = -15

// ResourceOperationService 资源运营服务业务id
const ResourceOperationService = 931

// ApplyGoroutinesNum 订单申请恢复服务开启协程数量
const ApplyGoroutinesNum = 5 // applyRecover协程数量

// RebornBizId reborn业务ID
const RebornBizId int64 = 213

// CrRelayModuleId CR Relay模块ID
const CrRelayModuleId int64 = 5069670

// DataToCleanedModule 业务数据清理模块ID
const DataToCleanedModule int64 = 16679

// Handler 业务处理人
const Handler = "dommyzhang;forestchen"

// DataPendingClean 数据待清理模块ID
const DataPendingClean int64 = 16679
