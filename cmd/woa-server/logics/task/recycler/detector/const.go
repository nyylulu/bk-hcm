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

// Package detector ...
package detector

import (
	"hcm/cmd/woa-server/dal/task/table"
)

// StepList detection task step list
var StepList = []*table.DetectStepCfg{
	{
		Sequence:    1,
		Name:        table.StepPreCheck,
		Description: "检查CC模块和负责人",
		Enable:      true,
		Retry:       5,
	},
	{
		Sequence:    2,
		Name:        table.StepCheckUwork,
		Description: "检查是否有Uwork故障单据",
		Enable:      true,
		Retry:       5,
	},
	{
		Sequence:    3,
		Name:        table.StepCheckGCS,
		Description: "检查是否有GCS记录",
		Enable:      true,
		Retry:       5,
	},
	{
		Sequence:    4,
		Name:        table.StepBasicCheck,
		Description: "tmp,tgw,tgw nat,l5策略检查",
		Enable:      true,
		Retry:       5,
	},
	{
		Sequence:    5,
		Name:        table.StepCvmCheck,
		Description: "检查cvm, docker on cvm的安全组与CLB策略",
		Enable:      true,
		Retry:       1,
	},
	{
		Sequence:    6,
		Name:        table.StepCheckSafety,
		Description: "安全基线检查",
		Enable:      true,
		Retry:       1,
	},
	{
		Sequence:    7,
		Name:        table.StepCheckReturn,
		Description: "检查是否有退回单据",
		Enable:      true,
		Retry:       1,
	},
	{
		Sequence:    8,
		Name:        table.StepCheckProcess,
		Description: "空闲检查",
		Enable:      true,
		Retry:       1,
	},
}
