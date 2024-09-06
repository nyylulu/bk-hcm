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

package xshipapi

// various task accept status
// 状态列表，0为已受理，1为处理中，2为流程中，3已完成，4为it异常处理中，5it异常重试中，6驳回，7已失效
const (
	AcceptStatusAccepted          AcceptStatus = 0
	AcceptStatusHandling          AcceptStatus = 1
	AcceptStatusProcessing        AcceptStatus = 2
	AcceptStatusDone              AcceptStatus = 3
	AcceptStatusExceptionHandling AcceptStatus = 4
	AcceptStatusRetrying          AcceptStatus = 5
	AcceptStatusRejected          AcceptStatus = 6
	AcceptStatusExpired           AcceptStatus = 7
)

const (
	CodeSuccess         string = "0"
	ReinstallLinkPrefix string = "https://xwing.woa.com/reinstall/reinstall"
	DftStarter          string = "icr"
)
